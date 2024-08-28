package examples_admin

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
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
	eb := mb.Editing("Title")
	eb.WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			r, err = in(obj, id, ctx)
			if err != nil {
				return
			}
			return
		}
	})
	// eb.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (interface{}, error) {
	// 	return gorm2op.DataOperator(db).Fetch(obj, id, ctx)
	// })
	// eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
	// 	return errors.New("测试错误")
	// })
	if customize != nil {
		customize(mb)
	}
	return b
}
