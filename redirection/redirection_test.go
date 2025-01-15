package redirection

import (
	"net/http"
	"net/url"
	"testing"
)

func TestCheckURLsBatch(t *testing.T) {
	urls := []string{
		"https://www.baidu.com",
		"https://bing.com/",
		"https://taobao.com/",
		"https://jd.com/",
	}
	r := http.Request{
		Host: "localhost:6000",
		URL: &url.URL{
			Path: "/index.html",
		},
	}
	success := checkURLsBatch(&r, urls)
	if !success {
		t.Fatalf("Expected all URLs to pass, but at least one failed: %v", urls)
		return
	}
	urls = []string{
		"https://www.baidu.com",
		"https://bing.com/",
		"https://taobao.com/",
		"https://jd.com/",
		"http://localhost:6000/",
		"/redirection/index.html",
	}
	success = checkURLsBatch(&r, urls)
	if success {
		t.Fatal("Expected http://localhost:6000/ failed")
		return
	}
	if urls[5] != "http://localhost:6000/redirection/index.html" {
		t.Fatalf("Expected http://localhost:6000/ failed got: %v", urls[5])
	}
}
