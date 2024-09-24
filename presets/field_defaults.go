package presets

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

type FieldDefaultBuilder struct {
	valType    reflect.Type
	compFunc   FieldComponentFunc
	setterFunc FieldSetterFunc
}

type FieldMode int

const (
	WRITE FieldMode = iota
	LIST
	DETAIL
)

func NewFieldDefault(t reflect.Type) (r *FieldDefaultBuilder) {
	r = &FieldDefaultBuilder{valType: t}
	return
}

func (b *FieldDefaultBuilder) ComponentFunc(v FieldComponentFunc) (r *FieldDefaultBuilder) {
	b.compFunc = v
	return b
}

func (b *FieldDefaultBuilder) SetterFunc(v FieldSetterFunc) (r *FieldDefaultBuilder) {
	b.setterFunc = v
	return b
}

var numberVals = []interface{}{
	int(0), int8(0), int16(0), int32(0), int64(0),
	uint(0), uint(8), uint16(0), uint32(0), uint64(0),
	float32(0.0), float64(0.0),
}

var stringVals = []interface{}{
	string(""),
	[]rune(""),
	[]byte(""),
}

var timeVals = []interface{}{
	time.Now(),
	ptrTime(time.Now()),
}

type FieldDefaults struct {
	mode             FieldMode
	fieldTypes       []*FieldDefaultBuilder
	excludesPatterns []string
}

func NewFieldDefaults(t FieldMode) (r *FieldDefaults) {
	r = &FieldDefaults{
		mode: t,
	}
	r.builtInFieldTypes()
	return
}

func (b *FieldDefaults) FieldType(v interface{}) (r *FieldDefaultBuilder) {
	return b.fieldTypeByTypeOrCreate(reflect.TypeOf(v))
}

func (b *FieldDefaults) Exclude(patterns ...string) (r *FieldDefaults) {
	b.excludesPatterns = patterns
	return b
}

func (b *FieldDefaults) InspectFields(val interface{}) (r *FieldsBuilder) {
	r, _ = b.inspectFieldsAndCollectName(val, nil)
	r.Model(val)
	return
}

func (b *FieldDefaults) inspectFieldsAndCollectName(val interface{}, collectType reflect.Type) (r *FieldsBuilder, names []string) {
	v := reflect.ValueOf(val)

	for v.Elem().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v = v.Elem()

	t := v.Type()

	r = &FieldsBuilder{
		model: val,
	}
	r.Defaults(b)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		ft := b.fieldTypeByType(f.Type)

		if !hasMatched(b.excludesPatterns, f.Name) && ft != nil {
			r.Field(f.Name).
				ComponentFunc(ft.compFunc).
				SetterFunc(ft.setterFunc)
		}

		if collectType != nil && f.Type == collectType {
			names = append(names, strcase.ToSnake(f.Name))
		}
	}

	return
}

func hasMatched(patterns []string, name string) bool {
	for _, p := range patterns {
		ok, err := filepath.Match(p, name)
		if err != nil {
			panic(err)
		}
		if ok {
			return true
		}
	}
	return false
}

func (b *FieldDefaults) String() string {
	var types []string
	for _, t := range b.fieldTypes {
		types = append(types, fmt.Sprint(t.valType))
	}
	return fmt.Sprintf("mode: %d, types %v", b.mode, types)
}

func (b *FieldDefaults) fieldTypeByType(tv reflect.Type) (r *FieldDefaultBuilder) {
	for _, ft := range b.fieldTypes {
		if ft.valType == tv {
			return ft
		}
	}
	return nil
}

func (b *FieldDefaults) fieldTypeByTypeOrCreate(tv reflect.Type) (r *FieldDefaultBuilder) {
	if r = b.fieldTypeByType(tv); r != nil {
		return
	}

	r = NewFieldDefault(tv)

	if b.mode == LIST {
		r.ComponentFunc(cfTextTd)
	} else {
		r.ComponentFunc(cfTextField)
	}
	b.fieldTypes = append(b.fieldTypes, r)
	return
}

func cfTextTd(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return h.Td(h.Text(field.StringValue(obj)))
}

func cfCheckbox(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return VCheckbox().
		Attr(web.VField(field.FormKey, reflectutils.MustGet(obj, field.Name).(bool))...).
		Label(field.Label).
		ErrorMessages(field.Errors...).
		Disabled(field.Disabled)
}

func cfNumber(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return VTextField().
		Type("number").
		Variant(FieldVariantUnderlined).
		Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
		Label(field.Label).
		ErrorMessages(field.Errors...).
		Disabled(field.Disabled)
}

func cfTime(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, CoreI18nModuleKey, Messages_en_US).(*Messages)
	val := ""
	if v := field.Value(obj); v != nil {
		switch vt := v.(type) {
		case time.Time:
			val = vt.Format("2006-01-02 15:04")
		case *time.Time:
			val = vt.Format("2006-01-02 15:04")
		default:
			panic(fmt.Sprintf("unknown time type: %T\n", v))
		}
	}
	return vuetifyx.VXDateTimePicker().
		Label(field.Label).
		Attr(web.VField(field.FormKey, val)...).
		Value(val).
		TimePickerProps(vuetifyx.TimePickerProps{
			Format:     "24hr",
			Scrollable: true,
		}).
		ClearText(msgr.Clear).
		OkText(msgr.OK).
		Disabled(field.Disabled)
}

func cfTimeSetter(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
	v := ctx.R.Form.Get(field.FormKey)
	if v == "" {
		return reflectutils.Set(obj, field.Name, nil)
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", v, time.Local)
	if err != nil {
		return err
	}
	return reflectutils.Set(obj, field.Name, t)
}
func cfTextField(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return CfTextField().Label(field.Label).
		Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
		ErrorMessages(field.Errors...).
		Disabled(field.Disabled)
}

func CfTextField() *vuetifyx.VXFieldBuilder {
	return vuetifyx.VXField()
}
func cfSelectField(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return CfSelectField().
		Label(field.Label).
		Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
		ErrorMessages(field.Errors...)
}

func CfSelectField() *vuetifyx.VXSelectBuilder {
	return vuetifyx.VXSelect()
}

func CfReadonlyText() *vuetifyx.VXReadonlyFieldBuilder {
	return vuetifyx.VXReadonlyField()
}
func cfReadonlyText(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return CfReadonlyText().
		Label(field.Label).
		Value(field.StringValue(obj))
}

func CfReadonlyCheckbox() *vuetifyx.VXReadonlyFieldBuilder {
	return vuetifyx.VXReadonlyField()
}

func cfReadonlyCheckbox(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return CfReadonlyCheckbox().
		Label(field.Label).
		Value(reflectutils.MustGet(obj, field.Name)).
		Checkbox(true)
}
func (b *FieldDefaults) builtInFieldTypes() {
	if b.mode == LIST {
		b.FieldType(true).
			ComponentFunc(cfTextTd)

		for _, v := range numberVals {
			b.FieldType(v).
				ComponentFunc(cfTextTd)
		}

		for _, v := range stringVals {
			b.FieldType(v).
				ComponentFunc(cfTextTd)
		}
		return
	}

	if b.mode == DETAIL {
		b.FieldType(true).
			ComponentFunc(cfReadonlyCheckbox)

		for _, v := range numberVals {
			b.FieldType(v).
				ComponentFunc(cfReadonlyText)
		}

		for _, v := range stringVals {
			b.FieldType(v).
				ComponentFunc(cfReadonlyText)
		}
		return
	}

	b.FieldType(true).
		ComponentFunc(cfCheckbox)

	for _, v := range numberVals {
		b.FieldType(v).
			ComponentFunc(cfNumber)
	}

	for _, v := range stringVals {
		b.FieldType(v).
			ComponentFunc(cfTextField)
	}

	for _, v := range timeVals {
		b.FieldType(v).
			ComponentFunc(cfTime).
			SetterFunc(cfTimeSetter)
	}

	b.Exclude("ID")
}
