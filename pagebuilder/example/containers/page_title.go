package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type PageTitle struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	HeroImage          media_library.MediaBox `sql:"type:text;"`
	NavigationLink     string
	NavigationLinkText string
	HeadingIcon        string
	Heading            string
	Text               string
	Tags               Tags
}

type Tags []*tag

func (this Tags) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *Tags) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func (*PageTitle) TableName() string {
	return "container_page_title"
}

func RegisterPageTitleContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("PageTitle").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*PageTitle)
			return PageTitleBody(v, input)
		})
	mb := vb.Model(&PageTitle{})

	eb := mb.Editing(
		"AddTopSpace", "AddBottomSpace", "AnchorID",
		"HeroImage", "NavigationLink", "NavigationLinkText",
		"HeadingIcon", "Heading", "Text", "Tags",
	)

	SetTagComponent(pb, eb)
}

func PageTitleBody(data *PageTitle, input *pagebuilder.RenderInput) (body HTMLComponent) {
	image := Div().Class("container-page_title-background").Style(fmt.Sprintf("background-image: url(%s)", data.HeroImage.URL()))
	wraper := Div(
		Div().Class("container-page_title-corner"),
		Div(
			Div(
				Div(
					If(data.NavigationLinkText != "", A(
						RawHTML(`<svg height=".72em" viewBox="0 0 12 15" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M10 2L3 7.5L10 13" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"/></svg>`),
						Span(data.NavigationLinkText),
					).Class("container-page_title-navigation").AttrIf("href", data.NavigationLink, data.NavigationLink != "")),
					Div(
						If(data.HeadingIcon != "", Div(RawHTML(data.HeadingIcon)).Class("container-page_title-icon")),
						H1(data.Heading),
					).Class("container-page_title-title"),
					If(data.Text != "", P(Text(data.Text)).Class("container-page_title-content p-large")),
				).Class("container-page_title-heading"),
				If(len(data.Tags) > 0, PageTitleTagsBody(data.Tags)),
			).Class("container-page_title-inner").AttrIf("data-has-navigation", "true", data.NavigationLinkText != "").AttrIf("data-has-icon", "true", data.HeadingIcon != ""),
		).Class("container-wrapper"),
	).Class("container-page_title-wrap")

	body = ContainerWrapper(
		data.AnchorID, "container-page_title container-lottie",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, "",
		image, wraper,
	)
	return
}

func PageTitleTagsBody(tags Tags) HTMLComponent {
	tagsDiv := Div().Class("container-page_title-tags-list")
	for _, t := range tags {
		tagsDiv.AppendChildren(
			A(
				getTagIconSVG(t.Icon),
				Span(t.Text),
			).Class("container-page_title-tags-item").
				AttrIf("href", t.Link, t.Link != "").
				Attr("data-font-color", t.FontColor).
				Attr("data-background-color", t.BackgroundColor),
		)
	}
	return Div(tagsDiv).Class("container-page_title-tags")
}
