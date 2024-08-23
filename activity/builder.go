package activity

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type User struct {
	ID     string
	Name   string
	Avatar string
}

// @snippet_begin(ActivityBuilder)
type Builder struct {
	models            []*ModelBuilder // registered model builders
	installedPresets  sync.Map        // installed presets builders for admin
	calledAutoMigrate atomic.Bool     // auto migrate flag

	dbPrimitive             *gorm.DB // primitive db
	db                      *gorm.DB // global db with table prefix scope
	tablePrefix             string
	logModelInstall         presets.ModelInstallFunc // admin preset install
	permPolicy              *perm.PolicyBuilder      // permission policy
	currentUserFunc         func(ctx context.Context) (*User, error)
	findUsersFunc           func(ctx context.Context, ids []string) (map[string]*User, error)
	findLogsForTimelineFunc func(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, error)
}

// @snippet_end

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	b.RegisterModel(m)
	return nil
}

func (ab *Builder) WrapLogModelInstall(w func(presets.ModelInstallFunc) presets.ModelInstallFunc) *Builder {
	ab.logModelInstall = w(ab.logModelInstall)
	return ab
}

func (ab *Builder) PermPolicy(v *perm.PolicyBuilder) *Builder {
	ab.permPolicy = v
	return ab
}

func (ab *Builder) FindUsersFunc(v func(ctx context.Context, ids []string) (map[string]*User, error)) *Builder {
	ab.findUsersFunc = v
	return ab
}

func (ab *Builder) FindLogsForTimelineFunc(v func(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, error)) *Builder {
	ab.findLogsForTimelineFunc = v
	return ab
}

func New(db *gorm.DB, currentUserFunc func(ctx context.Context) (*User, error)) *Builder {
	ab := &Builder{
		dbPrimitive:     db,
		db:              db,
		currentUserFunc: currentUserFunc,
		permPolicy: perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
			ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).
			On("*:activity_logs").On("*:activity_logs:*"),
	}
	ab.logModelInstall = ab.defaultLogModelInstall
	return ab
}

func (ab *Builder) TablePrefix(prefix string) *Builder {
	if ab.calledAutoMigrate.Load() {
		panic("please set table prefix before auto migrate")
	}
	ab.tablePrefix = prefix
	if prefix == "" {
		ab.db = ab.dbPrimitive
	} else {
		ab.db = ab.dbPrimitive.Scopes(ScopeWithTablePrefix(prefix)).Session(&gorm.Session{})
	}
	return ab
}

func (ab *Builder) RegisterModels(models ...any) *Builder {
	for _, model := range models {
		ab.RegisterModel(model)
	}
	return ab
}

func (ab *Builder) RegisterModel(m any) (amb *ModelBuilder) {
	if amb, exist := ab.GetModelBuilder(m); exist {
		return amb
	}

	model := m
	if preset, ok := m.(*presets.ModelBuilder); ok {
		model = preset.NewModel()
	}
	if model == nil {
		panic(fmt.Sprintf("%v is nil", m))
	}

	reflectType := reflect.Indirect(reflect.ValueOf(model)).Type()
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%v is not a struct", reflectType.Name()))
	}
	amb = &ModelBuilder{
		ref: reflect.New(reflectType).Interface(),
		typ: reflectType,
		ab:  ab,
	}

	primaryKeys := ParsePrimaryKeys(model)
	amb.Keys(primaryKeys...)
	amb.IgnoredFields(primaryKeys...)

	if mb, ok := m.(*presets.ModelBuilder); ok {
		amb.installPresetModelBuilder(mb)
	}

	ab.models = append(ab.models, amb)
	return amb
}

func (ab *Builder) GetModelBuilder(v any) (*ModelBuilder, bool) {
	if _, ok := v.(*presets.ModelBuilder); ok {
		return lo.Find(ab.models, func(amb *ModelBuilder) bool {
			return amb.presetModel == v
		})
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	return lo.Find(ab.models, func(amb *ModelBuilder) bool {
		return amb.typ == typ
	})
}

func (ab *Builder) MustGetModelBuilder(v any) *ModelBuilder {
	amb, ok := ab.GetModelBuilder(v)
	if !ok {
		panic(fmt.Sprintf("model %v is not registered", v))
	}
	return amb
}

func (ab *Builder) GetModelBuilders() []*ModelBuilder {
	return ab.models
}

func (b *Builder) AutoMigrate() (r *Builder) {
	if !b.calledAutoMigrate.CompareAndSwap(false, true) {
		panic("already migrated")
	}
	if err := AutoMigrate(b.dbPrimitive, b.tablePrefix); err != nil {
		panic(err)
	}
	return b
}

func AutoMigrate(db *gorm.DB, tablePrefix string) error {
	if tablePrefix != "" {
		db = db.Scopes(ScopeWithTablePrefix(tablePrefix)).Session(&gorm.Session{})
	}
	dst := []any{&ActivityLog{}, &ActivityUser{}}
	for _, v := range dst {
		err := db.Model(v).AutoMigrate(v)
		if err != nil {
			return errors.Wrap(err, "auto migrate")
		}
		if vv, ok := v.(interface {
			AfterMigrate(tx *gorm.DB, tablePrefix string) error
		}); ok {
			err := vv.AfterMigrate(db, tablePrefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ab *Builder) findUsers(ctx context.Context, ids []string) (map[string]*User, error) {
	if ab.findUsersFunc != nil {
		return ab.findUsersFunc(ctx, ids)
	}
	vs := []*ActivityUser{}
	err := ab.db.Where("id IN ?", ids).Find(&vs).Error
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(vs, func(item *ActivityUser) (string, *User) {
		id := fmt.Sprint(item.ID)
		return id, &User{
			ID:     id,
			Name:   item.Name,
			Avatar: item.Avatar,
		}
	}), nil
}

func (ab *Builder) supplyUsers(ctx context.Context, logs []*ActivityLog) error {
	userIDs := lo.Uniq(lo.Map(logs, func(log *ActivityLog, _ int) string {
		return log.UserID
	}))
	users, err := ab.findUsers(ctx, userIDs)
	if err != nil {
		return err
	}
	for _, log := range logs {
		if user, ok := users[log.UserID]; ok {
			log.User = *user
		}
	}
	return nil
}

func (ab *Builder) findLogsForTimeline(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, error) {
	if ab.findLogsForTimelineFunc != nil {
		logs, err := ab.findLogsForTimelineFunc(ctx, modelName, modelKeys)
		if err != nil {
			return nil, err
		}
		userAllFilled := lo.EveryBy(logs, func(log *ActivityLog) bool {
			return log.User.ID != ""
		})
		if userAllFilled {
			return logs, nil
		}
		if err := ab.supplyUsers(ctx, logs); err != nil {
			return nil, err
		}
		return logs, nil
	}

	return ab.getActivityLogs(ctx, modelName, modelKeys)
}

func (ab *Builder) getActivityLogs(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, error) {
	var logs []*ActivityLog
	err := ab.db.Where("hidden = FALSE AND model_name = ? AND model_keys = ?", modelName, modelKeys).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	if err := ab.supplyUsers(ctx, logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (ab *Builder) onlyModelBuilder(v any) (*ModelBuilder, error) {
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	ambs := lo.Filter(ab.models, func(amb *ModelBuilder, _ int) bool {
		return amb.typ == typ
	})
	if len(ambs) == 0 {
		return nil, errors.Errorf("can't find model builder for %v", v)
	}
	if len(ambs) > 1 {
		bare, ok := lo.Find(ambs, func(amb *ModelBuilder) bool { return amb.presetModel == nil })
		if ok {
			return bare, nil
		}
		return nil, errors.Errorf("multiple preset model builders found for %v", v)
	}
	return ambs[0], nil
}

func (ab *Builder) Log(ctx context.Context, action string, v any, detail any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(v)
	if err != nil {
		return nil, err
	}
	return amb.Log(ctx, action, v, detail)
}

func (ab *Builder) OnCreate(ctx context.Context, v any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(v)
	if err != nil {
		return nil, err
	}
	return amb.OnCreate(ctx, v)
}

func (ab *Builder) OnView(ctx context.Context, v any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(v)
	if err != nil {
		return nil, err
	}
	return amb.OnView(ctx, v)
}

func (ab *Builder) OnEdit(ctx context.Context, old, new any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(new)
	if err != nil {
		return nil, err
	}
	return amb.OnEdit(ctx, old, new)
}

func (ab *Builder) OnDelete(ctx context.Context, v any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(v)
	if err != nil {
		return nil, err
	}
	return amb.OnDelete(ctx, v)
}

func (ab *Builder) Note(ctx context.Context, v any, note *Note) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(v)
	if err != nil {
		return nil, err
	}
	return amb.Note(ctx, v, note)
}
