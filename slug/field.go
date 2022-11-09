package slug

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	syncp "sync"

	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/admin/presets"
	"github.com/gosimple/unidecode"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

type Slug string

const (
	syncEvent                  = "slug_sync"
	I18nSlugKey i18n.ModuleKey = "I18nSlugKey"
)

var once = syncp.Once{}

func Configure(b *presets.Builder, mb *presets.ModelBuilder) {
	once.Do(func() {
		b.I18n().
			RegisterForModule(language.English, I18nSlugKey, Messages_en_US).
			RegisterForModule(language.SimplifiedChinese, I18nSlugKey, Messages_zh_CN)
		b.GetWebBuilder().RegisterEventFunc(syncEvent, sync)
	})

	reflectType := reflect.Indirect(reflect.ValueOf(mb.NewModel())).Type()
	if reflectType.Kind() != reflect.Struct {
		panic("slug: model must be struct")
	}
	for i := 0; i < reflectType.NumField(); i++ {
		if reflectType.Field(i).Type != reflect.TypeOf(Slug("")) {
			continue
		}

		fieldName := reflectType.Field(i).Name
		relatedFieldName := strings.TrimSuffix(fieldName, "WithSlug")
		if _, ok := reflectType.FieldByName(relatedFieldName); ok {
			editingBuilder := mb.Editing()
			if f := editingBuilder.Field(relatedFieldName); f != nil {
				f.ComponentFunc(SlugEditingComponentFunc)
			}

			editingBuilder.Field(fieldName).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) { return })
			editingBuilder.Field(fieldName).SetterFunc(SlugEditingSetterFunc)
		}
	}
}

func SlugEditingComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nSlugKey, Messages_en_US).(*Messages)
	slugFieldName := field.Name + "WithSlug"
	slugLabel := field.Label + " Slug"
	return VSheet(
		VTextField().
			Type("text").
			FieldName(field.Name).
			Label(field.Label).
			Value(field.Value(obj)).
			Attr("v-debounce:input", "300").
			Attr("@input:debounced", web.Plaid().
				EventFunc(syncEvent).Query("field_name", field.Name).Query("slug_label", slugLabel).Go()),

		VRow(
			VCol(
				web.Portal(
					VTextField().
						Type("text").
						FieldName(slugFieldName).
						Label(slugLabel).
						Value(reflectutils.MustGet(obj, slugFieldName).(Slug))).Name(portalName(slugFieldName)),
			).Cols(8),
			VCol(
				VCheckbox().FieldName(checkBoxName(slugFieldName)).InputValue("checked").Label(fmt.Sprintf(msgr.Sync, strings.ToLower(field.Label))),
			).Cols(4),
		),
	)
}

func SlugEditingSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	v := ctx.R.FormValue(field.Name)
	err = reflectutils.Set(obj, field.Name, Slug(v))
	if err != nil {
		return
	}
	return
}

func sync(ctx *web.EventContext) (r web.EventResponse, err error) {
	fieldName := ctx.R.FormValue("field_name")
	if fieldName == "" {
		return
	}

	slugFieldName := fieldName + "WithSlug"
	if checked := ctx.R.FormValue(checkBoxName(slugFieldName)); checked != "checked" {
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: portalName(slugFieldName),
		Body: VTextField().
			Type("text").
			FieldName(slugFieldName).
			Label(ctx.R.FormValue("slug_label")).
			Value(slug(ctx.R.FormValue(fieldName))),
	})
	return
}

var (
	regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")
	regexpMultipleDashes     = regexp.MustCompile("-+")
)

func slug(value string) string {
	value = strings.TrimSpace(value)
	value = unidecode.Unidecode(value)
	value = strings.ToLower(value)
	value = regexpNonAuthorizedChars.ReplaceAllString(value, "-")
	value = regexpMultipleDashes.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_")
	return value
}

func portalName(field string) string {
	return fmt.Sprintf("%s_Portal", field)
}

func checkBoxName(field string) string {
	return fmt.Sprintf("%s_Checkbox", field)
}
