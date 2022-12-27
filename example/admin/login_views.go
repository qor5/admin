package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/qor5/admin/presets"
	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/login"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

const (
	wrapperClass = "d-flex pt-16 flex-column mx-auto"
	wrapperStyle = "max-width: 28rem;"
	labelClass   = "d-block mb-1 grey--text text--darken-2 text-sm-body-2"
)

func errNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return v.VAlert(Text(msg)).
		Dense(true).
		Class("text-center").
		Icon(false).
		Type("error")
}

func warnNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return v.VAlert(Text(msg)).
		Dense(true).
		Class("text-center").
		Icon(false).
		Type("warning")
}

func infoNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return v.VAlert(Text(msg)).
		Dense(true).
		Class("text-center").
		Icon(false).
		Type("info")
}

func input(
	id string,
	placeholder string,
	val string,
) *v.VTextFieldBuilder {
	return v.VTextField().
		Attr("name", id).
		Id(id).
		Placeholder(placeholder).
		Value(val).
		Outlined(true).
		HideDetails(true).
		Dense(true)
}

func passwordInput(
	id string,
	placeholder string,
	val string,
	canReveal bool,
) *v.VTextFieldBuilder {
	in := input(id, placeholder, val)
	if canReveal {
		varName := fmt.Sprintf(`show_%s`, id)
		in.Attr(":append-icon", fmt.Sprintf(`vars.%s ? "visibility_off" : "visibility"`, varName)).
			Attr(":type", fmt.Sprintf(`vars.%s ? "text" : "password"`, varName)).
			Attr("@click:append", fmt.Sprintf(`vars.%s = !vars.%s`, varName, varName)).
			Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, varName))
	}

	return in
}

func formSubmitBtn(
	label string,
) *v.VBtnBuilder {
	return v.VBtn(label).
		Color("primary").
		Block(true).
		Large(true).
		Type("submit").
		Class("mt-6")
}

func injectRecaptchaAssets(ctx *web.EventContext, formID string, tokenFieldID string) {
	ctx.Injector.HeadHTML(`
<style>
.grecaptcha-badge { visibility: hidden; }
</style>
    `)
	ctx.Injector.HeadHTML(fmt.Sprintf(`
<script>
function onSubmit(token) {
	document.getElementById("%s").value = token;
	document.getElementById("%s").submit();
}
</script>
    `, tokenFieldID, formID))
	ctx.Injector.TailHTML(`
<script src="https://www.google.com/recaptcha/api.js"></script>
    `)
}

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
			injectRecaptchaAssets(ctx, "login-form", "token")
		}

		var userPassHTML HTMLComponent
		if vh.UserPassEnabled() {
			userPassHTML = Div(
				Form(
					Div(
						Label(msgr.AccountLabel).Class(labelClass).For("account"),
						input("account", msgr.AccountPlaceholder, wIn.Account),
					),
					Div(
						Label(msgr.PasswordLabel).Class(labelClass).For("password"),
						passwordInput("password", msgr.PasswordPlaceholder, wIn.Password, true),
					).Class("mt-6"),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					formSubmitBtn(msgr.SignInBtn).
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
		).Class(wrapperClass).Style(wrapperStyle)

		demoUser := os.Getenv("DEMO_USERNAME")
		demoPass := os.Getenv("DEMO_PASSWORD")
		isDemo := demoUser != "" && demoPass != ""
		demoTips := Div(
			Div(
				P(Text(i18n.T(ctx.R, I18nExampleKey, "DemoUsernameLabel")), B(demoUser)),
				P(Text(i18n.T(ctx.R, I18nExampleKey, "DemoPasswordLabel")), B(demoPass)),
				P(B(i18n.T(ctx.R, I18nExampleKey, "DemoTips"))),
			).Class(wrapperClass).Style(wrapperStyle).
				Style("border: 1px solid #d0d0d0; border-radius: 8px; width: 530px; padding: 0px 24px 0px 24px; padding-top: 16px!important;"),
		).Class("pt-12")

		r.Body = Div(
			errNotice(fMsg),
			warnNotice(wMsg),
			infoNotice(iMsg),
			bodyForm,
			If(isDemo, demoTips),
		)

		return
	})
}
