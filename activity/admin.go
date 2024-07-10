package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey      i18n.ModuleKey = "I18nActivityKey"
	DetailFieldTimeline  string         = "Timeline"
	ListFieldUnreadNotes string         = "UnreadNotes"
)

func (ab *Builder) Install(b *presets.Builder) error {
	// TODO: 缺少日文？
	b.GetI18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN)

	if permB := b.GetPermission(); permB != nil {
		permB.CreatePolicies(ab.permPolicy)
	}

	mb := b.Model(&ActivityLog{}).MenuIcon("mdi-book-edit")
	return ab.logModelInstall(b, mb)
}

func (ab *Builder) defaultLogModelInstall(b *presets.Builder, mb *presets.ModelBuilder) error {
	// TODO: i18n ?
	var (
		listing   = mb.Listing("CreatedAt", "Creator", "Action", "ModelKeys", "ModelLabel", "ModelName")
		detailing = mb.Detailing("ModelDiffs").Drawer(true)
	)
	ab.lmb = mb

	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
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

	listing.Field("CreatedAt").Label(Messages_en_US.ModelCreatedAt).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).CreatedAt.Format(timeFormat)))
		},
	)
	listing.Field("Creator").Label(Messages_en_US.ModelCreator).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).Creator.Name))
		},
	)
	listing.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	listing.Field("ModelName").Label(Messages_en_US.ModelName)
	listing.Field("ModelLabel").Label(Messages_en_US.ModelLabel).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			if obj.(*ActivityLog).ModelLabel == "" {
				return h.Td(h.Text("-"))
			}
			return h.Td(h.Text(obj.(*ActivityLog).ModelLabel))
		},
	)

	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)

		var actionOptions []*vuetifyx.SelectItem
		for _, action := range DefaultActions {
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  action, // TODO: i18n label ?
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

		creatorIDs := []uint{}
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
				Value: fmt.Sprint(creator.ID),
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

	listing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		return []*presets.FilterTab{
			{
				Label: msgr.ActionAll,
				Query: url.Values{"action": []string{}},
			},
			{
				Label: msgr.ActionView,
				Query: url.Values{"action": []string{ActionView}},
			},
			{
				Label: msgr.ActionEdit,
				Query: url.Values{"action": []string{ActionEdit}},
			},
			{
				Label: msgr.ActionCreate,
				Query: url.Values{"action": []string{ActionCreate}},
			},
			{
				Label: msgr.ActionDelete,
				Query: url.Values{"action": []string{ActionDelete}},
			},
		}
	})

	// TODO: 这个 Label 最后会被 i18n 处理吗？
	detailing.Field("ModelDiffs").Label("Detail").ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			var (
				record = obj.(*ActivityLog)
				msgr   = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			)
			var detailElems []h.HTMLComponent
			detailElems = append(detailElems, VCard(
				VCardTitle(
					VBtn("").Children(
						VIcon("mdi-account").Class("pr-2").Size(SizeSmall),
					).Icon(true).Attr("@click", "window.history.back()"),
					h.Text(" "+msgr.DiffDetail),
				),
				VCardText(
					// vuetif.VAvatar().Class("mr-2").Children(
					//	VIcon("mdi-account").Size(SizeSmall),
					// ),
					h.Text(" "+msgr.DiffDetail),
				),
				VTable(
					h.Tbody(
						h.Tr(h.Td(h.Text(msgr.ModelCreator)), h.Td(h.Text(record.Creator.Name))),
						h.Tr(h.Td(h.Text(msgr.ModelUserID)), h.Td(h.Text(fmt.Sprint(record.CreatorID)))),
						h.Tr(h.Td(h.Text(msgr.ModelAction)), h.Td(h.Text(string(record.Action)))),
						h.Tr(h.Td(h.Text(msgr.ModelName)), h.Td(h.Text(record.ModelName))),
						h.Tr(
							h.Td(h.Text(msgr.ModelLabel)),
							h.Td(h.Text(func() string {
								if record.ModelLabel == "" {
									return "-"
								}
								return record.ModelLabel
							}())),
						),
						h.Tr(h.Td(h.Text(msgr.ModelKeys)), h.Td(h.Text(record.ModelKeys))),
						h.If(record.ModelLink != "", h.Tr(h.Td(h.Text(msgr.ModelLink)), h.Td(h.Text(record.ModelLink)))),
						h.Tr(h.Td(h.Text(msgr.ModelCreatedAt)), h.Td(h.Text(record.CreatedAt.Format(timeFormat)))),
					),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))

			if d := field.Value(obj).(string); d != "" {
				detailElems = append(detailElems, DiffComponent(d, ctx.R))
			}

			return h.Components(detailElems...)
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
			VCard(
				VCardTitle(h.Text(msgr.DiffAdd)),
				VTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
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
			VCard(
				VCardTitle(h.Text(msgr.DiffDelete)),
				VTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
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
			VCard(
				VCardTitle(h.Text(msgr.DiffChanges)),
				VTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffOld), h.Th(msgr.DiffNew))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
	}
	return h.Components(diffsElems...)
}
