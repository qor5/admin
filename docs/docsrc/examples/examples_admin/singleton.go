package examples_admin

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"gorm.io/gorm"
)

func SingletonExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return singletonExample(b, db, nil)
}

type WithSingletenProduct struct {
	gorm.Model
	Title string
}

func singletonExample(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder)) http.Handler {
	b.DataOperator(gorm2op.DataOperator(db))

	err := db.AutoMigrate(&WithSingletenProduct{})
	if err != nil {
		panic(err)
	}

	mb := b.Model(&WithSingletenProduct{}).Singleton(true)
	mb.Editing("Title")
	if customize != nil {
		customize(mb)
	}
	return b
}
