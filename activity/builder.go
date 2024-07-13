package activity

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type CurrentUserFunc func(ctx context.Context) *User

// @snippet_begin(ActivityBuilder)
// Builder struct contains all necessary fields
type Builder struct {
	models []*ModelBuilder // registered model builders

	db              *gorm.DB                 // global db
	logModelInstall presets.ModelInstallFunc // log model install
	permPolicy      *perm.PolicyBuilder      // permission policy
	currentUserFunc CurrentUserFunc
	findUsersFunc   func(ctx context.Context, ids []string) (map[string]*User, error)
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

func (ab *Builder) CurrentUserFunc(v CurrentUserFunc) *Builder {
	ab.currentUserFunc = v
	return ab
}

func (ab *Builder) FindUsersFunc(v func(ctx context.Context, ids []string) (map[string]*User, error)) *Builder {
	ab.findUsersFunc = v
	return ab
}

// New initializes a new Builder instance with a provided database connection and an optional activity log model.
func New(db *gorm.DB) *Builder {
	ab := &Builder{
		db: db,
		permPolicy: perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
			ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).
			On("*:activity_logs").On("*:activity_logs:*"),
	}
	ab.logModelInstall = ab.defaultLogModelInstall
	return ab
}

// RegisterModels register mutiple models
func (ab *Builder) RegisterModels(models ...any) *Builder {
	for _, model := range models {
		ab.RegisterModel(model)
	}
	return ab
}

// RegisterModel Model register a model and return model builder
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

	primaryKeys := getPrimaryKeys(reflectType)
	amb = &ModelBuilder{
		typ:           reflectType,
		ab:            ab,
		keys:          primaryKeys,
		ignoredFields: primaryKeys,
	}
	if mb, ok := m.(*presets.ModelBuilder); ok {
		amb.installPresetsModelBuilder(mb)
	}

	ab.models = append(ab.models, amb)
	return amb
}

// GetModelBuilder 	get model builder
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

// MustGetModelBuilder 	get model builder
func (ab *Builder) MustGetModelBuilder(v any) *ModelBuilder {
	amb, ok := ab.GetModelBuilder(v)
	if !ok {
		panic(fmt.Sprintf("model %v is not registered", v))
	}
	return amb
}

// GetModelBuilders get all model builders
func (ab *Builder) GetModelBuilders() []*ModelBuilder {
	return ab.models
}

func (b *Builder) AutoMigrate() (r *Builder) {
	if err := AutoMigrate(b.db); err != nil {
		panic(err)
	}
	return b
}

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(&ActivityLog{}, &ActivityUser{})
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

func (ab *Builder) supplyCreators(ctx context.Context, logs []*ActivityLog) error {
	creatorIDs := lo.Uniq(lo.Map(logs, func(log *ActivityLog, _ int) string {
		return log.CreatorID
	}))
	creators, err := ab.findUsers(ctx, creatorIDs)
	if err != nil {
		return err
	}
	for _, log := range logs {
		if creator, ok := creators[log.CreatorID]; ok {
			log.Creator = *creator
		}
	}
	return nil
}

func (ab *Builder) getActivityLogs(ctx context.Context, modelName, modelKeys string) ([]*ActivityLog, error) {
	var logs []*ActivityLog
	err := ab.db.Where("model_name = ? AND model_keys = ?", modelName, modelKeys).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	if err := ab.supplyCreators(ctx, logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// AddRecords add records log
func (ab *Builder) AddRecords(ctx context.Context, action string, vs ...any) error {
	if len(vs) == 0 {
		return errors.New("data are empty")
	}

	for _, v := range vs {
		if mb, ok := ab.GetModelBuilder(v); ok {
			if err := mb.AddRecords(ctx, action, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddCustomizedRecord add customized record
func (ab *Builder) AddCustomizedRecord(ctx context.Context, action string, diff bool, obj any) error {
	if mb, ok := ab.GetModelBuilder(obj); ok {
		return mb.AddCustomizedRecord(ctx, action, diff, obj)
	}

	return fmt.Errorf("can't find model builder for %v", obj)
}

// AddViewRecord add view record
func (ab *Builder) AddViewRecord(creator *User, v any) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddViewRecord(creator, v)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddDeleteRecord	add delete record
func (ab *Builder) AddDeleteRecord(creator *User, v any) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddDeleteRecord(creator, v)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddSaverRecord will save a create log or a edit log
func (ab *Builder) AddSaveRecord(creator *User, new any) error {
	if mb, ok := ab.GetModelBuilder(new); ok {
		return mb.AddSaveRecord(creator, new)
	}

	return fmt.Errorf("can't find model builder for %v", new)
}

// AddCreateRecord add create record
func (ab *Builder) AddCreateRecord(creator *User, v any) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddCreateRecord(creator, v)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecord(creator *User, new any) error {
	if mb, ok := ab.GetModelBuilder(new); ok {
		return mb.AddEditRecord(creator, new)
	}

	return fmt.Errorf("can't find model builder for %v", new)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecordWithOld(creator *User, old, new any) error {
	if mb, ok := ab.GetModelBuilder(new); ok {
		return mb.AddEditRecordWithOld(creator, old, new)
	}

	return fmt.Errorf("can't find model builder for %v", new)
}

// AddEditRecordWithOldAndContext add edit record
func (ab *Builder) AddEditRecordWithOldAndContext(ctx context.Context, old, new any) error {
	if mb, ok := ab.GetModelBuilder(new); ok {
		return mb.AddEditRecordWithOld(ab.currentUserFunc(ctx), old, new)
	}

	return fmt.Errorf("can't find model builder for %v", new)
}
