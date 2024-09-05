package examples

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/qor5/admin/v3/autocomplete"
	"github.com/qor5/admin/v3/presets"
)

var db *gorm.DB

var dbParamsString = osenv.Get("DB_PARAMS", "database connection string", "user=docs password=docs dbname=docs sslmode=disable host=localhost port=6532 TimeZone=Asia/Tokyo")

func ExampleDB() (r *gorm.DB) {
	if db != nil {
		return db
	}
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.Logger = db.Logger.LogMode(logger.Info)
	r = db
	return
}

func AddGA(ctx *web.EventContext) {
	if strings.Index(ctx.R.Host, "localhost") >= 0 {
		return
	}
	ctx.Injector.HeadHTML(`
<!-- Global site tag (gtag.js) - Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=UA-149605708-1"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', 'UA-149605708-1');
</script>
`)
}

func AddPresetExample(mux examples.Muxer, f func(*presets.Builder, *gorm.DB) http.Handler) {
	path := examples.URLPathByFunc(f)
	fmt.Println("Examples mounting path:", path)
	p := presets.New().AssetFunc(AddGA).URIPrefix(path)
	mux.Handle(path, f(p, ExampleDB()))
}

func AddPresetAutocompleteExample(mux examples.Muxer, f func(*presets.Builder, *autocomplete.Builder, *gorm.DB) http.Handler) {
	ab := autocomplete.New().Prefix("/examples/api/complete")
	// mux.Handle("/examples/api/complete", ab)
	path := examples.URLPathByFunc(f)
	fmt.Println("Examples mounting path:", path)
	p := presets.New().AssetFunc(AddGA).URIPrefix(path)
	mux.Handle(path, f(p, ab, ExampleDB()))
	// mux after models add
	ab.Mux(mux)
}

func AddPresetsLinkageSelectFilterItemRemoteExample(mux examples.Muxer, f func(*presets.Builder, examples.Muxer, *gorm.DB) http.Handler) {
	path := examples.URLPathByFunc(f)
	fmt.Println("Examples mounting path:", path)
	p := presets.New().AssetFunc(AddGA).URIPrefix(path)
	mux.Handle(path, f(p, mux, ExampleDB()))
}
