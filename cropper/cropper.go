package cropper

import (
	"context"

	h "github.com/theplant/htmlgo"
)

type CropperBuilder struct {
	tag *h.HTMLTagBuilder
}

func Cropper() (r *CropperBuilder) {
	r = &CropperBuilder{
		tag: h.Tag("vue-cropper"),
	}

	return
}

func (b *CropperBuilder) Src(v string) (r *CropperBuilder) {
	b.tag.Attr(":src", h.JSONString(v))
	return b
}

func (b *CropperBuilder) Alt(v string) (r *CropperBuilder) {
	b.tag.Attr(":alt", h.JSONString(v))
	return b
}

func (b *CropperBuilder) Value(v string) (r *CropperBuilder) {
	b.tag.Attr(":value", h.JSONString(v))
	return b
}

func (b *CropperBuilder) SetAttr(k string, v interface{}) {
	b.tag.SetAttr(k, v)
}

func (b *CropperBuilder) Attr(vs ...interface{}) (r *CropperBuilder) {
	b.tag.Attr(vs...)
	return b
}

func (b *CropperBuilder) Children(children ...h.HTMLComponent) (r *CropperBuilder) {
	b.tag.Children(children...)
	return b
}

func (b *CropperBuilder) AppendChildren(children ...h.HTMLComponent) (r *CropperBuilder) {
	b.tag.AppendChildren(children...)
	return b
}

func (b *CropperBuilder) PrependChildren(children ...h.HTMLComponent) (r *CropperBuilder) {
	b.tag.PrependChildren(children...)
	return b
}

func (b *CropperBuilder) Class(names ...string) (r *CropperBuilder) {
	b.tag.Class(names...)
	return b
}

func (b *CropperBuilder) ClassIf(name string, add bool) (r *CropperBuilder) {
	b.tag.ClassIf(name, add)
	return b
}

func (b *CropperBuilder) MarshalHTML(ctx context.Context) (r []byte, err error) {
	return b.tag.MarshalHTML(ctx)
}
