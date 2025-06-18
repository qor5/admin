package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// TestProfileLanguageWithoutIcons tests the language selector without icons
// This covers the commented icon functionality to maintain test coverage
func TestProfileLanguageWithoutIcons(t *testing.T) {
	testUser := &ProfileUser{
		Model:   gorm.Model{ID: 1},
		Email:   "test@theplant.jp",
		Name:    "Test User",
		Avatar:  "https://i.pravatar.cc/300",
		Role:    "Admin",
		Status:  "Active",
		Company: "Test Corp",
	}

	case1 := multipartestutils.TestCase{
		Name:  "Language Selector Without Icons",
		Debug: true,
		HandlerMaker: func() http.Handler {
			pb := presets.New()
			// Test multiple languages to trigger the language selector
			pb.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)
			return profileExample(pb, TestDB, testUser, func(pb *plogin.ProfileBuilder) {
				pb.DisableNotification(true).LogoutURL("auth/logout")
			})
		},
		ReqFunc: func() *http.Request {
			return httptest.NewRequest("GET", "/", http.NoBody)
		},
		ExpectPageBodyContainsInOrder: []string{
			"ProfileCompo",
			"vx-select", // Language selector component
			"English",   // English language option
			"简体中文",      // Chinese language option
			"日本語",       // Japanese language option
		},
		ExpectPageBodyNotContains: []string{
			// These icon-related elements were commented out
			"prepend-inner",          // Icon slot that was commented out
			"item.raw.Icon",          // Icon reference that was commented out
			"selectedItems[0]?.Icon", // Icon display that was commented out
		},
	}

	multipartestutils.RunCase(t, case1, case1.HandlerMaker())
}
