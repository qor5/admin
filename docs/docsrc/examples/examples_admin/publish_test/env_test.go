package publish_test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

var (
	DB             *gorm.DB
	SQLDB          *sql.DB
	PresetsBuilder *presets.Builder
)

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()

	DB = env.DB
	SQLDB, err = DB.DB()
	if err != nil {
		panic(err)
	}
	PresetsBuilder = presets.New().DataOperator(gorm2op.DataOperator(DB)).URIPrefix("/examples/publish-example")
	examples_admin.PublishExample(PresetsBuilder, DB)

	m.Run()
}

type Flow struct {
	db *gorm.DB
	h  http.Handler
}
