package emailbuilder

import (
	"fmt"
	"os"
	"strconv"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
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
	lb := mb.Listing("ID", "Subject")
	_ = lb
	dp := mb.Detailing("Demo")
	_ = dp
	eb := mb.Editing("Subject", "JSONBody", "HTMLBody")

	eb.Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		// JSON_Body
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			et := obj.(*MailTemplate)
			et.JSONBody = fmt.Sprintf(`{
  "subject": %s,
  "subTitle": "",
  "content": {
    "type": "page",
    "data": {
      "value": {
        "breakpoint": "480px",
        "headAttributes": "",
        "font-size": "14px",
        "font-weight": "400",
        "line-height": "1.7",
        "headStyles": [],
        "fonts": [],
        "responsive": true,
        "font-family": "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans','Helvetica Neue', sans-serif",
        "text-color": "#000000"
      }
    },
    "attributes": {
      "background-color": "#efeeea",
      "width": "600px"
    },
    "children": [
      {
        "type": "advanced_wrapper",
        "data": {
          "value": {}
        },
        "attributes": {
          "padding": "20px 0px 20px 0px",
          "border": "none",
          "direction": "ltr",
          "text-align": "center"
        },
        "children": []
      }
    ]
  }
}`, strconv.Quote(et.Subject))
			return in(obj, id, ctx)
		}
	})

	dp.Title(func(evCtx *web.EventContext, obj any, style presets.DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		titleCompo = h.Div(
			v.VToolbarTitle("Inbox"),
			v.VSpacer(),
			vx.VXBtn("Save").Variant("elevated").Attr("@click", web.Emit("save_mail")).Color("primary").Attr(":disabled", "vars.$EmailEditorLoading"),
			vx.VXBtn("Send Email").
				Variant("elevated").
				Color("secondary").
				Attr("@click", web.Emit("open_send_mail_dialog")).
				Color("secondary").
				Class("ml-2"),
		).Class("d-flex align-center w-100")

		return
	})

	dp.Field("Demo").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			et := obj.(*MailTemplate)
			modalName := "emailEditorDialog"

			return h.Div(
				web.Listen("save_mail",
					fmt.Sprintf(`() => { $refs.emailEditor.emit('getData').then(res=> {%s})}`,
						web.Plaid().EventFunc(actions.Update).Query(presets.ParamID, et.ID).Form(web.Var("res")).Go())),

				web.Listen("open_send_mail_dialog", ShowDialogScript(modalName, UtilDialogPayloadType{
					Title: "Please Enter a Email Address",
					ContentEl: vx.VXField().
						Label("To").
						Placeholder("Enter a email address").
						Attr("v-model", "vars.to_email_address").
						// Attr(":rules", "(v) => !!v || 'Email is required'").
						Required(true),
					OkAction:  web.Emit("send_mail"),
					LoadingOk: "vars.$EmailEditorLoading",
				})),

				web.Listen("send_mail", fmt.Sprintf(`(payload)=> {
						vars.$EmailEditorLoading = true

						$refs.emailEditor.emit('sendMail', {to: vars.to_email_address})
							.then(res=> { %s })
							.catch(err=> { %s }) 
							.finally(()=> { vars.$EmailEditorLoading = false })
					}`, presets.ShowSnackbarScript("Email successfully sent.", "success"), presets.ShowSnackbarScript("Email sent failed.", "error"))),
				vx.VXIframeEmailEditor().
					Ref("emailEditor").
					Src(fmt.Sprintf("/email_builder/editor?id=%d&userId=undefined", et.ID)).
					Class("flex-1 ml-2"),

				web.Portal().Name(modalName),
			).Class("d-flex").Style("height: calc(100vh - 100px - 20px);")
		})

	return mb
}
