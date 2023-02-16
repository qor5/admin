package admin

import (
	"fmt"
	"net/http"
	"net/url"

	plogin "github.com/qor5/admin/login"
	"github.com/qor5/admin/presets"
	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/login"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type languageItem struct {
	Label string
	Value string
}

func loginPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		// i18n start
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		i18nBuilder := vh.I18n()
		var langs []languageItem
		var currLangVal string
		if ls := i18nBuilder.GetSupportLanguages(); len(ls) > 1 {
			qn := i18nBuilder.GetQueryName()
			lang := ctx.R.FormValue(qn)
			if lang == "" {
				lang = i18nBuilder.GetCurrentLangFromCookie(ctx.R)
			}
			accept := ctx.R.Header.Get("Accept-Language")
			_, mi := language.MatchStrings(language.NewMatcher(ls), lang, accept)
			for i, l := range ls {
				u, _ := url.Parse(ctx.R.RequestURI)
				qs := u.Query()
				qs.Set(qn, l.String())
				u.RawQuery = qs.Encode()
				if i == mi {
					currLangVal = u.String()
				}
				langs = append(langs, languageItem{
					Label: display.Self.Name(l),
					Value: u.String(),
				})
			}
		}
		// i18n end

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)
		wMsg := vh.GetWarnFlashMessage(msgr, ctx.W, ctx.R)
		iMsg := vh.GetInfoFlashMessage(msgr, ctx.W, ctx.R)
		wIn := vh.GetWrongLoginInputFlash(ctx.W, ctx.R)

		if iMsg != "" && vh.GetInfoCodeFlash(ctx.W, ctx.R) == login.InfoCodePasswordSuccessfullyChanged {
			wMsg = ""
		}

		var oauthHTML HTMLComponent
		if vh.OAuthEnabled() {
			ul := Div().Class("d-flex flex-column justify-center mt-8 text-center")
			for _, provider := range vh.OAuthProviders() {
				ul.AppendChildren(
					v.VBtn("").
						Block(true).
						Large(true).
						Class("mt-4").
						Outlined(true).
						Href(fmt.Sprintf("%s?provider=%s", vh.OAuthBeginURL(), provider.Key)).
						Children(
							Div(
								provider.Logo,
							).Class("mr-2"),
							Text(provider.Text),
						),
				)
			}

			oauthHTML = Div(
				ul,
			)
		}

		isRecaptchaEnabled := vh.RecaptchaEnabled()
		if isRecaptchaEnabled {
			plogin.DefaultViewCommon.InjectRecaptchaAssets(ctx, "login-form", "token")
		}

		var userPassHTML HTMLComponent
		if vh.UserPassEnabled() {
			userPassHTML = Div(
				Form(
					Div(
						Label(msgr.AccountLabel).Class(plogin.DefaultViewCommon.LabelClass).For("account"),
						plogin.DefaultViewCommon.Input("account", msgr.AccountPlaceholder, wIn.Account),
					),
					Div(
						Label(msgr.PasswordLabel).Class(plogin.DefaultViewCommon.LabelClass).For("password"),
						plogin.DefaultViewCommon.PasswordInput("password", msgr.PasswordPlaceholder, wIn.Password, true),
					).Class("mt-6"),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					plogin.DefaultViewCommon.FormSubmitBtn(msgr.SignInBtn).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", vh.RecaptchaSiteKey(), isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled),
				).Id("login-form").Method(http.MethodPost).Action(vh.PasswordLoginURL()),
				If(!vh.NoForgetPasswordLink(),
					Div(
						A(Text(msgr.ForgetPasswordLink)).Href(vh.ForgetPasswordPageURL()).
							Class("grey--text text--darken-1"),
					).Class("text-right mt-2"),
				),
			)
		}

		r.PageTitle = "Sign In"
		var bodyForm HTMLComponent
		bodyForm = Div(
			userPassHTML,
			oauthHTML,
			If(len(langs) > 0,
				v.VSelect().
					Items(langs).
					ItemText("Label").
					ItemValue("Value").
					Attr(web.InitContextVars, fmt.Sprintf(`{currLangVal: '%s'}`, currLangVal)).
					Attr("v-model", `vars.currLangVal`).
					Attr("@change", `window.location.href=vars.currLangVal`).
					Outlined(true).
					Dense(true).
					Class("mt-12").
					HideDetails(true),
			),
		).Class(plogin.DefaultViewCommon.WrapperClass).Style(plogin.DefaultViewCommon.WrapperStyle)

		demoTips := Div(
			Div(
				P(B(i18n.T(ctx.R, I18nExampleKey, "DemoTips"))),
			).Class(plogin.DefaultViewCommon.WrapperClass).Style(plogin.DefaultViewCommon.WrapperStyle).
				Style("border: 1px solid #d0d0d0; border-radius: 8px; width: 530px; padding: 0px 24px 0px 24px; padding-top: 16px!important;"),
		).Class("pt-12")

		r.Body = Div(
			plogin.DefaultViewCommon.ErrNotice(fMsg),
			plogin.DefaultViewCommon.WarnNotice(wMsg),
			plogin.DefaultViewCommon.InfoNotice(iMsg),
			bodyForm,
			demoTips,
		)

		return
	})
}
