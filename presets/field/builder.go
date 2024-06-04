package field

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type Context struct {
	Parent  *Context
	Name    string
	Value   any
	Virtual bool // TODO: 应对可能的虚拟字段需求，或者换成通过虚拟字段的 Builder 自身来实现上下 Parent 衔接？哪种设计更好呢？

	scope    string
	once     sync.Once
	formKey  string
	ancestor *Context
}

func (c *Context) init() {
	c.once.Do(func() {
		if c.Parent != nil {
			if strings.HasPrefix(c.Name, "[") {
				c.formKey = c.Parent.FormKey() + c.Name
			} else {
				c.formKey = c.Parent.FormKey() + "." + c.Name
			}
			c.ancestor = c.Parent.Ancestor()
		} else {
			c.ancestor = c
		}
	})
}

func (c *Context) FormKey() string {
	c.init()
	return c.formKey
}

func (c *Context) Ancestor() *Context {
	c.init()
	return c.ancestor
}

// TODO: 可以依赖于此做组件内部 PortalName 的前缀来控制作用域？例如一个弹出页面里有两个 listing compo，那其 portal 前缀可以是 Scope()+portalName? 这个 scope 可以是类似以前 uriname 的存在?
// TODO: 或许应该设计为往 Parent 去找直到找到一个 scope 不为空的值？
func (c *Context) Scope() string {
	return c.Ancestor().scope
}

type Builder interface {
	Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error)
}

type BuilderFunc func(ctx *web.EventContext, field *Context) (h.HTMLComponent, error)

func (f BuilderFunc) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	return f(ctx, field)
}

type scopeBuilder struct {
	Builder
	scope string
}

// TODO: 或许不应该需要这个中间件，而是给 Context 提供一个设置方法更合理
func Scope(b Builder, scope string) Builder {
	return &scopeBuilder{b, scope}
}

func (b *scopeBuilder) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	field.scope = b.scope
	return b.Builder.Build(ctx, field)
}

type disengage struct {
	Builder
}

// TODO: 有些场景可能需要进行 Parent 断代
func Disengage(b Builder) Builder {
	return &disengage{b}
}

func (b *disengage) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	field.Parent = nil
	return b.Builder.Build(ctx, field)
}

type NamedBuilder struct {
	Builder
	Name string
}

func Named(b Builder, name string) *NamedBuilder {
	return &NamedBuilder{b, name}
}

func (b *NamedBuilder) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	field.Name = b.Name
	return b.Builder.Build(ctx, field)
}

type ScalarBuilder struct{}

func (b *ScalarBuilder) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	return h.Div(h.Text(fmt.Sprintf("%s: %v (FormKey: %s AncestorType: %T Scope: %s)", field.Name, field.Value, field.FormKey(), field.Ancestor().Value, field.Scope()))), nil
}

type SliceBuilder struct {
	ElemBuilder Builder
}

func (b *SliceBuilder) Build(ctx *web.EventContext, field *Context) (r h.HTMLComponent, err error) {
	children := []h.HTMLComponent{}
	i := 0
	reflectutils.ForEach(field.Value, func(elem interface{}) {
		defer func() { i++ }()
		if err != nil {
			return
		}
		compo, rerr := b.ElemBuilder.Build(ctx, &Context{
			Parent: field,
			Name:   fmt.Sprintf("[%d]", i),
			Value:  elem,
		})
		if rerr != nil {
			err = rerr
			return
		}
		children = append(children, compo)
	})
	if err != nil {
		return nil, err
	}
	return h.Div(children...).Class(field.Name), nil
}

// TIPS: 本可支持 map[string]Builder 但是因为其无序，就暂时没写，或许也可以参考 json 根据 key 做 order

type StructBuilder struct {
	model any // TODO: 只记录 reflect.Type 应该更合理些

	// TODO: 需要必填类似 FieldDefaults 的东西来输入 类型的默认实现 ，例如称之为 TypeDefaults，这个东西可能需要向下传递

	once          sync.Once
	fieldBuilders []*NamedBuilder
}

func Inspect(model any) *StructBuilder {
	fb := &StructBuilder{
		model: model,
	}
	return fb
}

func (b *StructBuilder) newSliceBuilder(elemType reflect.Type) Builder {
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	switch elemType.Kind() {
	case reflect.Slice, reflect.Array:
		return &SliceBuilder{ElemBuilder: b.newSliceBuilder(elemType.Elem())}
	case reflect.Struct:
		return &SliceBuilder{ElemBuilder: Inspect(reflect.New(elemType).Interface())}
	default:
		// TODO: 一些类型并非 scalar，还未处理
		return &SliceBuilder{ElemBuilder: &ScalarBuilder{}}
	}
}

func (b *StructBuilder) build() {
	b.once.Do(func() {
		rt := reflect.TypeOf(b.model)
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		for i := 0; i < rt.NumField(); i++ {
			fieldType := rt.Field(i).Type
			fieldName := rt.Field(i).Name
			for fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			switch fieldType.Kind() {
			case reflect.Slice, reflect.Array:
				b.fieldBuilders = append(b.fieldBuilders, Named(b.newSliceBuilder(fieldType.Elem()), fieldName))
			case reflect.Struct:
				b.fieldBuilders = append(b.fieldBuilders, Named(Inspect(reflect.New(fieldType).Interface()), fieldName))
			default:
				// TODO: 一些类型并非 scalar，还未处理
				b.fieldBuilders = append(b.fieldBuilders, Named(&ScalarBuilder{}, fieldName))
			}
		}
	})
}

func (b *StructBuilder) Build(ctx *web.EventContext, field *Context) (h.HTMLComponent, error) {
	// 最终使用时候执行 懒build ，以让之前的所有配置修改等等均此时才确认
	b.build()

	children := []h.HTMLComponent{}
	for _, v := range b.fieldBuilders {
		compo, err := v.Build(ctx, &Context{
			Parent: field,
			Value:  reflectutils.MustGet(field.Value, v.Name),
		})
		if err != nil {
			return nil, err
		}
		children = append(children, compo)
	}
	return h.Div(children...).Class(field.Name), nil
}

// TODO: 需要对 StructBuilder 的 wrapper，以满足类似以前 FieldsBuilder 里过滤字段或者给予虚拟字段的需求
// TODO: 这个 wrapper 也需要类似 FieldBuilder 的东西用以 按字段自定义实现
// TODO: 需要一个 FieldsLayout 的 interface ，其也实现 Builder ，以实现类似 FieldsSection 的需求
// TODO: 也可能需要一个 HijackFieldsBuilder 的中间件用以劫持这个 wrapper 里的某些字段到自身，供 FieldsLayout 和 wrapper 打配合
