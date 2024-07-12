package activity

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
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

const InjectorTop = "_actitivy_top_"

type User struct {
	ID     string `json:"id"`
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
		ab.installPresetsModelBuilder(mb, presetModel)
	}

	return mb
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

func (ab *Builder) installPresetsModelBuilder(amb *ModelBuilder, mb *presets.ModelBuilder) {
	amb.presetModel = mb
	amb.LinkFunc(func(a any) string {
		id := objectID(a)
		if id == "" {
			return id
		}
		return mb.Info().DetailingHref(id)
	})

	eb := mb.Editing()
	dp := mb.Detailing()
	lb := mb.Listing()

	eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if id == "" && amb.skip&Create == 0 {
				err := in(obj, id, ctx)
				if err != nil {
					return err
				}
				log, err := amb.createWithObj(ctx.R.Context(), ActionCreate, obj, nil)
				if err != nil {
					return err
				}
				presets.WrapEventFuncAddon(ctx, func(in presets.EventFuncAddon) presets.EventFuncAddon {
					return func(ctx *web.EventContext, r *web.EventResponse) (err error) {
						if err = in(ctx, r); err != nil {
							return
						}
						r.Emit(presets.NotifModelsCreated(&ActivityLog{}), presets.PayloadModelsCreated{
							Models: []any{log},
						})
						return
					}
				})
				return nil
			}

			if id != "" && amb.skip&Update == 0 {
				old, err := eb.Fetcher(mb.NewModel(), id, ctx)
				if err != nil {
					return errors.Wrap(err, "fetch old record")
				}
				if err := in(obj, id, ctx); err != nil {
					return err
				}
				log, err := amb.createEditLog(ctx.R.Context(), old, obj)
				if err != nil {
					return err
				}
				if log != nil {
					presets.WrapEventFuncAddon(ctx, func(in presets.EventFuncAddon) presets.EventFuncAddon {
						return func(ctx *web.EventContext, r *web.EventResponse) (err error) {
							if err = in(ctx, r); err != nil {
								return
							}
							r.Emit(presets.NotifModelsUpdated(&ActivityLog{}), presets.PayloadModelsUpdated{
								Ids:    []string{fmt.Sprint(log.ID)},
								Models: []any{log},
							})
							return
						}
					})
				}
				return nil
			}
			return in(obj, id, ctx)
		}
	})

	eb.WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
		return func(obj any, id string, ctx *web.EventContext) (err error) {
			if amb.skip&Delete != 0 {
				return in(obj, id, ctx)
			}
			old, err := eb.Fetcher(mb.NewModel(), id, ctx)
			if err != nil {
				return errors.Wrap(err, "fetch old record")
			}
			if err := in(obj, id, ctx); err != nil {
				return err
			}
			log, err := amb.createWithObj(ctx.R.Context(), ActionDelete, old, nil)
			if err != nil {
				return err
			}
			presets.WrapEventFuncAddon(ctx, func(in presets.EventFuncAddon) presets.EventFuncAddon {
				return func(ctx *web.EventContext, r *web.EventResponse) (err error) {
					err = in(ctx, r)
					if err != nil {
						return
					}
					r.Emit(presets.NotifModelsDeleted(&ActivityLog{}), presets.PayloadModelsDeleted{
						// TODO: 这个 payload 还是应该加上 Models
						Ids: []string{fmt.Sprint(log.ID)},
					})
					return
				}
			})
			return nil
		}
	})

	detailFieldTimeline := dp.GetField(DetailFieldTimeline)
	if detailFieldTimeline != nil {
		injectorName := fmt.Sprintf("__activity:%s__", mb.Info().URIName())
		dc := mb.GetPresetsBuilder().GetDependencyCenter()
		dc.RegisterInjector(injectorName)
		dc.MustProvide(injectorName, func() *Builder {
			return ab
		})
		dc.MustProvide(injectorName, func() *presets.ModelBuilder {
			return mb
		})
		detailFieldTimeline.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			keys := ab.MustGetModelBuilder(mb).KeysValue(obj)
			return dc.MustInject(injectorName, &TimelineCompo{
				ID:        mb.Info().URIName() + ":" + keys,
				ModelName: modelName(obj),
				ModelKeys: keys,
				ModelLink: amb.link(obj),
			})
		})
	}

	listFieldUnreadNotes := lb.GetField(ListFieldUnreadNotes)
	if listFieldUnreadNotes != nil {
		// TODO: 因为 UnreadCount 的存在，所以这里其实应该进行添加 listener 来进行更新对应值，进行细粒度更新
		// listFieldUnreadNotes.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// 	rt := modelName(mb.NewModel())
		// 	ri := amb.KeysValue(obj)
		// 	user := ab.currentUserFunc(ctx.R.Context())
		// 	count, _ := GetUnreadNotesCount(db, user.ID, rt, ri)

		// 	return h.Td(
		// 		h.If(count > 0,
		// 			v.VBadge().Content(count).Color("red"),
		// 		).Else(
		// 			h.Text(""),
		// 		),
		// 	)
		// }).Label("Unread Notes") // TODO: i18n
	}
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
