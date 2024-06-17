package compo

import (
	"context"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type Reloadable interface {
	h.HTMLComponent
	CompoName() string
}

type portalize struct {
	c        Reloadable
	children []h.HTMLComponent
}

type skipPortalNameCtxKey struct{}

func skipPortalize(c Reloadable) h.HTMLComponent {
	return h.ComponentFunc(func(ctx context.Context) (r []byte, err error) {
		ctx = context.WithValue(ctx, skipPortalNameCtxKey{}, c.CompoName())
		return c.MarshalHTML(ctx)
	})
}

func (p *portalize) MarshalHTML(ctx context.Context) ([]byte, error) {
	compoName := p.c.CompoName()
	skipName, _ := ctx.Value(skipPortalNameCtxKey{}).(string)
	if skipName == compoName {
		return h.Components(p.children...).MarshalHTML(ctx)
	}
	return web.Portal(p.children...).Name(compoName).MarshalHTML(ctx)
}

func Reloadify[T Reloadable](c T, children ...h.HTMLComponent) h.HTMLComponent {
	return &portalize{
		c:        c,
		children: children,
	}
}

const (
	actionMethodReload = "OnReload"
)

func ReloadAction[T Reloadable](c T, f func(cloned T)) *web.VueEventTagBuilder {
	cloned := MustClone(c)
	if f != nil {
		f(cloned)
	}
	return PlaidAction(cloned, actionMethodReload, struct{}{})
}

func ApplyReloadToResponse(r *web.EventResponse, c Reloadable) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: c.CompoName(),
		Body: skipPortalize(c),
	})
}

func OnReload(c Reloadable) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: c.CompoName(),
		Body: skipPortalize(c),
	})
	return
}
