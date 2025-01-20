package redirection

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3"
)

var (
	mockServer *httptest.Server
	successUrl string
	failedUrl  string
)

func TestMain(m *testing.M) {
	mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	successUrl = mockServer.URL + "/success"
	failedUrl = mockServer.URL + "/failure"
	defer mockServer.Close()
	m.Run()
}

func TestCheckURLsBatch(t *testing.T) {
	// Create a mock server

	// Define test URLs
	urls := map[string][]string{
		successUrl: {},
		failedUrl:  {},
	}

	// Run the function
	failedURLs := checkURLsBatch(urls)

	// Verify the results
	expectedFailed := []string{failedUrl}
	if len(failedURLs) != len(expectedFailed) {
		t.Errorf("Expected %d failed URLs, got %v", len(expectedFailed), failedURLs)
	}

	for i, failedURL := range failedURLs {
		if failedURL != expectedFailed[i] {
			t.Errorf("Expected failed URL %s, got %s", expectedFailed[i], failedURL)
		}
	}
}

type (
	CheckItems struct {
		Name   string
		Item   Redirection
		Except bool
	}
)

func TestCheckRecords(t *testing.T) {
	items := []CheckItems{
		{Name: "Source Has Http Prefix", Item: Redirection{Source: successUrl, Target: "/index2.html"}, Except: false},
		{Name: "Target is UnReachable", Item: Redirection{Source: "/3/index.html", Target: failedUrl}, Except: false},
		{Name: "Target is Reachable", Item: Redirection{Source: "/3/index.html", Target: successUrl}, Except: true},
		{Name: "Source Invalid Format", Item: Redirection{Source: "3/index.html", Target: failedUrl}, Except: false},
		{Name: "Target Invalid Format", Item: Redirection{Source: "/3/index.html", Target: "index2.html"}, Except: false},
	}
	var (
		passed bool
		b      = Builder{}
		r      web.EventResponse
	)
	for _, item := range items {
		t.Run(item.Name, func(t *testing.T) {
			passed = b.checkRecords(&r, Messages_en_US, []Redirection{item.Item})
			if passed != item.Except {
				t.Errorf("Expected %t, got %t", item.Except, passed)
			}
		})
	}
}
