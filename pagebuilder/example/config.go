package example

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
)

var dbParamsString = osenv.Get("DB_PARAMS", "page builder example database connection string", "")

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

//go:embed assets/images
var containerImages embed.FS

func ConfigPageBuilder(db *gorm.DB, prefix, style string, b *presets.Builder) *pagebuilder.Builder {
	err := db.AutoMigrate(
		&containers.WebHeader{},
		&containers.WebFooter{},
		&containers.VideoBanner{},
		&containers.Heading{},
		&containers.BrandGrid{},
		&containers.ListContent{},
		&containers.ImageContainer{},
		&containers.InNumbers{},
		&containers.ContactForm{},
		&containers.PageTitle{},
		&containers.ListContentLite{},
		&containers.ListContentWithImage{},
	)
	if err != nil {
		panic(err)
	}
	pb := pagebuilder.New(prefix, db, b).AutoMigrate()
	if style != "" {
		pb.PageStyle(h.RawHTML(style))
	}
	pb.GetPresetsBuilder().ExtraAsset("/tiptap.css", "text/css", tiptap.ThemeGithubCSSComponentsPack())

	fSys, _ := fs.Sub(containerImages, "assets/images")
	imagePrefix := "/assets/images"
	pb.Images(http.StripPrefix(path.Join(prefix, imagePrefix), http.FileServer(http.FS(fSys))), imagePrefix)
	containers.RegisterHeader(pb)
	containers.RegisterFooter(pb)
	containers.RegisterVideoBannerContainer(pb)
	containers.RegisterHeadingContainer(pb, db)
	containers.RegisterBrandGridContainer(pb, db)
	containers.RegisterListContentContainer(pb, db)
	containers.RegisterImageContainer(pb, db)
	containers.RegisterInNumbersContainer(pb, db)
	containers.RegisterContactFormContainer(pb, db)
	containers.RegisterPageTitleContainer(pb, db)
	containers.RegisterListContentLiteContainer(pb, db)
	containers.RegisterListContentWithImageContainer(pb, db)
	return pb
}
