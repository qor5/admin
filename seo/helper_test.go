package seo

import (
	"strings"

	_ "github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestQorSEOSetting struct {
	QorSEOSetting
}

func init() {
	if db, err := gorm.Open(sqlite.Open("/tmp/seo_helper.db"), &gorm.Config{}); err != nil {
		panic(err)
	} else {
		GlobalDB = db
	}
	GlobalDB.AutoMigrate(&TestQorSEOSetting{})
}

// @snippet_begin(SeoModelExample)
type Product struct {
	Name string
	SEO  Setting
}

// @snippet_end

func resetDB() {
	GlobalDB.Exec("truncate test_qor_seo_settings;")
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
