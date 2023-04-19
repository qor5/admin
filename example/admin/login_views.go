package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

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
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)
		loginMsgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
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
							Text(i18n.T(ctx.R, I18nExampleKey, provider.Text)),
						),
				)
			}

			oauthHTML = Div(
				ul,
			)
		}

		wIn := vh.GetWrongLoginInputFlash(ctx.W, ctx.R)
		isRecaptchaEnabled := vh.RecaptchaEnabled()
		if isRecaptchaEnabled {
			plogin.DefaultViewCommon.InjectRecaptchaAssets(ctx, "login-form", "token")
		}

		var logoSection HTMLComponent
		logo, _ := assets.ReadFile("assets/logo.svg")
		logoSection = Div(
			A(RawHTML(logo)).Href("https://qor5.com/").Target("_blank"),
		).Style("text-align: center;")

		var userPassHTML HTMLComponent
		if vh.UserPassEnabled() {
			userPassHTML = Div(
				Form(
					Div(
						Label(loginMsgr.AccountLabel).Class(plogin.DefaultViewCommon.LabelClass).For("account"),
						plogin.DefaultViewCommon.Input("account", loginMsgr.AccountPlaceholder, wIn.Account),
					),
					Div(
						Label(loginMsgr.PasswordLabel).Class(plogin.DefaultViewCommon.LabelClass).For("password"),
						plogin.DefaultViewCommon.PasswordInput("password", loginMsgr.PasswordPlaceholder, wIn.Password, true),
					).Class("mt-6"),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					plogin.DefaultViewCommon.FormSubmitBtn(loginMsgr.SignInBtn).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", vh.RecaptchaSiteKey(), isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled),
				).Id("login-form").Method(http.MethodPost).Action(vh.PasswordLoginURL()),
				If(!vh.NoForgetPasswordLink(),
					Div(
						A(Text(loginMsgr.ForgetPasswordLink)).Href(vh.ForgetPasswordPageURL()).
							Class("grey--text text--darken-1"),
					).Class("text-right mt-2"),
				),
			)
		}

		r.PageTitle = loginMsgr.LoginPageTitle
		var bodyForm HTMLComponent
		bodyForm = Div(
			logoSection,
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

		username := os.Getenv("LOGIN_INITIAL_USER_EMAIL")
		password := os.Getenv("LOGIN_INITIAL_USER_PASSWORD")
		isDemo := username != "" && password != ""
		demoTips := Div(
			Div(
				P(Text(msgr.DemoUsernameLabel), B(username)),
				P(Text(msgr.DemoPasswordLabel), B(password)),
				P(B(msgr.DemoTips)),
			).Class(plogin.DefaultViewCommon.WrapperClass).Style(plogin.DefaultViewCommon.WrapperStyle).
				Style("border: 1px solid #d0d0d0; border-radius: 8px; width: 530px; padding: 0px 24px 0px 24px; padding-top: 16px!important;"),
		).Class("py-12")

		r.Body = Div(
			plogin.DefaultViewCommon.Notice(vh, loginMsgr, ctx.W, ctx.R),
			bodyForm,
			If(isDemo, demoTips),
		)

		return
	})
}

func oauthCompleteInfoPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)
		loginMsgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		fMsg := vh.GetFailFlashMessage(loginMsgr, ctx.W, ctx.R)
		dvc := plogin.DefaultViewCommon

		r.PageTitle = "Complete Info"
		var bodyForm HTMLComponent
		bodyForm = Div(
			H1(msgr.OAuthCompleteInfoTitle).Class(dvc.TitleClass),

			Form(
				Div(
					Label(msgr.OAuthCompleteInfoPositionLabel).Class(dvc.LabelClass).For("position"),
					dvc.Input("position", "", ""),
				),
				Div(
					Input("agree").Type("checkbox").Id("agree").Style("margin-right: 8px; margin-left: 2px;"),
					Label(msgr.OAuthCompleteInfoAgreeLabel).For("agree"),
				).Class("mt-6"),
				dvc.FormSubmitBtn(loginMsgr.Confirm),
				v.VBtn(msgr.OAuthCompleteInfoBackLabel).Block(true).Large(true).Class("mt-6").Href("/auth/logout"),
			).Method(http.MethodPost).Action(oauthCompleteInfoActionURL),
		).Class(dvc.WrapperClass).Style(dvc.WrapperStyle)

		r.Body = Div(
			dvc.ErrNotice(fMsg),
			bodyForm,
		)
		return
	})
}
