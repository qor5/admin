package admin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

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
		Name           string
		FieldData      FieldData      `gorm:"type:json"`
		SelectData     SelectData     `gorm:"type:json"`
		CheckboxData   CheckboxData   `gorm:"type:json"`
		DatepickerData DatepickerData `gorm:"type:json"`
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
	DatepickerData struct {
		Date                 int64
		DateTime             int64
		DateRange            []int64
		DateRangeNeedConfirm []int64
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
func (c *DatepickerData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *DatepickerData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func configureDemoCase(b *presets.Builder, db *gorm.DB) {
	err := db.AutoMigrate(&DemoCase{})
	if err != nil {
		panic(err)
	}
	mb := b.Model(&DemoCase{})
	mb.Editing("Name").WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			if in != nil {
				in(obj, ctx)
			}
			p := obj.(*DemoCase)
			if p.ID == 0 {
				return
			}
			if len(p.FieldData.TextValidate) < 5 {
				err.FieldError(fmt.Sprintf("%s.%s.TextValidate", "FieldSection", "FieldData"), "input more than 5 chars")
			}
			if len(p.FieldData.TextareaValidate) < 10 {
				err.FieldError(fmt.Sprintf("%s.%s.TextareaValidate", "FieldSection", "FieldData"), "input more than 10 chars")
			}
			if len(p.SelectData.AutoComplete) == 1 {
				err.FieldError(fmt.Sprintf("%s.%s.AutoComplete", "SelectSection", "SelectData"), "select more than 1 item")
			}
			if p.SelectData.NormalSelect == 8 {
				err.FieldError(fmt.Sprintf("%s.%s.NormalSelect", "SelectSection", "SelectData"), "can`t select Trevor")
			}
			if p.DatepickerData.Date == 0 {
				err.FieldError(fmt.Sprintf("%s.%s.Date", "DatepickerSection", "DatepickerData"), "Date is required")
			}
			if p.DatepickerData.DateTime == 0 {
				err.FieldError(fmt.Sprintf("%s.%s.DateTime", "DatepickerSection", "DatepickerData"), "DateTime is required")
			}
			if p.DatepickerData.DateRange == nil || p.DatepickerData.DateRange[1] < p.DatepickerData.DateRange[0] {
				err.FieldError(fmt.Sprintf("%s.%s.DateRange", "DatepickerSection", "DatepickerData"), "End later than Start")
			}
			return
		}
	})
	mb.Listing("ID", "Name")
	detailing := mb.Detailing("FieldSection", "SelectSection", "CheckboxSection", "DatepickerSection", "DialogSection")
	configVxField(detailing, mb)
	configVxSelect(detailing, mb)
	configVxCheckBox(detailing, mb)
	configVxDialog(detailing, mb)
	configVxDatepicker(detailing, mb)
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

func configVxField(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	sectionName := "FieldSection"
	editField := "FieldData"
	label := "vx-field"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				v.VRow(
					v.VCol(
						vx.VXField().
							Label("Text(Disabled)").
							ModelValue("This is Disabled Vx-Field").
							Disabled(true),
					),
					v.VCol(
						vx.VXField().
							Label("Textarea(Disabled)").
							ModelValue("This is Readonly Vx-Field type Textarea").
							Disabled(true).
							Type("textarea"),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseTextField(obj, sectionName, editField, "Text", "Text", vErr).
							Tips("This is Tips").Clearable(true),
					),
					v.VCol(
						DemoCaseTextField(obj, sectionName, editField, "Textarea", "Textarea", vErr).
							Type("textarea").Clearable(true),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseTextField(obj, sectionName, editField, "TextValidate", "TextValidate(input more than 5 chars)", vErr).Required(true).Clearable(true),
					),
					v.VCol(
						DemoCaseTextField(obj, sectionName, editField, "TextareaValidate", "TextareaValidate(input more than 10 chars)", vErr).Required(true).
							Type("textarea").Clearable(true),
					),
				),
			)
		})
	detailing.Section(section)
}

func configVxSelect(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	sectionName := "SelectSection"
	editField := "SelectData"
	label := "vx-select"
	section := generateSection(detailing, mb, sectionName, editField, label).
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
						DemoCaseSelect(obj, sectionName, editField, "AutoComplete", "AutoComplete(select more than 1 item)", vErr, items).
							Type("autocomplete").Multiple(true).Chips(true).ClosableChips(true).Clearable(true),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseSelect(obj, sectionName, editField, "NormalSelect", "", vErr, items).
							Attr(":rules", `[(value) => value !== 8 || "can't select Trevor"]`).
							Type("autocomplete"),
					),
				),
			)
		})
	detailing.Section(section)
}

func configVxCheckBox(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	sectionName := "CheckboxSection"
	editField := "CheckboxData"
	label := "vx-checkbox"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Components(
				v.VRow(
					v.VCol(
						DemoCaseCheckBox(obj, sectionName, editField, "Checkbox", "Checkbox").
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
	detailing.Section(section)
}

func dialogCard(title string, comp ...h.HTMLComponent) h.HTMLComponent {
	rows := v.VRow()
	for _, c := range comp {
		rows.AppendChildren(v.VCol(c))
	}
	return v.VCard(
		v.VCardItem(
			rows,
		),
	).Title(title).Class("pa-2 my-4")
}

func configVxDialog(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	label := "vx-dialog"
	section := presets.NewSectionBuilder(mb, "DialogSection").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		text := "This is an info description line This is an info description lineThis is an info description lineThis is an info description lineThis is an info description line"
		textLarge := text
		for i := 0; i < 5; i++ {
			textLarge += text
		}
		return web.Scope(
			h.Div(
				h.Div(
					h.H2(label).Class("section-title"),
				).Class("section-title-wrap"),
				dialogCard("activator",
					h.Components(h.Div(h.Text("v-model")).Class("mb-2"),
						v.VBtn("Open Dialog").Color(v.ColorPrimary).
							Attr("@click", "locals.dialogVisible=true"),
						vx.VXDialog().
							Attr("v-model", "locals.dialogVisible").
							Title("ModelValue").
							Text(text),
					),
					dialogActivator("Open Dialog", "Activator Slot(HideClose)", text, v.ColorSecondary).HideClose(true).
						OkText("✅").CancelText("❌"),
				),

				dialogCard("type",
					dialogActivator("Open Dialog", "Info(HideCancel)", text, v.ColorInfo).Type("info").
						HideCancel(true).
						Title("Info"),
					dialogActivator("Open Dialog", "Success(HideOk)", text, v.ColorSuccess).Type("success").
						HideOk(true).
						Title("Success"),
					dialogActivator("Open Dialog", "Warning(HideFooter)", text, v.ColorWarning).Type("warn").
						HideFooter(true).
						Title("Warning"),
					dialogActivator("Open Dialog", "Error(DisableOk)", text, v.ColorError).Type("error").
						DisableOk(true).
						Title("Error"),
				),
				dialogCard("size",
					dialogActivator("Open Dialog", "Large", text, v.ColorSecondary).Size("large"),
					dialogActivator("Open Dialog", "Custom Width", textLarge, v.ColorSecondary).Width(200),
					dialogActivator("Open Dialog", "Custom Height", text, v.ColorSecondary).Height(400),
				),
			).Class("section-wrap with-border-b"),
		).VSlot("{locals}").Init("{dialogVisible:false}")
	})
	detailing.Section(section)
}

func generateSection(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder, section, editField, label string) *presets.SectionBuilder {
	return presets.NewSectionBuilder(mb, section).Label(label).Editing(editField).
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

func DemoCaseDatepicker(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors) *vx.VXDatePickerBuilder {
	fieldName := fmt.Sprintf("%s.%s", editField, field)
	formKey := fmt.Sprintf("%s.%s", section, fieldName)
	val := reflectutils.MustGet(obj, fieldName)
	return vx.VXDatepicker().
		Clearable(true).
		Label(label).
		Attr(web.VField(formKey, val)...).
		Placeholder(field).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}
func DemoCaseRangePicker(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors) *vx.VXRangePickerBuilder {
	fieldName := fmt.Sprintf("%s.%s", editField, field)
	formKey := fmt.Sprintf("%s.%s", section, fieldName)
	val := reflectutils.MustGet(obj, fieldName)
	return vx.VXRangePicker().
		Clearable(true).
		Label(label).
		Attr(web.VField(formKey, val)...).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}
func configVxDatepicker(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	label := "vx-datepicker"
	sectionName := "DatepickerSection"
	editField := "DatepickerData"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				v.VRow(
					v.VCol(
						DemoCaseDatepicker(obj, sectionName, editField, "Date", "date-picker(required,Within five days before and after)", vErr).
							DatePickerProps(map[string]string{"min": time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
								"max": time.Now().AddDate(0, 0, 5).Format("2006-01-02")}),
					),
					v.VCol(
						DemoCaseDatepicker(obj, sectionName, editField, "DateTime", "datetime-picker(required)", vErr).Type("datetimepicker"),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseRangePicker(obj, sectionName, editField, "DateRange", "range-picker(end>start)", vErr).Placeholder([]string{"Start", "End"}),
					),
					v.VCol(
						DemoCaseRangePicker(obj, sectionName, editField, "DateRangeNeedConfirm", "range-picker (needConfirm)", vErr).NeedConfirm(true).Placeholder([]string{"Begin", "End"}),
					),
				),
			)
		})
	detailing.Section(section)
}

func dialogActivator(btn, label, text, color string) *vx.VXDialogBuilder {
	return vx.VXDialog(
		web.Slot(
			h.Div(h.Text(label)).Class("mb-2"),
			v.VBtn(btn).Color(color).Attr("v-bind", "activatorProps"),
		).Name("activator").Scope("{props: { activatorProps }}"),
	).Text(text)
}
