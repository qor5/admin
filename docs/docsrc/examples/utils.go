package examples

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func URLPathByFunc(v interface{}) string {
	funcNameWithPkg := runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
	segs := strings.Split(funcNameWithPkg, ".")
	return "/samples/" + strcase.ToKebab(segs[len(segs)-1])
}

type Muxer interface {
	Handle(pattern string, handler http.Handler)
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

func AddPresetExample(mux Muxer, f func(*presets.Builder, *gorm.DB) http.Handler) {
	path := URLPathByFunc(f)
	fmt.Println("Examples mounting path:", path)
	p := presets.New().AssetFunc(AddGA).URIPrefix(path)
	mux.Handle(path, f(p, ExampleDB()))
}
