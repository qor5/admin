package activity

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	Create = 1 << iota
	Delete
	Update
)

type User struct {
	ID     uint
	Name   string
	Avatar string
}

type CurrentUserFunc func(ctx context.Context) *User

// @snippet_begin(ActivityBuilder)
// Builder struct contains all necessary fields
type Builder struct {
	db              *gorm.DB                  // global db
	lmb             *presets.ModelBuilder     // log model builder
	models          []*ModelBuilder           // registered model builders
	tabHeading      func(*ActivityLog) string // tab heading format
	permPolicy      *perm.PolicyBuilder       // permission policy
	logModelInstall presets.ModelInstallFunc  // log model install
	currentUserFunc CurrentUserFunc
}

// @snippet_end

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	b.RegisterModel(m)
	// TODO: 应该写到 RegisterModel 里面？
	m.RegisterEventFunc(eventCreateNote, createNote(b, m))
	m.RegisterEventFunc(eventUpdateNote, updateNote(b, m))
	m.RegisterEventFunc(eventDeleteNote, deleteNote(b, m))
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

// New initializes a new Builder instance with a provided database connection and an optional activity log model.
func New(db *gorm.DB) *Builder {
	ab := &Builder{
		db: db,
	}

	ab.logModelInstall = ab.defaultLogModelInstall

	ab.permPolicy = perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
		ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).On("*:activity_logs").On("*:activity_logs:*")

	return ab
}

func (ab *Builder) GetActivityLogs(m interface{}, keys string, db *gorm.DB) []*ActivityLog {
	if keys == "" {
		keys = ab.MustGetModelBuilder(m).KeysValue(m)
	}
	var logs []*ActivityLog
	err := db.Where("model_name = ? AND model_keys = ?", modelName(m), keys).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil
	}
	return logs
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

func humanContent(log *ActivityLog) string {
	return fmt.Sprintf("%s: %s", log.Action, log.Comment)
}

func objectID(obj interface{}) string {
	var id string
	if slugger, ok := obj.(presets.SlugEncoder); ok {
		id = slugger.PrimarySlug()
	} else {
		id = fmt.Sprint(reflectutils.MustGet(obj, "ID"))
	}
	return id
}

func (ab *Builder) installModelBuilder(mb *ModelBuilder, presetModel *presets.ModelBuilder) {
	mb.presetModel = presetModel
	mb.LinkFunc(func(a any) string {
		return presetModel.Info().DetailingHref(objectID(a))
	})

	editing := presetModel.Editing()
	d := presetModel.Detailing()
	lb := presetModel.Listing()

	db := ab.db

	d.Field(DetailFieldTimeline).ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return web.Portal(ab.timelineList(obj, "", db)).Name(TimelinePortalName)
	})

	lb.Field(ListFieldUnreadNotes).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		rt := modelName(presetModel.NewModel())
		ri := mb.KeysValue(obj)
		user := ab.currentUserFunc(ctx.R.Context())
		count, _ := GetUnreadNotesCount(db, user.ID, rt, ri)

		return h.Td(
			h.If(count > 0,
				v.VBadge().Content(count).Color("red"),
			).Else(
				h.Text(""),
			),
		)
	}).Label("Unread Notes") // TODO: i18n

	// TODO: 这个的话会丢失掉一些通过 action 来操作的信息，所以需要通过 emit 来做？但是通过 emit 的话貌似又是需要要求 emit 给到 ctx 信息
	// TODO: section 目前没使用 editing 的 SaveFunc
	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Update != 0 && mb.skip&Create != 0 {
				return in(obj, id, ctx)
			}

			old, ok := findOld(obj, db)
			if err = in(obj, id, ctx); err != nil {
				return err
			}

			if (!ok || id == "") && mb.skip&Create == 0 {
				return mb.AddRecords(ctx.R.Context(), ActionCreate, obj)
			}

			if ok && id != "" && mb.skip&Update == 0 {
				return mb.AddEditRecordWithOld(ab.currentUserFunc(ctx.R.Context()), old, obj, db)
			}

			return
		}
	})

	editing.WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Delete != 0 {
				return in(obj, id, ctx)
			}

			old, ok := findOldWithSlug(obj, id, db)
			if err = in(obj, id, ctx); err != nil {
				return err
			}

			if ok {
				return mb.AddRecords(ctx.R.Context(), ActionDelete, old)
			}

			return
		}
	})
}

func (ab *Builder) timelineList(obj any, keys string, db *gorm.DB) h.HTMLComponent {
	children := []h.HTMLComponent{
		// TODO: onClick
		v.VBtn("Add Notes").Class("text-none mb-4").Attr("prepend-icon", "mdi-plus").Attr("variant", "tonal").Attr("color", "grey-darken-3"), // TODO: i18n
	}

	logs := ab.GetActivityLogs(obj, keys, db)
	for _, item := range logs {
		creatorName := item.Creator.Name
		if creatorName == "" {
			creatorName = "Unknown" // i18n
		}
		avatarText := ""
		if item.Creator.Avatar == "" {
			avatarText = strings.ToUpper(creatorName[0:1])
		}
		// TODO: v.ColorXXX ?
		// TODO: 不需要支持非 Notes 吗？
		children = append(children,
			h.Div().Class("d-flex flex-column ga-1").Children(
				h.Div().Class("d-flex flex-row align-center ga-2").Children(
					h.Div().Style("width: 8px; height: 8px; background-color: #30a46c").Class("rounded-circle"),
					h.Div(h.Text(humanize.Time(item.CreatedAt))).Style("color: #757575"),
				),
				h.Div().Class("d-flex flex-row ga-2").Children(
					h.Div().Class("align-self-stretch").Style("width: 1px; margin-top: -6px; margin-bottom: -2px; margin-left: 3.5px; margin-right: 3.5px; background-color: #30a46c"),
					h.Div().Class("d-flex flex-column pb-3").Children(
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							v.VAvatar().Attr("style", "font-size: 12px; color: #3e63dd").Attr("color", "#E6EDFE").Attr("size", "x-small").Attr("density", "compact").Attr("rounded", true).Text(avatarText).Children(
								h.Iff(item.Creator.Avatar != "", func() h.HTMLComponent {
									return v.VImg().Attr("alt", creatorName).Attr("src", item.Creator.Avatar)
								}),
							),
							h.Div(h.Text(creatorName)).Style("font-weight: 500"),
						),
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							h.Div().Style("width: 16px"),
							h.Div(h.Text(humanContent(item))),
						),
					),
				),
			),
		)
	}
	return h.Div().Class("d-flex flex-column").Style("font-size: 14px").Children(
		children...,
	)
}

// GetModelBuilder 	get model builder
func (ab *Builder) GetModelBuilder(v any) (*ModelBuilder, bool) {
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
func (ab *Builder) MustGetModelBuilder(v any) *ModelBuilder {
	mb, ok := ab.GetModelBuilder(v)
	if !ok {
		panic(fmt.Sprintf("model %v is not registered", v))
	}
	return mb
}

// GetModelBuilders 	get all model builders
func (ab *Builder) GetModelBuilders() []*ModelBuilder {
	return ab.models
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
func (ab *Builder) AddViewRecord(creator *User, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddViewRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddDeleteRecord	add delete record
func (ab *Builder) AddDeleteRecord(creator *User, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddDeleteRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddSaverRecord will save a create log or a edit log
func (ab *Builder) AddSaveRecord(creator *User, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddSaveRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddCreateRecord add create record
func (ab *Builder) AddCreateRecord(creator *User, v any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(v); ok {
		return mb.AddCreateRecord(creator, v, db)
	}

	return fmt.Errorf("can't find model builder for %v", v)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecord(creator *User, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecord(creator, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddEditRecord add edit record
func (ab *Builder) AddEditRecordWithOld(creator *User, old, now any, db *gorm.DB) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecordWithOld(creator, old, now, db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

// AddEditRecordWithOldAndContext add edit record
func (ab *Builder) AddEditRecordWithOldAndContext(ctx context.Context, old, now any) error {
	if mb, ok := ab.GetModelBuilder(now); ok {
		return mb.AddEditRecordWithOld(ab.currentUserFunc(ctx), old, now, ab.db)
	}

	return fmt.Errorf("can't find model builder for %v", now)
}

func (b *Builder) AutoMigrate() (r *Builder) {
	if err := AutoMigrate(b.db); err != nil {
		panic(err)
	}
	return b
}

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(&ActivityLog{})
}

func modelName(v any) string {
	segs := strings.Split(reflect.TypeOf(v).String(), ".")
	return segs[len(segs)-1]
}
