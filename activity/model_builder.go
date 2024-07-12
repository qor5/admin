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
	"github.com/samber/lo"
	"github.com/spf13/cast"
	h "github.com/theplant/htmlgo"
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

// AddKeys add keys to the model builder
func (mb *ModelBuilder) AddKeys(keys ...string) *ModelBuilder {
	mb.keys = lo.Uniq(append(mb.keys, keys...))
	return mb
}

// Keys set keys for the model builder
func (mb *ModelBuilder) Keys(keys ...string) *ModelBuilder {
	mb.keys = keys
	return mb
}

// LinkFunc set the link that linked to the modified record
func (mb *ModelBuilder) LinkFunc(f func(any) string) *ModelBuilder {
	mb.link = f
	return mb
}

// SkipCreate skip the created action for preset.ModelBuilder
func (mb *ModelBuilder) SkipCreate() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipCreate method only supports presets.ModelBuilder")
	}

	if mb.skip&Create == 0 {
		mb.skip |= Create
	}
	return mb
}

// SkipUpdate skip the update action for preset.ModelBuilder
func (mb *ModelBuilder) SkipUpdate() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipUpdate method only supports presets.ModelBuilder")
	}

	if mb.skip&Update == 0 {
		mb.skip |= Update
	}
	return mb
}

// SkipDelete skip the delete action for preset.ModelBuilder
func (mb *ModelBuilder) SkipDelete() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipDelete method only supports presets.ModelBuilder")
	}

	if mb.skip&Delete == 0 {
		mb.skip |= Delete
	}
	return mb
}

// AddIgnoredFields append ignored fields to the default ignored fields, this would not overwrite the default ignored fields
func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = lo.Uniq(append(mb.ignoredFields, fields...))
	return mb
}

// IgnoredFields set ignored fields to replace the default ignored fields with the new set.
func (mb *ModelBuilder) IgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = fields
	return mb
}

// AddTypeHanders add type handers for the model builder
func (mb *ModelBuilder) AddTypeHanders(v any, f TypeHandler) *ModelBuilder {
	if mb.typeHandlers == nil {
		mb.typeHandlers = map[reflect.Type]TypeHandler{}
	}
	mb.typeHandlers[reflect.Indirect(reflect.ValueOf(v)).Type()] = f
	return mb
}

// KeysValue get model keys value
func (mb *ModelBuilder) KeysValue(v any) string {
	return keysValue(v, mb.keys)
}

// AddRecords add records log
func (mb *ModelBuilder) AddRecords(ctx context.Context, action string, vs ...any) error {
	if len(vs) == 0 {
		return errors.New("data are empty")
	}

	creator := mb.ab.currentUserFunc(ctx)

	switch action {
	case ActionView:
		for _, v := range vs {
			err := mb.AddViewRecord(creator, v)
			if err != nil {
				return err
			}
		}

	case ActionDelete:
		for _, v := range vs {
			err := mb.AddDeleteRecord(creator, v)
			if err != nil {
				return err
			}
		}
	case ActionCreate:
		for _, v := range vs {
			err := mb.AddCreateRecord(creator, v)
			if err != nil {
				return err
			}
		}
	case ActionEdit:
		for _, v := range vs {
			err := mb.AddEditRecord(creator, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddCustomizedRecord add customized record
func (mb *ModelBuilder) AddCustomizedRecord(ctx context.Context, action string, diff bool, obj any) error {
	creator := mb.ab.currentUserFunc(ctx)
	if !diff {
		return mb.save(creator, action, obj, "")
	}

	old, ok := fetchOld(obj, mb.ab.db)
	if !ok {
		return fmt.Errorf("can't find old data for %+v ", obj)
	}
	return mb.addDiff(action, creator, old, obj)
}

// AddViewRecord add view record
func (mb *ModelBuilder) AddViewRecord(creator *User, v any) error {
	return mb.save(creator, ActionView, v, "")
}

// AddDeleteRecord	add delete record
func (mb *ModelBuilder) AddDeleteRecord(creator *User, v any) error {
	return mb.save(creator, ActionDelete, v, "")
}

// AddSaverRecord will save a create log or a edit log
func (mb *ModelBuilder) AddSaveRecord(creator *User, new any) error {
	old, ok := fetchOld(new, mb.ab.db)
	if !ok {
		return mb.AddCreateRecord(creator, new)
	}
	return mb.AddEditRecordWithOld(creator, old, new)
}

// AddCreateRecord add create record
func (mb *ModelBuilder) AddCreateRecord(creator *User, v any) error {
	return mb.save(creator, ActionCreate, v, "")
}

// AddEditRecord add edit record
func (mb *ModelBuilder) AddEditRecord(creator *User, new any) error {
	old, ok := fetchOld(new, mb.ab.db)
	if !ok {
		return fmt.Errorf("can't find old data for %+v ", new)
	}
	return mb.AddEditRecordWithOld(creator, old, new)
}

// AddEditRecord add edit record
func (mb *ModelBuilder) AddEditRecordWithOld(creator *User, old, new any) error {
	return mb.addDiff(ActionEdit, creator, old, new)
}

func (mb *ModelBuilder) addDiff(action string, creator *User, old, new any) error {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return err
	}

	if len(diffs) == 0 {
		return nil
	}

	b, err := json.Marshal(diffs)
	if err != nil {
		return err
	}

	return mb.save(creator, action, new, string(b))
}

// Diff get diffs between old and new value
func (mb *ModelBuilder) Diff(old, new any) ([]Diff, error) {
	return NewDiffBuilder(mb).Diff(old, new)
}

func (mb *ModelBuilder) save(creator *User, action string, v any, diffs string) error {
	if mb.ab.findUsersFunc == nil { // TODO: 先简单处理吧
		user := &ActivityUser{
			Name:   creator.Name,
			Avatar: creator.Avatar,
		}
		var err error
		user.ID, err = cast.ToUintE(creator.ID)
		if err != nil {
			return errors.Wrap(err, "failed to cast creator ID")
		}
		// TODO: 这样 CreatedAt 会是空值
		if err := mb.ab.db.Save(user).Error; err != nil {
			return err
		}
	}

	log := &ActivityLog{}

	log.CreatedAt = time.Now()
	log.CreatorID = creator.ID
	log.Action = action
	log.ModelName = modelName(v)
	log.ModelKeys = mb.KeysValue(v)

	if mb.presetModel != nil && mb.presetModel.Info().URIName() != "" {
		log.ModelLabel = mb.presetModel.Info().URIName()
	} else {
		log.ModelLabel = "-"
	}

	if f := mb.link; f != nil {
		log.ModelLink = f(v)
	}

	if diffs == "" && action == ActionEdit {
		return nil
	}

	if action == ActionEdit {
		log.Detail = diffs
	}

	err := mb.ab.db.Save(log).Error
	if err != nil {
		return err
	}
	return nil
}

func (mb *ModelBuilder) createEditLog(ctx context.Context, old any, new any) (*ActivityLog, error) {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return mb.createWithObj(ctx, ActionEdit, new, diffs)
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

	if mb.ab.findUsersFunc == nil { // TODO: 先简单处理吧
		user := &ActivityUser{
			Name:   creator.Name,
			Avatar: creator.Avatar,
		}
		var err error
		user.ID, err = cast.ToUintE(creator.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to cast creator ID")
		}
		// TODO: 这样 CreatedAt 会是空值，回头再优化
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
