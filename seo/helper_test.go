package seo

import (
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/qor5/admin/v3/l10n"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbForTest    *gorm.DB
	testDBParams = osenv.Get("TEST_DB_PARAMS", "test database connection string", "user=test password=test dbname=test sslmode=disable host=localhost port=6432 TimeZone=Asia/Tokyo")
)

func init() {
	if db, err := gorm.Open(postgres.Open(testDBParams), &gorm.Config{}); err != nil {
		panic(err)
	} else {
		err := db.AutoMigrate(&QorSEOSetting{})
		if err != nil {
			panic("failed to migrate db")
		}
		dbForTest = db
	}
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

func TestMain(m *testing.M) {
	code := m.Run()
	resetDB()
	os.Exit(code)
}
