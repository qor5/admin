package activity

import (
	"cmp"
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

const DefaultMaxCountShowInTimeline = 10

// @snippet_begin(ActivityBuilder)
type Builder struct {
	models            []*ModelBuilder // registered model builders
	calledAutoMigrate atomic.Bool     // auto migrate flag

	dbPrimitive             *gorm.DB // primitive db
	db                      *gorm.DB // global db with table prefix scope
	tablePrefix             string
	logModelInstall         presets.ModelInstallFunc // admin preset install
	permPolicy              *perm.PolicyBuilder      // permission policy
	currentUserFunc         func(ctx context.Context) (*User, error)
	findUsersFunc           func(ctx context.Context, ids []string) (map[string]*User, error)
	maxCountShowInTimeline  int
	findLogsForTimelineFunc func(ctx context.Context, db *gorm.DB, modelName, modelKeys string) (logs []*ActivityLog, hasMore bool, err error)
	skipResPermCheck        bool
	mu                      sync.RWMutex
	logModelBuilders        map[*presets.Builder]*presets.ModelBuilder
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

func (ab *Builder) SkipResPermCheck(v bool) *Builder {
	ab.skipResPermCheck = v
	return ab
}

func (ab *Builder) FindUsersFunc(v func(ctx context.Context, ids []string) (map[string]*User, error)) *Builder {
	ab.findUsersFunc = v
	return ab
}

func (ab *Builder) MaxCountShowInTimeline(v int) *Builder {
	ab.maxCountShowInTimeline = v
	return ab
}

func (ab *Builder) FindLogsForTimelineFunc(v func(ctx context.Context, db *gorm.DB, modelName, modelKeys string) (logs []*ActivityLog, hasMore bool, err error)) *Builder {
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
			On("*:presets:activity_logs").On("*:presets:activity_logs:*"),
		maxCountShowInTimeline: DefaultMaxCountShowInTimeline,
		logModelBuilders:       map[*presets.Builder]*presets.ModelBuilder{},
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

	amb.Keys(ParsePrimaryKeys(model, false)...)
	amb.IgnoredFields(ParsePrimaryKeys(model, true)...)

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
	if len(logs) == 0 {
		return nil
	}
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

func (ab *Builder) findLogsForTimeline(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, bool, error) {
	if ab.findLogsForTimelineFunc != nil {
		logs, hasMore, err := ab.findLogsForTimelineFunc(ctx, ab.db, modelName, modelKeys)
		if err != nil {
			return nil, false, err
		}
		userAllFilled := lo.EveryBy(logs, func(log *ActivityLog) bool {
			return log.User.ID != ""
		})
		if userAllFilled {
			return logs, hasMore, nil
		}
		if err := ab.supplyUsers(ctx, logs); err != nil {
			return nil, false, err
		}
		return logs, hasMore, nil
	}

	return ab.getActivityLogs(ctx, modelName, modelKeys)
}

func (ab *Builder) getActivityLogs(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, bool, error) {
	maxCountShowInTimeline := cmp.Or(ab.maxCountShowInTimeline, DefaultMaxCountShowInTimeline)

	var logs []*ActivityLog
	err := ab.db.Where("hidden = FALSE AND model_name = ? AND model_keys = ?", modelName, modelKeys).
		Order("created_at DESC").Limit(maxCountShowInTimeline + 1).Find(&logs).Error
	if err != nil {
		return nil, false, err
	}
	if err := ab.supplyUsers(ctx, logs); err != nil {
		return nil, false, err
	}
	if len(logs) > maxCountShowInTimeline {
		return logs[:maxCountShowInTimeline], true, nil
	}
	return logs, false, nil
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

func (ab *Builder) Log(ctx context.Context, action string, v, detail any) (*ActivityLog, error) {
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

func (ab *Builder) OnEdit(ctx context.Context, oldObj, newObj any) (*ActivityLog, error) {
	amb, err := ab.onlyModelBuilder(newObj)
	if err != nil {
		return nil, err
	}
	return amb.OnEdit(ctx, oldObj, newObj)
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
