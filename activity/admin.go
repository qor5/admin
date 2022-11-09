package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/admin/presets"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey i18n.ModuleKey = "I18nActivityKey"
)

func (ab *ActivityBuilder) configureAdmin(b *presets.Builder) {
	b.I18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN)

	var (
		mb        = b.Model(ab.logModel).MenuIcon("receipt_long")
		listing   = mb.Listing("CreatedAt", "Creator", "Action", "ModelKeys", "ModelLabel", "ModelName")
		detailing = mb.Detailing("ModelDiffs")
	)
	ab.lmb = mb
	listing.Field("CreatedAt").Label(Messages_en_US.ModelCreatedAt).ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).CreatedAt.Format("2006-01-02 15:04:05 MST")))
		},
	)
	listing.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	listing.Field("ModelName").Label(Messages_en_US.ModelName)
	listing.Field("ModelLabel").Label(Messages_en_US.ModelLabel).ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			if obj.(*ActivityLog).ModelLabel == "" {
				return h.Td(h.Text("-"))
			}
			return h.Td(h.Text(obj.(*ActivityLog).ModelLabel))
		},
	)

	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		var (
			msgr      = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			contextDB = ab.getDBFromContext(ctx.R.Context())
		)

		creatorGroups := ab.NewLogModelSlice()
		contextDB.Select("creator").Group("creator").Find(creatorGroups)
		creatorGroupsValues := reflect.Indirect(reflect.ValueOf(creatorGroups))
		var creatorOptions []*vuetifyx.SelectItem
		for i := 0; i < creatorGroupsValues.Len(); i++ {
			creator := reflect.Indirect(creatorGroupsValues.Index(i)).FieldByName("Creator").String()
			creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
				Text:  creator,
				Value: creator,
			})
		}

		actionGroups := ab.NewLogModelSlice()
		contextDB.Select("action").Group("action").Order("action").Find(actionGroups)
		actionGroupsValues := reflect.Indirect(reflect.ValueOf(actionGroups))
		var actionOptions []*vuetifyx.SelectItem
		for i := 0; i < actionGroupsValues.Len(); i++ {
			creator := reflect.Indirect(actionGroupsValues.Index(i)).FieldByName("Action").String()
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  creator,
				Value: creator,
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
				Query: url.Values{"action": []string{ActivityEdit}},
			},
			{
				Label: msgr.ActionCreate,
				Query: url.Values{"action": []string{ActivityCreate}},
			},
			{
				Label: msgr.ActionDelete,
				Query: url.Values{"action": []string{ActivityDelete}},
			},
		}
	})

	detailing.Field("ModelDiffs").Label("Detail").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			var (
				record = obj.(ActivityLogInterface)
				msgr   = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			)

			var detailElems []h.HTMLComponent
			detailElems = append(detailElems, vuetify.VCard(
				vuetify.VCardTitle(h.Text(msgr.DiffDetail)),
				vuetify.VSimpleTable(
					h.Tbody(
						h.Tr(h.Td(h.Text(msgr.ModelCreator)), h.Td(h.Text(record.GetCreator()))),
						h.Tr(h.Td(h.Text(msgr.ModelUserID)), h.Td(h.Text(fmt.Sprintf("%v", record.GetUserID())))),
						h.Tr(h.Td(h.Text(msgr.ModelAction)), h.Td(h.Text(record.GetAction()))),
						h.Tr(h.Td(h.Text(msgr.ModelName)), h.Td(h.Text(record.GetModelName()))),
						h.Tr(h.Td(h.Text(msgr.ModelLabel)), h.Td(h.Text(record.GetModelLabel()))),
						h.Tr(h.Td(h.Text(msgr.ModelKeys)), h.Td(h.Text(record.GetModelKeys()))),
						h.If(record.GetModelLink() != "", h.Tr(h.Td(h.Text(msgr.ModelLink)), h.Td(h.Text(record.GetModelLink())))),
						h.Tr(h.Td(h.Text(msgr.ModelCreatedAt)), h.Td(h.Text(record.GetCreatedAt().Format("2006-01-02 15:04:05 MST")))),
					),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))

			if d := field.Value(obj).(string); d != "" {
				detailElems = append(detailElems, DiffComponent(d, ctx.R))
			}

			return h.Components(detailElems...)
		},
	)
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
			elems = append(elems, h.Tr(h.Td(h.Text(d.Field)), h.Td(h.Text(fixSpecialChars(d.Now)))))
		}

		diffsElems = append(diffsElems,
			vuetify.VCard(
				vuetify.VCardTitle(h.Text(msgr.DiffNew)),
				vuetify.VSimpleTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
	}

	if len(deletediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range deletediffs {
			elems = append(elems, h.Tr(h.Td(h.Text(d.Field)), h.Td(h.Text(fixSpecialChars(d.Old)))))
		}

		diffsElems = append(diffsElems,
			vuetify.VCard(
				vuetify.VCardTitle(h.Text(msgr.DiffDelete)),
				vuetify.VSimpleTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
	}

	if len(changediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range changediffs {
			elems = append(elems, h.Tr(h.Td(h.Text(d.Field)), h.Td(h.Text(fixSpecialChars(d.Old)), h.Td(h.Text(fixSpecialChars(d.Now))))))
		}

		diffsElems = append(diffsElems,
			vuetify.VCard(
				vuetify.VCardTitle(h.Text(msgr.DiffChanges)),
				vuetify.VSimpleTable(
					h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffOld), h.Th(msgr.DiffNow))),
					h.Tbody(elems...),
				),
			).Attr("style", "margin-top:15px;margin-bottom:15px;"))
	}
	return h.Components(diffsElems...)
}
