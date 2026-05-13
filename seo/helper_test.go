package seo

import (
	"context"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/x/v3/gormx"
	"gorm.io/gorm"
)

var dbForTest *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	suite := gormx.MustStartRawTestSuite(ctx)
	defer func() { _ = suite.Stop(ctx) }()
	dbForTest = suite.DB()
	if err := dbForTest.AutoMigrate(&QorSEOSetting{}); err != nil {
		panic("failed to migrate db")
	}

	code := m.Run()
	resetDB()
	os.Exit(code)
}

// @snippet_begin(SeoModelExample)
type Product struct {
	Name string
	SEO  Setting
	l10n.Locale
}

// @snippet_end

func resetDB() {
	dbForTest.Exec("truncate qor_seo_settings;")
}

func metaEqual(got, want string) bool {
	for _, s := range strings.Split(want, "\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.Contains(got, s) {
			return false
		}
	}
	return true
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
