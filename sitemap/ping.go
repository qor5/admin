package sitemap

import (
	"context"
	"fmt"
	"net/http"
)

type ToUrlInterface interface {
	ToUrl(context.Context) string
}

func PingBing(site ToUrlInterface, ctx context.Context) (err error) {
	_, err = http.Get(fmt.Sprintf("http://www.bing.com/webmaster/ping.aspx?siteMap=%s", site.ToUrl(ctx)))
	return
}

func PingGoogle(site ToUrlInterface, ctx context.Context) (err error) {
	_, err = http.Get(fmt.Sprintf("https://www.google.com/webmasters/sitemaps/ping?sitemap=%s", site.ToUrl(ctx)))
	return
}

func PingAll(site ToUrlInterface, ctx context.Context) (err error) {
	if err = PingGoogle(site, ctx); err != nil {
		return
	}

	if err = PingBing(site, ctx); err != nil {
		return
	}

	return
}
