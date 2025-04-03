package presets

import (
	"errors"
	"net/http"

	"github.com/qor5/web/v3"
)

var ErrRecordNotFound = errors.New("record not found")

type PageRenderIface interface {
	Render(ctx *web.EventContext) (r web.PageResponse, err error)
}

type PageRenderFunc func(ctx *web.EventContext) (r web.PageResponse, err error)

func (e PageRenderFunc) Render(ctx *web.EventContext) (r web.PageResponse, err error) {
	return e(ctx)
}

type errPageRender struct {
	PageRenderFunc
	Reason string
}

func (e *errPageRender) Error() string {
	return e.Reason
}

func ErrNotFound(reason string) error {
	return &errPageRender{
		Reason: reason,
		PageRenderFunc: func(ctx *web.EventContext) (r web.PageResponse, err error) {
			ctx.W.WriteHeader(http.StatusNotFound)
			return
		},
	}
}
