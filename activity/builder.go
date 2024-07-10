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
	"github.com/samber/lo"
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
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
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
	findUsersFunc   func(ctx context.Context, ids []uint) (map[uint]*User, error)
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

func (ab *Builder) FindUsersFunc(v func(ctx context.Context, ids []uint) (map[uint]*User, error)) *Builder {
	ab.findUsersFunc = v
	return ab
}

func (ab *Builder) findUsers(ctx context.Context, ids []uint) (map[uint]*User, error) {
	if ab.findUsersFunc != nil {
		return ab.findUsersFunc(ctx, ids)
	}
	vs := []*ActivityUser{}
	err := ab.db.Where("id IN ?", ids).Find(&vs).Error
	if err != nil {
		return nil, err
	}
	return lo.SliceToMap(vs, func(item *ActivityUser) (uint, *User) {
		return item.ID, &User{
			ID:     item.ID,
			Name:   item.Name,
			Avatar: item.Avatar,
		}
	}), nil
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

func (ab *Builder) supplyCreators(ctx context.Context, logs []*ActivityLog) error {
	creatorIDs := lo.Uniq(lo.Map(logs, func(log *ActivityLog, _ int) uint {
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

func (ab *Builder) GetActivityLogs(ctx context.Context, m any, keys string) ([]*ActivityLog, error) {
	if keys == "" {
		keys = ab.MustGetModelBuilder(m).KeysValue(m)
	}
	var logs []*ActivityLog
	err := ab.db.Where("model_name = ? AND model_keys = ?", modelName(m), keys).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	if err := ab.supplyCreators(ctx, logs); err != nil {
		return nil, err
	}
	return logs, nil
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
		typ: reflectType,
		ab:  ab,

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

func objectID(obj any) string {
	var id string
	if slugger, ok := obj.(presets.SlugEncoder); ok {
		id = slugger.PrimarySlug()
	} else {
		v, err := reflectutils.Get(obj, "ID")
		if err == nil {
			id = fmt.Sprint(v)
		}
	}
	return id
}

func (ab *Builder) installModelBuilder(mb *ModelBuilder, presetModel *presets.ModelBuilder) {
	mb.presetModel = presetModel
	mb.LinkFunc(func(a any) string {
		id := objectID(a)
		if id == "" {
			return id
		}
		return presetModel.Info().DetailingHref(id)
	})

	presetModel.RegisterEventFunc(eventCreateNote, createNote(ab, presetModel))
	// TODO: 可以修改和删除？
	presetModel.RegisterEventFunc(eventUpdateNote, updateNote(ab, presetModel))
	presetModel.RegisterEventFunc(eventDeleteNote, deleteNote(ab, presetModel))

	editing := presetModel.Editing()
	d := presetModel.Detailing()
	lb := presetModel.Listing()

	db := ab.db

	d.Field(DetailFieldTimeline).ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return web.Portal(ab.timelineList(ctx.R.Context(), obj, "")).Name(TimelinePortalName) // TODO: 这里没给到 keys，可以吗？
	})

	// TODO: 需要判断是否存在此 field
	lb.Field(ListFieldUnreadNotes).ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
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

	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Update != 0 && mb.skip&Create != 0 {
				return in(obj, id, ctx)
			}

			old, ok := fetchOld(obj, db)
			if err = in(obj, id, ctx); err != nil {
				return err
			}

			if (!ok || id == "") && mb.skip&Create == 0 {
				return mb.AddRecords(ctx.R.Context(), ActionCreate, obj)
			}

			if ok && id != "" && mb.skip&Update == 0 {
				return mb.AddEditRecordWithOld(ab.currentUserFunc(ctx.R.Context()), old, obj)
			}

			return
		}
	})

	editing.WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if mb.skip&Delete != 0 {
				return in(obj, id, ctx)
			}

			old, ok := fetchOldWithSlug(obj, id, db)
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

func (ab *Builder) timelineList(ctx context.Context, obj any, keys string) h.HTMLComponent {
	children := []h.HTMLComponent{
		// TODO: onClick
		v.VBtn("Add Notes").Class("text-none mb-4").Attr("prepend-icon", "mdi-plus").Attr("variant", "tonal").Attr("color", "grey-darken-3"), // TODO: i18n
	}

	logs, err := ab.GetActivityLogs(ctx, obj, keys)
	if err != nil {
		panic(err)
	}

	for i, log := range logs {
		creatorName := log.Creator.Name
		if creatorName == "" {
			creatorName = "Unknown" // i18n
		}
		avatarText := ""
		if log.Creator.Avatar == "" {
			avatarText = strings.ToUpper(creatorName[0:1])
		}
		dotColor := "#30a46c"
		if i != 0 {
			dotColor = "#e0e0e0"
		}
		// TODO: v.ColorXXX ?
		children = append(children,
			h.Div().Class("d-flex flex-column ga-1").Children(
				h.Div().Class("d-flex flex-row align-center ga-2").Children(
					h.Div().Style("width: 8px; height: 8px; background-color: "+dotColor).Class("rounded-circle"),
					h.Div(h.Text(humanize.Time(log.CreatedAt))).Style("color: #757575"),
				),
				h.Div().Class("d-flex flex-row ga-2").Children(
					h.Div().Class("align-self-stretch").Style("background-color: "+dotColor+"; width: 1px; margin-top: -6px; margin-bottom: -2px; margin-left: 3.5px; margin-right: 3.5px;"),
					h.Div().Class("d-flex flex-column pb-3").Children(
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							v.VAvatar().Attr("style", "font-size: 12px; color: #3e63dd").Attr("color", "#E6EDFE").Attr("size", "x-small").Attr("density", "compact").Attr("rounded", true).Text(avatarText).Children(
								h.Iff(log.Creator.Avatar != "", func() h.HTMLComponent {
									return v.VImg().Attr("alt", creatorName).Attr("src", log.Creator.Avatar)
								}),
							),
							h.Div(h.Text(creatorName)).Style("font-weight: 500"),
						),
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							h.Div().Style("width: 16px"),
							h.Div(h.Text(humanContent(log))),
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

func (b *Builder) AutoMigrate() (r *Builder) {
	if err := AutoMigrate(b.db); err != nil {
		panic(err)
	}
	return b
}

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(&ActivityLog{}, &ActivityUser{})
}

func modelName(v any) string {
	segs := strings.Split(reflect.TypeOf(v).String(), ".")
	return segs[len(segs)-1]
}
