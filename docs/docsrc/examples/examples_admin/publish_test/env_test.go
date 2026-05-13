package publish_test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"gorm.io/gorm"
	"github.com/qor5/x/v3/gormx"
	"gorm.io/driver/postgres"
	"context"
)

var (
	DB             *gorm.DB
	SQLDB          *sql.DB
	PresetsBuilder *presets.Builder
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	pgContainer, err := gormx.OpenContainer(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()
	DB, err = gorm.Open(postgres.Open(pgContainer.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
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
