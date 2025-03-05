package emailbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"path"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
	"github.com/go-faker/faker/v4"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
)

type ModelBuilder struct {
	mb    *presets.ModelBuilder
	b     *Builder
	IsTpl bool
}

const (
	ParamTemplateID = "template_id"

	EmailEditorDialogPortalName = "emailEditorDialog"
)

func (mb *ModelBuilder) Install(b *presets.Builder) (err error) {
	if mb.b.ab != nil {
		mb.mb.Use(mb.b.ab)
	}
	mb.mb.Editing().Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		// JSON_Body
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			var (
				templateID           = ctx.Param(ParamTemplateID)
				templateModelBuilder = mb.b.getTemplateModelBuilder()
				et                   = obj.(EmailDetailInterface).EmbedEmailDetail()
				template             = templateModelBuilder.mb.NewModel()
				db                   = mb.b.db
			)
			if templateID != "" && templateModelBuilder != nil {
				err = utils.PrimarySluggerWhere(db, template, templateID).First(template).Error
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return
				} else if errors.Is(err, gorm.ErrRecordNotFound) {
					mb.setEmailDefaultValue(et)
				} else {
					*et = *template.(EmailDetailInterface).EmbedEmailDetail()
				}
			} else {
				mb.setEmailDefaultValue(et)
			}
			return in(obj, id, ctx)
		}
	})
	b.HandleCustomPage(mb.editorPattern(), presets.NewCustomPage(b).Body(mb.emailBuilderBody).Menu(func(*web.EventContext) h.HTMLComponent {
		return nil
	}))
	dp := mb.mb.Detailing()
	if mb.IsTpl {
		mb.mb.Listing().WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
			return func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
				cell.SetAttr("@click", web.Plaid().URL(mb.editorUri(id)).PushState(true).Go())
				return cell, nil
			}
		})
	} else {
		dp.GetField(EmailDetailField).ComponentFunc(mb.mailDetailFieldCompoFunc())
		if mb.b.ab != nil {
			dp.SidePanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
				return mb.b.ab.MustGetModelBuilder(mb.mb).NewTimelineCompo(ctx, obj, "_side")
			})
		}
	}
	return
}

func (mb *ModelBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/create":
		if r.Method == http.MethodPost {
			mb.createTemplate(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/list":
		if r.Method == http.MethodGet {
			mb.getTemplate(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/send":
		if r.Method == http.MethodPost {
			mb.send(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "page not found", http.StatusNotFound)
	}
}

func (mb *ModelBuilder) createTemplate(w http.ResponseWriter, r *http.Request) {
	var (
		model = mb.mb.NewModel()
	)
	et := model.(EmailDetailInterface).EmbedEmailDetail()
	if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// validate template format
	if _, err := GetTemplate(et.Subject); err != nil {
		http.Error(w, "invalid subject", http.StatusBadRequest)
		return
	}
	// if _, err := GetTemplate(et.JSONBody); err != nil {
	// 	http.Error(w, "invalid json body", http.StatusBadRequest)
	// 	return
	// }
	// if _, err := GetTemplate(et.HTMLBody); err != nil {
	// 	http.Error(w, "invalid html body", http.StatusBadRequest)
	// 	return
	// }
	// save in db

	if err := mb.mb.Editing().Creating().Saver(et, "", web.MustGetEventContext(r.Context())); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	by, _ := json.Marshal(&UnifyResponse{Data: &et})
	_, _ = w.Write(by)
}

func (mb *ModelBuilder) getTemplate(w http.ResponseWriter, r *http.Request) {
	var (
		models = mb.mb.NewModelSlice()
		db     = mb.b.db
	)
	idsParam := r.URL.Query().Get("ids")
	if strings.TrimSpace(idsParam) == "" {
		if err := db.Order("created_at DESC").Limit(10).Find(models).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		idsStr := strings.Split(idsParam, ",")
		ids := make([]int, len(idsStr))
		for i, idStr := range idsStr {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "invalid id params", http.StatusBadRequest)
				return
			}
			ids[i] = id
		}

		if err := db.Where("id in ?", ids).Find(models).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	by, _ := json.Marshal(&UnifyResponse{Data: models})
	_, _ = w.Write(by)
}

func (mb *ModelBuilder) send(w http.ResponseWriter, r *http.Request) {
	var sendRequest SendRequest
	err := json.NewDecoder(r.Body).Decode(&sendRequest)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var (
		model = mb.mb.NewModel()
		db    = mb.b.db
		et    = model.(EmailDetailInterface).EmbedEmailDetail()
	)
	err = db.Where("id = ?", sendRequest.TemplateID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "template not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	var results []SendResult
	var hasErr bool
	for _, uid := range sendRequest.UserIds {
		//  fake username here(actually, should get it by uid)
		mailData := MailData{
			Name: faker.Name(),
		}

		subjectTmpl, err := GetTemplate(et.Subject)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		subject, err := GetContent(subjectTmpl, mailData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		htmlBodyTmpl, err := GetTemplate(et.HTMLBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		htmlBody, err := GetContent(htmlBodyTmpl, mailData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get ToEmailAddress by config, actually should get it by uid
		toEmailAddress := strings.ToLower(strings.TrimSpace(sendRequest.ToEmailAddress))
		if !EmailValidator(toEmailAddress) {
			http.Error(w, "invalid email address", http.StatusBadRequest)
			return
		}

		input := &sesv2.SendEmailInput{
			Content: &types.EmailContent{
				Raw: nil,
				Simple: &types.Message{
					Body: &types.Body{
						Html: &types.Content{
							Data:    aws.String(htmlBody),
							Charset: aws.String(mb.b.sender.Config.HTMLBodyCharset),
						},
						Text: nil,
					},
					Subject: &types.Content{
						Data:    aws.String(subject),
						Charset: aws.String(mb.b.sender.Config.SubjectCharset),
					},
					Headers: nil,
				},
				Template: nil,
			},
			ConfigurationSetName: nil,
			Destination: &types.Destination{
				BccAddresses: nil,
				CcAddresses:  nil,
				ToAddresses:  []string{toEmailAddress},
			},
			EmailTags:                      nil,
			EndpointId:                     nil,
			FeedbackForwardingEmailAddress: nil,
			FeedbackForwardingEmailAddressIdentityArn: nil,
			FromEmailAddress: aws.String((&mail.Address{
				Name:    mb.b.sender.Config.FromName,
				Address: mb.b.sender.Config.FromEmailAddress,
			}).String()),
			FromEmailAddressIdentityArn: nil,
			ListManagementOptions:       nil,
			ReplyToAddresses:            nil,
		}
		output, err := mb.b.sender.SendEmail(r.Context(), input)
		if err != nil {
			hasErr = true
			errMsg := "unknown error"
			var opErr *smithy.OperationError
			if errors.As(err, &opErr) {
				var apiErr smithy.APIError
				if errors.As(opErr.Err, &apiErr) {
					errMsg = apiErr.ErrorCode()
				}
			}
			results = append(results, SendResult{
				UserId:     uid,
				TemplateID: sendRequest.TemplateID,
				MessageID:  "",
				ErrMsg:     errMsg,
			})
		} else {
			results = append(results, SendResult{
				UserId:     uid,
				TemplateID: sendRequest.TemplateID,
				MessageID:  lo.FromPtr(output.MessageId),
				ErrMsg:     "",
			})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if hasErr {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	by, _ := json.Marshal(&UnifyResponse{Data: results})
	_, _ = w.Write(by)
}

func (mb *ModelBuilder) setEmailDefaultValue(et *EmailDetail) {
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
}

func (mb *ModelBuilder) mailDetailFieldCompoFunc() presets.FieldComponentFunc {

	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		db := mb.b.db
		tm := &MailTemplate{}
		db.First(&tm, 1)
		return h.Div(
			h.Div(
				h.Div(
					v.VSpacer(),
					v.VBtn("Publish").
						Variant(v.VariantElevated).Color(v.ColorPrimary).Height(36),
					v.VBtn("").Size(v.SizeSmall).Children(v.VIcon("mdi-alarm").Size(v.SizeXLarge)).Rounded("0").Class("rounded-e ml-abs-1").
						Variant(v.VariantElevated).Color(v.ColorPrimary).Width(36).Height(36),
				).Class("tagList-bar-warp pb-4"),

			),
			h.Div(
				h.Div(
					h.Iframe().Attr(":srcdoc", h.JSONString(tm.HTMLBody)).
						Attr("scrolling", "no", "frameborder", "0").
						Style(`pointer-events: none; 
 -webkit-mask-image: radial-gradient(circle, black 80px, transparent);
  mask-image: radial-gradient(circle, black 80px, transparent);
transform-origin: 0 0; transform:scale(0.5);width:200%;height:200%`),
				).Class(v.W100, v.H100, "overflow-hidden"),
				h.Div(
					h.Div().Class(fmt.Sprintf("bg-%s", v.ColorGreyLighten3)),
					v.VBtn("Editor").AppendIcon("mdi-pencil").Color(v.ColorBlack).
						Class("rounded").Height(36).Variant(v.VariantElevated),
				).Class("pa-6 w-100 d-flex justify-space-between align-center").Style(`position:absolute;bottom:0;left:0`),
			).Style(`position:relative;height:320px;width:100%`).Class("border-thin rounded-lg").
				Attr("@click",
					web.Plaid().URL(mb.b.getTemplateModelBuilder().editorUri("1")).Go(),
				),
		).Class("my-10")
	}
}

func (mb *ModelBuilder) editorPattern() string {
	return path.Join(mb.mb.Info().URIName(), "/email_builder/editor")
}

func (mb *ModelBuilder) editorUri(primarySlug string) string {
	return fmt.Sprintf("/%s?%s=%s", path.Join(mb.mb.GetPresetsBuilder().GetURIPrefix(), mb.editorPattern()), presets.ParamID, primarySlug)
}

func (mb *ModelBuilder) emailBuilderBody(ctx *web.EventContext) h.HTMLComponent {
	var (
		primarySlug = ctx.Param(presets.ParamID)
		exitHref    string
	)
	if mb.IsTpl {
		exitHref = mb.mb.Info().ListingHref()
	} else {
		exitHref = mb.mb.Info().DetailingHref(primarySlug)
	}

	return v.VContainer().Children(

		h.Div(
			h.Div(
				h.Div().Style("transform:rotateY(180deg)").Class("mr-4").Children(
					v.VIcon("mdi-exit-to-app").Attr("@click", fmt.Sprintf(`
						const last = vars.__history.last();
						if (last && last.url && last.url === %q) {
							$event.view.window.history.back();
							return;
						}
						%s`, exitHref, web.GET().URL(exitHref).PushState(true).Go(),
					)),
				),
				v.VToolbarTitle("Inbox"),
				v.VSpacer(),
				vx.VXBtn("Save").Variant("elevated").Attr("@click", web.Emit("save_mail")).Color("primary").Attr(":disabled", "vars.$EmailEditorLoading"),
				vx.VXBtn("Send Email").
					Variant("elevated").
					Color("secondary").
					Attr("@click", web.Emit("open_send_mail_dialog")).
					Color("secondary").
					Class("ml-2"),
			).Class("d-flex align-center w-100"),
		).Class("d-flex align-center pa-3  w-100"),

		h.Div().Class("d-flex flex-column").Children(
			h.Div(
				web.Listen("save_mail",
					fmt.Sprintf(`() => { $refs.emailEditor.emit('getData').then(res=> {%s})}`,
						web.Plaid().EventFunc(actions.Update).Query(presets.ParamID, primarySlug).Form(web.Var("res")).Go())),
				web.Listen("open_send_mail_dialog", ShowDialogScript(EmailEditorDialogPortalName, UtilDialogPayloadType{
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
					Src(fmt.Sprintf("/email_builder/editor?id=%s&userId=undefined", primarySlug)).
					Class("flex-1 ml-2"),

				web.Portal().Name(EmailEditorDialogPortalName),
			).Class("d-flex").Style("height: calc(100vh - 100px - 20px);"),
		),
	).Fluid(true).Class("px-0 detailing-page-wrap")

}
