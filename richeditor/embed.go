package richeditor

import (
	"bytes"
	"embed"

	"github.com/goplaid/web"
)

//go:embed redactor
var box embed.FS

func JSComponentsPack() web.ComponentsPack {
	v1, err := box.ReadFile("redactor/redactor.min.js")
	if err != nil {
		panic(err)
	}
	v2, err := box.ReadFile("redactor/vue-redactor.js")
	if err != nil {
		panic(err)
	}

	return web.ComponentsPack(bytes.Join([][]byte{v1, v2}, []byte("\n\n")))
}

func CSSComponentsPack() web.ComponentsPack {
	v, err := box.ReadFile("redactor/redactor.min.css")
	if err != nil {
		panic(err)
	}

	return web.ComponentsPack(v)
}
