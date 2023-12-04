package seo

import (
	"embed"

	"github.com/qor5/web"
)

//go:embed v_seo
var box embed.FS

func SeoJSComponentsPack() web.ComponentsPack {
	c, err := box.ReadFile("v_seo/vue-seo.min.js")
	if err != nil {
		panic(err)
	}
	return web.ComponentsPack(c)
}
