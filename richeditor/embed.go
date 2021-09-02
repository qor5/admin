package richeditor

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/goplaid/web"
)

//go:embed redactor
var box embed.FS

func JSComponentsPack() web.ComponentsPack {
	var js [][]byte
	j1, err := box.ReadFile("redactor/redactor.min.js")
	if err != nil {
		panic(err)
	}
	js = append(js, j1)
	for _, p := range Plugins {
		pj, err := box.ReadFile(fmt.Sprintf("redactor/%s.min.js", p))
		if err != nil {
			continue
		}
		js = append(js, pj)
	}
	if len(PluginsJS) > 0 {
		js = append(js, PluginsJS...)
	}

	v3, err := box.ReadFile("redactor/vue-redactor.js")
	if err != nil {
		panic(err)
	}
	js = append(js, v3)
	return web.ComponentsPack(bytes.Join(js, []byte("\n\n")))
}

func CSSComponentsPack() web.ComponentsPack {
	var css [][]byte
	c, err := box.ReadFile("redactor/redactor.min.css")
	if err != nil {
		panic(err)
	}
	css = append(css, c)
	if len(PluginsCSS) > 0 {
		css = append(css, PluginsCSS...)
	}
	return web.ComponentsPack(bytes.Join(css, []byte("\n\n")))
}
