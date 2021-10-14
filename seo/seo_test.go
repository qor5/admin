package seo

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/goplaid/x/presets"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db         *gorm.DB
	collection *Collection
)

func init() {
	db = initDB()
	db.AutoMigrate(&QorSEOSetting{})
	collection = New("Site SEO")
	collection.RegisterGlobalVariables(&SEOGlobalSetting{SiteName: "Qor Shop"})
	collection.RegisterSettingModel(&QorSEOSetting{})
	collection.RegisterSEO(&SEO{
		Name:      "Product",
		Variables: []string{"Name", "Code"},
		Context: func(objects ...interface{}) map[string]string {
			context := make(map[string]string)
			context["Name"] = "name"
			context["Code"] = "code"
			return context
		},
	})
}

func initDB() (db *gorm.DB) {
	var err error
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{}); err != nil {
		panic(err)
	}

	return
}

type Product struct {
	Name string
	SEO  Setting
}

func (product Product) GetSEO() *SEO {
	return collection.GetSEO("Product")
}

type RenderTestCase struct {
	SiteName   string
	SeoSetting Setting
	Settings   []interface{}
	Result     string
}

type SEOGlobalSetting struct {
	SiteName string
}

func TestSaveSEOSetting(t *testing.T) {
	admin := presets.New().URIPrefix("/admin")
	collection.Configure(admin, db)
	admin.Model(&QorSEOSetting{})
	server := httptest.NewServer(admin)

	// create
	db.Exec("truncate qor_seo_settings;")
	if req, err := http.Get(server.URL + "/admin/qor-seo-settings?__execute_event__=__reload__"); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Setting page should be exist, status code is %v", req.StatusCode)
		}

		var seoSetting QorSEOSetting
		err := db.First(&seoSetting, "name = ?", "Product").Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("SEO Setting should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}

	// update
	title := "title test"
	description := "description test"
	keyword := "keyword test"
	var form = &bytes.Buffer{}
	mwriter := multipart.NewWriter(form)
	mwriter.WriteField("__event_data__", `{"eventFuncId":{"id":"seo_save_collection","params":["Product","global_seo_loading"],"pushState":null},"event":{}}`)
	mwriter.WriteField("Product.Title", title)
	mwriter.WriteField("Product.Description", description)
	mwriter.WriteField("Product.Keywords", keyword)
	mwriter.Close()

	req, err := http.DefaultClient.Post(server.URL+"/admin/qor-seo-settings?__execute_event__=seo_save_collection", mwriter.FormDataContentType(), form)
	if err != nil {
		t.Fatal(err)
	}

	if req.StatusCode != 200 {
		t.Errorf("Save should be processed successfully, status code is %v", req.StatusCode)
	}

	var seoSetting QorSEOSetting
	err = db.First(&seoSetting, "name = ?", "Product").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("SEO Setting should be created successfully")
	}

	if seoSetting.Setting.Title != title || seoSetting.Setting.Description != description || seoSetting.Setting.Keywords != keyword {
		t.Errorf("SEOSetting should be Save correctly, its value %#v", seoSetting)
	}
}

func TestRender(t *testing.T) {
	db.Exec("truncate qor_seo_settings;")

	db.Save(&QorSEOSetting{
		Name: "Site SEO",
		Setting: Setting{
			GlobalSetting: map[string]string{"SiteName": "Qor Shop"},
		},
		IsGlobalSEO: true,
	})

	db.Save(&QorSEOSetting{
		Name: "Product",
		Setting: Setting{
			Title:        "{{SiteName}} | Product Title",
			Description:  "Product Description | {{SiteName}}",
			Keywords:     "Product Keywords",
			OpenGraphURL: "products",
		},
	})

	req := &http.Request{Host: "localhost:9000", URL: &url.URL{}}
	html, err := collection.Render(db, req, "Product").MarshalHTML(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, c := range []struct {
		field       string
		expectation string
	}{
		{"title", "<title>Qor Shop | Product Title</title>"},
		{"description", "<meta name='description' content='Product Description | Qor Shop'>"},
		{"keywords", "<meta name='keywords' content='Product Keywords'>"},
		{"og:url", "<meta property='og:url' name='og:url' content='http://localhost:9000/products'>"},
	} {
		if !strings.Contains(string(html), c.expectation) {
			t.Errorf("%s is incorrect, the rended content is %s", c.field, string(html))
		}
	}

	for _, c := range []struct {
		product   Product
		testcases []struct {
			field       string
			expectation string
		}
	}{
		{
			product: Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "{{SiteName}} | Product Detail 1",
					Description:      "product 1 description",
					Keywords:         "product 1 keywords",
					OpenGraphURL:     "products/1",
					EnabledCustomize: true,
				},
			},
			testcases: []struct {
				field       string
				expectation string
			}{
				{"title", "<title>Qor Shop | Product Detail 1</title>"},
				{"description", "<meta name='description' content='product 1 description'>"},
				{"keywords", "<meta name='keywords' content='product 1 keywords'>"},
				{"og:url", "<meta property='og:url' name='og:url' content='http://localhost:9000/products/1'>"},
			},
		},

		{
			product: Product{
				Name: "product 2",
				SEO: Setting{
					Title:            "{{SiteName}} | Product Detail 2",
					Description:      "product 2 description",
					Keywords:         "product 2 keywords",
					OpenGraphURL:     "products/2",
					EnabledCustomize: false,
				},
			},
			testcases: []struct {
				field       string
				expectation string
			}{
				{"title", "<title>Qor Shop | Product Title</title>"},
				{"description", "<meta name='description' content='Product Description | Qor Shop'>"},
				{"keywords", "<meta name='keywords' content='Product Keywords'>"},
				{"og:url", "<meta property='og:url' name='og:url' content='http://localhost:9000/products'>"},
			},
		},
	} {
		html, err := collection.Render(db, req, "Product", c.product).MarshalHTML(context.TODO())
		if err != nil {
			t.Errorf(err.Error())
		}
		for _, testcase := range c.testcases {
			if !strings.Contains(string(html), testcase.expectation) {
				t.Errorf("%s is incorrect, the rended content is %s", testcase.field, string(html))
			}
		}
	}
}
