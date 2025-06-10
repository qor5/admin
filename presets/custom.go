package presets

import (
	"github.com/qor5/web/v3"
)

type (
	CustomBuilder struct {
		b             *Builder
		pattern       string
		body          ComponentFunc
		menus         ComponentFunc
		pageTitleFunc func(ctx *web.EventContext) string
		layoutFunc    web.PageFunc
		web.EventsHub
	}
)

func NewCustomPage(b *Builder) *CustomBuilder {
	r := &CustomBuilder{
		b: b,
	}
	r.layoutFunc = r.defaultLayout
	r.menus = r.b.defaultLeftMenuComp
	return r
}

func (c *CustomBuilder) Menu(v ComponentFunc) *CustomBuilder {
	c.menus = v
	return c
}

func (c *CustomBuilder) WrapMenu(v func(ComponentFunc) ComponentFunc) *CustomBuilder {
	c.menus = v(c.menus)
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
	pr.Body = b.defaultLayoutCompo(ctx, c.menus(ctx), c.body(ctx))
	return
}
