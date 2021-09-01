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
	v1, err := box.ReadFile("redactor/redactor.min.js")
	if err != nil {
		panic(err)
	}
	js = append(js, v1)
	for _, p := range Plugins {
		v2, err := box.ReadFile(fmt.Sprintf("redactor/%s.min.js", p))
		if err != nil {
			panic(err)
		}
		js = append(js, v2)
	}

	v3, err := box.ReadFile("redactor/vue-redactor.js")
	if err != nil {
		panic(err)
	}
	js = append(js, v3)
	return web.ComponentsPack(bytes.Join(js, []byte("\n\n")))
}

func CSSComponentsPack() web.ComponentsPack {
	v, err := box.ReadFile("redactor/redactor.min.css")
	if err != nil {
		panic(err)
	}

	return web.ComponentsPack(v)
}
