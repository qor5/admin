package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
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
		listing   = mb.Listing("CreatedAt", "UserID", "Creator", "Action", "ModelKeys", "ModelName")
		detailing = mb.Detailing("ModelLink", "ModelDiffs")
	)

	listing.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	listing.Field("ModelName").Label(Messages_en_US.ModelName)

	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		var (
			logs      = ab.NewLogModelSlice()
			msgr      = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			contextDB = ab.getDBFromContext(ctx.R.Context())
		)

		contextDB.Select("creator").Group("creator").Find(logs)
		reflectVlaue := reflect.Indirect(reflect.ValueOf(logs))
		var creatorOptions []*vuetifyx.SelectItem
		for i := 0; i < reflectVlaue.Len(); i++ {
			creator := reflect.Indirect(reflectVlaue.Index(i)).FieldByName("Creator").String()
			creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
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
				Options: []*vuetifyx.SelectItem{
					{Text: msgr.ActionEdit, Value: ActivityEdit},
					{Text: msgr.ActionCreate, Value: ActivityCreate},
					{Text: msgr.ActionDelete, Value: ActivityDelete},
				},
			},
			{
				Key:          "created",
				Label:        msgr.FilterCreatedAt,
				ItemType:     vuetifyx.ItemTypeDate,
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

	detailing.Field("ModelLink").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			link := field.Value(obj).(string)
			if link == "" {
				return nil
			}

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
			return vuetify.VCard(
				vuetify.VCardTitle(h.Text(msgr.ModelLink)),
				vuetify.VCardText(h.A(h.Text(link)).Href(link)),
			)
		},
	)

	detailing.Field("ModelDiffs").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			record := obj.(interface {
				GetAction() string
				GetCreatedAt() time.Time
			})

			action := record.GetAction()
			if action == ActivityCreate || action == ActivityDelete {
				msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)

				title := msgr.DiffNew
				message := msgr.TheRecordWasCreatedAt
				if action == ActivityDelete {
					title = msgr.DiffDelete
					message = msgr.TheRecordWasDeletedAt
				}
				return vuetify.VCard(
					vuetify.VCardTitle(h.Text(title)),
					vuetify.VCardText(
						h.Text(fmt.Sprintf(message, record.GetCreatedAt().Format("2006-01-02 15:04:05 MST"))),
					),
				).Attr("style", "margin-top:15px;")
			}

			d := field.Value(obj).(string)
			if d == "" {
				return nil
			}
			return DiffComponent(d, ctx.R)
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
