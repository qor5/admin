package admin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"

	"github.com/qor5/admin/v3/presets"
)

type (
	DemoCase struct {
		gorm.Model
		Name         string
		FieldData    FieldData    `gorm:"type:json"`
		SelectData   SelectData   `gorm:"type:json"`
		CheckboxData CheckboxData `gorm:"type:json"`
	}
	FieldData struct {
		Text             string
		Textarea         string
		TextValidate     string
		TextareaValidate string
	}
	SelectData struct {
		AutoComplete []int
		NormalSelect int
	}
	CheckboxData struct {
		Checkbox bool
	}

	DemoSelectItem struct {
		ID   int
		Name string
	}
)

func (c *FieldData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *FieldData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *SelectData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *SelectData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CheckboxData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *CheckboxData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func configureDemoCase(b *presets.Builder, db *gorm.DB) {
	err := db.AutoMigrate(&DemoCase{})
	if err != nil {
		panic(err)
	}
	mb := b.Model(&DemoCase{})
	mb.Editing("Name")
	mb.Listing("ID", "Name")
	detailing := mb.Detailing("FieldSection", "SelectSection", "CheckboxSection", "DialogSection")
	configVxField(detailing)
	configVxSelect(detailing)
	configVxCheckBox(detailing)
	configVxDialog(detailing)
	return
}

func DemoCaseTextField(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors) *vx.VXFieldBuilder {
	fieldName := fmt.Sprintf("%s.%s", editField, field)
	formKey := fmt.Sprintf("%s.%s", section, fieldName)
	return vx.VXField().
		Label(label).
		Attr(web.VField(formKey, reflectutils.MustGet(obj, fieldName))...).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}

func DemoCaseSelect(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors, items interface{}) *vx.VXSelectBuilder {
	fieldName := fmt.Sprintf("%s.%s", editField, field)
	formKey := fmt.Sprintf("%s.%s", section, fieldName)
	return vx.VXSelect().
		Label(label).
		Items(items).
		ItemTitle("Name").
		ItemValue("ID").
		Attr(web.VField(formKey, reflectutils.MustGet(obj, fieldName))...).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}

func DemoCaseCheckBox(obj interface{}, section, editField, field, label string) *vx.VXCheckboxBuilder {
	fieldName := fmt.Sprintf("%s.%s", editField, field)
	formKey := fmt.Sprintf("%s.%s", section, fieldName)
	return vx.VXCheckbox().
		Label(label).
		Attr(web.VField(formKey, reflectutils.MustGet(obj, fieldName))...)
}

func configVxField(detailing *presets.DetailingBuilder) {
	section := "FieldSection"
	editField := "FieldData"
	label := "vx-field"
	generateSection(detailing, section, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				v.VRow(
					v.VCol(
						DemoCaseTextField(obj, section, editField, "Text", "Text", vErr).
							Tips("This is Tips").
							Autofocus(true),
					),
					v.VCol(
						DemoCaseTextField(obj, section, editField, "Textarea", "Textarea", vErr).
							Type("textarea"),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseTextField(obj, section, editField, "TextValidate", "TextValidate(input more than 5 chars)", vErr).Required(true),
					),
					v.VCol(
						DemoCaseTextField(obj, section, editField, "TextareaValidate", "TextareaValidate(input more than 10 chars)", vErr).Required(true).
							Type("textarea"),
					),
				),
			)
		}).WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			in(obj, ctx)
			p := obj.(*DemoCase)
			if len(p.FieldData.TextValidate) < 5 {
				err.FieldError(fmt.Sprintf("%s.%s.TextValidate", section, editField), "input more than 5 chars")
			}
			if len(p.FieldData.TextareaValidate) < 10 {
				err.FieldError(fmt.Sprintf("%s.%s.TextareaValidate", section, editField), "input more than 10 chars")
			}
			return
		}
	})
}

func configVxSelect(detailing *presets.DetailingBuilder) {
	section := "SelectSection"
	editField := "SelectData"
	label := "vx-select"
	generateSection(detailing, section, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			items := []DemoSelectItem{
				{ID: 1, Name: "Petter"},
				{ID: 2, Name: "John"},
				{ID: 3, Name: "Devi"},
				{ID: 4, Name: "Anna"},
				{ID: 5, Name: "Jane"},
				{ID: 6, Name: "Britta"},
				{ID: 7, Name: "Sandra"},
				{ID: 8, Name: "Trevor"},
			}
			return h.Components(
				v.VRow(
					v.VCol(
						DemoCaseSelect(obj, section, editField, "AutoComplete", "AutoComplete(select more than 1 item)", vErr, items).
							Type("autocomplete").Multiple(true).Chips(true),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseSelect(obj, section, editField, "NormalSelect", "NormalSelect(can`t select Trevor)", vErr, items).
							Type("autocomplete"),
					),
				),
			)
		}).WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			in(obj, ctx)
			p := obj.(*DemoCase)
			if len(p.SelectData.AutoComplete) == 1 {
				err.FieldError(fmt.Sprintf("%s.%s.AutoComplete", section, editField), "select more than 1 item")
			}
			if p.SelectData.NormalSelect == 8 {
				err.FieldError(fmt.Sprintf("%s.%s.NormalSelect", section, editField), "can`t select Trevor")
			}
			return
		}
	})
}

func configVxCheckBox(detailing *presets.DetailingBuilder) {
	section := "CheckboxSection"
	editField := "CheckboxData"
	label := "vx-checkbox"
	generateSection(detailing, section, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Components(
				v.VRow(
					v.VCol(
						DemoCaseCheckBox(obj, section, editField, "Checkbox", "Checkbox").
							TrueLabel("True").
							TrueIconColor(v.ColorPrimary).
							FalseLabel("False").
							Title("CheckboxTitle").
							FalseIcon("mdi-circle-outline").
							FalseIconColor(v.ColorError),
					),
				),
			)
		})
}

func configVxDialog(detailing *presets.DetailingBuilder) {
	section := "DialogSection"
	label := "vx-dialog"
	detailing.Section(section).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		text := "This is an info description line This is an info description lineThis is an info description lineThis is an info description lineThis is an info description line"
		return web.Scope(
			h.Div(
				h.Div(
					h.H2(label).Class("section-title"),
				).Class("section-title-wrap"),
				v.VRow(
					v.VCol(
						h.Div(h.Text("v-model")).Class("mb-2"),
						v.VBtn("Open Dialog").Color(v.ColorPrimary).
							Attr("@click", "locals.dialogVisible=true"),

						vx.VXDialog().
							Attr("v-model", "locals.dialogVisible").
							Title("ModelValue").
							Text(text),
					),
					v.VCol(
						vx.VXDialog(
							web.Slot(
								h.Div(h.Text("activator slot")).Class("mb-2"),
								v.VBtn("Open Dialog").Color(v.ColorSecondary).Attr("v-bind", "activatorProps"),
							).Name("activator").Scope("{props: { activatorProps }}")).
							Title("ActivatorSlot").
							Text(text),
					),
				),
			).Class("section-wrap with-border-b"),
		).VSlot("{locals}").Init("{dialogVisible:false}")
	})
}

func generateSection(detailing *presets.DetailingBuilder, section, editField, label string) *presets.SectionBuilder {
	return detailing.Section(section).Label(label).Editing(editField).
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			p := obj.(*DemoCase)
			j := jsoniter.Config{
				EscapeHTML: false,
			}.Froze()
			data := reflectutils.MustGet(p, editField)
			jsonBytes, _ := j.MarshalIndent(data, "", "    ")
			return vx.VXReadonlyField().Value(string(jsonBytes)).Label(editField)
		})
}
