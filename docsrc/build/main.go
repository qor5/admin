package main

import (
	"github.com/qor/qor5/docsrc"
	"github.com/theplant/docgo"
)

func main() {
	docgo.New().
		Assets("/assets/", docsrc.Assets).
		Home(docsrc.Home).
		SitePrefix("/qor5/").
		BuildStaticSite("../docs")
}
