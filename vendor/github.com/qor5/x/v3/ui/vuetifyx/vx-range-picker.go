package vuetifyx

import (
	"context"
	"fmt"

	h "github.com/theplant/htmlgo"
)

type VXRangePickerBuilder struct {
	tag *h.HTMLTagBuilder
}

func VXRangePicker(children ...h.HTMLComponent) (r *VXRangePickerBuilder) {
	r = &VXRangePickerBuilder{
		tag: h.Tag("vx-range-picker").Children(children...),
	}
	return
}

func (b *VXRangePickerBuilder) Label(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("label", v)
	return b
}

func (b *VXRangePickerBuilder) Type(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("type", v)
	return b
}

func (b *VXRangePickerBuilder) Name(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("name", v)
	return b
}

func (b *VXRangePickerBuilder) Id(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("id", v)
	return b
}

func (b *VXRangePickerBuilder) Placeholder(v interface{}) (r *VXRangePickerBuilder) {
	b.tag.Attr(":placeholder", h.JSONString(v))
	return b
}

func (b *VXRangePickerBuilder) Readonly(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":readonly", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) Width(v int) (r *VXRangePickerBuilder) {
	b.tag.Attr("width", h.JSONString(v))
	return b
}

func (b *VXRangePickerBuilder) PasswordVisibleToggle(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":password-visible-toggle", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) PasswordVisibleDefault(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":password-visible-default", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) Disabled(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":disabled", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) Format(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("format", v)
	return b
}

func (b *VXRangePickerBuilder) Required(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":required", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) Autofocus(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":autofocus", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) HideDetails(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":hide-details", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) NeedConfirm(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":need-confirm", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) Clearable(v bool) (r *VXRangePickerBuilder) {
	b.tag.Attr(":clearable", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) ModelValue(v interface{}) (r *VXRangePickerBuilder) {
	b.tag.Attr(":model-value", h.JSONString(v))
	return b
}

func (b *VXRangePickerBuilder) DatePickerProps(v interface{}) (r *VXRangePickerBuilder) {
	b.tag.Attr(":date-picker-props", h.JSONString(v))
	return b
}

func (b *VXRangePickerBuilder) Attr(vs ...interface{}) (r *VXRangePickerBuilder) {
	b.tag.Attr(vs...)
	return b
}

func (b *VXRangePickerBuilder) SetAttr(k string, v interface{}) {
	b.tag.SetAttr(k, v)
}

func (b *VXRangePickerBuilder) Children(children ...h.HTMLComponent) (r *VXRangePickerBuilder) {
	b.tag.Children(children...)
	return b
}

func (b *VXRangePickerBuilder) Class(names ...string) (r *VXRangePickerBuilder) {
	b.tag.Class(names...)
	return b
}

func (b *VXRangePickerBuilder) Tips(v string) (r *VXRangePickerBuilder) {
	b.tag.Attr("tips", fmt.Sprint(v))
	return b
}

func (b *VXRangePickerBuilder) On(name string, value string) (r *VXRangePickerBuilder) {
	b.tag.Attr(fmt.Sprintf("v-on:%s", name), value)
	return b
}

func (b *VXRangePickerBuilder) Bind(name string, value string) (r *VXRangePickerBuilder) {
	b.tag.Attr(fmt.Sprintf("v-bind:%s", name), value)
	return b
}

func (b *VXRangePickerBuilder) ErrorMessages(errMsgs ...string) (r *VXRangePickerBuilder) {
	b.tag.Attr(":error-messages", errMsgs)
	return b
}

func (b *VXRangePickerBuilder) MarshalHTML(ctx context.Context) (r []byte, err error) {
	return b.tag.MarshalHTML(ctx)
}
