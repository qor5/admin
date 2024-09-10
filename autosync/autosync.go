package autosync

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type InitialChecked int

const (
	InitialCheckedFalse = iota
	InitialCheckedTrue
	InitialCheckedAuto
)

type Config struct {
	SyncFromFromKey string
	CheckboxLabel   string
	InitialChecked  InitialChecked
	SyncCall        func(from string) string
}

func NewLazyWrapComponentFunc(config func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) *Config) func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
	return func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			compo := in(obj, field, ctx)
			if field.Disabled {
				return compo
			}

			cfg := config(obj, field, ctx)
			if cfg == nil {
				return compo
			}

			checkedFormKey := fmt.Sprintf("%s__AutoSync__", field.FormKey)
			var initialCheckedSet string
			if cfg.InitialChecked == InitialCheckedAuto {
				initialCheckedSet = fmt.Sprintf(`form[%q] = (%s) === form[%q]`, checkedFormKey, cfg.SyncCall(fmt.Sprintf("form[%q]", cfg.SyncFromFromKey)), field.FormKey)
			} else {
				initialCheckedSet = fmt.Sprintf(`form[%q] = %t`, checkedFormKey, cfg.InitialChecked == InitialCheckedTrue)
			}
			return h.Div().Class("d-flex align-center ga-2").Children(
				h.Div().Class("flex-grow-1").Children(compo),
				h.Div().Style("display:none").Attr("v-on-mounted", fmt.Sprintf(`({watch,window}) => {
							if (form[%q] === undefined) {
								%s
							}
							window.setTimeout(() => {
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
							}, 10)
						}`,
					checkedFormKey,
					initialCheckedSet,
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

func SyncCallSlug(from string) string {
	return fmt.Sprintf(`plaid().slug(%s||"")`, from)
}
