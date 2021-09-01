package richeditor

import (
	"context"

	h "github.com/theplant/htmlgo"
)

type RedactorBuilder struct {
	tag *h.HTMLTagBuilder
}

type RedactorConfig struct {
	Plugins []string `json:"plugins"`
}

func Redactor() (r *RedactorBuilder) {
	r = &RedactorBuilder{
		tag: h.Tag("redactor"),
	}

	return
}

func (b *RedactorBuilder) Value(v string) (r *RedactorBuilder) {
	b.tag.Attr(":value", h.JSONString(v))
	return b
}
func (b *RedactorBuilder) Placeholder(v string) (r *RedactorBuilder) {
	b.tag.Attr(":placeholder", h.JSONString(v))
	return b
}
func (b *RedactorBuilder) Config(v RedactorConfig) (r *RedactorBuilder) {
	b.tag.Attr(":config", h.JSONString(v))
	return b
}

func (b *RedactorBuilder) SetAttr(k string, v interface{}) {
	b.tag.SetAttr(k, v)
}

func (b *RedactorBuilder) Attr(vs ...interface{}) (r *RedactorBuilder) {
	b.tag.Attr(vs...)
	return b
}

func (b *RedactorBuilder) Children(children ...h.HTMLComponent) (r *RedactorBuilder) {
	b.tag.Children(children...)
	return b
}

func (b *RedactorBuilder) AppendChildren(children ...h.HTMLComponent) (r *RedactorBuilder) {
	b.tag.AppendChildren(children...)
	return b
}

func (b *RedactorBuilder) PrependChildren(children ...h.HTMLComponent) (r *RedactorBuilder) {
	b.tag.PrependChildren(children...)
	return b
}

func (b *RedactorBuilder) Class(names ...string) (r *RedactorBuilder) {
	b.tag.Class(names...)
	return b
}

func (b *RedactorBuilder) ClassIf(name string, add bool) (r *RedactorBuilder) {
	b.tag.ClassIf(name, add)
	return b
}

func (b *RedactorBuilder) MarshalHTML(ctx context.Context) (r []byte, err error) {
	return b.tag.MarshalHTML(ctx)
}
