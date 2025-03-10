package examples_admin

import (
	"context"
	"net/http"

	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
)

func PageBuilderCommonContainerExample(b *presets.Builder, db *gorm.DB) http.Handler {
	b.DataOperator(gorm2op.DataOperator(db))
	storage := filesystem.New("/tmp/publish")
	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:     "1",
			Name:   "John",
			Avatar: "https://i.pravatar.cc/300",
		}, nil
	}).AutoMigrate()

	puBuilder := publish.New(db, storage)
	if b.GetPermission() == nil {
		b.Permission(
			perm.New().Policies(
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			),
		)
	}
	b.Use(puBuilder)
	pb := pagebuilder.New(b.GetURIPrefix()+"/page_builder", db, b).
		AutoMigrate().
		Activity(ab).
		PreviewOpenNewTab(true).
		Publisher(puBuilder)

	commonContainer.Setup(pb, db, nil,
		commonContainer.HeroImageHorizontal,
		commonContainer.TailWindHeroList,
		commonContainer.TailWindHeroVertical,
		commonContainer.TailWindExampleHeader,
		commonContainer.TailWindExampleFooter)

	// use demo container and media etc. plugins
	b.Use(pb)
	return TestHandler(pb, b)
}
