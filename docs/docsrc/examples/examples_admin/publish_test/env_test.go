package publish_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/x/v3/gormx"
	"gorm.io/gorm"
)

var (
	DB             *gorm.DB
	SQLDB          *sql.DB
	PresetsBuilder *presets.Builder
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	testSuite := gormx.MustStartTestSuite(ctx)
	defer func() {
		if err := testSuite.Stop(context.Background()); err != nil {
			fmt.Printf("Error during teardown: %v\n", err)
		}
	}()

	DB = testSuite.DB()
	var err error
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
