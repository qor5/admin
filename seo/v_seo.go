package seo

import (
	"context"

	h "github.com/theplant/htmlgo"
)

type VSeoBuilder struct {
	tag *h.HTMLTagBuilder
}

func VSeo(children ...h.HTMLComponent) (r *VSeoBuilder) {
	r = &VSeoBuilder{
		tag: h.Tag("v-seo").Children(children...),
	}
	return
}

func (b *VSeoBuilder) Value(v string) (r *VSeoBuilder) {
	b.tag.Attr(":value", h.JSONString(v))
	return b
}

func (b *VSeoBuilder) Placeholder(v string) (r *VSeoBuilder) {
	b.tag.Attr(":placeholder", h.JSONString(v))
	return b
}

func (b *VSeoBuilder) SetAttr(k string, v interface{}) {
	b.tag.SetAttr(k, v)
}

func (b *VSeoBuilder) Attr(vs ...interface{}) (r *VSeoBuilder) {
	b.tag.Attr(vs...)
	return b
}

func (b *VSeoBuilder) Children(children ...h.HTMLComponent) (r *VSeoBuilder) {
	b.tag.Children(children...)
	return b
}

func (b *VSeoBuilder) AppendChildren(children ...h.HTMLComponent) (r *VSeoBuilder) {
	b.tag.AppendChildren(children...)
	return b
}

func (b *VSeoBuilder) PrependChildren(children ...h.HTMLComponent) (r *VSeoBuilder) {
	b.tag.PrependChildren(children...)
	return b
}

func (b *VSeoBuilder) Class(names ...string) (r *VSeoBuilder) {
	b.tag.Class(names...)
	return b
}

func (b *VSeoBuilder) ClassIf(name string, add bool) (r *VSeoBuilder) {
	b.tag.ClassIf(name, add)
	return b
}

func (b *VSeoBuilder) MarshalHTML(ctx context.Context) (r []byte, err error) {
	return b.tag.MarshalHTML(ctx)
}
