package presets

import (
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type (
	CustomBuilder struct {
		b             *Builder
		hideMenu      bool
		pattern       string
		body          ComponentFunc
		pageTitleFunc func(ctx *web.EventContext) string
		layoutFunc    web.PageFunc
	}
)

func NewCustomPage() *CustomBuilder {
	r := &CustomBuilder{}
	r.layoutFunc = r.defaultLayout
	return r
}
func (c *CustomBuilder) HideMenu(v bool) *CustomBuilder {
	c.hideMenu = v
	return c
}
func (c *CustomBuilder) PageTitleFunc(v func(ctx *web.EventContext) string) *CustomBuilder {
	c.pageTitleFunc = v
	return c
}
func (c *CustomBuilder) WrapLayoutFunc(w func(web.PageFunc) web.PageFunc) *CustomBuilder {
	c.layoutFunc = w(c.layoutFunc)
	return c
}

func (c *CustomBuilder) Body(v ComponentFunc) *CustomBuilder {
	c.body = v
	return c
}
func (c *CustomBuilder) WrapBody(w func(ComponentFunc) ComponentFunc) *CustomBuilder {
	c.body = w(c.body)
	return c
}

func (c *CustomBuilder) defaultLayout(ctx *web.EventContext) (pr web.PageResponse, err error) {
	b := c.b
	b.InjectAssets(ctx)
	pt := "-"
	if c.pageTitleFunc != nil {
		pt = c.pageTitleFunc(ctx)
	}
	pr.PageTitle = pt
	pr.Body = b.defaultLayoutCompo(ctx, c.menuCompo(ctx), c.body(ctx))
	return
}

func (c *CustomBuilder) menuCompo(ctx *web.EventContext) h.HTMLComponent {
	if c.hideMenu {
		return nil
	}
	b := c.b
	return b.defaultLeftMenuComp(ctx)
}
