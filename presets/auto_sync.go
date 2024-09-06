package presets

import (
	"fmt"

	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type AutoSyncConfig struct {
	SyncFromFromKey string
	CheckboxLabel   string
	InitialChecked  bool
	SyncCall        func(from string) string
}

func SyncCallSlug(from string) string {
	return fmt.Sprintf(`plaid().slug(%s||"")`, from)
}

func WrapperAutoSync(config func(obj interface{}, field *FieldContext, ctx *web.EventContext) *AutoSyncConfig) func(in FieldComponentFunc) FieldComponentFunc {
	return func(in FieldComponentFunc) FieldComponentFunc {
		return func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			compo := in(obj, field, ctx)

			cfg := config(obj, field, ctx)
			if cfg == nil {
				return compo
			}

			checkedFormKey := fmt.Sprintf("%s__AutoSync__", field.FormKey)
			return h.Div().Class("d-flex align-center ga-2").Children(
				h.Div().Class("flex-grow-1").Children(compo),
				h.Div().Style("display:none").Attr("v-on-mounted", fmt.Sprintf(`({watch}) => {
							if (form[%q] === undefined) {
								form[%q] = %t;
							}
							const _sync = () => {
								if (!!form[%q]) {
									form[%q] = %s;
								}
							}
							watch(() => form[%q], (value) => {
								_sync()
							}, { immediate: true })
							watch(() => form[%q], (value) => {
								_sync()
							})
						}`,
					checkedFormKey,
					checkedFormKey, cfg.InitialChecked,
					checkedFormKey,
					field.FormKey, cfg.SyncCall(fmt.Sprintf("form[%q]", cfg.SyncFromFromKey)),
					checkedFormKey,
					cfg.SyncFromFromKey,
				)),
				h.Div(
					h.Span("").Class("text-subtitle-2 text-high-emphasis section-filed-label mb-1 d-sm-inline-block"),
					v.VCheckbox().
						Attr("v-model", fmt.Sprintf("form[%q]", checkedFormKey)).
						Label(cfg.CheckboxLabel).
						Disabled(field.Disabled),
				).Class("section-field-wrap"),
			)
		}
	}
}
