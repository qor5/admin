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
	"gorm.io/gorm/schema"
)

type contextKey string

const (
	CreatorContextKey contextKey = "Creator"
	DBContextKey      contextKey = "DB"
)

type ActivityBuilder struct {
	useL10n           bool
	creatorContextKey interface{}
	dbContextKey      interface{}
	models            []*ModelBuilder
}

type ModelBuilder struct {
	name          string
	keys          []string
	link          func(interface{}) string
	ignoredFields []string
	typeHanders   map[reflect.Type]TypeHandle
}

func Activity() *ActivityBuilder {
	return &ActivityBuilder{
		creatorContextKey: CreatorContextKey,
		dbContextKey:      DBContextKey,
	}
}

func (ab *ActivityBuilder) UseL10n() *ActivityBuilder {
	ab.useL10n = true
	return ab
}

func (ab *ActivityBuilder) SetCreatorContextKey(key interface{}) *ActivityBuilder {
	ab.creatorContextKey = key
	return ab
}

func (ab *ActivityBuilder) SetDBContextKey(key interface{}) *ActivityBuilder {
	ab.dbContextKey = key
	return ab
}

func (ab *ActivityBuilder) RegisterModel(model interface{}) *ModelBuilder {
	mb := &ModelBuilder{name: reflect.Indirect(reflect.ValueOf(model)).Type().Name()}
	ab.models = append(ab.models, mb)
	return mb
}

func (mb *ModelBuilder) SetKeys(keys ...string) *ModelBuilder {
	mb.keys = append(mb.keys, keys...)
	return mb
}

func (mb *ModelBuilder) SetLink(f func(interface{}) string) *ModelBuilder {
	mb.link = f
	return mb
}

func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = append(mb.ignoredFields, fields...)
	return mb
}

func (mb *ModelBuilder) AddTypeHanders(v interface{}, f TypeHandle) *ModelBuilder {
	if mb.typeHanders == nil {
		mb.typeHanders = map[reflect.Type]TypeHandle{}
	}
	mb.typeHanders[reflect.Indirect(reflect.ValueOf(v)).Type()] = f
	return mb
}

func (mb *ModelBuilder) getModelKey(v interface{}) string {
	var (
		stringBuilder = strings.Builder{}
		reflectValue  = reflect.Indirect(reflect.ValueOf(v))
	)

	if len(mb.keys) == 0 {
		if !reflectValue.FieldByName("ID").IsZero() {
			stringBuilder.WriteString(fmt.Sprintf("%v", reflectValue.FieldByName("ID").Interface()))
		}
	}

	for _, key := range mb.keys {
		if !reflectValue.FieldByName(key).IsZero() {
			stringBuilder.WriteString(fmt.Sprintf("%v:", reflectValue.FieldByName(key).Interface()))
		}
	}

	return strings.TrimRight(stringBuilder.String(), ":")
}

func (ab *ActivityBuilder) GetModelBuilder(v interface{}) *ModelBuilder {
	name := reflect.Indirect(reflect.ValueOf(v)).Type().Name()
	for _, m := range ab.models {
		if m.name == name {
			return m
		}
	}
	return &ModelBuilder{}
}

func (ab *ActivityBuilder) GetActivityLogModel() ActivityLogInterface {
	if ab.useL10n {
		return &ActivityLocaleLog{}
	}
	return &ActivityLog{}
}

func (ab *ActivityBuilder) save(creator string, action string, v interface{}, db *gorm.DB, diffs string) error {
	var (
		mb  = ab.GetModelBuilder(v)
		log = ab.GetActivityLogModel()
	)

	log.SetCreatedAt(time.Now())
	log.SetCreator(creator)
	log.SetAction(action)
	log.SetModelName(mb.name)
	log.SetModelKeys(mb.getModelKey(v))

	if f := mb.link; f != nil {
		log.SetModelLink(f(v))
	}

	if diffs != "" && action == ActivityEdit {
		log.SetModelDiffs(diffs)
	}

	return db.Save(log).Error
}

func (ab *ActivityBuilder) AddCreateRecord(creator string, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityCreate, v, db, "")
}

func (ab *ActivityBuilder) AddViewRecord(creator string, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityView, v, db, "")
}

func (ab *ActivityBuilder) AddDeleteRecord(creator string, v interface{}, db *gorm.DB) error {
	return ab.save(creator, ActivityDelete, v, db, "")
}

func (ab *ActivityBuilder) AddEditRecord(creator string, old, now interface{}, db *gorm.DB) error {
	diffs, err := ab.Diff(old, now)
	if err != nil {
		return err
	}
	b, err := json.Marshal(diffs)
	if err != nil {
		return err
	}
	return ab.save(creator, ActivityEdit, now, db, string(b))
}

func (ab *ActivityBuilder) Diff(old, now interface{}) ([]Diff, error) {
	return NewDiffBuilder(ab.GetModelBuilder(old)).Diff(old, now)
}

func (ab *ActivityBuilder) AddRecords(action string, ctx context.Context, vs ...interface{}) error {
	if len(vs) == 0 {
		return errors.New("models are empty")
	}

	var (
		creator string
		db      *gorm.DB
	)

	if c, ok := ctx.Value(ab.creatorContextKey).(string); ok {
		creator = c
	}

	if d, ok := ctx.Value(ab.dbContextKey).(*gorm.DB); ok {
		db = d
	}

	if creator == "" || db == nil {
		return errors.New("creator and db cannot be found from the context")
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
		if len(vs) == 2 {
			return ab.AddEditRecord(creator, vs[0], vs[1], db)
		}
	}
	return nil
}

func (ab *ActivityBuilder) HasModel(v interface{}) bool {
	name := reflect.Indirect(reflect.ValueOf(v)).Type().Name()
	for _, m := range ab.models {
		if m.name == name {
			return true
		}
	}
	return false
}

func (ab *ActivityBuilder) RegisterCallbackOnDB(db *gorm.DB, creatorDBKey string) {
	if db.Callback().Create().Get("activity:create") == nil {
		db.Callback().Create().After("gorm:after_create").Register("activity:create", ab.record(ActivityCreate, creatorDBKey))
	}
	if db.Callback().Update().Get("activity:update") == nil {
		db.Callback().Update().Before("gorm:update").Register("activity:update", ab.record(ActivityEdit, creatorDBKey))
	}
	if db.Callback().Delete().Get("activity:delete") == nil {
		db.Callback().Delete().Before("gorm:after_delete").Register("activity:delete", ab.record(ActivityDelete, creatorDBKey))
	}
}

func (ab *ActivityBuilder) record(mode, creatorDBKey string) func(*gorm.DB) {
	if creatorDBKey == "" {
		panic("creatorDBKey cannot be empty")
	}

	return func(db *gorm.DB) {
		model := db.Statement.Model
		if !ab.HasModel(model) {
			return
		}

		var (
			userName string
		)

		if user, ok := db.Get(creatorDBKey); ok {
			if u, ok := user.(string); ok {
				userName = u
			}
		}

		switch mode {
		case ActivityCreate:
			ab.AddCreateRecord(userName, model, db.Session(&gorm.Session{NewDB: true}))
		case ActivityDelete:
			ab.AddDeleteRecord(userName, model, db)
		case ActivityEdit:
			modelBuilder := ab.GetModelBuilder(model)
			reflectValue := reflect.Indirect(reflect.ValueOf(model))
			old := reflect.New(db.Statement.ReflectValue.Type()).Interface()
			if len(modelBuilder.keys) == 0 {
				if !reflectValue.FieldByName("ID").IsZero() {
					db.Session(&gorm.Session{NewDB: true}).Where("id = ?", reflectValue.FieldByName("ID").Interface()).Find(old)
					ab.AddEditRecord(userName, old, model, db.Session(&gorm.Session{NewDB: true}))
				}
			} else {
				newdb := db.Session(&gorm.Session{NewDB: true})
				for _, key := range modelBuilder.keys {
					newdb = newdb.Where(fmt.Sprintf("%s = ?", (schema.NamingStrategy{}).ColumnName("", key)), reflectValue.FieldByName(key).Interface())
				}
				newdb.Find(old)
				ab.AddEditRecord(userName, old, model, db.Session(&gorm.Session{NewDB: true}))
			}
		}
	}
}

func ContextWithCreator(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, CreatorContextKey, name)
}

func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBContextKey, db)
}
