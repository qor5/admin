package containers

import (
	"strings"

	"github.com/qor5/admin/v3/media/media_library"

	. "github.com/theplant/htmlgo"
)

const (
	Blue   = "blue"
	Orange = "orange"
	White  = "white"
	Grey   = "grey"
)

var BackgroundColors = []string{White, Grey, Blue}

var FontColors = []string{Blue, Orange, White}

const LINK_ARROW_SVG = RawHTML(`<svg height=".7em" viewBox="0 0 10 12" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M0 11.4381V0.561882C0 0.315846 0.133038 0.14941 0.399113 0.0625736C0.67997 -0.0387357 0.938655 -0.017027 1.17517 0.1277L9.31264 4.99053C9.51959 5.12078 9.68219 5.26551 9.80044 5.42471C9.93348 5.58391 10 5.77929 10 6.01085C10 6.24242 9.93348 6.4378 9.80044 6.597C9.68219 6.74173 9.51959 6.87922 9.31264 7.00947L1.17517 11.8723C0.938655 12.017 0.67997 12.0387 0.399113 11.9374C0.133038 11.8361 0 11.6697 0 11.4381Z" fill="currentColor"/>
</svg>`)

func ContainerWrapper(anchorID, classes,
	backgroundColor, transitionBackgroundColor, fontColor,
	imagePosition string, addTopSpace, addBottomSpace bool,
	style string, comp ...HTMLComponent,
) HTMLComponent {
	return Div(comp...).
		Id(anchorID).
		Class("container-instance").ClassIf(classes, classes != "").
		AttrIf("data-background-color", backgroundColor, backgroundColor != "").
		AttrIf("data-transition-background-color", transitionBackgroundColor, transitionBackgroundColor != "").
		AttrIf("data-font-color", fontColor, fontColor != "").
		AttrIf("data-image-position", imagePosition, imagePosition != "").
		AttrIf("data-container-top-space", "true", addTopSpace).
		AttrIf("data-container-bottom-space", "true", addBottomSpace).
		Style("position:relative;").StyleIf(style, style != "")
}

func LinkTextWithArrow(text, link string, class ...string) HTMLComponent {
	if text == "" || link == "" {
		return nil
	}
	c := "link-arrow"
	if len(class) > 0 {
		class = append(class, c)
		c = strings.Join(class, " ")
	}
	return A(Span(text), LINK_ARROW_SVG).Class(c).Href(link)
}

func LazyImageHtml(m media_library.MediaBox, class ...string) HTMLComponent {
	class = append(class, "lazyload")
	return Img("").Attr("data-src", m.URL()).Alt(m.Description).Class(class...)
}

func ImageHtml(m media_library.MediaBox, class ...string) HTMLComponent {
	return Img("").Attr("src", m.URL()).Alt(m.Description).Class(class...)
}
