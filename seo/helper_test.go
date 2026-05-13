package seo

import (
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/qor5/admin/v3/l10n"
	"gorm.io/gorm"
	"github.com/qor5/x/v3/gormx"
	"gorm.io/driver/postgres"
	"context"
)

var dbForTest *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := gormx.OpenContainer(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()
	dbForTest, err = gorm.Open(postgres.Open(pgContainer.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dbForTest.AutoMigrate(&QorSEOSetting{})
	if err != nil {
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
