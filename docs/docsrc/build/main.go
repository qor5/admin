package main

import (
	"github.com/qor5/docs/v3/docsrc"
	"github.com/qor5/docs/v3/docsrc/assets"
	"github.com/theplant/docgo"
)

func main() {
	docgo.New().
		Assets("/assets/", assets.Assets).
		MainPageTitle("QOR5 Document").
		SitePrefix("/docs/").
		DocTree(docsrc.DocTree...).
		BuildStaticSite("../docs")
}
