package slug

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/gosimple/unidecode"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type Slug string

const (
	convertToSlugEvent = "slug_ConvertToSlugEvent"
)

func registerEventFuncs(hub web.EventFuncHub) {
	hub.RegisterEventFunc(convertToSlugEvent, convertToSlug)
}

func slugPortalName(field string) string {
	return fmt.Sprintf("%s_PortalData", field)
}

func slugCheckBoxName(field string) string {
	return fmt.Sprintf("%s_Checkbox", field)
}

func (slug Slug) ConfigureField(model *presets.ModelBuilder, field reflect.StructField) error {
	fieldName := strings.TrimSuffix(field.Name, "WithSlug")
	editingBuilder := model.Editing()
	if f := editingBuilder.Field(fieldName); f != nil {
		f.ComponentFunc(SlugEditingComponentFunc)
	}

	editingBuilder.Field(field.Name).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) { return })
	editingBuilder.Field(field.Name).SetterFunc(SlugEditingSetterFunc)
	return nil
}

func SlugEditingComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	registerEventFuncs(ctx.Hub)

	slugTitle := field.Name + "WithSlug"
	return VSheet(
		VTextField().
			Type("text").
			FieldName(field.Name).
			Label(field.Label).
			Value(field.Value(obj)).
			Attr("v-debounce:input", "300").
			Attr("@input:debounced", web.Plaid().
				EventFunc(convertToSlugEvent, slugTitle).Go()),

		VRow(
			VCol(
				web.Portal(
					VTextField().
						Type("text").
						FieldName(slugTitle).
						Label(slugTitle).
						Value(reflectutils.MustGet(obj, slugTitle).(Slug))).Name(slugPortalName(slugTitle)),
			).Cols(8),
			VCol(
				VCheckbox().FieldName(slugCheckBoxName(slugTitle)).InputValue("checked").Label("Sync from "+field.Name),
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

func convertToSlug(ctx *web.EventContext) (r web.EventResponse, err error) {
	if len(ctx.Event.Params) == 0 {
		return
	}

	slugFieldTitle := ctx.Event.Params[0]
	checked := ctx.R.FormValue(slugCheckBoxName(slugFieldTitle))
	if checked != "checked" {
		return
	}

	var (
		regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")
		regexpMultipleDashes     = regexp.MustCompile("-+")
		slug                     = ctx.Event.Value
	)

	slug = strings.TrimSpace(slug)
	slug = unidecode.Unidecode(slug)
	slug = strings.ToLower(slug)
	slug = regexpNonAuthorizedChars.ReplaceAllString(slug, "-")
	slug = regexpMultipleDashes.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-_")

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: slugPortalName(slugFieldTitle),
		Body: VTextField().
			Type("text").
			FieldName(slugFieldTitle).
			Label(slugFieldTitle).
			Value(slug),
	})
	return
}
