package activity

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
	"gorm.io/gorm"
)

const (
	Create = 1 << iota
	Delete
	Update
)

type contextKey int

const (
	CreatorContextKey contextKey = iota
	DBContextKey
)

// @snippet_begin(ActivityBuilder)
type ActivityBuilder struct {
	db                *gorm.DB    // global db
	creatorContextKey interface{} // get the creator from context
	dbContextKey      interface{} // get the db from context

	lmb      *presets.ModelBuilder // log model builder
	logModel ActivityLogInterface  // log model

	models     []*ModelBuilder                   // registered model builders
	tabHeading func(ActivityLogInterface) string // tab heading format
}

// @snippet_end

func New(b *presets.Builder, db *gorm.DB, logModel ...ActivityLogInterface) *ActivityBuilder {
	ab := &ActivityBuilder{
		db:                db,
		creatorContextKey: CreatorContextKey,
		dbContextKey:      DBContextKey,
	}

	if len(logModel) > 0 {
		ab.logModel = logModel[0]
	} else {
		ab.logModel = &ActivityLog{}
	}

	if err := db.AutoMigrate(ab.logModel); err != nil {
		panic(err)
	}

	ab.configureAdmin(b)
	return ab
}

// GetPresetModelBuilder return the preset model builder
func (ab ActivityBuilder) GetPresetModelBuilder() *presets.ModelBuilder {
	return ab.lmb
}

// GetActivityLogs get activity logs
func (ab ActivityBuilder) GetActivityLogs(m interface{}, db *gorm.DB) []*ActivityLog {
	objs := ab.GetCustomizeActivityLogs(m, db)
	if objs == nil {
		return nil
	}

	logs, ok := objs.(*[]*ActivityLog)
	if !ok {
		return nil
	}
	return *logs
}

// GetCustomizeActivityLogs get customize activity logs
func (ab ActivityBuilder) GetCustomizeActivityLogs(m interface{}, db *gorm.DB) interface{} {
	mb, ok := ab.GetModelBuilder(m)

	if !ok {
		return nil
	}

	if db == nil {
		db = ab.db
	}

	keys := mb.KeysValue(m)
	logs := ab.NewLogModelSlice()
	err := db.Where("model_name = ? AND model_keys = ?", mb.typ.Name(), keys).Find(logs).Error
	if err != nil {
		return nil
	}
	return logs
}

// NewLogModelData new a log model data
func (ab ActivityBuilder) NewLogModelData() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type()).Interface()
}

// NewLogModelSlice new a log model slice
func (ab ActivityBuilder) NewLogModelSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.PtrTo(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type()))
	slice := reflect.New(sliceType)
	slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))
	return slice.Interface()
}

// SetCreatorContextKey change the default creator context key
func (ab *ActivityBuilder) SetCreatorContextKey(key interface{}) *ActivityBuilder {
	ab.creatorContextKey = key
	return ab
}

// SetDBContextKey change the default db context key
func (ab *ActivityBuilder) SetDBContextKey(key interface{}) *ActivityBuilder {
	ab.dbContextKey = key
	return ab
}

func (ab *ActivityBuilder) SetTabHeading(f func(log ActivityLogInterface) string) *ActivityBuilder {
	ab.tabHeading = f
	return ab
}

// RegisterModels register mutiple models
func (ab *ActivityBuilder) RegisterModels(models ...interface{}) *ActivityBuilder {
	for _, model := range models {
		ab.RegisterModel(model)
	}
	return ab
}

// Model register a model and return model builder
func (ab *ActivityBuilder) RegisterModel(m interface{}) (mb *ModelBuilder) {
	if m, exist := ab.GetModelBuilder(m); exist {
		return m
	}

	model := getBasicModel(m)
	if model == nil {
		panic(fmt.Sprintf("%v is nil", m))
	}

	reflectType := reflect.Indirect(reflect.ValueOf(model)).Type()
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%v is not a struct", reflectType.Name()))
	}

	keys := getPrimaryKey(reflectType)
	mb = &ModelBuilder{
		typ:      reflectType,
		activity: ab,

		keys:          keys,
		ignoredFields: keys,
	}
	ab.models = append(ab.models, mb)

	if presetModel, ok := m.(*presets.ModelBuilder); ok {
		mb.presetModel = presetModel

		var (
			editing    = presetModel.Editing()
			oldSaver   = editing.Saver
			oldDeleter = editing.Deleter
		)

		editing.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Update != 0 && mb.skip&Create != 0 {
				return oldSaver(obj, id, ctx)
			}

			old, ok := findOld(obj, ab.getDBFromContext(ctx.R.Context()))
			if err = oldSaver(obj, id, ctx); err != nil {
				return err
			}

			if (!ok || id == "") && mb.skip&Create == 0 {
				return mb.AddRecords(ActivityCreate, ctx.R.Context(), obj)
			}

			if ok && id != "" && mb.skip&Update == 0 {
				return mb.AddEditRecordWithOld(ab.getCreatorFromContext(ctx.R.Context()), old, obj, ab.getDBFromContext(ctx.R.Context()))
			}

			return
		})

		editing.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Delete != 0 {
				return oldDeleter(obj, id, ctx)
			}

			old, ok := findOldWithSlug(obj, id, ab.getDBFromContext(ctx.R.Context()))
			if err = oldDeleter(obj, id, ctx); err != nil {
				return err
			}

			if ok {
				return mb.AddRecords(ActivityDelete, ctx.R.Context(), old)
			}

			return
		})
	}

	return mb
}

// GetModelBuilder 	get model builder
func (ab ActivityBuilder) GetModelBuilder(v interface{}) (*ModelBuilder, bool) {
	var isPreset bool
	if _, ok := v.(*presets.ModelBuilder); ok {
		isPreset = true
	}

	typ := reflect.Indirect(reflect.ValueOf(getBasicModel(v))).Type()
	for _, m := range ab.models {
		if m.typ == typ {
			if !isPreset {
				return m, true
			}

			if isPreset && m.presetModel == v {
				return m, true
			}
		}
	}
	return &ModelBuilder{}, false
}

// GetModelBuilder 	get model builder
func (ab ActivityBuilder) MustGetModelBuilder(v interface{}) *ModelBuilder {
	mb, ok := ab.GetModelBuilder(v)
	if !ok {
		panic(fmt.Sprintf("model %v is not registered", v))
	}
	return mb
}

// GetModelBuilders 	get all model builders
func (ab ActivityBuilder) GetModelBuilders() []*ModelBuilder {
	return ab.models
}

// AddRecords add records log
func (ab *ActivityBuilder) AddRecords(action string, ctx context.Context, vs ...interface{}) error {
	if len(vs) == 0 {
		return errors.New("data are empty")
	}

	for _, v := range vs {
		if mb, ok := ab.GetModelBuilder(v); ok {
			if err := mb.AddRecords(action, ctx, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddCustomizedRecord add customized record
func (ab *ActivityBuilder) AddCustomizedRecord(action string, diff bool, ctx context.Context, obj interface{}) error {
	if mb, ok := ab.GetModelBuilder(obj); ok {
		return mb.AddCustomizedRecord(action, diff, ctx, obj)
	}

	return fmt.Errorf("can't find model builder for %v", obj)
}

// AddViewRecord add view record
func (ab *ActivityBuilder) AddViewRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddViewRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddDeleteRecord	add delete record
func (ab *ActivityBuilder) AddDeleteRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddDeleteRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddSaverRecord will save a create log or a edit log
func (ab *ActivityBuilder) AddSaveRecord(creator interface{}, now interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddSaveRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddCreateRecord add create record
func (ab *ActivityBuilder) AddCreateRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddCreateRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddEditRecord add edit record
func (ab *ActivityBuilder) AddEditRecord(creator interface{}, now interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddEditRecord add edit record
func (ab *ActivityBuilder) AddEditRecordWithOld(creator interface{}, old, now interface{}, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecordWithOld(creator, old, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// GetDB get db from context
func (ab *ActivityBuilder) getDBFromContext(ctx context.Context) *gorm.DB {
	if contextdb := ctx.Value(ab.dbContextKey); contextdb != nil {
		return contextdb.(*gorm.DB)
	}
	return ab.db
}

// GetDB get creator from context
func (ab *ActivityBuilder) getCreatorFromContext(ctx context.Context) interface{} {
	if creator := ctx.Value(ab.creatorContextKey); creator != nil {
		return creator
	}
	return ""
}
