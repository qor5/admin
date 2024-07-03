package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qor5/admin/v3/docs/docsrc"
	"github.com/qor5/admin/v3/docs/docsrc/assets"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/web/v3/examples"
	"github.com/theplant/docgo"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbParamsString = osenv.Get("DB_PARAMS", "database connection string", "user=docs password=docs dbname=docs sslmode=disable host=localhost port=6532 TimeZone=Asia/Tokyo")
	port           = osenv.Get("PORT", "The port to serve on", "8800")
	envString      = osenv.Get("ENV", "environment flag", "development")
)

func main() {
	db, err := gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	go runAtMidNight(db)

	// @snippet_begin(HelloWorldMuxSample1)
	mux := http.NewServeMux()
	// @snippet_end
	examples.Mux(mux)
	examples_admin.Mux(mux)

	mux.Handle("/", docgo.New().
		MainPageTitle("QOR5 Document").
		Assets("/assets/", assets.Assets).
		DocTree(docsrc.DocTree...).
		Build(),
	)

	// @snippet_begin(HelloWorldMainSample)
	fmt.Println("Starting docs at :" + port)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
	// @snippet_end
}

func runAtMidNight(db *gorm.DB) {
	if envString == "development" {
		return
	}

	t := time.Tick(time.Hour)
	for range t {
		if time.Now().Hour() == 0 {
			truncateAllTables(db)
		}
	}
}

func truncateAllTables(db *gorm.DB) {
	if err := db.Exec(`DO
$do$
BEGIN
    EXECUTE
   (SELECT 'TRUNCATE TABLE ' || string_agg(oid::regclass::text, ', ') || ' CASCADE'
    FROM   pg_class
    WHERE  relkind = 'r'  -- only tables
    AND    relnamespace = 'public'::regnamespace
   );
END
$do$;`).
		Error; err != nil {
		panic(err)
	}
}
