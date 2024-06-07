package activity

import (
	"context"
	"errors"
	"fmt"
	"go4.org/sort"
	"reflect"
	"strings"
	"time"

	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
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
	DBContextKey                 = 1 // get the db from context
)

// @snippet_begin(ActivityBuilder)
type Builder struct {
	db                *gorm.DB                          // global db
	creatorContextKey any                               // get the creator from context
	dbContextKey      any                               // get the db from context
	lmb               *presets.ModelBuilder             // log model builder
	logModel          ActivityLogInterface              // log model
	models            []*ModelBuilder                   // registered model builders
	tabHeading        func(ActivityLogInterface) string // tab heading format
	permPolicy        *perm.PolicyBuilder               // permission policy
	logModelInstall   presets.ModelInstallFunc          // log model install
}

type TimelineItem struct {
	Timestamp   time.Time
	Description string
	User        string
	Icon        string
}

// @snippet_end

func (ab *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	ab.RegisterModel(m)
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

// New initializes a new Builder instance with a provided database connection and an optional activity log model.
func New(db *gorm.DB, logModel ...ActivityLogInterface) *Builder {
	ab := &Builder{
		db:                db,
		creatorContextKey: CreatorContextKey,
		dbContextKey:      DBContextKey,
		logModel:          &ActivityLog{},
	}

	if len(logModel) > 0 {
		ab.logModel = logModel[0]
	}

	if err := db.AutoMigrate(ab.logModel); err != nil {
		fmt.Printf("failed to migrate log model: %w", err)
		return nil
	}

	// Default permission policy
	ab.permPolicy = perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
		ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).On("*:activity_logs").On("*:activity_logs:*")

	return ab
}

// GetCustomizeActivityLogs get customize activity logs
func (ab Builder) GetCustomizeActivityLogs(m any, db *gorm.DB) any {
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
func (ab Builder) NewLogModelData() any {
	return reflect.New(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type()).Interface()
}

// NewLogModelSlice new a log model slice
func (ab Builder) NewLogModelSlice() any {
	sliceType := reflect.SliceOf(reflect.PointerTo(reflect.Indirect(reflect.ValueOf(ab.logModel)).Type()))
	slice := reflect.New(sliceType)
	slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))
	return slice.Interface()
}

// CreatorContextKey change the default creator context key
func (ab *Builder) CreatorContextKey(key any) *Builder {
	ab.creatorContextKey = key
	return ab
}

// DBContextKey change the default db context key
func (ab *Builder) DBContextKey(key any) *Builder {
	ab.dbContextKey = key
	return ab
}

func (ab *Builder) TabHeading(f func(log ActivityLogInterface) string) *Builder {
	ab.tabHeading = f
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
func (ab *Builder) RegisterModel(m any) (mb *ModelBuilder) {
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
		ab.installModelBuilder(mb, presetModel)
	}

	return mb
}

func (ab *Builder) installModelBuilder(mb *ModelBuilder, presetModel *presets.ModelBuilder) {
	mb.presetModel = presetModel

	editing := presetModel.Editing()
	d := presetModel.Detailing()

	d.Field(Timeline).ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// Fetch combined timeline data
		timelineData := fetchTimelineData(ctx)

		// Create VTimelineItems
		var timelineItems []h.HTMLComponent
		for _, item := range timelineData {
			timelineItems = append(timelineItems,
				h.Div(
					h.Div(
						h.Text(item.Timestamp.Format("2 hours ago")),
					),
					h.Div(
						h.Div(
							vuetify.VAvatar().Text(strings.ToUpper(string(item.User[0]))).Color("secondary").Class("text-h6 rounded-lg").Size("x-small"),
							h.Div(
								h.Strong(item.User).Class("ml-1").Style("width: 100%; height: 20px; font-family: SF Pro; font-style: normal; font-weight: 510; font-size: 14px; line-height: 20px; display: flex; align-items: center; color: #9e9e9e;"),
								h.Div(h.Text(item.Description)).Class("text-caption").Style("width: 100%; font-family: SF Pro; font-style: normal; font-weight: 400; font-size: 14px; line-height: 20px; color: #9e9e9e;"),
							).Class("detailsStyle").Style("display: flex; flex-direction: column; align-items: flex-start; padding: 0; width: 100%;"),
						).Class("contentStyle").Style("display: flex; flex-direction: row; align-items: flex-start; padding: 0; gap: 8px; width: 100%;"),
					),
				).Class("itemStyle").Style("width: 100%; color: #9e9e9e;"),
			)
		}

		return vuetify.VTimeline(
			timelineItems...,
		)
	})

	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Update != 0 && mb.skip&Create != 0 {
				return in(obj, id, ctx)
			}

			old, ok := findOld(obj, ab.getDBFromContext(ctx.R.Context()))
			if err = in(obj, id, ctx); err != nil {
				return err
			}

			if (!ok || id == "") && mb.skip&Create == 0 {
				return mb.AddRecords(ActivityCreate, ctx.R.Context(), obj)
			}

			if ok && id != "" && mb.skip&Update == 0 {
				return mb.AddEditRecordWithOld(ab.getCreatorFromContext(ctx.R.Context()), old, obj, ab.getDBFromContext(ctx.R.Context()))
			}

			return
		}
	})

	editing.WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Delete != 0 {
				return in(obj, id, ctx)
			}

			old, ok := findOldWithSlug(obj, id, ab.getDBFromContext(ctx.R.Context()))
			if err = in(obj, id, ctx); err != nil {
				return err
			}

			if ok {
				return mb.AddRecords(ActivityDelete, ctx.R.Context(), old)
			}

			return
		}
	})
}

func fetchTimelineData(ctx *web.EventContext) []TimelineItem {
	var timelineData []TimelineItem
	db := ctx.R.Context().Value(DBContextKey).(*gorm.DB)

	var logs []ActivityLog

	if err := db.Find(&logs).Error; err != nil {
		fmt.Printf("Database query failure: %v", err)
		return nil
	}

	for _, log := range logs {
		timelineData = append(timelineData, TimelineItem{
			Timestamp:   log.CreatedAt,
			Description: log.Action + ": " + log.Comments,
			User:        log.Creator,
			Icon:        "mdi-activity",
		})
	}

	sort.Slice(timelineData, func(i, j int) bool {
		return timelineData[i].Timestamp.Before(timelineData[j].Timestamp)
	})

	return timelineData
}

// GetModelBuilder 	get model builder
func (ab Builder) GetModelBuilder(v any) (*ModelBuilder, bool) {
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
func (ab Builder) MustGetModelBuilder(v any) *ModelBuilder {
	mb, ok := ab.GetModelBuilder(v)
	if !ok {
		panic(fmt.Sprintf("model %v is not registered", v))
	}
	return mb
}

// GetModelBuilders 	get all model builders
func (ab Builder) GetModelBuilders() []*ModelBuilder {
	return ab.models
}

// AddRecords add records log
func (ab *Builder) AddRecords(action string, ctx context.Context, vs ...any) error {
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
func (ab *Builder) AddCustomizedRecord(action string, diff bool, ctx context.Context, obj any) error {
	if mb, ok := ab.GetModelBuilder(obj); ok {
		return mb.AddCustomizedRecord(action, diff, ctx, obj)
	}

	return fmt.Errorf("can't find model builder for %v", obj)
}

// AddViewRecord add view record
func (ab *Builder) AddViewRecord(creator any, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddViewRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddDeleteRecord	add delete record
func (ab *Builder) AddDeleteRecord(creator any, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddDeleteRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddSaverRecord will save a create log or a edit log
func (ab *Builder) AddSaveRecord(creator any, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddSaveRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddCreateRecord add create record
func (ab *Builder) AddCreateRecord(creator any, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddCreateRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecord(creator any, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecordWithOld(creator any, old, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecordWithOld(creator, old, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddEditRecordWithOldAndContext add edit record
func (ab *Builder) AddEditRecordWithOldAndContext(ctx context.Context, old, now any) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecordWithOld(ab.getCreatorFromContext(ctx), old, now, ab.getDBFromContext(ctx))
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// GetDB get db from context
func (ab *Builder) getDBFromContext(ctx context.Context) *gorm.DB {
	if contextdb := ctx.Value(ab.dbContextKey); contextdb != nil {
		return contextdb.(*gorm.DB)
	}
	return ab.db
}

// GetDB get creator from context
func (ab *Builder) getCreatorFromContext(ctx context.Context) any {
	if creator := ctx.Value(ab.creatorContextKey); creator != nil {
		return creator
	}
	return ""
}
