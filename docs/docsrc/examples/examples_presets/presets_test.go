package examples_presets

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	TestDB *gorm.DB
	SqlDB  *sql.DB
)

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	SqlDB, _ = TestDB.DB()
	m.Run()
}

func TestPresetsCommon(t *testing.T) {
	err := TestDB.AutoMigrate(&Customer{})
	if err != nil {
		panic(err)
	}
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	pb.Model(&Customer{})

	// dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Not Found",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/samples/publish/products", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"page cannot be found"},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusNotFound {
					t.Errorf("Expected HTTP 404, got %v", w.Code)
				}
				if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
					t.Errorf("Expected text/html; charset=utf-8, got %v", w.Header().Get("Content-Type"))
				}
			},
		},
		{
			Name:  "Not Found Assets",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/assets/vuetify.min.js.map", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"404 page not found"},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusNotFound {
					t.Errorf("Expected HTTP 404, got %v", w.Code)
				}
			},
		},
		{
			Name:  "Found",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers", http.NoBody)
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusOK {
					t.Errorf("Expected HTTP 200, got %v", w.Code)
				}
				if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
					t.Errorf("Expected text/html; charset=utf-8, got %v", w.Header().Get("Content-Type"))
				}
			},
		},

		{
			Name: "javascript content type is still correct",
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/assets/main.js", http.NoBody)
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusOK {
					t.Errorf("Expected HTTP 200, got %v", w.Code)
				}
				if w.Header().Get("Content-Type") != "text/javascript" {
					t.Errorf("Expected text/javascript, got %v", w.Header().Get("Content-Type"))
				}
			},
		},

		{
			Name: "Ending slash is redirected",
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers/", http.NoBody)
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusMovedPermanently {
					t.Errorf("Expected HTTP 301, got %v", w.Code)
				}
				if w.Header().Get("Location") != "//example.com/customers" {
					t.Errorf("Expected //example.com/customers, got %v", w.Header().Get("Location"))
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
