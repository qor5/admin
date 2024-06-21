package example

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers"
	"github.com/qor5/admin/v3/richeditor"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func ConfigPageBuilder(db *gorm.DB, prefix, style string, i18nB *i18n.Builder) *pagebuilder.Builder {
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
	pb := pagebuilder.New(prefix, db, i18nB)
	if style != "" {
		pb.PageStyle(h.RawHTML(style))
	}

	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert"}
	pb.GetPresetsBuilder().ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	pb.GetPresetsBuilder().ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())

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
