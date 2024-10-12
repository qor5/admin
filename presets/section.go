package presets

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
)

const (
	SectionFieldName = "section"
	SectionIsCancel  = "isCancel"

	sectionListUnsavedKey        = "sectionListUnsaved"
	sectionListEditBtnKey        = "sectionListEditBtn"
	sectionListSaveBtnKey        = "sectionListSaveBtn"
	sectionListFieldDeleteBtnKey = "sectionListDeleteBtn"

	sectionListFieldEditing = "sectionListEditing"
)

type SectionsBuilder struct {
	mb       *ModelBuilder
	sections []*SectionBuilder
	FieldsBuilder
}

func (d *SectionsBuilder) Section(name string) *SectionBuilder {
	for _, v := range d.sections {
		if v.name == name {
			return v
		}
	}
	return d.appendNewSection(name)
}

func (d *SectionsBuilder) GetSections() []*SectionBuilder {
	return slices.Clone(d.sections)
}

func (d *SectionsBuilder) appendNewSection(name string) (r *SectionBuilder) {
	r = &SectionBuilder{
		NameLabel: NameLabel{
			name:  name,
			label: name,
		},
		validator:         nil,
		setter:            nil,
		componentViewFunc: nil,
		componentEditFunc: nil,
		father:            d,
		elementViewFunc:   nil,
		elementEditFunc:   nil,
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
	r.editingFB.Model(d.mb.model)
	r.editingFB.defaults = d.mb.writeFields.defaults
	r.viewingFB.Model(d.mb.model)
	r.viewingFB.defaults = d.mb.p.detailFieldDefaults
	r.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		return r.father.mb.editing.Saver(obj, id, ctx)
	})
	r.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		if r.father.mb.editing.Validator != nil {
			return r.father.mb.editing.Validator(obj, ctx)
		}
		return err
	})
	r.UnmarshalFunc(r.DefaultUnmarshalFunc)
	d.sections = append(d.sections, r)

	// d.Field(name).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	// 	panic("you must set ViewComponentFunc and EditComponentFunc if you want to use SectionsBuilder")
	// })

	return
}

// SectionBuilder is a builder for a section in the detail page.
// save: 	   fetcher => setter => saver
// show, edit: fetcher => setter
type SectionBuilder struct {
	NameLabel
	unmarshalFunc     func(ctx *web.EventContext, obj interface{}) error
	saver             SaveFunc
	setter            SetterFunc
	validator         ValidateFunc
	hiddenFuncs       []ObjectComponentFunc
	componentViewFunc FieldComponentFunc
	componentEditFunc FieldComponentFunc
	father            *SectionsBuilder

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

func (b *SectionBuilder) Tabs(tabName string) *SectionBuilder {
	tabField := b.father.GetField(tabName)
	if tabField == nil {
		panic(fmt.Sprintf("field named %s required", tabName))
	}
	tabField.AppendTabs(b.father.Field(b.name))
	b.father.GetField(b.name).hidden = true
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
	if b.father.model == nil {
		panic("model must be provided")
	}
	rt := reflectutils.GetType(b.father.model, b.name)
	if rt.Kind() != reflect.Slice {
		panic("field kind must be slice")
	}
	if reflect.TypeOf(reflect.New(rt.Elem()).Elem().Interface()) != reflect.TypeOf(v) {
		panic(fmt.Sprintf("%s not equal to %s", reflect.New(rt.Elem()).Elem().Interface(), reflect.TypeOf(v)))
	}

	r = b
	r.editingFB.Model(v)
	r.isList = true
	r.UnmarshalFunc(r.DefaultListUnmarshalFunc)
	r.elementUnmarshaler = r.DefaultElementUnmarshal()

	return
}

// Editing default saver only save these field
func (b *SectionBuilder) Editing(fields ...interface{}) (r *SectionBuilder) {
	r = b
	b.editingFB = *b.editingFB.Only(fields...)
	if b.componentEditFunc == nil {
		b.EditComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return b.editingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, b.name, ctx)
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
			return b.viewingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, b.name, ctx)
		})
	}
	if b.isList {
		if b.elementViewFunc == nil {
			b.ElementShowComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
				return b.viewingFB.toComponentWithModifiedIndexes(field.ModelInfo, obj, field.FormKey, ctx)
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

func (b *SectionBuilder) UnmarshalFunc(v func(ctx *web.EventContext, obj interface{}) error) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.unmarshalFunc = v
	return b
}

func (b *SectionBuilder) WrapSetterFunc(w func(in SetterFunc) SetterFunc) (r *SectionBuilder) {
	b.setter = w(b.setter)
	return b
}

func (b *SectionBuilder) SetterFunc(v SetterFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.setter = v
	return b
}

func (b *SectionBuilder) ValidateFunc(v ValidateFunc) (r *SectionBuilder) {
	if v == nil {
		panic("value required")
	}
	b.validator = v
	return b
}

func (b *SectionBuilder) WrapValidateFunc(w func(in ValidateFunc) ValidateFunc) (r *SectionBuilder) {
	b.validator = w(b.validator)
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
			return web.Portal(
				b.viewComponent(obj, field, ctx),
			).Name(b.FieldPortalName())
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
			return web.Portal(
				b.viewComponent(obj, field, ctx),
			).Name(b.FieldPortalName())
		})
	}
	return b
}

func (b *SectionBuilder) Label(label string) (r *SectionBuilder) {
	b.father.Field(b.name).Label(label)
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
					b.listComponent(obj, ctx, -1, -1, -1, false),
				).Name(b.FieldPortalName()),
			).VSlot("{ locals }").Init("{show:true}")
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
			return web.Portal(
				b.listComponent(obj, ctx, -1, -1, -1, false),
			).Name(b.FieldPortalName())
		})
	}
	return b
}

// ComponentFunc set FieldBuilder compFunc
func (b *SectionBuilder) ComponentFunc(v FieldComponentFunc) (r *FieldBuilder) {
	r = b.father.Field(b.name)
	return r.ComponentFunc(v)
}

func (b *SectionBuilder) ListFieldPrefix(index int) string {
	return fmt.Sprintf("%s[%b]", b.name, index)
}

func (b *SectionBuilder) viewComponent(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	id := ctx.Param(ParamID)
	if id == "" {
		if slugIf, ok := obj.(SlugEncoder); ok {
			id = slugIf.PrimarySlug()
		}
	}

	disableEditBtn := b.father.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	btn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Edit")).Variant(VariantFlat).Size(SizeXSmall).
		PrependIcon("mdi-pencil-outline").
		Attr("v-show", fmt.Sprintf("%t&&%t", b.componentEditBtnFunc(obj, ctx), !disableEditBtn)).
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(actions.DoEditDetailingField).
			Query(SectionFieldName, b.name).
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
		lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.father.mb.label, b.label)
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
		web.Scope(
			content,
		).VSlot("{ form }"),
		hiddenComp,
	)
}

func (b *SectionBuilder) editComponent(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	id := ctx.Param(ParamID)
	if id == "" {
		if slugIf, ok := obj.(SlugEncoder); ok {
			id = slugIf.PrimarySlug()
		}
	}
	onChangeEvent := fmt.Sprintf("if (vars.%s ){ vars.%s.section_%s=true };", VarsPresetsDataChanged, VarsPresetsDataChanged, b.name)
	cancelChangeEvent := fmt.Sprintf("if (vars.%s ){vars.%s.section_%s=false};", VarsPresetsDataChanged, VarsPresetsDataChanged, b.name)

	cancelBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Cancel")).Size(SizeSmall).Variant(VariantFlat).Color(ColorGreyLighten3).
		Attr("style", "text-transform: none;").
		Attr("@click", cancelChangeEvent+web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(actions.DoSaveDetailingField).
			Query(SectionFieldName, b.name).
			Query(ParamID, id).
			Query(SectionIsCancel, true).
			Go())

	disableEditBtn := b.father.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	saveBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Save")).PrependIcon("mdi-check").Size(SizeSmall).Variant(VariantFlat).Color(ColorPrimary).Disabled(disableEditBtn).
		Attr("style", "text-transform: none;").
		Attr("@click", cancelChangeEvent+web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(actions.DoSaveDetailingField).
			Query(SectionFieldName, b.name).
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
		lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.father.mb.label, b.label)
		content.AppendChildren(
			h.Div(
				h.H2(lb).Class("section-title"),
			).Class("section-title-wrap"),
		)
	}

	if b.componentEditFunc != nil {
		content.AppendChildren(
			h.Div(
				VCard(
					VCardText(
						h.Div(
							// detailFields
							h.Div(b.componentEditFunc(obj, field, ctx)).
								Class("flex-grow-1 pb-6"),
							h.Div(
								cancelBtn,
								saveBtn.Class("ml-2"),
							).Class("section-edit-area bottom-area"),
						).Class("d-flex flex-column"),
					),
				).Variant(VariantOutlined),
			).Class("section-body"),
		)
	}
	return h.Div(
		web.Scope(
			content,
			hiddenComp,
		).VSlot("{ form }").OnChange(onChangeEvent).UseDebounce(150),
	)
}

func (b *SectionBuilder) DefaultUnmarshalFunc(ctx *web.EventContext, obj interface{}) (err error) {
	if tf := reflect.TypeOf(obj).Kind(); tf != reflect.Ptr {
		return fmt.Errorf("model %#+v must be pointer", obj)
	}
	formObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()

	if err = b.DefaultElementUnmarshal()(obj, formObj, b.name, ctx); err != nil {
		return
	}
	return
}

func (b *SectionBuilder) DefaultListUnmarshalFunc(ctx *web.EventContext, obj interface{}) (err error) {
	var index int64
	index, err = strconv.ParseInt(ctx.Queries().Get(b.SaveBtnKey()), 10, 64)
	if err != nil {
		return
	}

	listObj := reflect.ValueOf(reflectutils.MustGet(obj, b.name))
	if !listObj.IsValid() || listObj.Len() == int(index) {
		b.appendElement(obj)
		listObj = reflect.ValueOf(reflectutils.MustGet(obj, b.name))
	}
	elementObj := listObj.Index(int(index)).Interface()
	formObj := reflect.New(reflect.TypeOf(b.editingFB.model).Elem()).Interface()
	if err = b.elementUnmarshaler(elementObj, formObj, b.ListFieldPrefix(int(index)), ctx); err != nil {
		return
	}
	listObj.Index(int(index)).Set(reflect.ValueOf(elementObj))
	return nil
}

func (b *SectionBuilder) listComponent(obj interface{}, ctx *web.EventContext, deletedID, editID, saveID int, unsaved bool) h.HTMLComponent {
	if b.elementHoverFunc != nil {
		b.elementHover = b.elementHoverFunc(obj, ctx)
	}
	if b.elementEditBtnFunc != nil {
		b.elementEditBtn = b.elementEditBtnFunc(obj, ctx)
	}
	if b.elementEditBtn {
		b.elementEditBtn = b.father.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() == nil
	}

	id := ctx.Param(ParamID)
	if id == "" {
		if slugIf, ok := obj.(SlugEncoder); ok {
			id = slugIf.PrimarySlug()
		}
	}

	list, err := reflectutils.Get(obj, b.name)
	if err != nil {
		panic(err)
	}
	listLen := 0
	if list != nil {
		listLen = reflect.ValueOf(list).Len()
	}

	lb := i18n.PT(ctx.R, ModelsI18nModuleKey, b.father.mb.label, b.label)
	label := h.Div(h.H2(lb).Class("section-title")).Class("section-title-wrap")
	rows := h.Div()

	if b.alwaysShowListLabel && !b.disableLabel {
		rows.AppendChildren(label)
	}

	if list != nil {
		i := 0
		reflectutils.ForEach(list, func(elementObj interface{}) {
			defer func() { i++ }()
			if i == 0 {
				if b.label != "" && !b.alwaysShowListLabel && !b.disableLabel {
					rows.AppendChildren(label)
				}
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
				rows.AppendChildren(b.editElement(elementObj, sortIndex, unsaved && i == listLen-1, ctx))
			} else if saveID == sortIndex {
				// if click save
				rows.AppendChildren(b.showElement(elementObj, sortIndex, ctx))
			} else {
				// default
				isEditing := ctx.R.FormValue(b.ListElementIsEditing(fromIndex)) != ""
				if !isEditing {
					rows.AppendChildren(b.showElement(elementObj, sortIndex, ctx))
				} else {
					formObj := reflect.New(reflect.TypeOf(b.editingFB.model).Elem()).Interface()
					err = b.elementUnmarshaler(elementObj, formObj, b.ListFieldPrefix(fromIndex), ctx)
					if err != nil {
						panic(err)
					}
					rows.AppendChildren(b.editElement(elementObj, sortIndex, unsaved && i == listLen-1, ctx))
				}
			}
		})
	}

	disableCreateBtn := b.father.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	disableCreateBtn = disableCreateBtn || (ctx.ParamAsBool(b.elementUnsavedKey()))
	if !b.disableElementCreateBtn && !disableCreateBtn {
		addBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "AddRow")).PrependIcon("mdi-plus-circle").Color("primary").Variant(VariantText).
			Class("mb-2 ml-4").
			Attr("@click", "locals.show=false;"+web.Plaid().
				URL(ctx.R.URL.Path).
				EventFunc(actions.DoCreateDetailingListField).
				Query(b.elementUnsavedKey(), true).
				Query(SectionFieldName, b.name).
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
	return fmt.Sprintf("%s_%s[%b].%s", deletedHiddenNamePrefix, b.name, index, sectionListFieldEditing)
}

func (b *SectionBuilder) ListElementPortalName(index int) string {
	return fmt.Sprintf("DetailElementPortal_%s_%b", b.name, index)
}

func (b *SectionBuilder) FieldPortalName() string {
	return fmt.Sprintf("DetailFieldPortal_%s", b.name)
}

func (b *SectionBuilder) showElement(obj any, index int, ctx *web.EventContext) h.HTMLComponent {
	editBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Edit")).Variant(VariantFlat).Size(SizeXSmall).
		PrependIcon("mdi-pencil-outline").
		Attr("v-show", fmt.Sprintf("%t", b.elementEditBtn)).
		Attr("@click", web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(actions.DoEditDetailingListField).
			Query(b.elementUnsavedKey(), ctx.Param(b.elementUnsavedKey())).
			Query(SectionFieldName, b.name).
			Query(ParamID, ctx.Param(ParamID)).
			Query(b.EditBtnKey(), strconv.Itoa(index)).
			Go())

	content := b.elementViewFunc(obj, &FieldContext{
		ModelInfo: b.father.mb.modelInfo,
		Name:      b.name,
		FormKey:   fmt.Sprintf("%s[%b]", b.name, index),
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
func (b *SectionBuilder) editElement(obj any, index int, isCreated bool, ctx *web.EventContext) h.HTMLComponent {
	showAddBtn := "locals.show=true;"
	unsaved := ctx.ParamAsBool(b.elementUnsavedKey())
	if isCreated {
		unsaved = false
	}

	deleteEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(actions.DoDeleteDetailingListField).
		Query(SectionFieldName, b.name).
		Query(ParamID, ctx.Param(ParamID)).
		Query(b.elementUnsavedKey(), unsaved).
		Query(b.DeleteBtnKey(), index).
		Go()
	if isCreated {
		deleteEvent = showAddBtn + deleteEvent
	}
	cancelEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(actions.DoSaveDetailingListField).
		Query(SectionFieldName, b.name).
		Query(b.elementUnsavedKey(), unsaved).
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
				ModelInfo: b.father.mb.modelInfo,
				Name:      fmt.Sprintf("%s[%b]", b.name, index),
				FormKey:   fmt.Sprintf("%s[%b]", b.name, index),
				Label:     fmt.Sprintf("%s[%b]", b.label, index),
			}, ctx),
		).Class("flex-grow-1"),
		h.Div(deleteBtn).Class("d-flex pl-3"),
	).Class("d-flex justify-space-between")

	cancelBtn := VBtn(i18n.T(ctx.R, CoreI18nModuleKey, "Cancel")).Size(SizeSmall).Variant(VariantFlat).Color(ColorGreyLighten3).
		Attr("style", "text-transform: none;").
		Attr("@click", cancelEvent)

	saveEvent := web.Plaid().
		URL(ctx.R.URL.Path).
		EventFunc(actions.DoSaveDetailingListField).
		Query(SectionFieldName, b.name).
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

		_ = ctx2.UnmarshalForm(formObj)
		for _, f := range b.editingFB.fields {
			name := f.name
			info := b.father.mb.modelInfo
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
