package sitemap

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddSitemapUrl(t *testing.T) {
	robot := Robots()
	robot.Agent(AllAgents).AddSitemapUrl(SiteMap().ToUrl(WithHost("https://qor5.dev.com")))
	s := robot.ToTxt()
	expected := "User-agent: *\nSitemap: https:/qor5.dev.com/sitemap.xml\n\n"
	if s != expected {
		t.Errorf("\n\tExpected value: \n%s \tbut got: \n%s", expected, s)
	}
}

func TestAllow(t *testing.T) {
	robot := Robots()
	robot.Agent(AllAgents).Allow("/product1", "/product2")
	s := robot.ToTxt()
	expected := "User-agent: *\nAllow: /product1\nAllow: /product2\n\n"
	if s != expected {
		t.Errorf("\n\tExpected value: \n%s \tbut got: \n%s", expected, s)
	}
}

func TestDisallow(t *testing.T) {
	robot := Robots()
	robot.Agent(AllAgents).Disallow("/product1", "/product2")
	s := robot.ToTxt()
	expected := "User-agent: *\nDisallow: /product1\nDisallow: /product2\n\n"
	if s != expected {
		t.Errorf("\n\tExpected value: \n%s \tbut got: \n%s", expected, s)
	}
}

func TestRobotsServeHTTP(t *testing.T) {
	robot := Robots()
	robot.Agent(GoogleAgent).Disallow("/admin", "/product").AddSitemapUrl(SiteMap().ToUrl(WithHost("https://qor5.dev.com")))
	robot.Agent(DuckDuckAgent).Allow("/admin1", "/product2").Disallow("/product1")

	serveMux := http.NewServeMux()
	robot.MountTo(serveMux)

	server := httptest.NewServer(serveMux)
	resp, err := http.Get(server.URL + "/robots.txt")
	if err != nil {
		t.Error(err)
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	expected := "User-agent: Googlebot\nDisallow: /admin\nDisallow: /product\nSitemap: https:/qor5.dev.com/sitemap.xml\n\nUser-agent: DuckDuckBot\nDisallow: /product1\nAllow: /admin1\nAllow: /product2\n\n"
	if string(s) != expected {
		t.Errorf("\n\tExpected value: \n%s \tbut got: \n%s", expected, s)
	}
}
