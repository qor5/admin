package example

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/media/oss"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/pagebuilder"
	"github.com/qor/qor5/richeditor"
	h "github.com/theplant/htmlgo"
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

type TextAndImage struct {
	ID    uint
	Text  string
	Image media_library.MediaBox
}

type MainContent struct {
	ID    uint
	Title string
	Body  string
}

func ConfigPageBuilder(db *gorm.DB) *pagebuilder.Builder {
	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: sess,
	})

	err := db.AutoMigrate(
		&TextAndImage{},
		&MainContent{},
	)
	if err != nil {
		panic(err)
	}
	pb := pagebuilder.New(db)

	media_view.Configure(pb.GetPresetsBuilder(), db)

	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert"}
	pb.GetPresetsBuilder().ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	pb.GetPresetsBuilder().ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())

	textAndImage := pb.RegisterContainer("text_and_image").
		RenderFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			tai := obj.(*TextAndImage)
			return h.Div(
				h.Text(tai.Text),
				h.Img(tai.Image.Url),
			)
		})

	ed := textAndImage.Model(&TextAndImage{}).Editing("Text", "Image")
	ed.Field("Image")

	mainContent := pb.RegisterContainer("main_content").
		RenderFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			mc := obj.(*MainContent)
			return h.Div(
				h.H1(mc.Title),
				h.RawHTML(mc.Body),
			)
		})

	mainEd := mainContent.Model(&MainContent{}).Editing("Title", "Body")

	mainEd.Field("Body").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Body").
			Plugins([]string{"alignment", "video", "imageinsert"}).
			Value(obj.(*MainContent).Body).
			Label(field.Label)
	})
	return pb
}
