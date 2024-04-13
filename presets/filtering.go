package presets

import (
	"net/url"

	"github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
)

func (b *ListingBuilder) FilterDataFunc(v FilterDataFunc) {
	if v == nil {
		b.filterDataFunc = nil
		return
	}

	b.filterDataFunc = func(ctx *web.EventContext) vuetifyx.FilterData {
		fd := v(ctx)
		for _, k := range fd {
			k.Key = "f_" + k.Key
		}
		return fd
	}
}

func (b *ListingBuilder) FilterTabsFunc(v FilterTabsFunc) {
	if v == nil {
		b.filterTabsFunc = nil
		return
	}

	b.filterTabsFunc = func(ctx *web.EventContext) []*FilterTab {
		fts := v(ctx)
		for _, ft := range fts {
			newQuery := make(url.Values)
			for k, _ := range ft.Query {
				newQuery["f_"+k] = ft.Query[k]
			}
			ft.Query = newQuery
		}
		return fts
	}
}
