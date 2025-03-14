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
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	"github.com/sunfmin/reflectutils"
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
	name  string
}

const (
	ParamTemplateID        = "mailbuilder_template_id"
	ParamChangeTemplate    = "mailbuilder_change_template"
	TemplateSelectionFiled = "mailbuilder_template_selection"

	paramModelName = "modelName"

	EmailEditorDialogPortalName = "mailbuilder_emailEditorDialog"

	TemplateSelectedPortalName  = "mailbuilder_templateSelectedPortal"
	ReloadSelectedTemplateEvent = "mailbuilder_template_ReloadSelectedTemplateEvent"

	dialogWidth            = "700"
	dialogHeight           = "476"
	dialogIframeCardHeight = 120
)

func (mb *ModelBuilder) Install(b *presets.Builder) (err error) {
	if mb.b.ab != nil {
		mb.mb.Use(mb.b.ab)
	}
	var (
		dp       = mb.mb.Detailing()
		editing  = mb.mb.Editing()
		creating = mb.mb.Editing().Creating()
		listing  = mb.mb.Listing()
		tm       = mb.b.getTemplateModelBuilder()
	)
	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			var (
				et = obj.(EmailDetailInterface).EmbedEmailDetail()
			)
			if !mb.IsTpl {
				// Parse the existing JsonBody as a JSON object
				var jsonBody map[string]interface{}
				if err := json.Unmarshal([]byte(et.JSONBody), &jsonBody); err == nil {
					// Add the Subject field to the JsonBody
					jsonBody["subject"] = et.Subject
					// Convert back to string
					if jsonData, err := json.Marshal(jsonBody); err == nil {
						et.JSONBody = string(jsonData)
					}
				}
			}
			return in(obj, id, ctx)
		}
	})

	creating.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
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
					te := template.(EmailDetailInterface).EmbedEmailDetail()
					et.HTMLBody = te.HTMLBody
					et.JSONBody = te.JSONBody
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

	if mb.IsTpl {
		mb.configTemplate()
	} else {
		mb.registerFunctions()
		dp.GetField(EmailDetailField).ComponentFunc(mb.mailDetailFieldCompoFunc())
		if mb.b.ab != nil {
			dp.SidePanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
				return mb.b.ab.MustGetModelBuilder(mb.mb).NewTimelineCompo(ctx, obj, "_side")
			})
		}
		filed := creating.GetField(TemplateSelectionFiled)
		if filed != nil && filed.GetCompFunc() == nil && tm != nil {
			listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
				msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, presets.Messages_en_US).(*presets.Messages)
				return h.Components(
					v.VBtn(msgr.New).
						Color(v.ColorPrimary).
						Variant(v.VariantElevated).
						Theme("light").Class("ml-2").
						Attr("@click", web.Plaid().URL(tm.mb.Info().ListingHref()).EventFunc(actions.OpenListingDialog).Query(presets.ParamOverlay, actions.Dialog).Go()),
				)
			})

			filed.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				return h.Components(
					web.Portal(mb.selectedTemplate(ctx)).Name(TemplateSelectedPortalName))
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
	model := mb.mb.NewModel()
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
		obj    = mb.mb.NewModel()
		db     = mb.b.db
		ctx    = &web.EventContext{R: r, W: w}
	)
	primarySlug := ctx.Param(presets.ParamID)
	if primarySlug != "" {
		if err := utils.PrimarySluggerWhere(db, obj, primarySlug).Find(models).Error; err != nil {
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
		primaySlug := fmt.Sprint(reflectutils.MustGet(obj, "ID"))
		p := obj.(EmailDetailInterface).EmbedEmailDetail()
		return h.Div(
			h.Div(
				h.Div(
					h.Iframe().Attr(":srcdoc", h.JSONString(p.HTMLBody)).
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
					web.Plaid().URL(mb.editorUri(primaySlug)).PushState(true).Go(),
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
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nEmailBuilderKey, Messages_en_US).(*Messages)
		pMsgr       = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, presets.Messages_en_US).(*presets.Messages)
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
				vx.VXBtn(pMsgr.Save).Variant(v.VariantElevated).Attr("@click", web.Emit("save_mail")).Color("primary").Attr(":disabled", "vars.$EmailEditorLoading"),
				vx.VXBtn(msgr.SendEmail).
					Variant(v.VariantElevated).
					Color(v.ColorSecondary).
					Attr("@click", web.Emit("open_send_mail_dialog")).
					Color(v.ColorSecondary).
					Class("ml-2"),
			).Class("d-flex align-center w-100"),
		).Class("d-flex align-center pa-3  w-100"),

		h.Div().Class("d-flex flex-column").Children(
			h.Div(
				web.Listen("save_mail",
					fmt.Sprintf(`() => { $refs.emailEditor.emit('getData').then(res=> {%s})}`,
						web.Plaid().URL(mb.mb.Info().ListingHref()).EventFunc(actions.Update).Query(presets.ParamID, primarySlug).Form(web.Var("res")).Go())),
				web.Listen("open_send_mail_dialog", ShowDialogScript(EmailEditorDialogPortalName, UtilDialogPayloadType{
					Title: msgr.EnterEmailAddressPlaceholder,
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
					Src(fmt.Sprintf("/email_builder/editor?%s=%s&%s=%s&userId=undefined", presets.ParamID, primarySlug, paramModelName, mb.name)).
					Class("flex-1 ml-2"),

				web.Portal().Name(EmailEditorDialogPortalName),
			).Class("d-flex").Style("height: calc(100vh - 100px - 20px);"),
		),
	).Fluid(true).Class("px-0 detailing-page-wrap")
}

func (mb *ModelBuilder) menusComp(ctx *web.EventContext, obj interface{}) []h.HTMLComponent {
	var (
		menus []h.HTMLComponent
		err   error
		ojID  interface{}
		pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, presets.Messages_en_US).(*presets.Messages)
	)
	if ojID, err = reflectutils.Get(obj, "ID"); err != nil {
		panic(err)
	}
	if mb.mb.Info().Verifier().Do(presets.PermDelete).WithReq(ctx.R).IsAllowed() == nil {
		menus = append(menus,
			v.VListItem(h.Text(pMsgr.Delete)).Attr("@click", web.Plaid().
				URL(mb.mb.Info().ListingHref()).
				EventFunc(actions.DeleteConfirmation).
				Query(presets.ParamID, ojID).
				Go(),
			))
	}
	if mb.mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() == nil {
		menus = append(menus,
			v.VListItem(h.Text(pMsgr.Edit)).Attr("@click", web.Plaid().
				URL(mb.mb.Info().ListingHref()).
				EventFunc(actions.Edit).
				Query(presets.ParamID, ojID).
				Go(),
			))
	}
	return menus
}

func (mb *ModelBuilder) selectedTemplate(ctx *web.EventContext) h.HTMLComponent {
	tm := mb.b.getTemplateModelBuilder()
	if tm == nil {
		return nil
	}
	var (
		template       = tm.mb.NewModel()
		selectID       = ctx.Param(ParamTemplateID)
		err            error
		name, htmlBody string
		msgr           = i18n.MustGetModuleMessages(ctx.R, I18nEmailBuilderKey, Messages_en_US).(*Messages)
	)
	if selectID != "" {
		if err = utils.PrimarySluggerWhere(mb.b.db, template, selectID).First(template).Error; err != nil {
			panic(err)
		}
		p := template.(*EmailTemplate)
		name = p.Name
		htmlBody = p.HTMLBody
	}
	return h.Div(
		h.Div(
			h.Span(msgr.ModelLabelTemplate).Class("text-body-1"),
		),
		v.VBtn(msgr.ChangeTemplate).Color(v.ColorPrimary).
			Variant(v.VariantTonal).
			PrependIcon("mdi-cached").
			Attr("@click", web.Plaid().URL(tm.mb.Info().ListingHref()).EventFunc(actions.OpenListingDialog).Query(ParamChangeTemplate, "1").Query(presets.ParamOverlay, actions.Dialog).Go()),
		v.VCard(
			v.VCardText(
				h.Iframe().Attr(":srcdoc", h.JSONString(htmlBody)).
					Attr("scrolling", "no", "frameborder", "0").
					Style(`pointer-events: none;transform-origin: 0 0; transform:scale(0.2);width:500%;height:500%`),
			).Class("pa-0", v.H100, "border-xl"),
		).Height(106).Width(224).Elevation(0).Class("mt-2"),
		h.Div(
			h.Span(name).Class("text-caption"),
		).Class("mt-2"),
	).Class("mb-6").Attr(web.VAssign("form", fmt.Sprintf(`{%s:%q}`, ParamTemplateID, selectID))...)
}

func (mb *ModelBuilder) configTemplate() {
	mb.mb.Editing("Name").Creating("Name").WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			if err = in(obj, ctx); err.HaveErrors() {
				return
			}
			p := obj.(*EmailTemplate)
			if p.Name == "" {
				err.FieldError("Name", "Name Is Required")
				return
			}
			return
		}
	})
	listing := mb.mb.Listing().DialogWidth(dialogWidth).DialogHeight(dialogHeight).SearchColumns("Subject")
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		lc := presets.ListingCompoFromContext(ctx.R.Context())
		if mb.mb.Info().Verifier().Do(presets.PermCreate).WithReq(ctx.R).IsAllowed() != nil || lc.Popup {
			return nil
		}
		msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, presets.Messages_en_US).(*presets.Messages)
		return v.VBtn(msgr.New).
			Color(v.ColorPrimary).
			Variant(v.VariantElevated).
			Theme("light").Class("ml-2").
			Attr("@click", web.Plaid().EventFunc(actions.New).Go())
	})
	listing.DataTableFunc(func(ctx *web.EventContext, searchParam *presets.SearchParams, searchResult *presets.SearchResult) h.HTMLComponent {
		var (
			lc             = presets.ListingCompoFromContext(ctx.R.Context())
			rows           = v.VRow()
			inDialog       = lc.Popup
			cols           = 3
			changeTemplate = ctx.Param(ParamChangeTemplate)
			cardClickEvent string
			msgr           = i18n.MustGetModuleMessages(ctx.R, I18nEmailBuilderKey, Messages_en_US).(*Messages)
		)
		if inDialog {
			cols = 4
			cardHeight = dialogIframeCardHeight
			if changeTemplate != "" {
				cardClickEvent = web.Plaid().EventFunc(ReloadSelectedTemplateEvent).ThenScript("vars.presetsListingDialog=false").Go()
			} else {
				cardClickEvent = web.Plaid().EventFunc(actions.New).Query(presets.ParamOverlay, actions.Dialog).ThenScript("vars.presetsListingDialog=false").Go()
			}
			if searchParam.Page == 1 {
				rows.AppendChildren(v.VCol(
					v.VCard(
						v.VCardItem(
							v.VCard(
								v.VCardText(
									v.VIcon("mdi-plus").Class("mr-1"), h.Text(msgr.AddBlankPage),
								).Class("pa-0", v.H100, "text-"+v.ColorPrimary, "text-body-2", "d-flex", "justify-center", "align-center"),
							).Height(cardHeight).Elevation(0).Class("bg-"+v.ColorGreyLighten4),
						).Class("pa-0", v.W100),
						v.VCardText(h.Text(msgr.BlankPage)).Class("text-caption"),
					).Attr("@click", cardClickEvent).Elevation(0),
				).Cols(cols))
			}

		}
		reflectutils.ForEach(searchResult.Nodes, func(obj interface{}) {
			var (
				p          = obj.(*EmailTemplate)
				menus      = mb.menusComp(ctx, obj)
				clickEvent = web.Plaid().PushState(true).URL(mb.editorUri(fmt.Sprint(p.ID))).Go()
			)
			if inDialog {
				if changeTemplate != "" {
					clickEvent = web.Plaid().EventFunc(ReloadSelectedTemplateEvent).ThenScript("vars.presetsListingDialog=false").Query(ParamTemplateID, p.ID).Go()
				} else {
					clickEvent = web.Plaid().EventFunc(actions.New).Query(presets.ParamOverlay, actions.Dialog).ThenScript("vars.presetsListingDialog=false").Query(ParamTemplateID, p.ID).Go()
				}
			}

			rows.AppendChildren(v.VCol(
				v.VCard(
					v.VCardItem(
						v.VCard(
							v.VCardText(
								h.Iframe().Attr(":srcdoc", h.JSONString(p.HTMLBody)).
									Attr("scrolling", "no", "frameborder", "0").
									Style(`pointer-events: none;transform-origin: 0 0; transform:scale(0.2);width:500%;height:500%`),
							).Class("pa-0", v.H100, "bg-"+v.ColorGreyLighten4),
						).Height(cardHeight).Elevation(0),
					).Class("pa-0", v.W100),
					h.If(!inDialog,
						v.VCardItem(
							v.VCard(
								v.VCardItem(
									h.Div(
										h.Text(p.Name),
										h.If(!inDialog && len(menus) > 0,
											v.VMenu(
												web.Slot(
													v.VBtn("").Children(
														v.VIcon("mdi-dots-horizontal"),
													).Attr("v-bind", "props").Variant(v.VariantText).Size(v.SizeSmall),
												).Name("activator").Scope("{ props }"),
												v.VList(
													menus...,
												),
											),
										),
									).Class(v.W100, "d-flex", "justify-space-between", "align-center"),
								).Class("pa-2"),
							).Color(v.ColorGreyLighten5).Height(cardContentHeight),
						).Class("pa-0"),
					),
					h.If(inDialog, v.VCardTitle(h.Text(p.Name)).Class("text-caption")),
				).Elevation(0).Attr("@click", clickEvent),
			).Cols(cols))
		})
		return v.VContainer(
			rows,
		).Fluid(true)
	})
}
