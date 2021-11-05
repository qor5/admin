package sitemap

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRegisterRawString(t *testing.T) {
	s := SiteMap().RegisterRawString("/admin").EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/admin</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestRegisterURL(t *testing.T) {
	s := SiteMap().RegisterURL(URL{Loc: "/admin"}).EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/admin</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestRegisterURLAndRawString(t *testing.T) {
	s := SiteMap().RegisterRawString("/admin1").RegisterURL(URL{Loc: "/admin2"}).EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/admin1</loc></url><url><loc>/admin2</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestRegisterContextFunc(t *testing.T) {
	s := SiteMap().RegisterContextFunc(func(context.Context) []URL {
		return []URL{{Loc: "/admin1"}, {Loc: "/admin2"}}
	}).EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/admin1</loc></url><url><loc>/admin2</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

type post struct {
}

func (p post) Sitemap(ctx context.Context) []URL {
	return []URL{
		{Loc: "/post1"},
		{Loc: "/post2"},
	}
}
func TestRegisterModel(t *testing.T) {
	s := SiteMap().RegisterModel(post{}).EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/post1</loc></url><url><loc>/post2</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestSiteMapIndex(t *testing.T) {
	s := SiteMapIndex().RegisterSiteMap(SiteMap(), SiteMap("product"), SiteMap("admin")).EncodeToXml(context.TODO())
	expected := `<?xml version="1.0" encoding="UTF-8"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><sitemap><loc>/sitemap.xml</loc></sitemap><sitemap><loc>/product.xml</loc></sitemap><sitemap><loc>/admin.xml</loc></sitemap></sitemapindex>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestEncodeToXmlWithContext(t *testing.T) {
	s := SiteMap().RegisterRawString("/admin", "https://qor5-1.dev.com/product").EncodeToXml(WithHost("https://qor5.dev.com"))
	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https:/qor5.dev.com/admin</loc></url><url><loc>https://qor5-1.dev.com/product</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestRequestHost(t *testing.T) {
	u, _ := url.Parse("https://qor5.dev.com/sitemap.xml")
	s := EncodeToXmlByRequest(&http.Request{URL: u}, SiteMap().RegisterRawString("/admin", "https://qor5-1.dev.com/product"))

	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https:/qor5.dev.com/admin</loc></url><url><loc>https://qor5-1.dev.com/product</loc></url></urlset>`
	if s != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}

func TestServeHTTP(t *testing.T) {
	site := SiteMap().RegisterRawString("/admin", "https://qor5-1.dev.com/product")
	serveMux := http.NewServeMux()
	site.MountTo(serveMux)

	server := httptest.NewServer(serveMux)
	resp, err := http.Get(server.URL + "/sitemap.xml")
	if err != nil {
		t.Error(err)
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	expected := `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/admin</loc></url><url><loc>https://qor5-1.dev.com/product</loc></url></urlset>`
	if string(s) != expected {
		t.Errorf("\n\tExpected value: %s\n \tbut got: %s", expected, s)
	}
}
