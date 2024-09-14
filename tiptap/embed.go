package tiptap

import (
	"bytes"
	"embed"

	"github.com/qor5/web/v3"
)

//go:embed embed
var box embed.FS

func ThemeGithubCSSComponentsPack() web.ComponentsPack {
	var css [][]byte
	custom, err := box.ReadFile("embed/github.css")
	if err != nil {
		panic(err)
	}
	css = append(css, custom)
	return web.ComponentsPack(bytes.Join(css, []byte("\n\n")))
}
