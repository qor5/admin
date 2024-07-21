package activity

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	FieldTimeline       string = "__ActivityTimeline__"
	ListFieldNotes      string = "__ActivityNotes__"
	ListFieldLabelNotes string = "Notes"
)

const (
	Create = 1 << iota
	Edit
	Delete
)

// @snippet_begin(ActivityModelBuilder)
// a unique model builder is consist of typ and presetModel
type ModelBuilder struct {
	once sync.Once

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
}

// @snippet_end

type ctxKeyUnreadCounts struct{}

func NotifiLastViewedAtUpdated(modelName string) string {
	return fmt.Sprintf("activity_NotifModelsCreated_%s", modelName)
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

func (amb *ModelBuilder) NewTimelineCompo(evCtx *web.EventContext, obj any, idSuffix string) h.HTMLComponent {
	if amb.presetModel == nil {
		panic("NewTimelineCompo method only supports presets.ModelBuilder")
	}
	mb := amb.presetModel

	injectorName := fmt.Sprintf("__activity:%s__", mb.Info().URIName())
	dc := mb.GetPresetsBuilder().GetDependencyCenter()
	amb.once.Do(func() {
		dc.RegisterInjector(injectorName)
		dc.MustProvide(injectorName, func() *Builder {
			return amb.ab
		})
		dc.MustProvide(injectorName, func() *ModelBuilder {
			return amb
		})
		dc.MustProvide(injectorName, func() *presets.ModelBuilder {
			return mb
		})
	})
	return h.ComponentFunc(func(ctx context.Context) (r []byte, err error) {
		modelName := ParseModelName(obj)

		log, err := amb.Log(ctx, ActionLastView, obj, nil)
		if err != nil {
			panic(err)
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
		return dc.MustInject(injectorName, &TimelineCompo{
			ID:        mb.Info().URIName() + ":" + keys + idSuffix,
			ModelName: modelName,
			ModelKeys: keys,
			ModelLink: amb.link(obj),
		}).MarshalHTML(ctx)
	})
}

func (amb *ModelBuilder) installPresetsModelBuilder(mb *presets.ModelBuilder) {
	amb.presetModel = mb
	amb.LinkFunc(func(a any) string {
		id := ObjectID(a)
		if id == "" {
			return ""
		}
		if mb.HasDetailing() {
			return mb.Info().DetailingHref(id)
		}
		return ""
	})

	eb := mb.Editing()
	eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
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
			return func(model any, params *presets.SearchParams, ctx *web.EventContext) (r any, totalCount int, err error) {
				r, totalCount, err = in(model, params, ctx)
				if err != nil {
					return
				}
				user, uerr := amb.ab.currentUserFunc(ctx.R.Context())
				if uerr != nil {
					return
				}
				var modelName string
				modelKeyses := []string{}
				reflectutils.ForEach(r, func(obj any) {
					if modelName == "" {
						modelName = ParseModelName(obj)
					}
					modelKeyses = append(modelKeyses, amb.ParseModelKeys(obj))
				})
				if len(modelKeyses) > 0 {
					counts, err := GetNotesCounts(amb.ab.db, user.ID, modelName, modelKeyses)
					if err != nil {
						return r, totalCount, err
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

			// TODO: Because of the filter of hasUnreadNotes, it seems that you need to reload the list directly.
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

func (mb *ModelBuilder) Log(ctx context.Context, action string, obj any, detail any) (*ActivityLog, error) {
	return mb.create(ctx, action, ParseModelName(obj), mb.ParseModelKeys(obj), mb.modelLink(obj), detail)
}

func (mb *ModelBuilder) OnCreate(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionCreate, v, nil)
}

func (mb *ModelBuilder) OnView(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionView, v, nil)
}

func (mb *ModelBuilder) OnEdit(ctx context.Context, old any, new any) (*ActivityLog, error) {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return mb.Log(ctx, ActionEdit, new, diffs)
}

func (mb *ModelBuilder) OnDelete(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.Log(ctx, ActionDelete, v, nil)
}

func (mb *ModelBuilder) Note(ctx context.Context, v any, note *Note) (*ActivityLog, error) {
	return mb.Log(ctx, ActionNote, v, note)
}

func (mb *ModelBuilder) Diff(old, new any) ([]Diff, error) {
	return NewDiffBuilder(mb).Diff(old, new)
}

func (mb *ModelBuilder) modelLink(v any) string {
	if f := mb.link; f != nil {
		return f(v)
	}
	return ""
}

func (mb *ModelBuilder) create(
	ctx context.Context,
	action string,
	modelName, modelKeys, modelLink string,
	detail any,
) (*ActivityLog, error) {
	user, err := mb.ab.currentUserFunc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	if mb.ab.findUsersFunc == nil {
		user := &ActivityUser{
			CreatedAt: mb.ab.db.NowFunc(),
			ID:        user.ID,
			Name:      user.Name,
			Avatar:    user.Avatar,
		}
		if mb.ab.db.Where("id = ?", user.ID).Select("*").Omit("created_at").Updates(user).RowsAffected == 0 {
			if err := mb.ab.db.Create(user).Error; err != nil {
				return nil, errors.Wrap(err, "failed to create user")
			}
		}
	}

	log := &ActivityLog{
		CreatorID:  user.ID,
		Action:     action,
		ModelName:  modelName,
		ModelKeys:  modelKeys,
		ModelLabel: "-",
		ModelLink:  modelLink,
	}
	if mb.presetModel != nil {
		log.ModelLabel = cmp.Or(mb.presetModel.Info().URIName(), log.ModelLabel)
	}
	log.CreatedAt = mb.ab.db.NowFunc()

	detailJson, err := json.Marshal(detail)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal detail")
	}
	log.Detail = string(detailJson)

	if action == ActionLastView {
		log.Hidden = true
		r := &ActivityLog{}
		if err := mb.ab.db.
			Where("creator_id = ? AND model_name = ? AND model_keys = ? AND action = ?", user.ID, modelName, modelKeys, action).
			Assign(log).FirstOrCreate(r).Error; err != nil {
			return nil, err
		}
		return r, nil

		// Why not use this ? Because id will be empty when the record is already created
		// if mb.ab.db.Where("creator_id = ? AND model_name = ? AND model_keys = ? AND action = ?", log.CreatorID, log.ModelName, log.ModelKeys, log.Action).
		// 	Select("*").Updates(log).RowsAffected == 0 {
		// 	if err := mb.ab.db.Create(log).Error; err != nil {
		// 		return nil, errors.Wrap(err, "failed to create log")
		// 	}
		// }

		// Why not use Upsert ?
		// https://github.com/go-gorm/gorm/issues/6512
		// https://github.com/jackc/pgx/issues/1234
		// if err := mb.ab.db.Clauses(clause.OnConflict{
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

	if err := mb.ab.db.Create(log).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create log")
	}

	return log, nil
}
