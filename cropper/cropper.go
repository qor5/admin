package cropper

import (
	"context"

	h "github.com/theplant/htmlgo"
)

type CropperBuilder struct {
	tag *h.HTMLTagBuilder
}

const (
	VIEW_MODE_NO_RESTRICTIONS      = 0
	VIEW_MODE_RESTRICT_CROP_BOX    = 1
	VIEW_MODE_FIT_WITHIN_CONTAINER = 2
	VIEW_MODE_FILL_FIT_CONTAINER   = 3
)

// {"x":1141.504660866477,"y":540.6135919744316,"width":713.7745472301137,"height":466.93834339488643,"rotate":0,"scaleX":1,"scaleY":1}
type Value struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	//Rotate float64 `json:"rotate"`
	//ScaleX float64 `json:"scaleX"`
	//ScaleY float64 `json:"scaleY"`
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

func (b *CropperBuilder) AspectRatio(width float64, height float64) (r *CropperBuilder) {
	b.tag.Attr(":aspect-ratio", width/height)
	return b
}

func (b *CropperBuilder) ViewMode(viewMode int) (r *CropperBuilder) {
	b.tag.Attr(":view-mode", viewMode)
	return b
}

func (b *CropperBuilder) AutoCropArea(autoCropArea float64) (r *CropperBuilder) {
	b.tag.Attr(":auto-crop-area", autoCropArea)
	return b
}

func (b *CropperBuilder) Alt(v string) (r *CropperBuilder) {
	b.tag.Attr(":alt", h.JSONString(v))
	return b
}

func (b *CropperBuilder) Value(v Value) (r *CropperBuilder) {
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
