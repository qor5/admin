package views

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/publish"
	"github.com/qor5/admin/utils"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func draftCountFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var count int64
		modelSchema, err := schema.Parse(obj, &sync.Map{}, db.NamingStrategy)
		if err != nil {
			return h.Td(h.Text("0"))
		}
		publish.SetPrimaryKeysConditionWithoutVersion(db.Model((reflect.New(modelSchema.ModelType).Interface())), obj, modelSchema).
			Where("status = ?", publish.StatusDraft).Count(&count)

		return h.Td(h.Text(fmt.Sprint(count)))
	}
}

func onlineFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var count int64
		modelSchema, err := schema.Parse(obj, &sync.Map{}, db.NamingStrategy)
		if err != nil {
			return h.Td(h.Text("0"))
		}
		publish.SetPrimaryKeysConditionWithoutVersion(db.Model((reflect.New(modelSchema.ModelType).Interface())), obj, modelSchema).
			Where("status = ?", publish.StatusOnline).Count(&count)

		c := h.Text("-")
		if count > 0 {
			c = VBadge().Color("green")
		}
		return h.Td(c)
	}
}

func StatusListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		if s, ok := obj.(publish.StatusInterface); ok {
			return h.Td(VChip(h.Text(GetStatusText(s.GetStatus(), msgr))).Color(GetStatusColor(s.GetStatus())).Dark(true))
		}
		return nil
	}
}

func StatusEditFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		s, ok := obj.(publish.StatusInterface)
		if !ok || s.GetStatus() == "" {
			return nil
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		var btn h.HTMLComponent
		switch s.GetStatus() {
		case publish.StatusDraft, publish.StatusOffline:
			btn = h.Div(
				VBtn(msgr.Publish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, PublishEvent)),
			)
		case publish.StatusOnline:
			btn = h.Div(
				VBtn(msgr.Unpublish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, UnpublishEvent)),
				VBtn(msgr.Republish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, RepublishEvent)),
			)
		}

		paramID := obj.(presets.SlugEncoder).PrimarySlug()

		return web.Scope(
			VStepper(
				VStepperHeader(
					VStepperStep(h.Text(msgr.StatusDraft)).Step(0).Complete(s.GetStatus() == publish.StatusDraft),
					VDivider(),
					VStepperStep(h.Text(msgr.StatusOnline)).Step(0).Complete(s.GetStatus() == publish.StatusOnline),
				),
			),
			h.Br(),
			btn,
			h.Br(),
			utils.ConfirmDialog(msgr.Areyousure, web.Plaid().EventFunc(web.Var("locals.action")).
				Query(presets.ParamID, paramID).Go(),
				utilsMsgr),
		).Init(`{ action: "", commonConfirmDialog: false}`).VSlot("{ locals }")
	}
}

// need empty setterFunc here to avoid set status to empty when update
func StatusEditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	return
}

func GetStatusColor(status string) string {
	switch status {
	case publish.StatusDraft:
		return "orange"
	case publish.StatusOnline:
		return "green"
	case publish.StatusOffline:
		return "grey"
	}
	return ""
}
