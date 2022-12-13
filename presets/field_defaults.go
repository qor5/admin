package presets

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"
	"unsafe"

	"github.com/iancoleman/strcase"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
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
		FieldName(field.FormKey).
		Label(field.Label).
		InputValue(reflectutils.MustGet(obj, field.Name).(bool)).
		ErrorMessages(field.Errors...).
		Disabled(field.Disabled)
}

func cfNumber(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return VTextField().
		Type("number").
		FieldName(field.FormKey).
		Label(field.Label).
		Value(fmt.Sprint(reflectutils.MustGet(obj, field.Name))).
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
		FieldName(field.FormKey).
		Value(val).
		TimePickerProps(vuetifyx.TimePickerProps{
			Format:     "24hr",
			Scrollable: true,
		}).
		ClearText(msgr.Clear).
		OkText(msgr.OK)
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
	return VTextField().
		Type("text").
		FieldName(field.FormKey).
		Label(field.Label).
		Value(fmt.Sprint(reflectutils.MustGet(obj, field.Name))).
		ErrorMessages(field.Errors...).
		Disabled(field.Disabled)
}

func cfReadonlyText(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return vuetifyx.VXReadonlyField().
		Label(field.Label).
		Value(field.StringValue(obj))
}

func cfReadonlyCheckbox(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return vuetifyx.VXReadonlyField().
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
	return
}

type autoFillComps interface {
	h.HTMLTagBuilder |
		VSelectBuilder | VAutocompleteBuilder | VTextFieldBuilder
}

// WithDefaults will automatic filling of
// FieldName, Label, and Value attrs for autoFillComps.
func WithDefaults[T autoFillComps](comp *T, obj any, field *FieldContext) *T {
	r := reflect.Indirect(reflect.ValueOf(comp))
	if r.Kind() != reflect.Invalid {
		if r.Type() == reflect.TypeOf(h.HTMLTagBuilder{}) {
			// HTML component
			if fmt.Sprint(r.FieldByName("tag")) != "input" {
				return comp
			}
			attrs := r.FieldByName("attrs")
			newType := attrs.Type().Elem().Elem()
			var newAttrs []reflect.Value
			{
				// FieldName
				vs := web.VFieldName(field.FormKey)
				newAttrs = append(newAttrs, setAttr(newType, fmt.Sprint(vs[0]), fmt.Sprint(vs[1])).Addr())
				// Value
				newAttrs = append(newAttrs, setAttr(newType, "value", fmt.Sprint(field.Value(obj))).Addr())
			}
			attrs = reflect.NewAt(attrs.Type(), unsafe.Pointer(attrs.UnsafeAddr())).Elem()
			attrs.Set(reflect.Append(attrs, newAttrs...))
		} else {
			// Vuetify component
			tag := reflect.Indirect(r.FieldByName("tag"))
			attrs := tag.FieldByName("attrs")
			newType := attrs.Type().Elem().Elem()
			var newAttrs []reflect.Value
			{
				// FieldName
				vs := web.VFieldName(field.FormKey)
				newAttrs = append(newAttrs, setAttr(newType, fmt.Sprint(vs[0]), fmt.Sprint(vs[1])).Addr())
				// Label
				newAttrs = append(newAttrs, setAttr(newType, "label", field.Label).Addr())
				// Value
				newAttrs = append(newAttrs, setAttr(newType, ":value", h.JSONString(field.Value(obj))).Addr())
			}
			attrs = reflect.NewAt(attrs.Type(), unsafe.Pointer(attrs.UnsafeAddr())).Elem()
			attrs.Set(reflect.Append(attrs, newAttrs...))
		}
	}

	return comp
}

func setAttr(t reflect.Type, key string, value string) reflect.Value {
	newAttr := reflect.New(t).Elem()
	reflect.NewAt(newAttr.FieldByName("key").Type(), unsafe.Pointer(newAttr.FieldByName("key").UnsafeAddr())).Elem().
		Set(reflect.ValueOf(key))
	reflect.NewAt(newAttr.FieldByName("value").Type(), unsafe.Pointer(newAttr.FieldByName("value").UnsafeAddr())).Elem().
		Set(reflect.ValueOf(value))
	return newAttr
}
