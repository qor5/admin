package presets

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
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
	r.Model(val) // WARN: 多此一举，并且也不应该有 Model 方法，其在 FieldsBuilder 应为必选且无法修改项
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

		// WARN: excludesPatterns 只有在 WRITE 的时候才会被默认 Exclude ID
		// WARN: 这个应该会要求想要自定义 FieldDefaultBuilder 的话得在执行 b.Model 方法之前吧
		// WARN: 其实总感觉这个事不应该让 FieldDefaults 来做，它存一存默认实现就好啦，而且现在这么提前加载的话，我们要再对 FieldsBuilder 修改 Defaults ，新的修改也会因为这块逻辑无法应用体现
		// WARN: 当然了，如果 FieldDefaults 不负责 inspectFields ，那 excludesPatterns 也不应该它来处理
		if !hasMatched(b.excludesPatterns, f.Name) && ft != nil {
			r.Field(f.Name).
				ComponentFunc(ft.compFunc).
				SetterFunc(ft.setterFunc)
		}

		if collectType != nil && f.Type == collectType {
			// WARN: 不理解，目前只是被 newEditing 用到，给到的是字符串类型，只是收集所有字符串类型作为 listing 的 search columns ，但是为什么这个事是 newEditing 来做？其实觉得这个 collect 还是应该另外开个方法单独来做比较合适，现在这个太 trick
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
		r.ComponentFunc(cfTextField) // WARN: 是不是太武断了，会有影响吗？是因为它很通用所以这样？但是目前 DETAIL 可不简单是 Edit 哦
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
		DialogWidth(640).
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
		Variant(FieldVariantUnderlined).
		Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
		Label(field.Label).
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
