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
		Name              string
		FieldData         FieldData         `gorm:"type:json"`
		FieldTextareaData FieldTextareaData `gorm:"type:json"`
		FieldPasswordData FieldPasswordData `gorm:"type:json"`
		FieldNumberData   FieldNumberData   `gorm:"type:json"`
		SelectData        SelectData        `gorm:"type:json"`
		CheckboxData      CheckboxData      `gorm:"type:json"`
		DatepickerData    DatepickerData    `gorm:"type:json"`
	}
	FieldData struct {
		Text         string
		TextValidate string
	}
	FieldTextareaData struct {
		Textarea         string
		TextareaValidate string
	}
	FieldPasswordData struct {
		Password        string
		PasswordDefault string
	}
	FieldNumberData struct {
		Number         int
		NumberValidate int
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

func (c *FieldTextareaData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *FieldTextareaData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *FieldPasswordData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *FieldPasswordData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *FieldNumberData) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, c)
	}
	return nil
}

func (c *FieldNumberData) Value() (driver.Value, error) {
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
	mb.Editing().WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			if in != nil {
				in(obj, ctx)
			}
			p := obj.(*DemoCase)
			if p.ID == 0 {
				return
			}
			if len(p.FieldTextareaData.TextareaValidate) < 10 {
				err.FieldError("FieldTextareaSection.FieldTextareaData.TextareaValidate", "input more than 10 chars")
			}
			if len(p.FieldPasswordData.Password) < 5 {
				err.FieldError("FieldPasswordSection.FieldPasswordData.Password", "password more than 5 chars")
			}
			if p.FieldNumberData.NumberValidate <= 0 {
				err.FieldError("FieldNumberSection.FieldNumberData.NumberValidate", "input greater than 0")
			}
			if len(p.SelectData.AutoComplete) == 1 {
				err.FieldError("SelectSection.SelectData.editField", "select more than 1 item")
			}
			if p.SelectData.NormalSelect == 8 {
				err.FieldError("SelectSection.SelectData.NormalSelect", "can`t select Trevor")
			}
			if p.DatepickerData.Date == 0 {
				err.FieldError("DatepickerSection.DatepickerData.Date", "Date is required")
			}
			if p.DatepickerData.DateTime == 0 {
				err.FieldError("DatepickerSection.DatepickerData.DateTime", "DateTime is required")
			}
			if p.DatepickerData.DateRange == nil || p.DatepickerData.DateRange[1] < p.DatepickerData.DateRange[0] {
				err.FieldError("DatepickerSection.DatepickerData.DateRange", "End later than Start")
			}

			return
		}
	})
	mb.Listing("ID", "Name")
	detailing := mb.Detailing(
		"FieldSection",
		"FieldTextareaSection",
		"FieldPasswordSection",
		"FieldNumberSection",
		"SelectSection",
		"CheckboxSection",
		"DatepickerSection",
		"DialogSection",
		"AvatarSection",
	)
	editing := mb.Editing(
		"Name",
		"FieldSection",
		"FieldTextareaSection",
		"FieldPasswordSection",
		"FieldNumberSection",
		"SelectSection",
		"CheckboxSection",
		"DatepickerSection",
	)

	configVxField(detailing, editing, mb)
	configVxFieldArea(detailing, editing, mb)
	configVxFieldPassword(detailing, editing, mb)
	configVxFieldNumber(detailing, editing, mb)
	configVxSelect(detailing, editing, mb)
	configVxCheckBox(detailing, editing, mb)
	configVxDatepicker(detailing, editing, mb)
	configVxDialog(detailing, mb)
	configVxAvatar(detailing, mb)
	return
}

func DemoCaseTextField(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors) *vx.VXFieldBuilder {
	formKey := fmt.Sprintf("%s.%s", editField, field)
	return vx.VXField().
		Label(label).
		Attr(web.VField(formKey, reflectutils.MustGet(obj, formKey))...).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}

func DemoCaseSelect(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors, items interface{}) *vx.VXSelectBuilder {
	formKey := fmt.Sprintf("%s.%s", editField, field)
	return vx.VXSelect().
		Label(label).
		Items(items).
		ItemTitle("Name").
		ItemValue("ID").
		Attr(web.VField(formKey, reflectutils.MustGet(obj, formKey))...).
		ErrorMessages(vErr.GetFieldErrors(formKey)...)
}

func DemoCaseCheckBox(obj interface{}, section, editField, field, label string) *vx.VXCheckboxBuilder {
	formKey := fmt.Sprintf("%s.%s", editField, field)
	return vx.VXCheckbox().
		Label(label).
		Attr(web.VField(formKey, reflectutils.MustGet(obj, formKey))...)
}

func configVxField(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
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
				cardRows("", 2,
					vx.VXField().
						Label("Text(Disabled)").
						ModelValue("This is Disabled Vx-Field").
						Disabled(true),
					vx.VXField().
						Label("Text(Readonly)").
						ModelValue("This is Readonly Vx-Field").
						Readonly(true),
					DemoCaseTextField(obj, sectionName, editField, "Text", "Text", vErr).
						Tips("This is Tips").Clearable(true),
					DemoCaseTextField(obj, sectionName, editField, "TextValidate", "TextValidate(input more than 5 chars)", vErr).Required(true).Clearable(true),
				),
			)
		}).
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				p := obj.(*DemoCase)
				if len(p.FieldData.TextValidate) < 5 {
					err.FieldError(fmt.Sprintf("%s.%s.TextValidate", sectionName, editField), "input more than 5 chars")
				}
				return
			}
		})
	detailing.Section(section)
	editing.Section(section.Clone())
}

func configVxFieldArea(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
	sectionName := "FieldTextareaSection"
	editField := "FieldTextareaData"
	label := "vx-field(type textarea)"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				cardRows("", 2,
					vx.VXField().
						Label("Textarea(Disabled)").
						ModelValue("This is Disabled Vx-Field type Textarea").
						Disabled(true).
						Type("textarea"),
					vx.VXField().
						Label("Textarea(Readonly)").
						ModelValue("This is Readonly Vx-Field type Textarea").
						Readonly(true).
						Type("textarea"),
					DemoCaseTextField(obj, sectionName, editField, "Textarea", "Textarea", vErr).
						Tips("This is Textarea Tips").
						Type("textarea").Clearable(true),
					DemoCaseTextField(obj, sectionName, editField, "TextareaValidate", "TextareaValidate(input more than 10 chars)", vErr).Required(true).
						Type("textarea").Clearable(true),
				),
			)
		})
	detailing.Section(section)
	editing.Section(section.Clone())
}

func configVxFieldPassword(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
	sectionName := "FieldPasswordSection"
	editField := "FieldPasswordData"
	label := "vx-field(type password)"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				cardRows("", 2,
					vx.VXField().
						Label("Password(Readonly)").
						ModelValue("This is Readonly Vx-Field type Password").
						Readonly(true).
						Type("password"),
					vx.VXField().
						Label("Password(Disabled)").
						ModelValue("This is Disabled Vx-Field type Password").
						Disabled(true).
						Type("password"),
					DemoCaseTextField(obj, sectionName, editField, "Password", "Password(More Than 5 chars)", vErr).
						Tips("Password tips").
						Type("password").
						Clearable(true).
						PasswordVisibleToggle(true).
						ErrorMessages(vErr.GetFieldErrors("FieldPasswordSection.FieldPasswordData.Password")...),
					DemoCaseTextField(obj, sectionName, editField, "PasswordDefault", "PasswordDefault", vErr).
						Tips("PasswordDefault tips").
						Clearable(true).
						Type("password").
						PasswordVisibleDefault(true).
						PasswordVisibleToggle(true),
				),
			)
		})
	detailing.Section(section)
	editing.Section(section.Clone())
}

func configVxFieldNumber(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
	sectionName := "FieldNumberSection"
	editField := "FieldNumberData"
	label := "vx-field(type number)"
	section := generateSection(detailing, mb, sectionName, editField, label).
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return h.Components(
				cardRows("Number", 2,
					vx.VXField().
						Label("Number(Disabled)").
						ModelValue("This is Disabled Vx-Field type Number").
						Disabled(true),
					vx.VXField().
						Label("Number(Readonly)").
						ModelValue("This is Readonly Vx-Field type Number").
						Readonly(true),
					DemoCaseTextField(obj, sectionName, editField, "Number", "Number", vErr).
						Tips("Number tips").
						Clearable(true).
						Type("number"),
					DemoCaseTextField(obj, sectionName, editField, "NumberValidate", "NumberValidate( > 0)", vErr).
						Tips("NumberValidate tips").
						Clearable(true).
						Type("number").
						ErrorMessages(vErr.GetFieldErrors("FieldNumberSection.FieldNumberData.NumberValidate")...),
				),
			)
		})
	detailing.Section(section)
	editing.Section(section.Clone())
}

func configVxSelect(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
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
							Type("autocomplete").Multiple(true).Chips(true).ClosableChips(true).Clearable(true).
							ErrorMessages(vErr.GetFieldErrors("SelectSection.SelectData.AutoComplete")...),
					),
				),
				v.VRow(
					v.VCol(
						DemoCaseSelect(obj, sectionName, editField, "NormalSelect", "", vErr, items).
							Attr(":rules", `[(value) => value !== 8 || "can't select Trevor"]`).
							Type("autocomplete").
							ErrorMessages(vErr.GetFieldErrors("SelectSection.SelectData.NormalSelect")...),
					),
				),
			)
		})
	detailing.Section(section)
	editing.Section(section.Clone())
}

func configVxCheckBox(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
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
	editing.Section(section.Clone())
}

func cardRows(title string, splitCols int, comp ...h.HTMLComponent) *v.VCardBuilder {
	var (
		rows   []h.HTMLComponent
		result int
		row    = v.VRow()
	)

	for i, c := range comp {
		if i/splitCols == result {
			if i%splitCols == 0 {
				row = v.VRow()
				rows = append(rows, row)
			} else if i%splitCols == splitCols-1 {
				result++
			}
			row.AppendChildren(v.VCol(c).Class("text-center"))
		}
	}
	return v.VCard(
		v.VCardItem(
			rows...,
		),
	).Title(title).Class("pa-2 my-4")
}

func configVxDialog(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	label := "vx-dialog"
	sectionName := "DialogSection"
	section := presets.NewSectionBuilder(mb, sectionName).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		text := "This is an info description line This is an info description lineThis is an info description lineThis is an info description lineThis is an info description line"
		textLarge := text
		for i := 0; i < 30; i++ {
			textLarge += text
		}
		return web.Scope(
			h.Div(
				h.Div(
					h.H2(label).Class("section-title"),
				).Class("section-title-wrap"),
				cardRows("Activator", 5,
					h.Components(h.Div(h.Text("v-model")).Class("mb-2"),
						v.VBtn("Open Dialog").Color(v.ColorPrimary).
							Attr("@click", "locals.dialogVisible=true"),
						vx.VXDialog().
							Attr("v-model", "locals.dialogVisible").
							Title("ModelValue").
							Text(text),
					),
					dialogActivator("Open Dialog", "Activator Slot", text, v.ColorSecondary).Title("Conform"),
				),

				cardRows("Type", 5,
					dialogActivator("Open Dialog", "Default", text, v.ColorSecondary).
						Title("Default"),
					dialogActivator("Open Dialog", "Info", text, v.ColorInfo).Type("info").
						Title("Info"),
					dialogActivator("Open Dialog", "Success", text, v.ColorSuccess).Type("success").
						Title("Success"),
					dialogActivator("Open Dialog", "Warning", text, v.ColorWarning).Type("warn").
						Title("Warning"),
					dialogActivator("Open Dialog", "Error", text, v.ColorError).Type("error").
						Title("Error"),
				),
				cardRows("Size", 5,
					dialogActivator("Open Dialog", "Default", text, v.ColorSecondary).Title("Confirm"),
					dialogActivator("Open Dialog", "Large", text, v.ColorSecondary).Title("Confirm").Size("large"),
					dialogActivator("Open Dialog", "Custom Width", text, v.ColorSecondary).Title("Confirm").Width(1200),
					dialogActivator("Open Dialog", "Content Scroll Bar(Custom Height)", textLarge, v.ColorSecondary).Size("large").Title("Confirm").Height(400),
					dialogActivator("Open Dialog", "Content Scroll Bar", textLarge, v.ColorSecondary).Title("Confirm").Width(300),
				),
				cardRows("Button&Event", 2,
					dialogActivator("Open Dialog", "Custom Event", text, v.ColorSecondary).Title("Confirm").
						Attr("@click:ok", presets.ShowSnackbarScript("click ok", v.ColorSuccess)).
						Attr("@click:cancel", presets.ShowSnackbarScript("click cancel", v.ColorWarning)),
					dialogActivator("Open Dialog", "Custom Button Text", text, v.ColorSecondary).Title("Confirm").
						CancelText("取消").OkText("确定"),
				),
				cardRows("Hide&Show", 5,
					dialogActivator("Open Dialog", "HideCancel", text, v.ColorSecondary).Title("Confirm").
						HideCancel(true),
					dialogActivator("Open Dialog", "HideOk", text, v.ColorSecondary).Title("Confirm").
						HideOk(true),
					dialogActivator("Open Dialog", "HideClose", text, v.ColorSecondary).Title("Confirm").
						HideClose(true),
					dialogActivator("Open Dialog", "HideFooter", text, v.ColorSecondary).Title("Confirm").
						HideFooter(true),
					dialogActivator("Open Dialog", "DisableOk", text, v.ColorSecondary).Title("Confirm").
						DisableOk(true),
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
	formKey := fmt.Sprintf("%s.%s", editField, field)
	val := reflectutils.MustGet(obj, formKey)
	return vx.VXDatepicker().
		Clearable(true).
		Label(label).
		Attr(web.VField(formKey, val)...).
		Placeholder(field).
		ErrorMessages(vErr.GetFieldErrors(fmt.Sprintf("%s.%s", section, formKey))...)
}

func DemoCaseRangePicker(obj interface{}, section, editField, field, label string, vErr web.ValidationErrors) *vx.VXRangePickerBuilder {
	formKey := fmt.Sprintf("%s.%s", editField, field)
	val := reflectutils.MustGet(obj, formKey)
	return vx.VXRangePicker().
		Clearable(true).
		Label(label).
		Attr(web.VField(formKey, val)...).
		ErrorMessages(vErr.GetFieldErrors(fmt.Sprintf("%s.%s", section, formKey))...)
}

func configVxDatepicker(detailing *presets.DetailingBuilder, editing *presets.EditingBuilder, mb *presets.ModelBuilder) {
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
							DatePickerProps(map[string]string{
								"min": time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
								"max": time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
							}),
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
	editing.Section(section.Clone())
}

func dialogActivator(btn, label, text, color string) *vx.VXDialogBuilder {
	return vx.VXDialog(
		web.Slot(
			h.Div(h.Text(label)).Class("mb-2"),
			v.VBtn(btn).Color(color).Attr("v-bind", "activatorProps"),
		).Name("activator").Scope("{props: { activatorProps }}"),
	).Text(text)
}

func avatarView[T comparable](sizes []T, show func(T) string) (comps []h.HTMLComponent) {
	for _, size := range sizes {
		comps = append(comps, h.Components(h.Div(h.Text(show(size))).Class("mb-2"), vx.VXAvatar().Name("ShaoXing").Size(fmt.Sprint(size))))
	}
	return
}

func configVxAvatar(detailing *presets.DetailingBuilder, mb *presets.ModelBuilder) {
	label := "vx-avatar"
	sectionName := "AvatarSection"
	section := presets.NewSectionBuilder(mb, sectionName).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return web.Scope(
			h.Div(
				h.Div(
					h.H2(label).Class("section-title"),
				).Class("section-title-wrap"),
				cardRows("Size", 5,
					avatarView([]string{v.SizeXSmall, v.SizeSmall, v.SizeDefault, v.SizeLarge, v.SizeXLarge}, func(s string) string {
						return s
					})...,
				),
				cardRows("Custom Size", 5,
					avatarView([]int{16, 24, 32, 40, 48, 64, 80, 96, 128, 160}, func(s int) string {
						return fmt.Sprintf("%vpx", s)
					})...,
				),
			).Class("section-wrap with-border-b"),
		).VSlot("{locals}").Init("{dialogVisible:false}")
	})
	detailing.Section(section)
}
