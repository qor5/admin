package activity

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dustin/go-humanize"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey    i18n.ModuleKey = "I18nActivityKey"
	paramHideModelLink                = "hide_link"
)

func (ab *Builder) Install(b *presets.Builder) error {
	b.GetI18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nActivityKey, Messages_ja_JP)

	if permB := b.GetPermission(); permB != nil {
		permB.CreatePolicies(ab.permPolicy)
	}

	mb := b.Model(&ActivityLog{}).MenuIcon("mdi-book-edit")
	return ab.logModelInstall(b, mb)
}

func defaultActionLabels(msgr *Messages) map[string]string {
	return map[string]string{
		"":           msgr.ActionAll,
		ActionView:   msgr.ActionView,
		ActionEdit:   msgr.ActionEdit,
		ActionCreate: msgr.ActionCreate,
		ActionDelete: msgr.ActionDelete,
		ActionNote:   msgr.ActionNote,
	}
}

func (ab *Builder) defaultLogModelInstall(b *presets.Builder, mb *presets.ModelBuilder) error {
	var (
		lb = mb.Listing("CreatedAt", "Creator", "Action", "ModelKeys", "ModelLabel", "ModelName")
		dp = mb.Detailing("Detail").Drawer(true)
	)

	dp.WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(model interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			r, err = in(model, id, ctx)
			if err != nil {
				return
			}
			log := r.(*ActivityLog)
			if err := ab.supplyCreators(ctx.R.Context(), []*ActivityLog{log}); err != nil {
				return nil, err
			}
			return log, nil
		}
	})

	lb.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
			r, totalCount, err = in(model, params, ctx)
			if totalCount <= 0 {
				return
			}
			logs := r.([]*ActivityLog)
			if err := ab.supplyCreators(ctx.R.Context(), logs); err != nil {
				return nil, 0, err
			}
			return logs, totalCount, nil
		}
	})

	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })

	// TODO: should be able to delete log ?
	lb.RowMenu().RowMenuItem("Delete").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		return nil
	})

	lb.Field("CreatedAt").Label(Messages_en_US.ModelCreatedAt).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).CreatedAt.Format(timeFormat)))
		},
	)
	lb.Field("Creator").Label(Messages_en_US.ModelCreator).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).Creator.Name))
		},
	)
	lb.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	lb.Field("ModelName").Label(Messages_en_US.ModelName)
	lb.Field("ModelLabel").Label(Messages_en_US.ModelLabel).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			if obj.(*ActivityLog).ModelLabel == "" {
				return h.Td(h.Text("-"))
			}
			return h.Td(h.Text(obj.(*ActivityLog).ModelLabel))
		},
	)

	lb.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		actionLabels := defaultActionLabels(msgr)

		var actionOptions []*vuetifyx.SelectItem
		for _, action := range DefaultActions {
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  actionLabels[action],
				Value: action,
			})
		}
		actions := []string{}
		err := ab.db.Model(&ActivityLog{}).Select("DISTINCT action AS action").Pluck("action", &actions).Error
		if err != nil {
			panic(err)
		}
		for _, action := range actions {
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  string(action),
				Value: string(action),
			})
		}
		actionOptions = lo.UniqBy(actionOptions, func(item *vuetifyx.SelectItem) string { return item.Value })

		creatorIDs := []string{}
		err = ab.db.Model(&ActivityLog{}).Select("DISTINCT creator_id AS id").Pluck("id", &creatorIDs).Error
		if err != nil {
			panic(err)
		}
		creators, err := ab.findUsers(ctx.R.Context(), creatorIDs)
		if err != nil {
			panic(err)
		}
		var creatorOptions []*vuetifyx.SelectItem
		for _, creator := range creators {
			creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
				Text:  creator.Name,
				Value: creator.ID,
			})
		}

		var modelOptions []*vuetifyx.SelectItem
		for _, m := range ab.models {
			modelOptions = append(modelOptions, &vuetifyx.SelectItem{
				Text:  m.typ.Name(),
				Value: m.typ.Name(),
			})
		}

		filterData := []*vuetifyx.FilterItem{
			{
				Key:          "action",
				Label:        msgr.FilterAction,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `action %s ?`,
				Options:      actionOptions,
			},
			{
				Key:          "created",
				Label:        msgr.FilterCreatedAt,
				ItemType:     vuetifyx.ItemTypeDatetimeRange,
				SQLCondition: `created_at %s ?`,
			},
		}
		if len(creatorOptions) > 0 {
			filterData = append(filterData, &vuetifyx.FilterItem{
				Key:          "creator_id",
				Label:        msgr.FilterCreator,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `creator_id %s ?`,
				Options:      creatorOptions,
			})
		}
		if len(modelOptions) > 0 {
			filterData = append(filterData, &vuetifyx.FilterItem{
				Key:          "model",
				Label:        msgr.FilterModel,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `model_name %s ?`,
				Options:      modelOptions,
			})
		}
		return filterData
	})

	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		actionLabels := defaultActionLabels(msgr)
		return lo.Map(append([]string{""}, DefaultActions...), func(action string, _ int) *presets.FilterTab {
			filterTab := &presets.FilterTab{
				Label: actionLabels[action],
				Query: url.Values{"action": []string{action}},
			}
			if action == "" {
				filterTab.Query.Del("action")
			}
			return filterTab
		})
	})

	dp.Field("Detail").ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			var (
				log           = obj.(*ActivityLog)
				msgr          = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
				hideModelLink = cast.ToBool(ctx.R.Form.Get(paramHideModelLink))
			)
			var children []h.HTMLComponent
			children = append(children, VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(
					h.Text(" "+msgr.DiffDetail),
				),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Tbody(
							h.Tr(h.Td(h.Text(msgr.ModelCreator)), h.Td(h.Text(log.Creator.Name))),
							h.Tr(h.Td(h.Text(msgr.ModelUserID)), h.Td(h.Text(fmt.Sprint(log.CreatorID)))),
							h.Tr(h.Td(h.Text(msgr.ModelAction)), h.Td(h.Text(log.Action))),
							h.Tr(h.Td(h.Text(msgr.ModelName)), h.Td(h.Text(log.ModelName))),
							h.Tr(h.Td(h.Text(msgr.ModelLabel)), h.Td(h.Text(cmp.Or(log.ModelLabel, "-")))),
							h.Tr(h.Td(h.Text(msgr.ModelKeys)), h.Td(h.Text(log.ModelKeys))),
							h.If(!hideModelLink && log.ModelLink != "", h.Tr(h.Td(h.Text(msgr.ModelLink)), h.Td(
								v.VBtn(msgr.MoreInfo).Class("text-none text-overline d-flex align-center").
									Variant(v.VariantTonal).Color(v.ColorPrimary).Size(v.SizeXSmall).PrependIcon("mdi-open-in-new").
									Attr("@click", web.POST().
										EventFunc(actions.DetailingDrawer).
										Query(presets.ParamOverlay, actions.Dialog).
										URL(log.ModelLink).
										Go(),
									),
							))),
							h.Tr(h.Td(h.Text(msgr.ModelCreatedAt)), h.Td(h.Text(log.CreatedAt.Format(timeFormat)))),
						),
					),
				),
			))

			if d := field.Value(obj).(string); d != "" {
				switch log.Action {
				case ActionCreate, ActionView, ActionEdit, ActionDelete:
					children = append(children, DiffComponent(d, ctx.R))
				case ActionNote:
					note := &Note{}
					if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
						panic(err)
					}
					children = append(children, VCard().Elevation(0).Children(
						VCardTitle().Class("pa-0").Children(
							h.Text(msgr.ActionNote),
						),
						VCardText().Class("mt-3 pa-3 border-thin rounded").Children(
							h.Div().Class("d-flex flex-column").Children(
								h.Pre(note.Note).Style("white-space: pre-wrap"),
								h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
									return h.Div().Class("text-caption font-italic").Style("color: #757575").Children(
										h.Text(msgr.LastEditedAt(humanize.Time(note.LastEditedAt))),
									)
								}),
							),
						),
					))
				default:
					children = append(children, h.Text(msgr.PerformAction(log.Action, log.Detail)))
				}
			}

			return h.Div().Class("d-flex flex-column ga-3").Children(children...)
		},
	)
	return nil
}

func DiffComponent(diffstr string, req *http.Request) h.HTMLComponent {
	var diffs []Diff
	err := json.Unmarshal([]byte(diffstr), &diffs)
	if err != nil {
		return nil
	}

	if len(diffs) == 0 {
		return nil
	}

	var (
		addediffs   []Diff
		changediffs []Diff
		deletediffs []Diff
	)

	for _, diff := range diffs {
		if diff.New == "" && diff.Old != "" {
			deletediffs = append(deletediffs, diff)
			continue
		}

		if diff.New != "" && diff.Old == "" {
			addediffs = append(addediffs, diff)
			continue
		}

		if diff.New != "" && diff.Old != "" {
			changediffs = append(changediffs, diff)
			continue
		}
	}
	var diffsElems []h.HTMLComponent
	msgr := i18n.MustGetModuleMessages(req, I18nActivityKey, Messages_en_US).(*Messages)

	if len(addediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range addediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.New),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffAdd)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
						h.Tbody(elems...),
					),
				),
			))
	}

	if len(deletediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range deletediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.Old),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffDelete)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
						h.Tbody(elems...),
					),
				),
			))
	}

	if len(changediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range changediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.Old),
				h.Td().Attr("v-pre", true).Text(d.New),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffChanges)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffOld), h.Th(msgr.DiffNew))),
						h.Tbody(elems...),
					),
				),
			))
	}
	return h.Components(diffsElems...)
}
