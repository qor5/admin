package emailbuilder

import (
	"fmt"
	"os"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get("DB_PARAMS", "email builder example database connection string", "")

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

func LoadToEmailAddress() string {
	to := os.Getenv("TO_ADDRESS")
	return to
}

func LoadSenderConfig() (config SESDriverConfig) {
	from := os.Getenv("FROM_ADDRESS")
	return SESDriverConfig{
		FromEmailAddress:               from,
		FromName:                       "",
		SubjectCharset:                 "UTF-8",
		HTMLBodyCharset:                "UTF-8",
		TextBodyCharset:                "UTF-8",
		ConfigurationSetName:           "",
		FeedbackForwardingEmailAddress: "",
		FeedbackForwardingEmailAddressIdentityArn: "",
		FromEmailAddressIdentityArn:               "",
		ContactListName:                           "",
		TopicName:                                 "",
	}
}

func ConfigMailTemplate(pb *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	mb := pb.Model(&MailTemplate{})
	lb := mb.Listing("ID", "Subject").NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return h.Div(vx.VXBtn("New").Href("//localhost:9500/email_builder/editor"))
	})
	_ = lb
	dp := mb.Detailing("Demo")
	_ = dp
	dp.Field("Demo").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		et := obj.(*MailTemplate)
		// fmt.Println(et.ID)
		// fmt.Println(et.Subject)
		return h.Div(
			h.Iframe().
				Class("flex-1").
				Attr("frameborder", "0").
				Attr("width", "100%").
				Src(fmt.Sprintf("//localhost:9500/email_builder/editor?id=%d&userId=undefined", et.ID)),
		).Class("d-flex").Style("height: calc(100vh - 100px - 16px);")
	})
	// eb := mb.Editing()
	// eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
	// 	return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
	// 		return in(obj, id, ctx)
	// 	}
	// })
	return mb
}
