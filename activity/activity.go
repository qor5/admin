package activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

type contextKey string

const (
	Create = 1 << iota
	Delete
	Update
	All = Create | Delete | Update
)

var (
	GlobalDB          *gorm.DB
	CreatorContextKey contextKey = "Creator"
	DBContextKey      contextKey = "DB"
)

// @snippet_begin(ActivityBuilder)
type ActivityBuilder struct {
	creatorContextKey interface{}          // get the creator from context
	dbContextKey      interface{}          // get the db from context
	logModel          ActivityLogInterface // log model
	models            []*ModelBuilder      // registered model builders
}

// @snippet_end

// @snippet_begin(ActivityModelBuilder)
type ModelBuilder struct {
	typ               reflect.Type
	keys              []string                     // primary keys
	disableOnCallback uint8                        // disable the callback depends on the mode
	link              func(interface{}) string     // display the model link on the admin detail page
	ignoredFields     []string                     // ignored fields
	typeHanders       map[reflect.Type]TypeHandler // type handlers
}

// @snippet_end

func Activity() *ActivityBuilder {
	return &ActivityBuilder{
		logModel:          &ActivityLog{},
		creatorContextKey: CreatorContextKey,
		dbContextKey:      DBContextKey,
	}
}

// SetLogModel change the default log model
func (ab *ActivityBuilder) SetLogModel(model ActivityLogInterface) *ActivityBuilder {
	ab.logModel = model
	return ab
}

// NewLogModelData new a log model data
func (ab ActivityBuilder) NewLogModelData() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type()).Interface()
}

// NewLogModelSlice new a log model slice
func (ab ActivityBuilder) NewLogModelSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type())
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

// getPrimaryKey get primary keys from a model
func getPrimaryKey(t reflect.Type) (keys []string) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primary") {
			keys = append(keys, t.Field(i).Name)
			continue
		}

		if t.Field(i).Type.Kind() == reflect.Ptr && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type.Elem())...)
		}

		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type)...)
		}
	}
	return
}

// RegisterModels register mutiple models
func (ab *ActivityBuilder) RegisterModels(models ...interface{}) *ActivityBuilder {
	for _, model := range models {
		ab.RegisterModel(model)
	}
	return ab
}

// RegisterModel register a model and return model builder
func (ab *ActivityBuilder) RegisterModel(model interface{}) *ModelBuilder {
	reflectType := reflect.Indirect(reflect.ValueOf(model)).Type()
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%v is not a struct", reflectType.Name()))
	}

	keys := getPrimaryKey(reflectType)
	mb := &ModelBuilder{
		typ:           reflectType,
		keys:          keys,
		ignoredFields: keys,
	}
	ab.models = append(ab.models, mb)
	return mb
}

// GetModelBuilder 	get model builder
func (ab ActivityBuilder) GetModelBuilder(v interface{}) (*ModelBuilder, bool) {
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	for _, m := range ab.models {
		if m.typ == typ {
			return m, true
		}
	}
	return &ModelBuilder{}, false
}

// AddKeys add keys to the model builder
func (mb *ModelBuilder) AddKeys(keys ...string) *ModelBuilder {
	mb.keys = append(mb.keys, keys...)
	return mb
}

// SetKeys set keys for the model builder
func (mb *ModelBuilder) SetKeys(keys ...string) *ModelBuilder {
	mb.keys = keys
	return mb
}

func (mb *ModelBuilder) SetLink(f func(interface{}) string) *ModelBuilder {
	mb.link = f
	return mb
}

func (mb *ModelBuilder) DisableOnCallback(modes ...uint8) *ModelBuilder {
	var mode uint8
	for _, m := range modes {
		mode |= m
	}
	mb.disableOnCallback = mode
	return mb
}

// AddIgnoredFields add ignored fields to the model builder
func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = append(mb.ignoredFields, fields...)
	return mb
}

// SetIgnoredFields set ignored fields for the model builder
func (mb *ModelBuilder) SetIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = append(mb.ignoredFields, fields...)
	return mb
}

// AddTypeHanders add type handers for the model builder
func (mb *ModelBuilder) AddTypeHanders(v interface{}, f TypeHandler) *ModelBuilder {
	if mb.typeHanders == nil {
		mb.typeHanders = map[reflect.Type]TypeHandler{}
	}
	mb.typeHanders[reflect.Indirect(reflect.ValueOf(v)).Type()] = f
	return mb
}

// GetModelKeyValues get model key values
func (mb *ModelBuilder) GetModelKeyValue(v interface{}) string {
	var (
		stringBuilder = strings.Builder{}
		reflectValue  = reflect.Indirect(reflect.ValueOf(v))
		reflectType   = reflectValue.Type()
	)

	for _, key := range mb.keys {
		if fields, ok := reflectType.FieldByName(key); ok {
			if reflectValue.FieldByName(key).IsZero() {
				continue
			}
			if fields.Anonymous {
				stringBuilder.WriteString(fmt.Sprintf("%v:", reflectValue.FieldByName(key).FieldByName(key).Interface()))
			} else {
				stringBuilder.WriteString(fmt.Sprintf("%v:", reflectValue.FieldByName(key).Interface()))
			}
		}
	}

	return strings.TrimRight(stringBuilder.String(), ":")
}

// AddCreateRecord add create record
func (ab *ActivityBuilder) AddCreateRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityCreate, v, db, "")
}

// AddViewRecord add view record
func (ab *ActivityBuilder) AddViewRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityView, v, db, "")
}

// AddDeleteRecord	add delete record
func (ab *ActivityBuilder) AddDeleteRecord(creator interface{}, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityDelete, v, db, "")
}

// AddEditRecord add edit record
func (ab *ActivityBuilder) AddEditRecord(creator interface{}, now interface{}, db *gorm.DB) error {
	old, ok := findOld(now)
	if !ok {
		return errors.New("can not find old value")
	}
	return ab.AddEditRecordWithOld(creator, old, now, db)
}

// AddEditRecord add edit record
func (ab *ActivityBuilder) AddEditRecordWithOld(creator interface{}, old, now interface{}, db *gorm.DB) error {
	diffs, err := ab.Diff(old, now)
	if err != nil {
		return err
	}

	if len(diffs) == 0 {
		return ab.save(creator, ActivityEdit, now, db, "")
	}

	b, err := json.Marshal(diffs)
	if err != nil {
		return err
	}

	return ab.save(creator, ActivityEdit, now, db, string(b))
}

// save log into db
func (ab *ActivityBuilder) save(creator interface{}, action string, v interface{}, db *gorm.DB, diffs string) error {
	mb, ok := ab.GetModelBuilder(v)
	if !ok {
		return errors.New("model not found")
	}

	var m = ab.NewLogModelData()
	log, ok := m.(ActivityLogInterface)
	if !ok {
		return errors.New("invalid activity log model")
	}

	log.SetCreatedAt(time.Now())
	switch user := creator.(type) {
	case string:
		log.SetCreator(user)
	case CreatorInferface:
		log.SetCreator(user.GetName())
		log.SetUserID(user.GetID())
	default:
		log.SetCreator("unknown")
	}

	log.SetAction(action)
	log.SetModelName(mb.typ.Name())
	log.SetModelKeys(mb.GetModelKeyValue(v))

	if f := mb.link; f != nil {
		log.SetModelLink(f(v))
	}

	if diffs == "" && action == ActivityEdit {
		return nil
	}

	if action == ActivityEdit {
		log.SetModelDiffs(diffs)
	}

	if db.Save(log).Error != nil {
		return db.Error
	}
	return nil
}

// Diff get diffs between old and now value
func (ab *ActivityBuilder) Diff(old, now interface{}) ([]Diff, error) {
	mb, ok := ab.GetModelBuilder(old)
	if !ok {
		return nil, errors.New("can not find model builder")
	}

	if mb.GetModelKeyValue(old) != mb.GetModelKeyValue(now) {
		return nil, errors.New("primary keys value are different") //// ignore diffs if the primary keys value are different, this situation mostly occurs when a new version is created and localized a page to the other locale
	}

	return NewDiffBuilder(mb).Diff(old, now)
}

// AddRecords add records log
func (ab *ActivityBuilder) AddRecords(action string, ctx context.Context, vs ...interface{}) error {
	if len(vs) == 0 {
		return errors.New("models are empty")
	}

	var (
		creator interface{}
		db      = ab.getDBFromContext(ctx)
	)

	if c := ctx.Value(ab.creatorContextKey); c != nil {
		creator = c
	}

	if d, ok := ctx.Value(ab.dbContextKey).(*gorm.DB); ok {
		db = d
	}

	if creator == "" || db == nil {
		return errors.New("creator or db cannot be found")
	}

	switch action {
	case ActivityView:
		for _, v := range vs {
			err := ab.AddViewRecord(creator, v, db)
			if err != nil {
				return err
			}
		}

	case ActivityDelete:
		for _, v := range vs {
			err := ab.AddDeleteRecord(creator, v, db)
			if err != nil {
				return err
			}
		}
	case ActivityCreate:
		for _, v := range vs {
			err := ab.AddCreateRecord(creator, v, db)
			if err != nil {
				return err
			}
		}
	case ActivityEdit:
		for _, v := range vs {
			err := ab.AddEditRecord(creator, v, db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterCallbackOnDB register callback on db
func (ab *ActivityBuilder) RegisterCallbackOnDB(db *gorm.DB, creatorDBKey string) {
	if creatorDBKey == "" {
		panic("creatorDBKey cannot be empty")
	}
	if db.Callback().Create().Get("activity:create") == nil {
		db.Callback().Create().After("gorm:after_create").Register("activity:create", ab.record(Create, creatorDBKey))
	}
	if db.Callback().Update().Get("activity:update") == nil {
		db.Callback().Update().Before("gorm:update").Register("activity:update", ab.record(Update, creatorDBKey))
	}
	if db.Callback().Delete().Get("activity:delete") == nil {
		db.Callback().Delete().Before("gorm:after_delete").Register("activity:delete", ab.record(Delete, creatorDBKey))
	}
}

func (ab *ActivityBuilder) record(mode uint8, creatorDBKey string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		now := db.Statement.Dest
		mb, ok := ab.GetModelBuilder(now)
		if !ok || mode&mb.disableOnCallback != 0 {
			return
		}

		creator, _ := db.Get(creatorDBKey)
		switch mode {
		case Create:
			ab.AddCreateRecord(creator, now, db.Session(&gorm.Session{NewDB: true}))
		case Delete:
			if mb.GetModelKeyValue(now) == "" { //handle the delete action from presets/gorm2op
				old, ok := findDeletedOldByWhere(db)
				if ok {
					ab.AddDeleteRecord(creator, old, db.Session(&gorm.Session{NewDB: true}))
				}
			} else {
				ab.AddDeleteRecord(creator, now, db.Session(&gorm.Session{NewDB: true}))
			}
		case Update:
			ab.AddEditRecord(creator, now, db.Session(&gorm.Session{NewDB: true}))
		}
	}
}

// GetDB get db from context
func (ab *ActivityBuilder) getDBFromContext(ctx context.Context) *gorm.DB {
	if contextdb := ctx.Value(ab.dbContextKey); contextdb != nil {
		return contextdb.(*gorm.DB)
	}
	return GlobalDB
}
