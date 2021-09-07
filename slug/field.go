package slug

import (
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type Slug struct {
	Slug string
}

// Scan scan value into Slug
func (slug *Slug) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		slug.Slug = string(bytes)
	} else if str, ok := value.(string); ok {
		slug.Slug = str
	} else if strs, ok := value.([]string); ok {
		slug.Slug = strs[0]
	}
	return nil
}

// Value get slug's Value
func (slug Slug) Value() (driver.Value, error) {
	return slug.Slug, nil
}

// This interface method is best to be called here. https://github.com/goplaid/x/blob/master/presets/field_defaults.go#L132 It is not necessary to configure this field for each model and project.
func (slug Slug) ConfigureField(model *presets.ModelBuilder, field reflect.StructField) error {
	fieldName := strings.TrimSuffix(field.Name, "WithSlug")
	editingBuilder := model.Editing()
	if f := editingBuilder.Field(fieldName); f != nil {
		f.ComponentFunc(SlugEditingComponentFunc)
	}

	editingBuilder.Field(field.Name).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) { return })
	editingBuilder.Field(field.Name).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.FormValue(field.Name)
		err = reflectutils.Set(obj, field.Name, Slug{Slug: v})
		if err != nil {
			return
		}
		return
	})

	listingBuilder := model.Listing()
	listingBuilder.Field(field.Name).ComponentFunc(SlugListingComponentFunc)
	return nil
}

func SlugEditingComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	ctx.Hub.RegisterEventFunc("slug_sync", syncSlug)

	slugTitle := field.Name + "WithSlug"
	return VSheet(
		VTextField().
			Type("text").
			FieldName(field.Name).
			Label(field.Label).
			Value(reflectutils.MustGet(obj, field.Name).(string)).Attr("@input", web.Plaid().
			EventFunc("slug_sync", slugTitle).Go()),

		VRow(
			VCol(
				web.Portal(
					VTextField().
						Type("text").
						FieldName(slugTitle).
						Label(slugTitle).
						Value(reflectutils.MustGet(obj, slugTitle).(Slug).Slug)).Name("slug_sync_data"),
			).Cols(8),
			VCol(
				VCheckbox().FieldName("SlugSync").InputValue("checked").Label("Sync from "+field.Name),
			).Cols(4),
		),
	)
}
func SlugListingComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	slug := field.Value(obj).(Slug)
	return h.Td(h.Text(slug.Slug))
}

func syncSlug(ctx *web.EventContext) (r web.EventResponse, err error) {
	checked := ctx.R.FormValue("SlugSync")
	if checked == "" {
		return
	}

	slugTitle := ctx.Event.Params[0]

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "slug_sync_data",
		Body: VTextField().
			Type("text").
			FieldName(slugTitle).
			Label(slugTitle).
			Value(ctx.Event.Value),
	})
	return
}
