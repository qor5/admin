package activity

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
)

const (
	DetailFieldTimeline  string = "Timeline"
	ListFieldUnreadNotes string = "UnreadNotes"
)

const (
	Create = 1 << iota
	View
	Edit
	Delete
)

// @snippet_begin(ActivityModelBuilder)
// a unique model builder is consist of typ and presetModel
type ModelBuilder struct {
	typ           reflect.Type                 // model type
	ab            *Builder                     // activity builder
	presetModel   *presets.ModelBuilder        // preset model builder
	skip          uint8                        // skip the refined data operator of the presetModel
	keys          []string                     // primary keys
	ignoredFields []string                     // ignored fields
	typeHandlers  map[reflect.Type]TypeHandler // type handlers
	link          func(any) string             // display the model link on the admin detail page
}

// @snippet_end

func (amb *ModelBuilder) installPresetsModelBuilder(mb *presets.ModelBuilder) {
	amb.presetModel = mb
	amb.LinkFunc(func(a any) string {
		id := ObjectID(a)
		if id == "" {
			return ""
		}
		return mb.Info().DetailingHref(id)
	})

	eb := mb.Editing()
	dp := mb.Detailing()
	lb := mb.Listing()

	emitLogCreated := func(evCtx *web.EventContext, log *ActivityLog) {
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
				emitLogCreated(ctx, log)
				return nil
			}

			if id != "" && amb.skip&Edit == 0 {
				old, err := eb.Fetcher(mb.NewModel(), id, ctx)
				if err != nil {
					return errors.Wrap(err, "fetch old record")
				}
				if err := in(obj, id, ctx); err != nil {
					return err
				}
				log, err := amb.Edit(ctx.R.Context(), old, obj)
				if err != nil {
					return err
				}
				emitLogCreated(ctx, log)
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
			log, err := amb.Delete(ctx.R.Context(), old)
			if err != nil {
				return err
			}
			emitLogCreated(ctx, log)
			return nil
		}
	})

	dp.WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj any, id string, ctx *web.EventContext) (r any, err error) {
			r, err = in(obj, id, ctx)
			if err != nil {
				return
			}
			if amb.skip&View != 0 {
				return
			}
			log, err := amb.View(ctx.R.Context(), r)
			if err != nil {
				return r, err
			}
			emitLogCreated(ctx, log)
			return r, nil
		}
	})

	detailFieldTimeline := dp.GetField(DetailFieldTimeline)
	if detailFieldTimeline != nil {
		injectorName := fmt.Sprintf("__activity:%s__", mb.Info().URIName())
		dc := mb.GetPresetsBuilder().GetDependencyCenter()
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
		detailFieldTimeline.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			keys := amb.KeysValue(obj)
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
		listFieldUnreadNotes.ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			// rt := modelName(mb.NewModel())
			// ri := amb.KeysValue(obj)
			// user := ab.currentUserFunc(ctx.R.Context())
			// count, _ := GetUnreadNotesCount(db, user.ID, rt, ri)
			count := 10
			return h.Td(
				// web.Listen(
				// 	presets.NotifModelsUpdated(&ActivityLog{}), ``,
				// ),
				h.If(count > 0,
					v.VBadge().Inline(true).Content(count).Color(v.ColorError),
				).Else(
					h.Text(""),
				),
			)
		}).Label("Unread Notes")
	}
}

func (mb *ModelBuilder) AddKeys(keys ...string) *ModelBuilder {
	mb.keys = lo.Uniq(append(mb.keys, keys...))
	return mb
}

func (mb *ModelBuilder) Keys(keys ...string) *ModelBuilder {
	mb.keys = keys
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

func (mb *ModelBuilder) SkipView() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipView method only supports presets.ModelBuilder")
	}

	if mb.skip&View == 0 {
		mb.skip |= View
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

func (mb *ModelBuilder) KeysValue(v any) string {
	return keysValue(v, mb.keys)
}

func (mb *ModelBuilder) Log(ctx context.Context, action string, obj any, detail any) (*ActivityLog, error) {
	return mb.createWithObj(ctx, action, obj, detail)
}

func (mb *ModelBuilder) Create(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.createWithObj(ctx, ActionCreate, v, nil)
}

func (mb *ModelBuilder) View(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.createWithObj(ctx, ActionView, v, nil)
}

func (mb *ModelBuilder) Edit(ctx context.Context, old any, new any) (*ActivityLog, error) {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return mb.createWithObj(ctx, ActionEdit, new, diffs)
}

func (mb *ModelBuilder) Delete(ctx context.Context, v any) (*ActivityLog, error) {
	return mb.createWithObj(ctx, ActionDelete, v, nil)
}

func (mb *ModelBuilder) Note(ctx context.Context, v any, note *Note) (*ActivityLog, error) {
	return mb.createWithObj(ctx, ActionNote, v, note)
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

func (mb *ModelBuilder) createWithObj(ctx context.Context, action string, obj any, detail any) (*ActivityLog, error) {
	return mb.create(ctx, action, modelName(obj), mb.KeysValue(obj), mb.modelLink(obj), detail)
}

func (mb *ModelBuilder) create(
	ctx context.Context,
	action string,
	modelName, modelKeys, modelLink string,
	detail any,
) (*ActivityLog, error) {
	creator := mb.ab.currentUserFunc(ctx)
	if creator == nil {
		return nil, errors.New("current user is nil")
	}

	if mb.ab.findUsersFunc == nil {
		user := &ActivityUser{
			CreatedAt: mb.ab.db.NowFunc(),
			ID:        creator.ID,
			Name:      creator.Name,
			Avatar:    creator.Avatar,
		}
		if err := mb.ab.db.Save(user).Error; err != nil {
			return nil, errors.Wrap(err, "failed to save user")
		}
	}

	log := &ActivityLog{
		CreatorID:  creator.ID,
		Action:     action,
		ModelName:  modelName,
		ModelKeys:  modelKeys,
		ModelLabel: "-",
		ModelLink:  modelLink,
	}
	if mb.presetModel != nil {
		log.ModelLabel = cmp.Or(mb.presetModel.Info().URIName(), log.ModelLabel)
	}
	log.CreatedAt = time.Now()

	detailJson, err := json.Marshal(detail)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal detail")
	}
	log.Detail = string(detailJson)

	if err := mb.ab.db.Create(log).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create log")
	}
	return log, nil
}
