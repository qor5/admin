package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	FieldTimeline       string = "__ActivityTimeline__"
	ListFieldNotes      string = "__ActivityNotes__"
	ListFieldLabelNotes string = "Notes"

	NopModelLabel = "-"
)

const (
	Create = 1 << iota
	Edit
	Delete
)

// @snippet_begin(ActivityModelBuilder)
type ModelBuilder struct {
	ref           any                          // model ref
	typ           reflect.Type                 // model type
	ab            *Builder                     // activity builder
	presetModel   *presets.ModelBuilder        // preset model builder
	skip          uint8                        // skip the refined data operator of the presetModel
	keys          []string                     // primary keys
	keyColumns    []string                     // primary field columns
	ignoredFields []string                     // ignored fields
	typeHandlers  map[reflect.Type]TypeHandler // type handlers
	link          func(any) string             // display the model link on the admin detail page
	label         func() string                // display the model label on the admin detail page
	beforeCreate  func(ctx context.Context, log *ActivityLog) error
}

// @snippet_end

type ctxKeyUnreadCounts struct{}

func NotifiLastViewedAtUpdated(modelName string) string {
	return fmt.Sprintf("activity_NotifiLastViewedAtUpdated_%s", modelName)
}

type PayloadLastViewedAtUpdated struct {
	Log *ActivityLog `json:"log"`
}

func emitLogCreated(evCtx *web.EventContext, log *ActivityLog) {
	if log == nil {
		return
	}
	presets.WrapEventFuncAddon(evCtx, func(in presets.EventFuncAddon) presets.EventFuncAddon {
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
}

func injectorName(mb *presets.ModelBuilder) string {
	return fmt.Sprintf("__activity:%s__", mb.Info().URIName())
}

func (amb *ModelBuilder) NewTimelineCompo(evCtx *web.EventContext, obj any, idSuffix string) h.HTMLComponent {
	if amb.presetModel == nil {
		panic("NewTimelineCompo method only supports presets.ModelBuilder")
	}
	mb := amb.presetModel

	if mb.Info().Verifier().Do(PermListNotes).ObjectOn(obj).WithReq(evCtx.R).IsAllowed() != nil {
		return nil
	}

	injectorName := injectorName(mb)
	dc := mb.GetPresetsBuilder().GetDependencyCenter()
	modelName := ParseModelName(obj)

	log, err := amb.Log(evCtx.R.Context(), ActionLastView, obj, nil)
	if err != nil {
		panic(fmt.Errorf("failed to log last view: %w", err))
	}
	presets.WrapEventFuncAddon(evCtx, func(in presets.EventFuncAddon) presets.EventFuncAddon {
		return func(ctx *web.EventContext, r *web.EventResponse) (err error) {
			if err = in(ctx, r); err != nil {
				return
			}
			// this action is special, so we use a separate notification
			r.Emit(NotifiLastViewedAtUpdated(modelName), PayloadLastViewedAtUpdated{Log: log})
			return
		}
	})
	keys := amb.ParseModelKeys(obj)
	return h.ComponentFunc(func(ctx context.Context) (r []byte, err error) {
		return dc.MustInject(injectorName, &TimelineCompo{
			ID:        mb.Info().URIName() + ":" + keys + idSuffix,
			ModelName: modelName,
			ModelKeys: keys,
			ModelLink: amb.link(obj),
		}).MarshalHTML(ctx)
	})
}

func (amb *ModelBuilder) WrapperSaveFunc(in presets.SaveFunc) presets.SaveFunc {
	mb := amb.presetModel
	eb := mb.Editing()
	return func(obj any, id string, ctx *web.EventContext) (err error) {
		if amb.skip&Create != 0 && amb.skip&Edit != 0 {
			return in(obj, id, ctx)
		}

		fetchOld := func(id string) (_ any, xerr error) {
			if id == "" {
				return nil, gorm.ErrRecordNotFound
			}
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					xerr = errors.Wrap(err, "Panic")
				}
			}()
			return eb.Fetcher(mb.NewModel(), id, ctx)
		}
		old, err := fetchOld(id)
		edit := err == nil && old != nil

		if err := in(obj, id, ctx); err != nil {
			return err
		}

		if !edit && amb.skip&Create == 0 {
			log, err := amb.OnCreate(ctx.R.Context(), obj)
			if err != nil {
				return err
			}
			emitLogCreated(ctx, log)
			return nil
		}

		if edit && amb.skip&Edit == 0 {
			log, err := amb.OnEdit(ctx.R.Context(), old, obj)
			if err != nil {
				return err
			}
			emitLogCreated(ctx, log)
			return nil
		}

		return nil
	}
}

func (amb *ModelBuilder) installPresetModelBuilder(mb *presets.ModelBuilder) {
	amb.presetModel = mb
	amb.link = func(a any) string {
		id := presets.ObjectID(a)
		if id == "" {
			return ""
		}
		if mb.HasDetailing() {
			return mb.Info().DetailingHref(id)
		}
		return ""
	}
	amb.label = func() string {
		return mb.Info().URIName()
	}

	pb := mb.GetPresetsBuilder()
	if amb.ab.GetLogModelBuilder(pb) == nil {
		pb.GetI18n().
			RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
			RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN).
			RegisterForModule(language.Japanese, I18nActivityKey, Messages_ja_JP)
	}
	dc := mb.GetPresetsBuilder().GetDependencyCenter()
	injectorName := injectorName(mb)
	dc.RegisterInjector(injectorName)
	dc.MustProvide(injectorName, func() (*Builder, *presets.ModelBuilder, *ModelBuilder) {
		return amb.ab, mb, amb
	})

	eb := mb.Editing()
	eb.WrapSaveFunc(amb.WrapperSaveFunc)

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
			log, err := amb.OnDelete(ctx.R.Context(), old)
			if err != nil {
				return err
			}
			emitLogCreated(ctx, log)
			return nil
		}
	})

	eb.Creating().Except(FieldTimeline)
	editFieldTimeline := eb.GetField(FieldTimeline)
	if editFieldTimeline != nil && editFieldTimeline.GetCompFunc() == nil {
		editFieldTimeline.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return amb.NewTimelineCompo(ctx, obj, "_edit_"+FieldTimeline)
		})
	}

	if mb.HasDetailing() {
		dp := mb.Detailing()
		detailFieldTimeline := dp.GetField(FieldTimeline)
		if detailFieldTimeline != nil && detailFieldTimeline.GetCompFunc() == nil {
			detailFieldTimeline.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				return amb.NewTimelineCompo(ctx, obj, "_detail_"+FieldTimeline)
			})
		}
	}

	lb := mb.Listing()
	listFieldNotes := lb.GetField(ListFieldNotes)
	if listFieldNotes != nil && listFieldNotes.GetCompFunc() == nil {
		lb.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
			return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
				result, err = in(ctx, params)
				if err != nil {
					return
				}
				var modelName string
				var modelKeyses []string
				reflectutils.ForEach(result.Nodes, func(obj any) {
					if modelName == "" {
						modelName = ParseModelName(obj)
					}
					modelKeyses = append(modelKeyses, amb.ParseModelKeys(obj))
				})
				if len(modelKeyses) > 0 {
					counts, err := amb.ab.GetNotesCounts(ctx.R.Context(), modelName, modelKeyses)
					if err != nil {
						return nil, err
					}
					m := lo.SliceToMap(counts, func(v *NoteCount) (string, *NoteCount) {
						return v.ModelKeys, v
					})
					ctx.WithContextValue(ctxKeyUnreadCounts{}, m)
				}
				return
			}
		})
		listFieldNotes.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			counts, ok := ctx.ContextValue(ctxKeyUnreadCounts{}).(map[string]*NoteCount)
			if !ok {
				return h.Text("")
			}
			modelKeys := amb.ParseModelKeys(obj)
			count := counts[modelKeys]
			if count == nil {
				count = &NoteCount{ModelKeys: modelKeys, UnreadNotesCount: 0, TotalNotesCount: 0}
			}
			var totalNotesCountText string
			if count.TotalNotesCount > 99 {
				totalNotesCountText = "99+"
			} else {
				totalNotesCountText = fmt.Sprintf("%d", count.TotalNotesCount)
			}

			total := h.Div().Class("text-caption bg-grey-lighten-3 rounded px-1").Text(totalNotesCountText)
			return h.Td(
				web.Scope().VSlot("{locals}").Init(fmt.Sprintf("{ unreadNotesCount: %d }", count.UnreadNotesCount)).Children(
					web.Listen(
						NotifiLastViewedAtUpdated(ParseModelName(obj)),
						fmt.Sprintf(`if (payload.log.ModelKeys === %q) { locals.unreadNotesCount = 0 }`, modelKeys),
					),
					v.VBadge().Attr("v-if", "locals.unreadNotesCount>0").Color(v.ColorError).Dot(true).Bordered(true).OffsetX("-3").OffsetY("-3").Children(
						total,
					),
					h.Div().Attr("v-else", true).Class("d-flex flex-row").Children(
						total,
					),
				),
			)
		}).Label(ListFieldLabelNotes)
		lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
			msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
			return map[string]string{
				ListFieldNotes: msgr.HeaderNotes,
			}, nil
		}))
	}
}

func (mb *ModelBuilder) AddKeys(keys ...string) *ModelBuilder {
	mb.keys = lo.Uniq(append(mb.keys, keys...))
	mb.keyColumns = mb.parseColumns(keys)
	return mb
}

func (mb *ModelBuilder) Keys(keys ...string) *ModelBuilder {
	mb.keys = keys
	mb.keyColumns = mb.parseColumns(keys)
	return mb
}

func (mb *ModelBuilder) parseColumns(keys []string) []string {
	s, err := ParseSchema(mb.ref)
	if err != nil {
		return lo.Map(keys, func(v string, _ int) string { return strcase.ToSnake(v) })
	}

	columns := make([]string, 0, len(keys))
	m := lo.SliceToMap(s.Fields, func(f *schema.Field) (string, *schema.Field) {
		return f.Name, f
	})
	for _, key := range keys {
		f, ok := m[key]
		if !ok {
			panic(fmt.Errorf("invalid key %q", key))
		}
		columns = append(columns, f.DBName)
	}
	return columns
}

func (mb *ModelBuilder) LabelFunc(f func() string) *ModelBuilder {
	mb.label = f
	return mb
}

func (mb *ModelBuilder) LinkFunc(f func(any) string) *ModelBuilder {
	mb.link = f
	return mb
}

func (mb *ModelBuilder) SkipCreate() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipCreate method only supports presets.ModelBuilder")
	}

	if mb.skip&Create == 0 {
		mb.skip |= Create
	}
	return mb
}

func (mb *ModelBuilder) SkipEdit() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipEdit method only supports presets.ModelBuilder")
	}

	if mb.skip&Edit == 0 {
		mb.skip |= Edit
	}
	return mb
}

func (mb *ModelBuilder) SkipDelete() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipDelete method only supports presets.ModelBuilder")
	}

	if mb.skip&Delete == 0 {
		mb.skip |= Delete
	}
	return mb
}

func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = lo.Uniq(append(mb.ignoredFields, fields...))
	return mb
}

func (mb *ModelBuilder) IgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = fields
	return mb
}

func (mb *ModelBuilder) BeforeCreate(f func(ctx context.Context, log *ActivityLog) error) *ModelBuilder {
	mb.beforeCreate = f
	return mb
}

func (mb *ModelBuilder) AddTypeHanders(v any, f TypeHandler) *ModelBuilder {
	if mb.typeHandlers == nil {
		mb.typeHandlers = map[reflect.Type]TypeHandler{}
	}
	mb.typeHandlers[reflect.Indirect(reflect.ValueOf(v)).Type()] = f
	return mb
}

const ModelKeysSeparator = ":"

func (mb *ModelBuilder) ParseModelKeys(v any) string {
	return KeysValue(v, mb.keys, ModelKeysSeparator)
}

func (mb *ModelBuilder) Log(ctx context.Context, action string, obj, detail any) (*ActivityLog, error) {
	return mb.create(ctx, action, ParseModelName(obj), mb.ParseModelKeys(obj), mb.modelLink(obj), detail)
}

func (mb *ModelBuilder) OnCreate(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionCreate, v, nil)
}

func (mb *ModelBuilder) OnView(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionView, v, nil)
}

func (mb *ModelBuilder) OnEdit(ctx context.Context, oldObj, newObj any) (*ActivityLog, error) {
	diffs, err := mb.Diff(oldObj, newObj)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return mb.Log(ctx, ActionEdit, newObj, diffs)
}

func (mb *ModelBuilder) OnDelete(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionDelete, v, nil)
}

func (mb *ModelBuilder) Note(ctx context.Context, v any, note *Note) (*ActivityLog, error) {
	return mb.Log(ctx, ActionNote, v, note)
}

func (mb *ModelBuilder) Diff(oldObj, newObj any) ([]Diff, error) {
	return NewDiffBuilder(mb).Diff(oldObj, newObj)
}

func (mb *ModelBuilder) modelLink(v any) string {
	if f := mb.link; f != nil {
		return f(v)
	}
	return ""
}

type ctxKeyScope struct{}

func ContextWithScope(ctx context.Context, scope string) context.Context {
	return context.WithValue(ctx, ctxKeyScope{}, scope)
}

func ScopeWithOwner(owner string) string {
	return fmt.Sprintf(",owner:%s,", owner)
}

type ctxKeyDB struct{}

func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKeyDB{}, db)
}

func (mb *ModelBuilder) create(
	ctx context.Context,
	action string,
	modelName, modelKeys, modelLink string,
	detail any,
) (*ActivityLog, error) {
	db, ok := ctx.Value(ctxKeyDB{}).(*gorm.DB)
	if !ok {
		db = mb.ab.db
	} else if mb.ab.tablePrefix != "" {
		db = db.Scopes(ScopeWithTablePrefix(mb.ab.tablePrefix)).Session(&gorm.Session{})
	}

	user, err := mb.ab.currentUserFunc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	if mb.ab.findUsersFunc == nil {
		user := &ActivityUser{
			CreatedAt: db.NowFunc(),
			ID:        user.ID,
			Name:      user.Name,
			Avatar:    user.Avatar,
		}
		if db.Where("id = ?", user.ID).Select("*").Omit("created_at").Updates(user).RowsAffected == 0 {
			if err := db.Create(user).Error; err != nil {
				return nil, errors.Wrap(err, "failed to create user")
			}
		}
	}

	scope, _ := ctx.Value(ctxKeyScope{}).(string)
	if scope == "" {
		if action == ActionCreate {
			scope = ScopeWithOwner(user.ID)
		} else {
			createdLog := &ActivityLog{}
			if err := db.Where("model_name = ? AND model_keys = ? AND action = ? ", modelName, modelKeys, ActionCreate).
				Order("created_at ASC").First(&createdLog).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.Wrap(err, "failed to find created log")
				}
			}
			if createdLog.ID != 0 && createdLog.UserID != "" {
				scope = createdLog.Scope
			}
		}
	}

	log := &ActivityLog{
		UserID:     user.ID,
		Action:     action,
		ModelName:  modelName,
		ModelKeys:  modelKeys,
		ModelLabel: "",
		ModelLink:  modelLink,
		Scope:      scope,
	}
	if mb.label != nil {
		log.ModelLabel = mb.label()
	}
	log.CreatedAt = db.NowFunc()

	detailJson, err := json.Marshal(detail)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal detail")
	}
	log.Detail = string(detailJson)

	if mb.beforeCreate != nil {
		if err := mb.beforeCreate(ctx, log); err != nil {
			return nil, errors.Wrap(err, "failed to run before create")
		}
	}

	if action == ActionLastView {
		log.Hidden = true
		r := &ActivityLog{}
		if err := db.
			Where("user_id = ? AND model_name = ? AND model_keys = ? AND action = ?", user.ID, modelName, modelKeys, action).
			Assign(log).FirstOrCreate(r).Error; err != nil {
			return nil, err
		}
		return r, nil

		// Why not use this ? Because log.id is empty although the record is already created, there is no advance fetch of the original id here .
		// if db.Where("user_id = ? AND model_name = ? AND model_keys = ? AND action = ?", log.UserID, log.ModelName, log.ModelKeys, log.Action).
		// 	Select("*").Updates(log).RowsAffected == 0 {
		// 	if err := db.Create(log).Error; err != nil {
		// 		return nil, errors.Wrap(err, "failed to create log")
		// 	}
		// }

		// Why not use Upsert ?
		// https://github.com/go-gorm/gorm/issues/6512
		// https://github.com/jackc/pgx/issues/1234
		// if err := db.Clauses(clause.OnConflict{
		// 	Columns: []clause.Column{
		// 		{Name: "model_name"},
		// 		{Name: "model_keys"},
		// 	},
		// 	TargetWhere: clause.Where{Exprs: []clause.Expression{
		// 		gorm.Expr("action = ?", ActionLastView),
		// 		gorm.Expr("deleted_at IS NULL"),
		// 	}},
		// 	UpdateAll: true,
		// }).Create(log).Error; err != nil {
		// 	return nil, errors.Wrap(err, "failed to create log")
		// }
		// return log, nil
	}

	if err := db.Create(log).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create log")
	}

	return log, nil
}
