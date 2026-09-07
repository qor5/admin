package redirection

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/gormx"
	s3x "github.com/qor5/x/v3/oss/s3"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	mockServer *httptest.Server
	successUrl string
	failedUrl  string
	TestDB     *gorm.DB
	b          *Builder
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
	ctx := context.Background()
	testSuite := gormx.MustStartTestSuite(ctx)
	defer func() {
		if err := testSuite.Stop(context.Background()); err != nil {
			fmt.Printf("Error during teardown: %v\n", err)
		}
	}()
	TestDB = testSuite.DB()
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	b = &Builder{db: TestDB}
	b.AutoMigrate()
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

var redirectionData = gofixtures.Data(gofixtures.Sql(`
`, []string{"redirections"}))

func TestCreateEmptyTargetRecord(t *testing.T) {
	dbr, _ := TestDB.DB()
	redirectionData.TruncatePut(dbr)
	b.createEmptyTargetRecord("/index_empty.html")
	m := Redirection{}
	TestDB.Order("id desc").First(&m)
	if m.Source != "/index_empty.html" {
		t.Fatalf("create record failed source:%v", m.Source)
		return
	}
	if m.Target != "" {
		t.Fatalf("create record failed targe:%v", m.Target)
		return
	}
}

func TestCheckObjects(t *testing.T) {
	ctx := &web.EventContext{
		R: &http.Request{},
	}
	r := &web.EventResponse{}
	if !b.checkObjects(ctx, r, Messages_en_US, []Redirection{}) {
		t.Fatalf("No Objects Is Passed")
		return
	}
}

func newTestS3Client(t *testing.T, handler http.Handler) *s3x.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &s3x.Client{
		S3: awss3.New(awss3.Options{
			BaseEndpoint: aws.String(server.URL),
			UsePathStyle: true,
			Region:       "us-east-1",
			Credentials:  credentials.NewStaticCredentialsProvider("test", "test", ""),
		}),
		Config: &s3x.Config{Bucket: "test-bucket"},
	}
}

func TestCheckTargetExists(t *testing.T) {
	client := newTestS3Client(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		existing := map[string]bool{
			"/test-bucket/index.html":     true,
			"/test-bucket/503/index.html": true,
		}
		if r.Method == http.MethodHead && existing[r.URL.Path] {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	builder := &Builder{s3Client: client}

	cases := []struct {
		name   string
		target string
		expect bool
	}{
		{name: "existing object", target: "/503/index.html", expect: true},
		{name: "directory form resolves to index document", target: "/503/", expect: true},
		{name: "root resolves to index document", target: "/", expect: true},
		{name: "missing object", target: "/missing.html", expect: false},
		{name: "directory form without index document", target: "/missing/", expect: false},
		{name: "no trailing slash is not treated as directory", target: "/503", expect: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := builder.checkTargetExists(context.Background(), c.target); got != c.expect {
				t.Errorf("checkTargetExists(%q) = %t, want %t", c.target, got, c.expect)
			}
		})
	}
}

func TestRedirectionKeepsDirectoryTargetSlash(t *testing.T) {
	redirectLocations := make(map[string]string)
	client := newTestS3Client(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPut:
			redirectLocations[r.URL.Path] = r.Header.Get("x-amz-website-redirect-location")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	builder := &Builder{s3Client: client}

	cases := []struct {
		name   string
		source string
		target string
		expect string
	}{
		{name: "directory form keeps trailing slash", source: "/old/a.html", target: "/503/", expect: "/503/"},
		{name: "file target unchanged", source: "/old/b.html", target: "/503/index.html", expect: "/503/index.html"},
		{name: "absolute url unchanged", source: "/old/c.html", target: "https://example.com/x/", expect: "https://example.com/x/"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := builder.redirection(context.Background(), &Redirection{Source: c.source, Target: c.target}); err != nil {
				t.Fatalf("redirection() error: %v", err)
			}
			key := "/test-bucket" + c.source
			if got := redirectLocations[key]; got != c.expect {
				t.Errorf("WebsiteRedirectLocation for %s = %q, want %q", c.source, got, c.expect)
			}
		})
	}
}
