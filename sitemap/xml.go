package sitemap

import (
	"context"
	"fmt"
	neturl "net/url"
	"path"
	"strings"
)

type EncodeToXmlInterface interface {
	EncodeToXml(ctx context.Context) string
}

func (s SiteMapBuilder) EncodeToXml(ctx context.Context) string {
	b := strings.Builder{}
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	var hostWithScheme string
	if h, ok := ctx.Value(hostWithSchemeKey).(string); ok {
		hostWithScheme = h
	}

	var urls = make([]URL, len(s.urls))
	copy(urls, s.urls)

	for _, contextfunc := range s.contextFuncs {
		urls = append(urls, contextfunc(ctx)...)
	}

	for _, url := range urls {
		u, err := neturl.Parse(url.Loc)
		if err != nil {
			continue
		}

		if u.Host == "" {
			url.Loc = path.Join(hostWithScheme, url.Loc)
		}

		b.WriteString(`<url>`)
		b.WriteString(fmt.Sprintf(`<loc>%s</loc>`, url.Loc))
		if url.LastMod != "" {
			b.WriteString(fmt.Sprintf(`<lastmod>%s</lastmod>`, url.LastMod))
		}
		if url.Changefreq != "" {
			b.WriteString(fmt.Sprintf(`<changefreq>%s</changefreq>`, url.Changefreq))
		}
		if url.Priority != 0.0 {
			b.WriteString(fmt.Sprintf(`<priority>%f</priority>`, url.Priority))
		}
		b.WriteString(`</url>`)
	}

	b.WriteString(`</urlset>`)
	return b.String()
}

func (s SiteMapIndexBuilder) EncodeToXml(ctx context.Context) string {
	b := strings.Builder{}
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	for _, site := range s.siteMaps {
		b.WriteString(`<sitemap>`)
		b.WriteString(fmt.Sprintf(`<loc>%s</loc>`, site.ToUrl(ctx)))
		b.WriteString(`</sitemap>`)
	}

	b.WriteString(`</sitemapindex>`)
	return b.String()
}
