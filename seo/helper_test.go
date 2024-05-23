package seo

import (
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/qor5/admin/v3/l10n"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

var dbForTest *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	dbForTest = env.DB
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
