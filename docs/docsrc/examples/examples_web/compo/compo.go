package compo

import h "github.com/theplant/htmlgo"

type Named interface {
	h.HTMLComponent
	CompoName() string
}
