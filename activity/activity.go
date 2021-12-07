package activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	vuetify "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	Create = 1 << iota
	Delete
	Update
	All = Create | Delete | Update
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

	logModel   ActivityLogInterface              // log model
	models     []*ModelBuilder                   // registered model builders
	tabHeading func(ActivityLogInterface) string // tab heading format
}

// @snippet_end

// @snippet_begin(ActivityModelBuilder)
type ModelBuilder struct {
	typ      reflect.Type
	mb       *presets.ModelBuilder
	activity *ActivityBuilder

	keys          []string                     // primary keys
	ignoredFields []string                     // ignored fields
	typeHanders   map[reflect.Type]TypeHandler // type handlers
	link          func(interface{}) string     // display the model link on the admin detail page
	skip          uint8                        // skip the default data operator
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
func (ab *ActivityBuilder) Models(models ...*presets.ModelBuilder) *ActivityBuilder {
	for _, model := range models {
		ab.Model(model)
	}
	return ab
}

// Model register a model and return model builder
func (ab *ActivityBuilder) Model(mb *presets.ModelBuilder) *ModelBuilder {
	reflectType := reflect.Indirect(reflect.ValueOf(mb.NewModel())).Type()
	if reflectType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%v is not a struct", reflectType.Name()))
	}

	keys := getPrimaryKey(reflectType)
	model := &ModelBuilder{
		typ:           reflectType,
		mb:            mb,
		activity:      ab,
		keys:          keys,
		ignoredFields: keys,
	}

	ab.models = append(ab.models, model)

	editing := mb.Editing()
	oldSaver := editing.Saver
	oldDeleter := editing.Deleter
	editing.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		if model.skip&Update != 0 && model.skip&Create != 0 {
			return oldSaver(obj, id, ctx)
		}

		old, ok := findOld(obj, ab.getDBFromContext(ctx.R.Context()))
		if err = oldSaver(obj, id, ctx); err != nil {
			return err
		}

		if (!ok || id == "") && model.skip&Create == 0 {
			return ab.AddRecords(ActivityCreate, ctx.R.Context(), obj)
		}

		if ok && id != "" && model.skip&Update == 0 {
			return ab.AddEditRecordWithOld(ctx.R.Context().Value(ab.creatorContextKey), old, obj, ab.getDBFromContext(ctx.R.Context()))
		}

		return
	})

	editing.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		if model.skip&Delete != 0 {
			return oldDeleter(obj, id, ctx)
		}

		old, ok := findOldWithSlug(obj, id, ab.getDBFromContext(ctx.R.Context()))
		if err = oldDeleter(obj, id, ctx); err != nil {
			return err
		}

		if ok {
			return ab.AddRecords(ActivityDelete, ctx.R.Context(), old)
		}

		return
	})

	return model
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

func (mb *ModelBuilder) SkipCreate() *ModelBuilder {
	if mb.skip&Create == 0 {
		mb.skip |= Create
	}
	return mb
}

func (mb *ModelBuilder) SkipUpdate() *ModelBuilder {
	if mb.skip&Update == 0 {
		mb.skip |= Update
	}
	return mb
}

func (mb *ModelBuilder) SkipDelete() *ModelBuilder {
	if mb.skip&Delete == 0 {
		mb.skip |= Delete
	}
	return mb
}

func (mb *ModelBuilder) UseDefaultTab() *ModelBuilder {
	editing := mb.mb.Editing()
	editing.AppendTabsPanelFunc(func(obj interface{}, ctx *web.EventContext) (c h.HTMLComponent) {
		logs := mb.activity.GetCustomizeActivityLogs(obj, mb.activity.getDBFromContext(ctx.R.Context()))
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		logList := h.HTMLComponents{}

		logsvalues := reflect.Indirect(reflect.ValueOf(logs))
		for i := 0; i < logsvalues.Len(); i++ {
			log := logsvalues.Index(i).Interface().(ActivityLogInterface)
			var headerText string
			if mb.activity.tabHeading != nil {
				headerText = mb.activity.tabHeading(log)
			} else {
				headerText = fmt.Sprintf("%s %s at %s", log.GetCreator(), log.GetAction(), log.GetCreatedAt().String())
			}
			record := h.Div(
				vuetify.VExpansionPanels(
					vuetify.VExpansionPanel(
						vuetify.VExpansionPanelHeader(h.Span(headerText)),
						vuetify.VExpansionPanelContent(DiffComponent(log.GetModelDiffs(), ctx.R)),
					),
				),
			)
			logList = append(logList, record)
		}

		return h.Components(
			vuetify.VTab(h.Text(msgr.Activities)),
			vuetify.VTabItem(logList),
		)
	})

	return mb
}

// AddIgnoredFields add ignored fields to the model builder
func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = append(mb.ignoredFields, fields...)
	return mb
}

// SetIgnoredFields set ignored fields for the model builder
func (mb *ModelBuilder) SetIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = fields
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

// KeysValue get model keys value
func (mb *ModelBuilder) KeysValue(v interface{}) string {
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

// AddRecords add records log
func (ab *ActivityBuilder) AddRecords(action string, ctx context.Context, vs ...interface{}) error {
	if len(vs) == 0 {
		return errors.New("data are empty")
	}

	var (
		creator = ctx.Value(ab.creatorContextKey)
		db      = ab.getDBFromContext(ctx)
	)

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
	old, ok := findOld(now, db)
	if !ok {
		return ab.AddCreateRecord(creator, now, db)
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

// Diff get diffs between old and now value
func (ab *ActivityBuilder) Diff(old, now interface{}) ([]Diff, error) {
	mb, ok := ab.GetModelBuilder(old)
	if !ok {
		return nil, fmt.Errorf("can not find type %T on activity", old)
	}

	return NewDiffBuilder(mb).Diff(old, now)
}

// GetDB get db from context
func (ab *ActivityBuilder) getDBFromContext(ctx context.Context) *gorm.DB {
	if contextdb := ctx.Value(ab.dbContextKey); contextdb != nil {
		return contextdb.(*gorm.DB)
	}
	return ab.db
}

// save log into db
func (ab *ActivityBuilder) save(creator interface{}, action string, v interface{}, db *gorm.DB, diffs string) error {
	mb, ok := ab.GetModelBuilder(v)
	if !ok {
		return fmt.Errorf("can not find type %T on activity", v)
	}

	var m = ab.NewLogModelData()
	log, ok := m.(ActivityLogInterface)
	if !ok {
		return fmt.Errorf("model %T is not implement ActivityLogInterface", m)
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
	log.SetModelKeys(mb.KeysValue(v))

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
