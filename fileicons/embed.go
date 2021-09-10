package fileicons

import (
	"embed"
	"encoding/base64"
	"fmt"
	"strings"

	h "github.com/theplant/htmlgo"
)

//go:embed icons
var icons embed.FS

var svgs = map[string]string{}

func init() {
	files, err := icons.ReadDir("icons")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		//fmt.Println(f.Name())
		svg, err := icons.ReadFile(fmt.Sprintf("icons/%s", f.Name()))
		if err != nil {
			panic(err)
		}
		filetype := strings.Split(f.Name(), ".")[0]
		svgs[filetype] = base64.StdEncoding.EncodeToString(svg)
	}

	//fmt.Println(svgs)

}

func Icon(ext string) *h.HTMLTagBuilder {
	img, ok := svgs[ext]
	if !ok {
		img = svgs["file"]
	}

	return h.Img(fmt.Sprintf("data:image/svg+xml;base64,%s", img))
}
