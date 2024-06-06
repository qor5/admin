package main

import (
	"github.com/qor5/admin/v3/docs/docsrc"
	"github.com/qor5/admin/v3/docs/docsrc/assets"
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
