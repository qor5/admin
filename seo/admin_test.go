package seo

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/goplaid/x/presets"
	"gorm.io/gorm"
)

func TestAdmin(t *testing.T) {
	var (
		admin  = presets.New().URIPrefix("/admin")
		server = httptest.NewServer(admin)
	)

	collection := NewCollection().SetSettingModel(&TestQorSEOSetting{}).RegisterSEOByNames("Product Detail", "Product")
	collection.Configure(admin, GlobalDB)

	// should create all seo setting in the first time
	resetDB()
	if req, err := http.Get(server.URL + "/admin/test-qor-seo-settings?__execute_event__=__reload__"); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Setting page should be exist, status code is %v", req.StatusCode)
		}

		var seoSetting = collection.NewSettingModelSlice()
		err := GlobalDB.Find(seoSetting, "name in (?)", []string{"Product Detail", "Product", collection.globalName}).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("SEO Setting should be created successfully")
		}

		if reflect.Indirect(reflect.ValueOf(seoSetting)).Len() != 3 {
			t.Errorf("SEO Setting should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}

	// save seo setting
	var (
		title       = "title test"
		description = "description test"
		keyword     = "keyword test"
	)

	var form = &bytes.Buffer{}
	mwriter := multipart.NewWriter(form)
	mwriter.WriteField("Product.Title", title)
	mwriter.WriteField("Product.Description", description)
	mwriter.WriteField("Product.Keywords", keyword)
	mwriter.Close()

	req, err := http.DefaultClient.Post(server.URL+"/admin/test-qor-seo-settings?__execute_event__=seo_save_collection&name=Product", mwriter.FormDataContentType(), form)
	if err != nil {
		t.Fatal(err)
	}

	if req.StatusCode != 200 {
		t.Errorf("Save should be processed successfully, status code is %v", req.StatusCode)
	}

	var seoSetting = collection.NewSettingModelInstance().(QorSEOSettingInterface)
	err = GlobalDB.First(&seoSetting, "name = ?", "Product").Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("SEO Setting should be created successfully")
	}

	if seoSetting.GetSEOSetting().Title != title || seoSetting.GetSEOSetting().Description != description || seoSetting.GetSEOSetting().Keywords != keyword {
		t.Errorf("SEOSetting should be Save correctly, its value %#v", seoSetting)
	}
}
