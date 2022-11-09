package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"net/url"

	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/login"
	"github.com/pquerna/otp"
	"github.com/qor5/admin/presets"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"gorm.io/gorm"
)

const (
	wrapperClass = "d-flex pt-16 flex-column mx-auto"
	wrapperStyle = "max-width: 28rem;"
	titleClass   = "text-h5 mb-6 font-weight-bold"
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

func errorBody(msg string) HTMLComponent {
	return Div(
		Text(msg),
	)
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

// need to import zxcvbn js
// func PasswordInputWithStrengthMeter(in *v.VTextFieldBuilder, id string, val string) HTMLComponent {
// 	passVar := fmt.Sprintf(`password_%s`, id)
// 	meterScoreVar := fmt.Sprintf(`meter_score_%s`, id)
// 	in.Attr("v-model", fmt.Sprintf(`vars.%s`, passVar)).
// 		Attr(":loading", fmt.Sprintf(`!!vars.%s`, passVar)).
// 		On("input", fmt.Sprintf(`vars.%s = vars.%s ? zxcvbn(vars.%s).score + 1 : 0`, meterScoreVar, passVar, passVar))
// 	return Div(
// 		in.Children(
// 			RawHTML(fmt.Sprintf(`
//         <template v-slot:progress>
//           <v-progress-linear
//             :value="vars.%s * 20"
//             :color="['grey', 'red', 'deep-orange', 'amber', 'yellow', 'light-green'][vars.%s]"
//             absolute
//           ></v-progress-linear>
//         </template>
//             `, meterScoreVar, meterScoreVar)),
// 		),
// 	).Attr(web.InitContextVars, fmt.Sprintf(`{%s: "%s", %s: "%s" ? zxcvbn("%s").score + 1 : 0}`, passVar, val, meterScoreVar, val, val))
// }

// need to import zxcvbn.js
func passwordInputWithStrengthMeter(in *v.VTextFieldBuilder, id string, val string) HTMLComponent {
	passVar := fmt.Sprintf(`password_%s`, id)
	meterScoreVar := fmt.Sprintf(`meter_score_%s`, id)
	in.Attr("v-model", fmt.Sprintf(`vars.%s`, passVar)).
		On("input", fmt.Sprintf(`vars.%s = vars.%s ? zxcvbn(vars.%s).score + 1 : 0`, meterScoreVar, passVar, passVar))
	return Div(
		in,
		v.VProgressLinear().
			Class("mt-2").
			Attr(":value", fmt.Sprintf(`vars.%s * 20`, meterScoreVar)).
			Attr(":color", fmt.Sprintf(`["grey", "red", "deep-orange", "amber", "yellow", "light-green"][vars.%s]`, meterScoreVar)).
			Attr("v-show", fmt.Sprintf(`!!vars.%s`, passVar)),
	).Attr(web.InitContextVars, fmt.Sprintf(`{%s: "%s", %s: "%s" ? zxcvbn("%s").score + 1 : 0}`, passVar, val, meterScoreVar, val, val))
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

// requirements:
// - submit button
//   - add class `g-recaptcha`
//   - add attr `data-sitekey=<key>`
//   - add attr `data-callback=onSubmit`
//
// - add token field like `Input("token").Id("token").Type("hidden")`
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

func injectZxcvbn(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(fmt.Sprintf(`
<script src="%s"></script>
    `, login.ZxcvbnJSURL))
}

type languageItem struct {
	Label string
	Value string
}

func defaultLoginPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
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

		r.Body = Div(
			errNotice(fMsg),
			warnNotice(wMsg),
			infoNotice(iMsg),
			bodyForm,
		)

		return
	})
}

func defaultForgetPasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)
		wIn := vh.GetWrongForgetPasswordInputFlash(ctx.W, ctx.R)
		secondsToResend := vh.GetSecondsToRedoFlash(ctx.W, ctx.R)
		activeBtnText := msgr.SendResetPasswordEmailBtn
		inactiveBtnText := msgr.ResendResetPasswordEmailBtn
		inactiveBtnTextWithInitSeconds := fmt.Sprintf("%s (%d)", inactiveBtnText, secondsToResend)

		doTOTP := ctx.R.URL.Query().Get("totp") == "1"
		actionURL := vh.SendResetPasswordLinkURL()
		if doTOTP {
			actionURL = login.MustSetQuery(actionURL, "totp", "1")
		}

		isRecaptchaEnabled := vh.RecaptchaEnabled()
		if isRecaptchaEnabled {
			injectRecaptchaAssets(ctx, "forget-form", "token")
		}

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			errNotice(fMsg),
			If(secondsToResend > 0,
				warnNotice(msgr.SendEmailTooFrequentlyNotice),
			),
			Div(
				H1(msgr.ForgotMyPasswordTitle).Class(titleClass),
				Form(
					Div(
						Label(msgr.ForgetPasswordEmailLabel).Class(labelClass).For("account"),
						input("account", msgr.ForgetPasswordEmailPlaceholder, wIn.Account),
					),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(labelClass).For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					formSubmitBtn(inactiveBtnTextWithInitSeconds).
						Attr("id", "disabledBtn").
						ClassIf("d-none", secondsToResend <= 0),
					formSubmitBtn(activeBtnText).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", vh.RecaptchaSiteKey(), isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled).
						Attr("id", "submitBtn").
						ClassIf("d-none", secondsToResend > 0),
				).Id("forget-form").Method(http.MethodPost).Action(actionURL),
			).Class(wrapperClass).Style(wrapperStyle),
		)

		if secondsToResend > 0 {
			ctx.Injector.TailHTML(fmt.Sprintf(`
<script>
(function(){
    var secondsToResend = %d;
    var btnText = "%s";
    var disabledBtn = document.getElementById("disabledBtn");
    var submitBtn = document.getElementById("submitBtn");
    var interv = setInterval(function(){
        secondsToResend--;
        if (secondsToResend === 0) {
            clearInterval(interv);
            disabledBtn.classList.add("d-none");
            submitBtn.innerText = btnText;
            submitBtn.classList.remove("d-none");
            return;
        }
        disabledBtn.innerText = btnText + " (" + secondsToResend + ")" ;
    }, 1000);
})();
</script>
        `, secondsToResend, inactiveBtnText))
		}
		return
	})
}

func defaultResetPasswordLinkSentPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		a := ctx.R.URL.Query().Get("a")

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Div(
				H1(fmt.Sprintf("%s %s.", msgr.ResetPasswordLinkWasSentTo, a)).Class("text-h5"),
				H2(msgr.ResetPasswordLinkSentPrompt).Class("text-body-1 mt-2"),
			).Class(wrapperClass).Style(wrapperStyle),
		)
		return
	})
}

func defaultResetPasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)
		if fMsg == "" {
			fMsg = vh.GetCustomErrorMessageFlash(ctx.W, ctx.R)
		}
		wIn := vh.GetWrongResetPasswordInputFlash(ctx.W, ctx.R)

		doTOTP := ctx.R.URL.Query().Get("totp") == "1"
		actionURL := vh.ResetPasswordURL()
		if doTOTP {
			actionURL = login.MustSetQuery(actionURL, "totp", "1")
		}

		var user interface{}

		r.PageTitle = "Reset Password"

		query := ctx.R.URL.Query()
		id := query.Get("id")
		if id == "" {
			r.Body = Div(Text("user not found"))
			return r, nil
		} else {
			user, err = vh.FindUserByID(id)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					r.Body = Div(Text("user not found"))
					return r, nil
				}
				r.Body = Div(Text("system error"))
				return r, nil
			}
		}
		token := query.Get("token")
		if token == "" {
			r.Body = Div(Text("invalid token"))
			return r, nil
		} else {
			storedToken, _, expired := user.(login.UserPasser).GetResetPasswordToken()
			if expired {
				r.Body = Div(Text("token expired"))
				return r, nil
			}
			if token != storedToken {
				r.Body = Div(Text("invalid token"))
				return r, nil
			}
		}

		injectZxcvbn(ctx)

		r.Body = Div(
			errNotice(fMsg),
			Div(
				H1(msgr.ResetYourPasswordTitle).Class(titleClass),
				Form(
					Input("user_id").Type("hidden").Value(id),
					Input("token").Type("hidden").Value(token),
					Div(
						Label(msgr.ResetPasswordLabel).Class(labelClass).For("password"),
						passwordInputWithStrengthMeter(passwordInput("password", msgr.ResetPasswordLabel, wIn.Password, true), "password", wIn.Password),
					),
					Div(
						Label(msgr.ResetPasswordConfirmLabel).Class(labelClass).For("confirm_password"),
						passwordInput("confirm_password", msgr.ResetPasswordConfirmPlaceholder, wIn.ConfirmPassword, true),
					).Class("mt-6"),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(labelClass).For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					formSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(actionURL),
			).Class(wrapperClass).Style(wrapperStyle),
		)
		return
	})
}

func defaultChangePasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)
		if fMsg == "" {
			fMsg = vh.GetCustomErrorMessageFlash(ctx.W, ctx.R)
		}
		wIn := vh.GetWrongChangePasswordInputFlash(ctx.W, ctx.R)

		injectZxcvbn(ctx)

		r.PageTitle = "Change Password"

		r.Body = Div(
			errNotice(fMsg),
			Div(
				H1(msgr.ChangePasswordTitle).Class(titleClass),
				Form(
					Div(
						Label(msgr.ChangePasswordOldLabel).Class(labelClass).For("old_password"),
						passwordInput("old_password", msgr.ChangePasswordOldPlaceholder, wIn.OldPassword, true),
					),
					Div(
						Label(msgr.ChangePasswordNewLabel).Class(labelClass).For("password"),
						passwordInputWithStrengthMeter(passwordInput("password", msgr.ChangePasswordNewPlaceholder, wIn.NewPassword, true), "password", wIn.NewPassword),
					).Class("mt-6"),
					Div(
						Label(msgr.ChangePasswordNewConfirmLabel).Class(labelClass).For("confirm_password"),
						passwordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, wIn.ConfirmPassword, true),
					).Class("mt-6"),
					If(vh.TOTPEnabled(),
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(labelClass).For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					formSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(vh.ChangePasswordURL()),
			).Class(wrapperClass).Style(wrapperStyle),
		)
		return
	})
}

func changePasswordDialog(vh *login.ViewHelper, ctx *web.EventContext, showVar string, content HTMLComponent) HTMLComponent {
	pmsgr := presets.MustGetMessages(ctx.R)
	return v.VDialog(
		v.VCard(
			content,
			v.VCardActions(
				v.VSpacer(),
				v.VBtn(pmsgr.Cancel).
					Depressed(true).
					Class("ml-2").
					On("click", fmt.Sprintf("vars.%s = false", showVar)),

				v.VBtn(pmsgr.OK).
					Color("primary").
					Depressed(true).
					Dark(true).
					Attr("@click", web.Plaid().EventFunc("login_changePassword").Go()),
			),
		),
	).MaxWidth("600px").
		Attr("v-model", fmt.Sprintf("vars.%s", showVar)).
		Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, showVar))
}

func defaultChangePasswordDialogContent(vh *login.ViewHelper, pb *presets.Builder) func(ctx *web.EventContext) HTMLComponent {
	return func(ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		return Div(
			v.VCardTitle(Text(msgr.ChangePasswordTitle)),
			v.VCardText(
				Div(
					passwordInput("old_password", msgr.ChangePasswordOldPlaceholder, "", true).
						Outlined(false).
						Label(msgr.ChangePasswordOldLabel).
						FieldName("old_password"),
				),
				Div(
					passwordInputWithStrengthMeter(
						passwordInput("password", msgr.ChangePasswordNewPlaceholder, "", true).
							Outlined(false).
							Label(msgr.ChangePasswordNewLabel).
							FieldName("password"),
						"password", ""),
				).Class("mt-12"),
				Div(
					passwordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, "", true).
						Outlined(false).
						Label(msgr.ChangePasswordNewConfirmLabel).
						FieldName("confirm_password"),
				).Class("mt-12"),
				If(vh.TOTPEnabled(),
					Div(
						input("otp", msgr.TOTPValidateCodePlaceholder, "").
							Outlined(false).
							Label(msgr.TOTPValidateCodeLabel).
							FieldName("otp"),
					).Class("mt-12"),
				),
			),
		)
	}
}

func defaultTOTPSetupPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)

		user := login.GetCurrentUser(ctx.R)
		u := user.(login.UserPasser)

		var QRCode bytes.Buffer

		// Generate key from TOTPSecret
		var key *otp.Key
		totpSecret := u.GetTOTPSecret()
		if len(totpSecret) == 0 {
			r.Body = errorBody("need setup totp")
			return
		}
		key, err = otp.NewKeyFromURL(
			fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s",
				url.PathEscape(vh.TOTPIssuer()),
				url.PathEscape(u.GetAccountName()),
				url.QueryEscape(vh.TOTPIssuer()),
				url.QueryEscape(totpSecret),
			),
		)

		img, err := key.Image(200, 200)
		if err != nil {
			r.Body = errorBody(err.Error())
			return
		}

		err = png.Encode(&QRCode, img)
		if err != nil {
			r.Body = errorBody(err.Error())
			return
		}

		r.PageTitle = "TOTP Setup"
		r.Body = Div(
			errNotice(fMsg),
			Div(
				Div(
					H1(msgr.TOTPSetupTitle).
						Class(titleClass),
					Label(msgr.TOTPSetupScanPrompt),
				),
				Div(
					Img(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(QRCode.Bytes()))),
				).Class("d-flex justify-center my-2"),
				Div(
					Label(msgr.TOTPSetupSecretPrompt),
				),
				Div(Label(u.GetTOTPSecret())).Class("font-weight-bold my-4"),
				Form(
					Label(msgr.TOTPSetupEnterCodePrompt),
					input("otp", msgr.TOTPSetupCodePlaceholder, "").Class("mt-6"),
					formSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(vh.ValidateTOTPURL()),
			).Class(wrapperClass).Style(wrapperStyle).Class("text-center"),
		)

		return
	})
}

func defaultTOTPValidatePage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		fMsg := vh.GetFailFlashMessage(msgr, ctx.W, ctx.R)

		r.PageTitle = "TOTP Validate"
		r.Body = Div(
			errNotice(fMsg),
			Div(
				Div(
					H1(msgr.TOTPValidateTitle).
						Class(titleClass),
					Label(msgr.TOTPValidateEnterCodePrompt),
				),
				Form(
					input("otp", msgr.TOTPValidateCodePlaceholder, "").Autofocus(true).Class("mt-6"),
					formSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(vh.ValidateTOTPURL()),
			).Class(wrapperClass).Style(wrapperStyle).Class("text-center"),
		)

		return
	})
}
