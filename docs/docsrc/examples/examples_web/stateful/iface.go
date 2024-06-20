package stateful

import h "github.com/theplant/htmlgo"

type Named interface {
	h.HTMLComponent
	CompoName() string
}

type Unwrapable interface {
	Unwrap() h.HTMLComponent
}

func Unwrap(c h.HTMLComponent) h.HTMLComponent {
	if u, ok := c.(Unwrapable); ok {
		return Unwrap(u.Unwrap())
	}
	return c
}
