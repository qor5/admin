package admin

import (
	"fmt"
	"net/http"

	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	v "github.com/qor5/x/v3/ui/vuetify"
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
		qn := i18nBuilder.GetQueryName()
		if ls := i18nBuilder.GetSupportLanguages(); len(ls) > 1 {
			lang := ctx.R.FormValue(qn)
			if lang == "" {
				lang = i18nBuilder.GetCurrentLangFromCookie(ctx.R)
			}
			accept := ctx.R.Header.Get("Accept-Language")
			_, mi := language.MatchStrings(language.NewMatcher(ls), lang, accept)
			for i, l := range ls {
				if i == mi {
					currLangVal = l.String()
				}

				langs = append(langs, languageItem{
					Label: display.Self.Name(l),
					Value: l.String(),
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
						Size(v.SizeLarge).
						Class("mt-4").
						Variant(v.VariantOutlined).
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
				web.Scope(
					v.VSelect().
						Items(langs).
						ItemTitle("Label").
						ItemValue("Value").
						Attr("v-model", `selectLocals.currLangVal`).
						Attr("@update:model-value", web.Plaid().Query(qn, web.Var("selectLocals.currLangVal")).PushState(true).Go()).
						Variant(v.VariantOutlined).
						Density(v.DensityCompact).
						Class("mt-12").
						HideDetails(true),
				).VSlot("{locals:selectLocals}").Init(fmt.Sprintf(`{currLangVal: '%s'}`, currLangVal)),
			),
		).Class(plogin.DefaultViewCommon.WrapperClass).Style(plogin.DefaultViewCommon.WrapperStyle)

		username := loginInitialUserEmail
		password := loginInitialUserPassword
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
