package vuetifyx

import (
	"context"
	"fmt"

	h "github.com/theplant/htmlgo"
)

type VXDatePickerBuilder struct {
	tag *h.HTMLTagBuilder
}

func VXDatepicker(children ...h.HTMLComponent) (r *VXDatePickerBuilder) {
	r = &VXDatePickerBuilder{
		tag: h.Tag("vx-date-picker").Children(children...),
	}
	return
}

func (b *VXDatePickerBuilder) Label(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("label", v)
	return b
}

func (b *VXDatePickerBuilder) Type(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("type", v)
	return b
}

func (b *VXDatePickerBuilder) Name(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("name", v)
	return b
}

func (b *VXDatePickerBuilder) Id(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("id", v)
	return b
}

func (b *VXDatePickerBuilder) Format(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("format", v)
	return b
}

func (b *VXDatePickerBuilder) Placeholder(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("placeholder", v)
	return b
}

func (b *VXDatePickerBuilder) Width(v int) (r *VXDatePickerBuilder) {
	b.tag.Attr("width", h.JSONString(v))
	return b
}

func (b *VXDatePickerBuilder) Disabled(v bool) (r *VXDatePickerBuilder) {
	b.tag.Attr(":disabled", fmt.Sprint(v))
	return b
}

func (b *VXDatePickerBuilder) Required(v bool) (r *VXDatePickerBuilder) {
	b.tag.Attr(":required", fmt.Sprint(v))
	return b
}

func (b *VXDatePickerBuilder) HideDetails(v bool) (r *VXDatePickerBuilder) {
	b.tag.Attr(":hide-details", fmt.Sprint(v))
	return b
}

func (b *VXDatePickerBuilder) Clearable(v bool) (r *VXDatePickerBuilder) {
	b.tag.Attr(":clearable", fmt.Sprint(v))
	return b
}

func (b *VXDatePickerBuilder) DatePickerProps(v interface{}) (r *VXDatePickerBuilder) {
	b.tag.Attr(":date-picker-props", h.JSONString(v))
	return b
}

func (b *VXDatePickerBuilder) ModelValue(v interface{}) (r *VXDatePickerBuilder) {
	b.tag.Attr(":model-value", h.JSONString(v))
	return b
}

func (b *VXDatePickerBuilder) Attr(vs ...interface{}) (r *VXDatePickerBuilder) {
	b.tag.Attr(vs...)
	return b
}

func (b *VXDatePickerBuilder) SetAttr(k string, v interface{}) {
	b.tag.SetAttr(k, v)
}

func (b *VXDatePickerBuilder) Children(children ...h.HTMLComponent) (r *VXDatePickerBuilder) {
	b.tag.Children(children...)
	return b
}

func (b *VXDatePickerBuilder) Class(names ...string) (r *VXDatePickerBuilder) {
	b.tag.Class(names...)
	return b
}

func (b *VXDatePickerBuilder) Tips(v string) (r *VXDatePickerBuilder) {
	b.tag.Attr("tips", fmt.Sprint(v))
	return b
}

func (b *VXDatePickerBuilder) On(name string, value string) (r *VXDatePickerBuilder) {
	b.tag.Attr(fmt.Sprintf("v-on:%s", name), value)
	return b
}

func (b *VXDatePickerBuilder) Bind(name string, value string) (r *VXDatePickerBuilder) {
	b.tag.Attr(fmt.Sprintf("v-bind:%s", name), value)
	return b
}

func (b *VXDatePickerBuilder) ErrorMessages(errMsgs ...string) (r *VXDatePickerBuilder) {
	b.tag.Attr(":error-messages", errMsgs)
	return b
}

func (b *VXDatePickerBuilder) MarshalHTML(ctx context.Context) (r []byte, err error) {
	return b.tag.MarshalHTML(ctx)
}
