package activity

import (
	"encoding/json"
	"net/url"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const (
	I18nActivityKey i18n.ModuleKey = "I18nActivityKey"
)

func (ab *ActivityBuilder) Configure(b *presets.Builder, db *gorm.DB) {
	if err := db.AutoMigrate(ab.GetActivityLogModel()); err != nil {
		panic(err)
	}

	b.I18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN)

	var (
		mb        = b.Model(ab.GetActivityLogModel())
		listing   = mb.Listing("CreatedAt", "Creator", "ModelKeys", "ModelName")
		detailing = mb.Detailing("ModelLink", "ModelDiff")
	)

	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)

		var creatorOptions []*vuetifyx.SelectItem

		if ab.useL10n {
			var logs []ActivityLocaleLog
			db.Select("creator").Group("creator").Find(&logs)
			for _, log := range logs {
				creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
					Text:  log.Creator,
					Value: log.Creator,
				})
			}
		} else {
			var logs []ActivityLog
			db.Select("creator").Group("creator").Find(&logs)
			for _, log := range logs {
				creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
					Text:  log.Creator,
					Value: log.Creator,
				})
			}
		}

		var modelOptions []*vuetifyx.SelectItem
		for _, m := range ab.models {
			modelOptions = append(modelOptions, &vuetifyx.SelectItem{
				Text:  m.name,
				Value: m.name,
			})
		}

		return []*vuetifyx.FilterItem{
			{
				Key:          "action",
				Label:        msgr.FilterAction,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `action %s ?`,
				Options: []*vuetifyx.SelectItem{
					{Text: msgr.ActivityEdit, Value: ActivityEdit},
					{Text: msgr.ActivityCreate, Value: ActivityCreate},
					{Text: msgr.ActivityDelete, Value: ActivityDelete},
					{Text: msgr.ActivityView, Value: ActivityView},
				},
			},
			{
				Key:          "created",
				Label:        msgr.FilterCreatedAt,
				ItemType:     vuetifyx.ItemTypeDate,
				SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
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
				Label: msgr.ActivityAll,
				Query: url.Values{"action": []string{}},
			},
			{
				Label: msgr.ActivityEdit,
				Query: url.Values{"action": []string{ActivityEdit}},
			},
			{
				Label: msgr.ActivityCreate,
				Query: url.Values{"action": []string{ActivityCreate}},
			},
			{
				Label: msgr.ActivityDelete,
				Query: url.Values{"action": []string{ActivityDelete}},
			},
			{
				Label: msgr.ActivityView,
				Query: url.Values{"action": []string{ActivityView}},
			},
		}
	})
	detailing.Field("ModelLink").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			link := field.Value(obj).(string)
			if link == "" {
				return nil
			}

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			return VCard(
				VCardTitle(h.Text(msgr.Link)),
				VCardText(h.A(h.Text(link)).Href(link)),
			)
		},
	)

	detailing.Field("ModelDiffs").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			d := field.Value(obj).(string)
			if d == "" {
				return nil
			}

			var diffs []Diff
			err := json.Unmarshal([]byte(d), &diffs)
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

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			return VCard(
				VCardTitle(h.Text(msgr.Diffs)),
				VList(
					VListItem(h.H4("1"),
						VSimpleTable(
							h.Thead(h.Tr(h.Th("11"), h.Th("22"))),
							h.Tbody(h.Tr(h.Td(h.Text("11aa")), h.Td(h.Text("22c")))),
						),
					),
					VListItem(h.Text("2"), VSimpleTable(
						h.Thead(h.Tr(h.Th("11"), h.Th("22"))),
						h.Tbody(h.Tr(h.Td(h.Text("11aa")), h.Td(h.Text("22c")))),
					)),
					VListItem(h.Text("3"), VSimpleTable(
						h.Thead(h.Tr(h.Th("11"), h.Th("22"))),
						h.Tbody(h.Tr(h.Td(h.Text("11aa")), h.Td(h.Text("22c")))),
					)),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;")
		},
	)
}
