package sitemap

import (
	"context"
	"path"
	"strings"
)

const (
	FreqNever   = "never"
	FreqYearly  = "yearly"
	FreqMonthly = "monthly"
	FreqWeekly  = "weekly"
	FreqDaily   = "daily"
	FreqHourly  = "hourly"
	FreqAlways  = "always"
)

type ContextFunc func(context.Context) []URL
type ModelInferface interface {
	Sitemap(context.Context) []URL
}

type SiteMapIndexBuilder struct {
	pathName string
	siteMaps []*SiteMapBuilder
}

type SiteMapBuilder struct {
	pathName     string
	urls         []URL
	contextFuncs []ContextFunc
	models       []ModelInferface
}

type URL struct {
	Loc        string
	LastMod    string
	Changefreq string
	Priority   float32
}

func SiteMapIndex(names ...string) (s *SiteMapIndexBuilder) {
	var namePath string
	if len(names) == 0 {
		namePath = "/sitemap.xml"
	} else {
		if names[0] == "" {
			namePath = "/sitemap.xml"
		}
		if !strings.HasPrefix(names[0], "/") {
			namePath = "/" + names[0]
		}
		if !strings.HasSuffix(names[0], ".xml") {
			namePath = names[0] + ".xml"
		}
	}

	return &SiteMapIndexBuilder{
		pathName: namePath,
	}
}

func (index *SiteMapIndexBuilder) RegisterSiteMap(sites ...*SiteMapBuilder) (s *SiteMapIndexBuilder) {
	index.siteMaps = append(index.siteMaps, sites...)
	return index
}

func (index *SiteMapIndexBuilder) ToUrl(ctx context.Context) string {
	if h, ok := ctx.Value(hostWithSchemeKey).(string); ok {
		return path.Join(h, index.pathName)
	}
	return index.pathName
}

func SiteMap(names ...string) (s *SiteMapBuilder) {
	var namePath string
	if len(names) == 0 {
		namePath = "/sitemap.xml"
	} else {
		namePath = names[0]
		if namePath == "" {
			namePath = "/sitemap.xml"
		}
		if !strings.HasPrefix(namePath, "/") {
			namePath = "/" + namePath
		}
		if !strings.HasSuffix(namePath, ".xml") {
			namePath = namePath + ".xml"
		}
	}

	return &SiteMapBuilder{
		pathName: namePath,
	}
}

func (site *SiteMapBuilder) RegisterRawString(rs ...string) (s *SiteMapBuilder) {
	for _, s := range rs {
		site.urls = append(site.urls, URL{Loc: s})
	}
	return site
}

func (site *SiteMapBuilder) RegisterURL(urls ...URL) (s *SiteMapBuilder) {
	site.urls = append(site.urls, urls...)
	return site
}


func (site *SiteMapBuilder) RegisterContextFunc(funcs ...ContextFunc) (s *SiteMapBuilder) {
	site.contextFuncs = append(site.contextFuncs, funcs...)
	return site
}

func (site *SiteMapBuilder) RegisterModel(models ...ModelInferface) (s *SiteMapBuilder) {
	site.models = append(site.models, models...)
	return site
}

func (site *SiteMapBuilder) ToUrl(ctx context.Context) string {
	if h, ok := ctx.Value(hostWithSchemeKey).(string); ok {
		return path.Join(h, site.pathName)
	}
	return site.pathName
}

func WithHost(host string, ctxs ...context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.WithValue(context.TODO(), hostWithSchemeKey, host)
	}
	return context.WithValue(ctxs[0], hostWithSchemeKey, host)
}
