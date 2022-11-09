package example

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/media/oss"
	media_view "github.com/qor5/admin/media/views"
	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/admin/pagebuilder/example/containers"
	"github.com/qor5/admin/pagebuilder/example/layouts"
	"github.com/qor5/admin/richeditor"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

//go:embed assets/images
var containerImages embed.FS

func ConfigPageBuilder(db *gorm.DB, prefix, style string, i18nB *i18n.Builder) *pagebuilder.Builder {
	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: sess,
	})

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
	)
	if err != nil {
		panic(err)
	}
	pb := pagebuilder.New(db, i18nB)
	if prefix != "" {
		pb.Prefix(prefix)
	}
	if style != "" {
		pb.PageStyle(h.RawHTML(style))
	}

	media_view.Configure(pb.GetPresetsBuilder(), db)

	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert"}
	pb.GetPresetsBuilder().ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	pb.GetPresetsBuilder().ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
	pb.PageLayout(layouts.DefaultPageLayoutFunc)
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
	return pb
}
