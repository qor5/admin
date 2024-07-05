package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey     i18n.ModuleKey = "I18nActivityKey"
	DetailFieldTimeline string         = "Timeline"
	ListFieldNotes      string         = "Notes"
)

func (ab *Builder) Install(b *presets.Builder) error {
	b.GetI18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN)

	// if permB := b.GetPermission(); permB != nil {
	// 	permB.CreatePolicies(ab.permPolicy)
	// }
	mb := b.Model(&ActivityLog{}).MenuIcon("mdi-book-edit")
	return ab.logModelInstall(b, mb)
}

func (ab *Builder) defaultLogModelInstall(b *presets.Builder, mb *presets.ModelBuilder) error {
	var (
		listing   = mb.Listing("CreatedAt", "Creator", "Action", "ModelKeys", "ModelLabel", "ModelName")
		detailing = mb.Detailing("ModelDiffs")
	)
	ab.lmb = mb
	listing.Field("CreatedAt").Label(Messages_en_US.ModelCreatedAt).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).CreatedAt.Format("2006-01-02 15:04:05 MST")))
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
		var (
			msgr      = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			contextDB = ab.db
		)

		creatorGroups := []*ActivityLog{}
		err := contextDB.Select("creator").Group("creator").Find(&creatorGroups).Error
		if err != nil {
			panic(err)
		}
		var creatorOptions []*vuetifyx.SelectItem

		for _, creator := range creatorGroups {
			creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
				Text:  creator.Creator.Name,
				Value: fmt.Sprint(creator.Creator.ID),
			})
		}

		actionGroups := []*ActivityLog{}
		err = contextDB.Select("action").Group("action").Order("action").Find(&actionGroups).Error
		if err != nil {
			panic(err)
		}
		var actionOptions []*vuetifyx.SelectItem

		for _, action := range actionGroups {
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  string(action.Action),
				Value: string(action.Action),
			})
		}

		var modelOptions []*vuetifyx.SelectItem
		for _, m := range ab.models {
			modelOptions = append(modelOptions, &vuetifyx.SelectItem{
				Text:  m.typ.Name(),
				Value: m.typ.Name(),
			})
		}

		return []*vuetifyx.FilterItem{
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
			{
				Key:          "creator",
				Label:        msgr.FilterCreator,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `creator %s ?`,
				Options:      creatorOptions,
			},
			{
				Key:          "model",
				Label:        msgr.FilterModel,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `model_name %s ?`,
				Options:      modelOptions,
			},
		}
	})

	listing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		return []*presets.FilterTab{
			{
				Label: msgr.ActionAll,
				Query: url.Values{"action": []string{}},
			},
			{
				Label: msgr.ActionEdit,
				Query: url.Values{"action": []string{string(ActionEdit)}},
			},
			{
				Label: msgr.ActionCreate,
				Query: url.Values{"action": []string{string(ActionCreate)}},
			},
			{
				Label: msgr.ActionDelete,
				Query: url.Values{"action": []string{string(ActionDelete)}},
			},
		}
	})

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
						h.Tr(h.Td(h.Text(msgr.ModelUserID)), h.Td(h.Text(fmt.Sprintf("%v", record.UserID)))),
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
						h.Tr(h.Td(h.Text(msgr.ModelCreatedAt)), h.Td(h.Text(record.CreatedAt.Format("2006-01-02 15:04:05 MST")))),
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

func fixSpecialChars(str string) string {
	str = strings.Replace(str, "{", "[", -1)
	str = strings.Replace(str, "}", "]", -1)
	return str
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
		newdiffs    []Diff
		changediffs []Diff
		deletediffs []Diff
	)

	for _, diff := range diffs {
		if diff.Now == "" && diff.Old != "" {
			deletediffs = append(deletediffs, diff)
			continue
		}

		if diff.Now != "" && diff.Old == "" {
			newdiffs = append(newdiffs, diff)
			continue
		}

		if diff.Now != "" && diff.Old != "" {
			changediffs = append(changediffs, diff)
			continue
		}
	}
	var diffsElems []h.HTMLComponent
	msgr := i18n.MustGetModuleMessages(req, I18nActivityKey, Messages_en_US).(*Messages)

	if len(newdiffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range newdiffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Text(fixSpecialChars(d.Now)),
			))
		}

		diffsElems = append(diffsElems,
			VCard(
				VCardTitle(h.Text(msgr.DiffNew)),
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
				h.Td().Text(fixSpecialChars(d.Old)),
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
				h.Td().Text(fixSpecialChars(d.Old)),
				h.Td().Text(fixSpecialChars(d.Now)),
			))
		}

		diffsElems = append(diffsElems,
			VCard(
				VCardTitle(h.Text(msgr.DiffChanges)),
				VTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffOld), h.Th(msgr.DiffNow))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
	}
	return h.Components(diffsElems...)
}
