package containers

import (
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"

	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type VideoBanner struct {
	ID                    uint
	AddTopSpace           bool
	AddBottomSpace        bool
	AnchorID              string
	Video                 media_library.MediaBox `sql:"type:text;"`
	BackgroundVideo       media_library.MediaBox `sql:"type:text;"`
	MobileBackgroundVideo media_library.MediaBox `sql:"type:text;"`
	VideoCover            media_library.MediaBox `sql:"type:text;"`
	MobileVideoCover      media_library.MediaBox `sql:"type:text;"`
	Heading               string
	PopupText             string
	Text                  string
	LinkText              string
	Link                  string
}

func (*VideoBanner) TableName() string {
	return "container_video_banners"
}

func RegisterVideoBannerContainer(pb *pagebuilder.Builder) {
	vb := pb.RegisterContainer("Video Banner").Group("Content").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*VideoBanner)
			return VideoBannerBody(v, input)
		})
	ed := vb.Model(&VideoBanner{}).Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Video", "BackgroundVideo", "MobileBackgroundVideo", "VideoCover", "MobileVideoCover", "Heading", "PopupText", "Text", "LinkText", "Link")
	ed.Field("Heading").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("textarea")
	})
	ed.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("textarea")
	})
}

func VideoBannerBody(data *VideoBanner, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		data.AnchorID, "container-video_banner",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, "",
		Div().Class("container-video_banner-mask"), VideoBannerHeadBody(data), VideoBannerFootBody(data),
		// If(data.PopupText != "", VideoBannerPopupBody(data)),
	)
	return
}

func VideoBannerHeadBody(data *VideoBanner) HTMLComponent {
	return Div(
		Div().Class("container-video_banner-background container-video_banner-background-image"),
		Video(
			Source("").Src(data.BackgroundVideo.URL()),
		).Class("container-video_banner-background container-video_banner-background-desktop").
			Attr("preload", "none").Attr("loop", "true").Attr("muted", "true").Attr("playsinline", "true").Attr("webkit-playsinline", "true").Attr("data-cover-image-url", data.VideoCover.URL()),
		Video(
			Source("").Src(data.MobileBackgroundVideo.URL()),
		).Class("container-video_banner-background container-video_banner-background-mobile").
			Attr("preload", "none").Attr("loop", "true").Attr("muted", "true").Attr("playsinline", "true").Attr("webkit-playsinline", "true").Attr("data-cover-image-url", data.MobileVideoCover.URL()),
		Div(
			If(data.Heading != "", H1(data.Heading).Class("container-video_banner-heading")),
			// 	If(data.PopupText != "", A(Span(data.PopupText), LINK_ARROW_SVG).Class("container-video_banner-full link-arrow")),
		).Class("container-video_banner-head-wrap container-wrapper").Style("display:none;"),
	).Class("container-video_banner-head")
}

func VideoBannerFootBody(data *VideoBanner) HTMLComponent {
	return Div(
		Div(
			P(Text(data.Text)).Class("container-video_banner-text p-large"),
			LinkTextWithArrow(data.LinkText, data.Link),
		).Class("container-wrapper"),
	).Class("container-video_banner-foot")
}
