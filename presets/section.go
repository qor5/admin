package presets

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

const (
	SectionIsCancel = "isCancel"

	sectionListUnsavedKey        = "sectionListUnsaved"
	sectionListEditBtnKey        = "sectionListEditBtn"
	sectionListSaveBtnKey        = "sectionListSaveBtn"
	sectionListFieldDeleteBtnKey = "sectionListDeleteBtn"

	sectionListFieldEditing = "sectionListEditing"

	DisabledKeyButtonSave = "disabledKeyButtonSave"
)

func NewSectionBuilder(mb *ModelBuilder, name string) (r *SectionBuilder) {
	r = &SectionBuilder{
		NameLabel: NameLabel{
			name:  name,
			label: name,
		},
		mb: mb,
		saver: func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			return mb.editing.Saver(obj, id, ctx)
		},
		validator: func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			return
		},
		componentViewFunc: nil,
		componentEditFunc: nil,
		saveBtnFunc: func(obj interface{}, ctx *web.EventContext) bool {
			return true
		},
		elementViewFunc: nil,
		elementEditFunc: nil,
		componentEditBtnFunc: func(obj interface{}, ctx *web.EventContext) bool {
			return true
		},
		componentHoverFunc: func(obj interface{}, ctx *web.EventContext) bool {
			return true
		},
		disableElementDeleteBtn: false,
		disableElementCreateBtn: false,
		elementEditBtnFunc: func(obj interface{}, ctx *web.EventContext) bool {
			return true
		},
		elementEditBtn: true,
		elementHoverFunc: func(obj interface{}, ctx *web.EventContext) bool {
			return true
		},
		elementHover:        true,
		alwaysShowListLabel: false,
		isList:              false,
		disableLabel:        false,
	}
	r.editingFB.Model(mb.model)
	r.editingFB.defaults = mb.writeFields.defaults
	r.viewingFB.Model(mb.model)
	r.viewingFB.defaults = mb.p.detailFieldDefaults
	// r.setter = r.defaultUnmarshalFunc
	// d.Field(name).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	// 	panic("you must set ViewComponentFunc and EditComponentFunc if you want to use SectionsBuilder")
	// })

	return
}

// func (b *SectionsBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
// 	return b.comp
// }

// SectionBuilder is a builder for a section in the detail page.
// save: 	   fetcher => setter => saver
// show, edit: fetcher => setter
type SectionBuilder struct {
	NameLabel
	mb           *ModelBuilder
	isUsed       bool
	isRegistered bool
	isEdit       bool
	comp         FieldComponentFunc
	saver        SaveFunc
	setter       func(obj interface{}, ctx *web.EventContext) error
	// validate object in save section event
	validator         ValidateFunc
	hiddenFuncs       []ObjectComponentFunc
	componentViewFunc FieldComponentFunc
	componentEditFunc FieldComponentFunc
	saveBtnFunc       ObjectBoolFunc
	// father            *SectionsBuilder

	isList       bool
	disableLabel bool
	// Only when isList is false, the following param will take effect
	// control Delete button in the show component
	componentEditBtnFunc ObjectBoolFunc
	// control Hover in the show component
	componentHoverFunc ObjectBoolFunc

	// Only when isList is true, the following param will take effect
	// Disable Delete button in edit element
	disableElementDeleteBtn bool
	// Disable Create button in element list
	disableElementCreateBtn bool
	// Disable Edit button in show element
	elementEditBtnFunc ObjectBoolFunc
	// This is the return value of elementEditBtnFunc
	elementEditBtn bool
	// Disable Hover in show element
	elementHoverFunc ObjectBoolFunc
	// This is the return value of elementHoverFunc
	elementHover bool
	// By default, the title will only be displayed if the list is not empty.
	// If alwaysShowListLabel is true, the label will show anyway
	alwaysShowListLabel bool
	elementViewFunc     FieldComponentFunc
	elementEditFunc     FieldComponentFunc
	elementUnmarshaler  func(toObj, formObj any, prefix string, ctx *web.EventContext) error

	editingFB FieldsBuilder
	viewingFB FieldsBuilder
}

type ObjectBoolFunc func(obj interface{}, ctx *web.EventContext) bool

func (s *SectionBuilder) Clone() *SectionBuilder {
	newSection := *s
	newSection.isUsed = false

	if s.hiddenFuncs != nil {
		newSection.hiddenFuncs = make([]ObjectComponentFunc, len(s.hiddenFuncs))
		copy(newSection.hiddenFuncs, s.hiddenFuncs)
	}

	return &newSection
}

func (b *SectionBuilder) FieldComponent(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return b.comp(obj, field, ctx)
}

func (b *SectionBuilder) WrapValidator(w func(in ValidateFunc) ValidateFunc) (r *SectionBuilder) {
	b.validator = w(b.validator)
	return b
}

func (b *SectionBuilder) ComponentEditBtnFunc(v ObjectBoolFunc) *SectionBuilder {
	if v == nil {
		panic("value required")
	}
	b.componentEditBtnFunc = v
	return b
}

func (b *SectionBuilder) WrapComponentEditBtnFunc(w func(in ObjectBoolFunc) ObjectBoolFunc) (r *SectionBuilder) {
	b.componentEditBtnFunc = w(b.componentEditBtnFunc)
	return b
}

func (b *SectionBuilder) ComponentHoverFunc(v ObjectBoolFunc) *SectionBuilder {
	if v == nil {
		panic("value required")
	}
	b.componentHoverFunc = v
	return b
}

func (b *SectionBuilder) WrapComponentHoverFunc(w func(in ObjectBoolFunc) ObjectBoolFunc) (r *SectionBuilder) {
	b.componentHoverFunc = w(b.componentHoverFunc)
	return b
}

func (b *SectionBuilder) WrapSaveBtnFunc(w func(in ObjectBoolFunc) ObjectBoolFunc) (r *SectionBuilder) {
	b.saveBtnFunc = w(b.componentHoverFunc)
	return b
}

func (b *SectionBuilder) DisableElementDeleteBtn() *SectionBuilder {
	b.disableElementDeleteBtn = true
	return b
}

func (b *SectionBuilder) DisableElementCreateBtn() *SectionBuilder {
	b.disableElementCreateBtn = true
	return b
}

func (b *SectionBuilder) ElementEditBtnFunc(v ObjectBoolFunc) *SectionBuilder {
	if v == nil {
		panic("value required")
	}
	b.elementEditBtnFunc = v
	return b
}

func (b *SectionBuilder) WrapElementEditBtnFunc(w func(in ObjectBoolFunc) ObjectBoolFunc) (r *SectionBuilder) {
	b.elementEditBtnFunc = w(b.elementEditBtnFunc)
	return b
}

func (b *SectionBuilder) ElementHoverFunc(v ObjectBoolFunc) *SectionBuilder {
	if v == nil {
		panic("value required")
	}
	b.elementHoverFunc = v
	return b
}

func (b *SectionBuilder) WrapElementHoverFunc(w func(in ObjectBoolFunc) ObjectBoolFunc) (r *SectionBuilder) {
	b.elementHoverFunc = w(b.elementHoverFunc)
	return b
}

func (b *SectionBuilder) AlwaysShowListLabel() *SectionBuilder {
	b.alwaysShowListLabel = true
	return b
}

func (b *SectionBuilder) IsList(v interface{}) (r *SectionBuilder) {
	if b.mb == nil {
		panic("model must be provided")
	}
	rt := reflectutils.GetType(b.mb.model, b.name)
	if rt.Kind() != reflect.Slice {
		panic("field kind must be slice")
	}
	if reflect.TypeOf(reflect.New(rt.Elem()).Elem().Interface()) != reflect.TypeOf(v) {
		panic(fmt.Sprintf("%s not equal to %s", reflect.New(rt.Elem()).Elem().Interface(), reflect.TypeOf(v)))
	}

	r = b
	r.editingFB.Model(v)
	r.isList = true
	r.elementUnmarshaler = r.DefaultElementUnmarshal()

	return
}

// Editing default saver only save these field
func (b *SectionBuilder) Editing(fields ...interface{}) (r *SectionBuilder) {
	r = b
	b.editingFB = *b.editingFB.Only(fields...)
	if b.componentEditFunc == nil {
		b.EditComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return b.editingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, "", ctx)
		})
	}
	if b.isList {
		if b.elementEditFunc == nil {
			b.ElementEditComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
				return b.editingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, field.FormKey, ctx)
			})
		}
	}
	b.Viewing(fields...)
	return
}

func (b *SectionBuilder) Viewing(fields ...interface{}) (r *SectionBuilder) {
	r = b
	b.viewingFB = *b.viewingFB.Only(fields...)
	if b.componentViewFunc == nil {
		b.ViewComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return b.viewingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, "", ctx)
		})
	}
	if b.isList {
		if b.elementViewFunc == nil {
			b.ElementShowComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
				return b.viewingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, "", ctx)
			})
		}
	}
	return
}

func (b *SectionBuilder) EditingField(name string) (r *FieldBuilder) {
	return b.editingFB.Field(name)
}

func (b *SectionBuilder) ViewingField(name string) (r *FieldBuilder) {
	return b.viewingFB.Field(name)
}

func (b *SectionBuilder) SaveFunc(v SaveFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.saver = v
	return b
}

func (b *SectionBuilder) WrapSaveFunc(w func(in SaveFunc) SaveFunc) (r *SectionBuilder) {
	b.saver = w(b.saver)
	return b
}

func (b *SectionBuilder) WrapSetterFunc(w func(in func(obj interface{}, ctx *web.EventContext) error) func(obj interface{}, ctx *web.EventContext) error) (r *SectionBuilder) {
	b.setter = w(b.setter)
	return b
}

func (b *SectionBuilder) SetterFunc(v func(obj interface{}, ctx *web.EventContext) error) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.setter = v
	return b
}

func (b *SectionBuilder) HiddenFuncs(funcs ...ObjectComponentFunc) (r *SectionBuilder) {
	for _, f := range funcs {
		if f == nil {
			panic("value required")
		}
		b.hiddenFuncs = append(b.hiddenFuncs, f)
	}
	return b
}

func (b *SectionBuilder) ElementUnmarshalFunc(v func(toObj, formObj any, prefix string, ctx *web.EventContext) error) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.elementUnmarshaler = v
	return b
}

func (b *SectionBuilder) ViewComponentFunc(v FieldComponentFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.componentViewFunc = v
	if b.componentEditFunc != nil {
		b.ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return web.Scope(
				web.Portal(
					b.viewComponent(obj, field, ctx),
				).Name(b.FieldPortalName()),
			).VSlot("{ form, dash }").DashInit("{errorMessages:{},disabled:{}}")
		})
	}
	return b
}

func (b *SectionBuilder) EditComponentFunc(v FieldComponentFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.componentEditFunc = v
	if b.componentViewFunc != nil {
		b.ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return web.Scope(
				web.Portal(
					b.viewComponent(obj, field, ctx),
				).Name(b.FieldPortalName()),
			).VSlot("{ form, dash }").DashInit("{errorMessages:{},disabled:{}}")
		})
	}
	return b
}

func (b *SectionBuilder) Label(label string) (r *SectionBuilder) {
	b.label = label
	return b
}

func (b *SectionBuilder) DisableLabel() (r *SectionBuilder) {
	b.disableLabel = true
	return b
}

func (b *SectionBuilder) ElementShowComponentFunc(v FieldComponentFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.elementViewFunc = v
	if b.elementEditFunc != nil {
		b.ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return web.Scope(
				web.Portal(
					b.listComponent(obj, ctx, -1, -1, -1, false, false),
				).Name(b.FieldPortalName()),
			).VSlot("{ locals, dash }").DashInit("{errorMessages:{},disabled:{}}\n").Init("{show:true}")
		})
	}
	return b
}

func (b *SectionBuilder) ElementEditComponentFunc(v FieldComponentFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.elementEditFunc = v
	if b.elementViewFunc != nil {
		b.ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return web.Scope(
				web.Portal(
					b.listComponent(obj, ctx, -1, -1, -1, false, false),
				).Name(b.FieldPortalName()),
			).VSlot("{ locals, dash }").DashInit("{errorMessages:{},disabled:{}}\n").Init("{show:true}")
		})
	}
	return b
}

// ComponentFunc set FieldBuilder compFunc
func (b *SectionBuilder) ComponentFunc(v FieldComponentFunc) *SectionBuilder {
	b.comp = v
	return b
}

func (b *SectionBuilder) ListFieldPrefix(index int) string {
	return fmt.Sprintf("%s[%d]", b.name, index)
}

func (b *SectionBuilder) registerEvent() {
	if b.isRegistered {
		return
	}
	b.isRegistered = true
	b.mb.RegisterEventFunc(b.EventSave(), b.SaveDetailField)
	b.mb.RegisterEventFunc(b.EventEdit(), b.EditDetailField)
	b.mb.RegisterEventFunc(b.EventValidate(), b.ValidateDetailField)
	b.mb.RegisterEventFunc(b.EventDelete(), b.DeleteDetailListField)
	b.mb.RegisterEventFunc(b.EventCreate(), b.CreateDetailListField)
	b.mb.RegisterEventFunc(b.EventReload(), b.ReloadDetailField)
}

func (b *SectionBuilder) EventEdit() string {
	return fmt.Sprintf("section_edit_%s", b.name)
}

func (b *SectionBuilder) EventSave() string {
	return fmt.Sprintf("section_save_%s", b.name)
}

func (b *SectionBuilder) EventValidate() string {
	return fmt.Sprintf("section_validate_%s", b.name)
}

func (b *SectionBuilder) EventDelete() string {
	return fmt.Sprintf("section_delete_%s", b.name)
}

func (b *SectionBuilder) EventCreate() string {
	return fmt.Sprintf("section_create_%s", b.name)
}

func (b *SectionBuilder) EventReload() string {
	return fmt.Sprintf("section_reload_%s", b.name)
}

func (b *SectionBuilder) viewComponent(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	id := b.getObjectID(ctx, obj)
	initDataChanged := fmt.Sprintf("if (vars.%s ){vars.%s.section_%s=false};", VarsPresetsDataChanged, VarsPresetsDataChanged, b.name)

	disableEditBtn := b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	btn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Edit")).Variant(VariantFlat).Size(SizeXSmall).
		PrependIcon("mdi-pencil-outline").
		Attr("v-show", fmt.Sprintf("%t&&%t", b.componentEditBtnFunc(obj, ctx), !disableEditBtn)).
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(b.EventEdit()).
			// Query(SectionFieldName, b.name).
			Query(ParamID, id).
			Go())

	hiddenComp := h.Div()
	if len(b.hiddenFuncs) > 0 {
		for _, f := range b.hiddenFuncs {
			hiddenComp.AppendChildren(f(obj, ctx))
		}
	}
	content := h.Div().Class("section-wrap with-border-b").ClassIf("can-edit", b.componentEditBtnFunc(obj, ctx) && !disableEditBtn)
	if b.label != "" {
		lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.label)
		content.AppendChildren(
			h.Div(
				h.If(!b.disableLabel, h.H2(lb).Class("section-title")),
			).Class("section-title-wrap"),
		)
	}

	showComponent := b.componentViewFunc(obj, field, ctx)
	if showComponent != nil {
		content.AppendChildren(
			h.Div(
				VCard(
					VCardText(
						h.Div(
							h.Div(btn).Class("section-edit-area top-area").Style("z-index:2;"),
							h.Div(showComponent).
								Class("flex-grow-1"),
						).Class("section-content"),
					),
				).Variant(VariantFlat),
			).Class("section-body"),
		)
	}

	return h.Div(
		content,
		hiddenComp,
	).Attr("v-on-mounted", fmt.Sprintf(`()=>{%s}`, initDataChanged)).Class("mb-10")
}

func (b *SectionBuilder) editComponent(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	id := b.getObjectID(ctx, obj)

	onChangeEvent := fmt.Sprintf("if (vars.%s ){ vars.%s.section_%s=true };", VarsPresetsDataChanged, VarsPresetsDataChanged, b.name)

	cancelBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Cancel")).Size(SizeSmall).Variant(VariantFlat).Color(ColorGreyLighten3).
		Attr("style", "text-transform: none;").
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(b.EventSave()).
			// Query(SectionFieldName, b.name).
			Query(ParamID, id).
			Query(SectionIsCancel, true).
			Go())

	disableEditBtn := b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	saveBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Save")).PrependIcon("mdi-check").
		Size(SizeSmall).Variant(VariantFlat).Color(ColorPrimary).
		Attr(":disabled", fmt.Sprintf("%v || dash.disabled.%v", disableEditBtn, DisabledKeyButtonSave)).
		Attr("style", "text-transform: none;").
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(b.EventSave()).
			// Query(SectionFieldName, b.name).
			Query(ParamID, id).
			Go())

	hiddenComp := h.Div()
	if len(b.hiddenFuncs) > 0 {
		for _, f := range b.hiddenFuncs {
			hiddenComp.AppendChildren(f(obj, ctx))
		}
	}

	content := h.Div().Class("section-wrap edit-view with-border-b")

	if b.label != "" && !b.disableLabel {
		lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.label)
		content.AppendChildren(
			h.Div(
				h.H2(lb).Class("section-title"),
			).Class("section-title-wrap"),
		)
	}

	disableSaveBtn := !b.saveBtnFunc(obj, ctx)
	if b.componentEditFunc != nil {
		content.AppendChildren(
			h.Div(
				VCard(
					VCardText(
						h.Div(
							// detailFields
							h.Div(b.componentEditFunc(obj, field, ctx)).
								Class("flex-grow-1 pb-6"),
							h.If(
								!disableSaveBtn,
								h.Div(
									cancelBtn,
									saveBtn.Class("ml-2"),
								).Class("section-edit-area bottom-area"),
							),
						).Class("d-flex flex-column"),
					),
				).Variant(VariantOutlined),
			).Class("section-body"),
		)
	}
	operateID := fmt.Sprint(time.Now().UnixNano())
	onChangeEvent += checkFormChangeScript + setValidateKeysScript +
		web.Plaid().URL(ctx.R.URL.Path).
			BeforeScript(fmt.Sprintf(`dash.__ValidateOperateID=%q`, operateID)).
			EventFunc(b.EventValidate()).
			Query(ParamID, id).
			Query(ParamOperateID, operateID).
			Go()

	comps := h.Components(
		web.Listen(b.mb.NotifModelsSectionValidate(b.name),
			setFieldErrorsScript))
	if b.isEdit {
		return h.Div(
			web.Scope(
				comps,
				content,
				hiddenComp,
			).OnChange(onChangeEvent).UseDebounce(500),
		)
	}
	return h.Div(
		web.Scope(
			comps,
			content,
			hiddenComp,
		).VSlot("{ form }").OnChange(onChangeEvent).UseDebounce(500),
	).Class("mb-10")
}

func (b *SectionBuilder) defaultUnmarshalFunc(obj interface{}, ctx *web.EventContext) (err error) {
	if tf := reflect.TypeOf(obj).Kind(); tf != reflect.Ptr {
		return fmt.Errorf("model %#+v must be pointer", obj)
	}
	formObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()

	if err = b.DefaultElementUnmarshal()(obj, formObj, b.name, ctx); err != nil {
		return
	}
	return
}

func (b *SectionBuilder) buildElementRows(list interface{}, deletedID, editID, saveID, listLen int, unsaved, isReload bool, ctx *web.EventContext) *h.HTMLTagBuilder {
	rows := h.Div()
	if b.alwaysShowListLabel && !b.disableLabel {
		lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.label)
		label := h.Div(h.H2(lb).Class("section-title")).Class("section-title-wrap")
		rows.AppendChildren(label)
	}

	if list != nil {
		i := 0
		reflectutils.ForEach(list, func(elementObj interface{}) {
			defer func() { i++ }()
			if i == 0 && b.label != "" && !b.alwaysShowListLabel && !b.disableLabel {
				lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.label)
				label := h.Div(h.H2(lb).Class("section-title")).Class("section-title-wrap")
				rows.AppendChildren(label)
			}
			// set fieldSetting to ctx.R.Form by sortIndex
			sortIndex := i
			// find fieldSetting from ctx.R.Form by fromIndex
			fromIndex := i
			if deletedID != -1 && i >= deletedID {
				// if last event is click deleteBtn, fromIndex should add one
				fromIndex++
			}
			if editID == sortIndex {
				// if click edit
				rows.AppendChildren(b.editElement(elementObj, sortIndex, unsaved && i == listLen-1, unsaved, ctx))
			} else if saveID == sortIndex || isReload {
				// if click save
				rows.AppendChildren(b.showElement(elementObj, sortIndex, unsaved, ctx))
			} else {
				// default
				isEditing := ctx.R.FormValue(b.ListElementIsEditing(fromIndex)) != ""
				if !isEditing {
					rows.AppendChildren(b.showElement(elementObj, sortIndex, unsaved, ctx))
				} else {
					formObj := reflect.New(reflect.TypeOf(b.editingFB.model).Elem()).Interface()
					err := b.elementUnmarshaler(elementObj, formObj, b.ListFieldPrefix(fromIndex), ctx)
					if err != nil {
						panic(err)
					}
					rows.AppendChildren(b.editElement(elementObj, sortIndex, unsaved && i == listLen-1, unsaved, ctx))
				}
			}
		})
	}

	return rows
}

func (b *SectionBuilder) listComponent(obj interface{}, ctx *web.EventContext, deletedID, editID, saveID int, unsaved, isReload bool) h.HTMLComponent {
	if b.elementHoverFunc != nil {
		b.elementHover = b.elementHoverFunc(obj, ctx)
	}
	if b.elementEditBtnFunc != nil {
		b.elementEditBtn = b.elementEditBtnFunc(obj, ctx)
	}
	if b.elementEditBtn {
		b.elementEditBtn = b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() == nil
	}

	id := b.getObjectID(ctx, obj)

	list, err := reflectutils.Get(obj, b.name)
	if err != nil {
		panic(err)
	}
	listLen := 0
	if list != nil {
		listLen = reflect.ValueOf(list).Len()
	}

	rows := b.buildElementRows(list, deletedID, editID, saveID, listLen, unsaved, isReload, ctx)

	disableCreateBtn := b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	// Respect caller-provided unsaved to avoid drifting back to request param
	// Only hide "Add Item" button when there's an unsaved new item (unsaved=true)
	disableCreateBtn = !isReload && (disableCreateBtn || unsaved)
	if !b.disableElementCreateBtn && !disableCreateBtn {
		addBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "AddRow")).PrependIcon("mdi-plus-circle").Color("primary").Variant(VariantText).
			Class("mb-2 ml-4").
			Attr("@click", "locals.show=false;"+web.Plaid().
				URL(ctx.R.URL.Path).
				EventFunc(b.EventCreate()).
				Query(b.elementUnsavedKey(), true).
				// Query(SectionFieldName, b.name).
				Query(ParamID, id).
				Go())
		rows.AppendChildren(addBtn)
	}

	hiddenComp := h.Div()
	if len(b.hiddenFuncs) > 0 {
		for _, f := range b.hiddenFuncs {
			hiddenComp.AppendChildren(f(obj, ctx))
		}
	}

	// TODO editing list section
	// if b.isEdit {
	// 	return h.Div(
	// 		// element and addBtn have mb-2, so the real effect is mb-6
	// 		h.Div(rows).Class("mb-4"),
	// 		hiddenComp,
	// 	)
	// }
	return h.Div(
		web.Scope(
			// element and addBtn have mb-2, so the real effect is mb-6
			h.Div(rows).Class("mb-4"),
		).VSlot("{ form }"),
		hiddenComp,
	)
}

func (b *SectionBuilder) elementUnsavedKey() string {
	return fmt.Sprintf("%s_%s", sectionListUnsavedKey, b.name)
}

func (b *SectionBuilder) EditBtnKey() string {
	return fmt.Sprintf("%s_%s", sectionListEditBtnKey, b.name)
}

func (b *SectionBuilder) SaveBtnKey() string {
	return fmt.Sprintf("%s_%s", sectionListSaveBtnKey, b.name)
}

func (b *SectionBuilder) DeleteBtnKey() string {
	return fmt.Sprintf("%s_%s", sectionListFieldDeleteBtnKey, b.name)
}

func (b *SectionBuilder) ListElementIsEditing(index int) string {
	return fmt.Sprintf("%s_%s[%d].%s", deletedHiddenNamePrefix, b.name, index, sectionListFieldEditing)
}

func (b *SectionBuilder) ListElementPortalName(index int) string {
	return fmt.Sprintf("DetailElementPortal_%s_%d", b.name, index)
}

func (b *SectionBuilder) FieldPortalName() string {
	return fmt.Sprintf("DetailFieldPortal_%s", b.name)
}

func (b *SectionBuilder) showElement(obj any, index int, unsaved bool, ctx *web.EventContext) h.HTMLComponent {
	editBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Edit")).Variant(VariantFlat).Size(SizeXSmall).
		PrependIcon("mdi-pencil-outline").
		Attr("v-show", fmt.Sprintf("%t", b.elementEditBtn)).
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(b.EventEdit()).
			Query(b.elementUnsavedKey(), unsaved).
			// Query(SectionFieldName, b.name).
			Query(ParamID, ctx.Param(ParamID)).
			Query(b.EditBtnKey(), strconv.Itoa(index)).
			Go())

	content := b.elementViewFunc(obj, &FieldContext{
		ModelInfo: b.mb.modelInfo,
		Name:      b.name,
		FormKey:   fmt.Sprintf("%s[%d]", b.name, index),
		Label:     b.label,
	}, ctx)

	return web.Portal(
		web.Slot(
			VCard(
				VCardText(
					h.Div(
						h.Div(editBtn).Class("section-edit-area top-area"),
						h.Div(content).Class("flex-grow-1 pr-3"),
					).Class("d-flex justify-space-between section-content"),
				),
			).Class("mb-4 section-body").ClassIf("can-edit", b.elementEditBtn).
				Variant(VariantFlat).
				Attr("v-bind", "props"),
		).Name("default").Scope("{ isHovering, props }"),
	).Name(b.ListElementPortalName(index))
}

// If you have created this element but not save, isCreated = true
func (b *SectionBuilder) editElement(obj any, index int, isCreated bool, unsaved bool, ctx *web.EventContext) h.HTMLComponent {
	showAddBtn := "locals.show=true;"
	// use unsaved from caller; do not override based on ctx or isCreated

	deleteEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(b.EventDelete()).
		// Query(SectionFieldName, b.name).
		Query(ParamID, ctx.Param(ParamID)).
		Query(b.elementUnsavedKey(), unsaved && !isCreated).
		Query(b.DeleteBtnKey(), index).
		Go()
	if isCreated {
		deleteEvent = showAddBtn + deleteEvent
	}
	cancelEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(b.EventSave()).
		// Query(SectionFieldName, b.name).
		Query(b.elementUnsavedKey(), unsaved && !isCreated).
		Query(SectionIsCancel, true).
		Query(ParamID, ctx.Param(ParamID)).
		Query(b.SaveBtnKey(), strconv.Itoa(index)).
		Go()
	if isCreated {
		cancelEvent = showAddBtn + cancelEvent
		deleteEvent = cancelEvent
	}

	deleteBtn := VBtn("").Size(SizeXSmall).Variant("text").
		Rounded("0").
		Icon("mdi-delete-outline").
		Attr("v-show", fmt.Sprintf("%t", !b.disableElementDeleteBtn)).
		Attr("@click", deleteEvent)

	contentDiv := h.Div(
		h.Div(
			b.elementEditFunc(obj, &FieldContext{
				ModelInfo: b.mb.modelInfo,
				Name:      fmt.Sprintf("%s[%d]", b.name, index),
				FormKey:   fmt.Sprintf("%s[%d]", b.name, index),
				Label:     fmt.Sprintf("%s[%d]", b.label, index),
			}, ctx),
		).Class("flex-grow-1"),
		h.Div(deleteBtn).Class("d-flex pl-3"),
	).Class("d-flex justify-space-between")

	cancelBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Cancel")).Size(SizeSmall).Variant(VariantFlat).Color(ColorGreyLighten3).
		Attr("style", "text-transform: none;").
		Attr("@click", cancelEvent)

	saveEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(b.EventSave()).
		// Query(SectionFieldName, b.name).
		Query(b.elementUnsavedKey(), unsaved).
		Query(ParamID, ctx.Param(ParamID)).
		Query(b.SaveBtnKey(), strconv.Itoa(index)).
		Go()
	if isCreated {
		saveEvent = showAddBtn + saveEvent
	}
	saveBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Save")).PrependIcon("mdi-check").Size(SizeSmall).Variant(VariantFlat).Color(ColorPrimary).
		Attr("style", "text-transform: none;").
		Attr("@click", saveEvent)

	card := VCard(
		VCardText(
			h.Div(
				contentDiv,
				h.Div(
					cancelBtn,
					saveBtn.Class("ml-2"),
				).Class("section-edit-area bottom-area"),
			).Class("d-flex flex-column"),
		),
		h.Input("").Type("hidden").Attr(web.VField(b.ListElementIsEditing(index), true)...),
	).Variant(VariantOutlined).Class("mb-4 section-body")

	return web.Portal(
		h.Div(card).Class("section-wrap edit-view"),
	).Name(b.ListElementPortalName(index))
}

func (b *SectionBuilder) DefaultElementUnmarshal() func(toObj, formObj any, prefix string, ctx *web.EventContext) error {
	return func(toObj, formObj any, prefix string, ctx *web.EventContext) (err error) {
		if b.isList {
			if tf := reflect.TypeOf(toObj).Kind(); tf != reflect.Ptr {
				return fmt.Errorf("model %#+v must be pointer", toObj)
			}
			oldForm := &multipart.Form{
				Value: (map[string][]string)(http.Header(ctx.R.MultipartForm.Value).Clone()),
			}
			newForm := &multipart.Form{
				Value: make(map[string][]string),
			}
			// prefix.key => key
			for k, v := range oldForm.Value {
				if strings.HasPrefix(k, prefix+".") {
					newForm.Value[strings.TrimPrefix(k, prefix+".")] = v
				}
			}
			ctx2 := &web.EventContext{R: new(http.Request)}
			ctx2.R.MultipartForm = newForm
			ctx = ctx2
		}
		_ = ctx.UnmarshalForm(formObj)
		for _, f := range b.editingFB.fields {
			name := f.name
			info := b.mb.modelInfo
			if info != nil {
				if info.Verifier().Do(PermCreate).ObjectOn(formObj).SnakeOn("f_"+name).WithReq(ctx.R).IsAllowed() != nil && info.Verifier().Do(PermUpdate).ObjectOn(formObj).SnakeOn("f_"+name).WithReq(ctx.R).IsAllowed() != nil {
					continue
				}
			}
			if v, err := reflectutils.Get(formObj, f.name); err == nil {
				reflectutils.Set(toObj, f.name, v)
			}
			keyPath := fmt.Sprintf("%s.%s", prefix, f.name)
			err := f.lazySetterFunc()(toObj, &FieldContext{
				ModelInfo: info,
				FormKey:   keyPath,
				Name:      f.name,
				Label:     b.editingFB.getLabel(f.NameLabel),
			}, ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (b *SectionBuilder) appendElement(obj interface{}) (listLen int, err error) {
	if !b.isList {
		return
	}
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return 0, fmt.Errorf("obj %#+v must be pointer", obj)
	}
	var list any
	if list, err = reflectutils.Get(obj, b.name); err != nil {
		return
	}

	if list != nil {
		listValue := reflect.ValueOf(list)
		if listValue.Kind() != reflect.Slice {
			err = fmt.Errorf("the kind of list field is %s, not slice", listValue.Kind())
			return
		}
		listLen = listValue.Len()
	}

	return listLen + 1, reflectutils.Set(obj, b.name+"[]", b.editingFB.model)
}

func (b *SectionBuilder) removeElement(obj interface{}, index int) (err error) {
	// delete from slice
	var list any
	if list, err = reflectutils.Get(obj, b.name); err != nil {
		return
	}
	listValue := reflect.ValueOf(list)
	if listValue.Kind() != reflect.Slice {
		err = errors.New("field is not a slice")
		return
	}
	newList := reflect.MakeSlice(reflect.TypeOf(list), 0, 0)
	for i := 0; i < listValue.Len(); i++ {
		if i != int(index) {
			newList = reflect.Append(newList, listValue.Index(i))
		}
	}

	return reflectutils.Set(obj, b.name, newList.Interface())
}

// EditDetailField EventFunc: click detail field component edit button
func (b *SectionBuilder) EditDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.isList {
		return b.EditDetailListField(ctx)
	}

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.editComponent(obj, &FieldContext{
			ModelInfo: b.mb.modelInfo,
			FormKey:   b.name,
			Name:      b.name,
			Label:     b.label,
		}, ctx),
	})
	return r, nil
}

// SaveDetailField EventFunc: click save button
func (b *SectionBuilder) SaveDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.isList {
		return b.SaveDetailListField(ctx)
	}
	id := ctx.Param(ParamID)
	isCancel := ctx.ParamAsBool(SectionIsCancel)
	field := &FieldContext{
		ModelInfo: b.mb.modelInfo,
		FormKey:   b.name,
		Name:      b.name,
		Label:     b.label,
	}

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	if isCancel {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: b.FieldPortalName(),
			Body: b.viewComponent(obj, field, ctx),
		})
		return
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}
	vErrSetter := b.editingFB.Unmarshal(obj, b.mb.Info(), true, ctx)
	if vErrSetter.HaveErrors() && vErrSetter.HaveGlobalErrors() {
		ShowMessage(&r, vErrSetter.Error(), "warning")
		return
	}

	if b.setter != nil {
		b.setter(obj, ctx)
	}

	needSave := true
	if b.mb.editing.Validator != nil {
		vErr := b.mb.editing.Validator(obj, ctx)
		newVErrSetter := vErrSetter
		_ = newVErrSetter.Merge(&vErr)
		vErr = newVErrSetter
		if vErr.HaveErrors() {
			ctx.Flash = &vErr
			needSave = false
			if vErr.GetGlobalError() != "" {
				ShowMessage(&r, vErr.GetGlobalError(), "warning")
			}
		}

	}
	vErr := b.validator(obj, ctx)
	_ = vErrSetter.Merge(&vErr)
	vErr = vErrSetter
	if vErr.HaveErrors() {
		ctx.Flash = &vErr
		needSave = false
		if vErr.GetGlobalError() != "" {
			ShowMessage(&r, vErr.GetGlobalError(), "warning")
		}
	}

	if needSave {
		err := b.saver(obj, id, ctx)
		if err != nil {
			var ve *web.ValidationErrors
			if errors.As(err, &ve) {
				ctx.Flash = ve
				if ve.GetGlobalError() != "" {
					ShowMessage(&r, ve.GetGlobalError(), "warning")
				}
			} else {
				ShowMessage(&r, err.Error(), "warning")
				return r, nil
			}
		}
	}

	if _, ok := ctx.Flash.(*web.ValidationErrors); ok {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: b.FieldPortalName(),
			Body: b.editComponent(obj, field, ctx),
		})
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.viewComponent(obj, field, ctx),
	})

	r.Emit(b.mb.NotifModelsUpdated(), PayloadModelsUpdated{
		Ids:    []string{id},
		Models: map[string]any{id: obj},
	})
	return r, nil
}

func (b *SectionBuilder) ValidateDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		id        = ctx.Param(ParamID)
		operateID = ctx.Param(ParamOperateID)
		obj       = b.mb.NewModel()
		vErr      web.ValidationErrors
	)

	defer func() {
		web.AppendRunScripts(&r,
			fmt.Sprintf(`if (dash.__ValidateOperateID==%q){%s}`, operateID,
				web.Emit(
					b.mb.NotifModelsSectionValidate(b.name),
					PayloadModelsSetter{
						FieldErrors: vErr.FieldErrors(),
						Id:          id,
						Passed:      !vErr.HaveErrors(),
					},
				),
			),
		)
		if vErr.HaveErrors() && vErr.HaveGlobalErrors() {
			web.AppendRunScripts(&r, ShowSnackbarScript(strings.Join(vErr.GetGlobalErrors(), ";"), ColorWarning))
		}
	}()

	if id != "" {
		var err1 error
		obj, err1 = b.mb.editing.Fetcher(obj, id, ctx)
		if err1 != nil {
			vErr.GlobalError(err1.Error())
			return
		}
	}
	if b.setter != nil {
		b.setter(obj, ctx)
	}

	vErr = b.editingFB.Unmarshal(obj, b.mb.Info(), false, ctx)
	if vErr.HaveErrors() && vErr.HaveGlobalErrors() {
		return
	}
	vErrSetter := vErr
	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		vErr.GlobalError(perm.PermissionDenied.Error())
		return
	}
	if b.validator != nil {
		vErr = b.validator(obj, ctx)
		_ = vErrSetter.Merge(&vErr)
		if vErrSetter.HaveErrors() {
			vErr = vErrSetter
			return
		}
	}

	return
}

// EditDetailListField Event: click detail list field element edit button
func (b *SectionBuilder) EditDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var index int64
	unsaved := ctx.ParamAsBool(b.elementUnsavedKey())
	index, err = strconv.ParseInt(ctx.Queries().Get(b.EditBtnKey()), 10, 32)
	if err != nil {
		return
	}

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	if unsaved {
		if _, err := b.appendElement(obj); err != nil {
			panic(err)
		}
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.listComponent(obj, ctx, -1, int(index), -1, unsaved, false),
	})

	return
}

// SaveDetailListField Event: click detail list field element Save button
func (b *SectionBuilder) SaveDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		index    int
		isCancel bool
	)

	isCancel = ctx.ParamAsBool(SectionIsCancel)

	unsaved := ctx.ParamAsBool(b.elementUnsavedKey())
	index = ctx.ParamAsInt(b.SaveBtnKey())

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	listObj := reflect.ValueOf(reflectutils.MustGet(obj, b.name))
	var isUnsavedAdded bool
	if listObj.IsValid() && listObj.Len() == int(index) {
		isUnsavedAdded = true
	}

	if isCancel {
		if unsaved && !isUnsavedAdded {
			if _, err = b.appendElement(obj); err != nil {
				panic(err)
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: b.FieldPortalName(),
			Body: b.listComponent(obj, ctx, -1, -1, index, unsaved, false),
		})
		return
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	// Determine if current save is for a newly created element before potential append
	wasCreated := !listObj.IsValid() || listObj.Len() == int(index)
	if wasCreated {
		b.appendElement(obj)
		listObj = reflect.ValueOf(reflectutils.MustGet(obj, b.name))
	}
	elementObj := listObj.Index(int(index)).Interface()
	newForm := new(multipart.Form)
	newForm.Value = make(map[string][]string)
	for key, val := range ctx.R.MultipartForm.Value {
		prefix := fmt.Sprintf("%s[%d].", b.name, index)
		if strings.HasPrefix(key, prefix) {
			newForm.Value[strings.TrimPrefix(key, prefix)] = val
		}
	}
	oldForm := ctx.R.MultipartForm
	ctx.R.MultipartForm = newForm
	if Verr := b.editingFB.Unmarshal(elementObj, b.mb.Info(), true, ctx); Verr.HaveErrors() {
		ShowMessage(&r, Verr.Error(), "warning")
		return r, nil
	}
	ctx.R.MultipartForm = oldForm
	listObj.Index(int(index)).Set(reflect.ValueOf(elementObj))

	if b.setter != nil {
		b.setter(obj, ctx)
	}

	needSave := true
	if b.mb.editing.Validator != nil {
		if vErr := b.mb.editing.Validator(obj, ctx); vErr.HaveErrors() {
			ctx.Flash = &vErr
			needSave = false
			if vErr.GetGlobalError() != "" {
				ShowMessage(&r, vErr.GetGlobalError(), "warning")
			}
		}
	}
	if vErr := b.validator(obj, ctx); vErr.HaveErrors() {
		ctx.Flash = &vErr
		needSave = false
		if vErr.GetGlobalError() != "" {
			ShowMessage(&r, vErr.GetGlobalError(), "warning")
		}
	}

	if needSave {
		err = b.saver(obj, ctx.Queries().Get(ParamID), ctx)
		if err != nil {
			ShowMessage(&r, err.Error(), "warning")
			return r, nil
		}
	}

	// Append a new empty element only when caller requests keeping an unsaved slot,
	// and only after a successful save of an existing element.
	// Avoid duplicating when saving a newly created element or when validation failed.
	if ctx.ParamAsBool(b.elementUnsavedKey()) && err == nil && !wasCreated {
		if _, err := b.appendElement(obj); err != nil {
			panic(err)
		}
	}

	if _, ok := ctx.Flash.(*web.ValidationErrors); ok {
		editIDForRender := -1
		if wasCreated {
			// Keep unsaved state only when the newly created element fails validation
			unsaved = true
			// For newly created items, pass editID to show in edit mode
			editIDForRender = index
		}
		// For existing items (wasCreated=false), pass editID=-1 so the item is shown
		// in edit mode via form value, but "Add Item" button remains visible
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: b.FieldPortalName(),
			Body: b.listComponent(obj, ctx, -1, editIDForRender, -1, unsaved, false),
		})
		return
	}

	if isUnsavedAdded {
		unsaved = false
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.listComponent(obj, ctx, -1, -1, index, unsaved, false),
	})

	return
}

// DeleteDetailListField Event: click detail list field element Delete button
func (b *SectionBuilder) DeleteDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var index int64

	unsaved := ctx.ParamAsBool(b.elementUnsavedKey())
	index, err = strconv.ParseInt(ctx.Queries().Get(b.DeleteBtnKey()), 10, 32)
	if err != nil {
		return
	}

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	err = b.removeElement(obj, int(index))
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	needSave := true
	if b.mb.editing.Validator != nil {
		if vErr := b.mb.editing.Validator(obj, ctx); vErr.HaveErrors() {
			ctx.Flash = &vErr
			needSave = false
			if vErr.GetGlobalError() != "" {
				ShowMessage(&r, vErr.GetGlobalError(), "warning")
			}
		}
	}

	if needSave {
		err = b.saver(obj, ctx.Queries().Get(ParamID), ctx)
		if err != nil {
			ShowMessage(&r, err.Error(), "warning")
			return r, nil
		}
	}

	if unsaved {
		if _, err := b.appendElement(obj); err != nil {
			panic(err)
		}
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.listComponent(obj, ctx, int(index), -1, -1, unsaved, false),
	})

	return
}

// CreateDetailListField Event: click detail list field element Add row button
func (b *SectionBuilder) CreateDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	var listLen int
	if ctx.ParamAsBool(b.elementUnsavedKey()) {
		if listLen, err = b.appendElement(obj); err != nil {
			panic(err)
		}
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.listComponent(obj, ctx, -1, listLen-1, -1, true, false),
	})

	return
}

func (b *SectionBuilder) ReloadDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.isList {
		return b.ReloadDetailListField(ctx)
	}

	field := &FieldContext{
		ModelInfo: b.mb.modelInfo,
		FormKey:   b.name,
		Name:      b.name,
		Label:     b.label,
	}

	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.viewComponent(obj, field, ctx),
	})
	return
}

func (b *SectionBuilder) ReloadDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	obj := b.mb.NewModel()
	obj, err = b.mb.editing.Fetcher(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if b.mb.editing.Setter != nil {
		b.mb.editing.Setter(obj, ctx)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: b.FieldPortalName(),
		Body: b.listComponent(obj, ctx, -1, -1, -1, false, true),
	})
	return
}

func (*SectionBuilder) getObjectID(ctx *web.EventContext, obj interface{}) string {
	id := ctx.Param(ParamID)
	if id == "" {
		if slugIf, ok := obj.(SlugEncoder); ok {
			id = slugIf.PrimarySlug()
		}
	}
	return id
}
