package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"net/url"

	v "github.com/goplaid/ui/vuetify"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/qor/qor5/presets"
	"github.com/pquerna/otp"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"gorm.io/gorm"
)

func getFailCodeText(msgr *Messages, code FailCode) string {
	switch code {
	case FailCodeSystemError:
		return msgr.ErrorSystemError
	case FailCodeCompleteUserAuthFailed:
		return msgr.ErrorCompleteUserAuthFailed
	case FailCodeUserNotFound:
		return msgr.ErrorUserNotFound
	case FailCodeIncorrectAccountNameOrPassword:
		return msgr.ErrorIncorrectAccountNameOrPassword
	case FailCodeUserLocked:
		return msgr.ErrorUserLocked
	case FailCodeAccountIsRequired:
		return msgr.ErrorAccountIsRequired
	case FailCodePasswordCannotBeEmpty:
		return msgr.ErrorPasswordCannotBeEmpty
	case FailCodePasswordNotMatch:
		return msgr.ErrorPasswordNotMatch
	case FailCodeIncorrectPassword:
		return msgr.ErrorIncorrectPassword
	case FailCodeInvalidToken:
		return msgr.ErrorInvalidToken
	case FailCodeTokenExpired:
		return msgr.ErrorTokenExpired
	case FailCodeIncorrectTOTPCode:
		return msgr.ErrorIncorrectTOTPCode
	case FailCodeTOTPCodeHasBeenUsed:
		return msgr.ErrorTOTPCodeReused
	case FailCodeIncorrectRecaptchaToken:
		return msgr.ErrorIncorrectRecaptchaToken
	}

	return ""
}

func getWarnCodeText(msgr *Messages, code WarnCode) string {
	switch code {
	case WarnCodePasswordHasBeenChanged:
		return msgr.WarnPasswordHasBeenChanged
	}
	return ""
}

func getInfoCodeText(msgr *Messages, code InfoCode) string {
	switch code {
	case InfoCodePasswordSuccessfullyReset:
		return msgr.InfoPasswordSuccessfullyReset
	case InfoCodePasswordSuccessfullyChanged:
		return msgr.InfoPasswordSuccessfullyChanged
	}
	return ""
}

const (
	otpKeyFormat = "otpauth://totp/%s:%s?issuer=%s&secret=%s"
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

func injectLoginAssets(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(fmt.Sprintf(`
<link rel="stylesheet" href="%s">
    `, styleCSSURL))
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
    `, zxcvbnJSURL))
}

type languageItem struct {
	Label string
	Value string
}

func defaultLoginPage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		// i18n start
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)
		var langs []languageItem
		var currLangVal string
		if ls := b.i18nBuilder.GetSupportLanguages(); len(ls) > 1 {
			qn := b.i18nBuilder.GetQueryName()
			lang := ctx.R.FormValue(qn)
			if lang == "" {
				lang = b.i18nBuilder.GetCurrentLangFromCookie(ctx.R)
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

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)
		wcFlash := GetWarnCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		wcText := getWarnCodeText(msgr, wcFlash)
		icFlash := GetInfoCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		icText := getInfoCodeText(msgr, icFlash)
		wlFlash := GetWrongLoginInputFlash(b.cookieConfig, ctx.W, ctx.R)

		if icFlash == InfoCodePasswordSuccessfullyChanged {
			wcText = ""
		}

		injectLoginAssets(ctx)

		wrapperClass := "tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto"

		var oauthHTML HTMLComponent
		if b.oauthEnabled {
			ul := Div().Class("tw-flex tw-flex-col tw-justify-center tw-mt-8 tw-text-center")
			for _, provider := range b.providers {
				ul.AppendChildren(
					v.VBtn("").
						Block(true).
						Large(true).
						Class("mt-4").
						Outlined(true).
						Href(fmt.Sprintf("%s?provider=%s", b.OAuthBeginURL, provider.Key)).
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

		isRecaptchaEnabled := b.recaptchaEnabled
		if isRecaptchaEnabled {
			injectRecaptchaAssets(ctx, "login-form", "token")
		}

		var userPassHTML HTMLComponent
		if b.userPassEnabled {
			wrapperClass += " tw-pt-16"
			userPassHTML = Div(
				Form(
					Div(
						Label(msgr.AccountLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("account"),
						input("account", msgr.AccountPlaceholder, wlFlash.Account),
					),
					Div(
						Label(msgr.PasswordLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("password"),
						passwordInput("password", msgr.PasswordPlaceholder, wlFlash.Password, true),
					).Class("tw-mt-6"),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					formSubmitBtn(msgr.SignInBtn).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", b.recaptchaConfig.SiteKey, isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled),
				).Id("login-form").Method(http.MethodPost).Action(b.PasswordLoginURL),
				If(!b.noForgetPasswordLink,
					Div(
						A(Text(msgr.ForgetPasswordLink)).Href(b.ForgetPasswordPageURL).
							Class("grey--text text--darken-1"),
					).Class("tw-text-right tw-mt-2"),
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
		).Class(wrapperClass)

		if b.defaultLoginPageFormWrapFunc != nil {
			bodyForm = b.defaultLoginPageFormWrapFunc(bodyForm)
		}

		r.Body = Div(
			errNotice(fcText),
			warnNotice(wcText),
			infoNotice(icText),
			bodyForm,
		)

		return
	})
}

func defaultForgetPasswordPage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)
		inputFlash := GetWrongForgetPasswordInputFlash(b.cookieConfig, ctx.W, ctx.R)
		secondsToResend := GetSecondsToRedoFlash(b.cookieConfig, ctx.W, ctx.R)
		activeBtnText := msgr.SendResetPasswordEmailBtn
		inactiveBtnText := msgr.ResendResetPasswordEmailBtn
		inactiveBtnTextWithInitSeconds := fmt.Sprintf("%s (%d)", inactiveBtnText, secondsToResend)

		doTOTP := ctx.R.URL.Query().Get("totp") == "1"
		actionURL := b.SendResetPasswordLinkURL
		if doTOTP {
			actionURL = mustSetQuery(actionURL, "totp", "1")
		}

		injectLoginAssets(ctx)

		isRecaptchaEnabled := b.recaptchaEnabled
		if isRecaptchaEnabled {
			injectRecaptchaAssets(ctx, "forget-form", "token")
		}

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			errNotice(fcText),
			If(secondsToResend > 0,
				warnNotice(msgr.SendEmailTooFrequentlyNotice),
			),
			Div(
				H1(msgr.ForgotMyPasswordTitle).Class("tw-leading-tight tw-text-3xl tw-mt-0 tw-mb-6"),
				Form(
					Div(
						Label(msgr.ForgetPasswordEmailLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("account"),
						input("account", msgr.ForgetPasswordEmailPlaceholder, inputFlash.Account),
					),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, inputFlash.TOTP),
						).Class("tw-mt-6"),
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
						AttrIf("data-sitekey", b.recaptchaConfig.SiteKey, isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled).
						Attr("id", "submitBtn").
						ClassIf("d-none", secondsToResend > 0),
				).Id("forget-form").Method(http.MethodPost).Action(actionURL),
			).Class("tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-pt-16"),
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

func defaultResetPasswordLinkSentPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		a := ctx.R.URL.Query().Get("a")

		injectLoginAssets(ctx)

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Div(
				H1(fmt.Sprintf("%s %s.", msgr.ResetPasswordLinkWasSentTo, a)).Class("tw-leading-tight tw-text-2xl tw-mt-0 tw-mb-4"),
				H2(msgr.ResetPasswordLinkSentPrompt).Class("tw-leading-tight tw-text-1xl tw-mt-0"),
			).Class("tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-pt-16"),
		)
		return
	}
}

func defaultResetPasswordPage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		errMsg := getFailCodeText(msgr, fcFlash)
		if errMsg == "" {
			errMsg = GetCustomErrorMessageFlash(b.cookieConfig, ctx.W, ctx.R)
		}
		wrpiFlash := GetWrongResetPasswordInputFlash(b.cookieConfig, ctx.W, ctx.R)

		doTOTP := ctx.R.URL.Query().Get("totp") == "1"
		actionURL := b.ResetPasswordURL
		if doTOTP {
			actionURL = mustSetQuery(actionURL, "totp", "1")
		}

		var user interface{}

		r.PageTitle = "Reset Password"

		query := ctx.R.URL.Query()
		id := query.Get("id")
		if id == "" {
			r.Body = Div(Text("user not found"))
			return r, nil
		} else {
			user, err = b.findUserByID(id)
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
			storedToken, _, expired := user.(UserPasser).GetResetPasswordToken()
			if expired {
				r.Body = Div(Text("token expired"))
				return r, nil
			}
			if token != storedToken {
				r.Body = Div(Text("invalid token"))
				return r, nil
			}
		}

		injectLoginAssets(ctx)
		injectZxcvbn(ctx)

		r.Body = Div(
			errNotice(errMsg),
			Div(
				H1(msgr.ResetYourPasswordTitle).Class("tw-leading-tight tw-text-3xl tw-mt-0 tw-mb-6"),
				Form(
					Input("user_id").Type("hidden").Value(id),
					Input("token").Type("hidden").Value(token),
					Div(
						Label(msgr.ResetPasswordLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("password"),
						passwordInput("password", msgr.ResetPasswordLabel, wrpiFlash.Password, true),
						passwordInputWithStrengthMeter(passwordInput("password", msgr.ResetPasswordLabel, wrpiFlash.Password, true), "password", wrpiFlash.Password),
					),
					Div(
						Label(msgr.ResetPasswordConfirmLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("confirm_password"),
						passwordInput("confirm_password", msgr.ResetPasswordConfirmPlaceholder, wrpiFlash.ConfirmPassword, true),
					).Class("tw-mt-6"),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, wrpiFlash.TOTP),
						).Class("tw-mt-6"),
					),
					formSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(actionURL),
			).Class("tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-pt-16"),
		)
		return
	})
}

func defaultChangePasswordPage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		errMsg := getFailCodeText(msgr, fcFlash)
		if errMsg == "" {
			errMsg = GetCustomErrorMessageFlash(b.cookieConfig, ctx.W, ctx.R)
		}
		inputFlash := GetWrongChangePasswordInputFlash(b.cookieConfig, ctx.W, ctx.R)

		injectLoginAssets(ctx)
		injectZxcvbn(ctx)

		r.PageTitle = "Change Password"

		r.Body = Div(
			errNotice(errMsg),
			Div(
				H1(msgr.ChangePasswordTitle).Class("tw-leading-tight tw-text-3xl tw-mt-0 tw-mb-6"),
				Form(
					Div(
						Label(msgr.ChangePasswordOldLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("old_password"),
						passwordInput("old_password", msgr.ChangePasswordOldPlaceholder, inputFlash.OldPassword, true),
					),
					Div(
						Label(msgr.ChangePasswordNewLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("password"),
						passwordInputWithStrengthMeter(passwordInput("password", msgr.ChangePasswordNewPlaceholder, inputFlash.NewPassword, true), "password", inputFlash.NewPassword),
					).Class("tw-mt-6"),
					Div(
						Label(msgr.ChangePasswordNewConfirmLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("confirm_password"),
						passwordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, inputFlash.ConfirmPassword, true),
					).Class("tw-mt-6"),
					If(b.totpEnabled,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class("tw-block tw-mb-2 tw-text-sm tw-text-gray-600 dark:tw-text-gray-200").For("otp"),
							input("otp", msgr.TOTPValidateCodePlaceholder, inputFlash.TOTP),
						).Class("tw-mt-6"),
					),
					formSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(b.ChangePasswordURL),
			).Class("tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-pt-16"),
		)
		return
	})
}

func changePasswordDialog(b *Builder, ctx *web.EventContext, showVar string, content HTMLComponent) HTMLComponent {
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

func defaultChangePasswordDialogContent(b *Builder) HTMLContentFunc {
	return func(ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)
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
				If(b.totpEnabled,
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

func defaultTOTPSetupPage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)

		user := GetCurrentUser(ctx.R)
		u := user.(UserPasser)

		var QRCode bytes.Buffer

		// Generate key from TOTPSecret
		var key *otp.Key
		totpSecret := u.GetTOTPSecret()
		if len(totpSecret) == 0 {
			r.Body = errorBody("need setup totp")
			return
		}
		key, err = otp.NewKeyFromURL(
			fmt.Sprintf(otpKeyFormat,
				url.PathEscape(b.totpIssuer),
				url.PathEscape(u.GetAccountName()),
				url.QueryEscape(b.totpIssuer),
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

		injectLoginAssets(ctx)

		wrapperClass := "tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-relative tw-text-center"
		labelClass := "tw-w-80 tw-text-sm tw-mb-8 tw-font-semibold tw-text-gray-700 tw-tracking-wide"

		r.PageTitle = "TOTP Setup"
		r.Body = Div(
			errNotice(fcText),
			Div(
				Div(
					H1(msgr.TOTPSetupTitle).
						Class("tw-text-3xl tw-font-bold tw-mb-4"),
					Label(msgr.TOTPSetupScanPrompt).
						Class(labelClass),
				),
				Div(
					Img(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(QRCode.Bytes()))),
				).Class("tw-my-2 tw-flex tw-items-center tw-justify-center"),
				Div(
					Label(msgr.TOTPSetupSecretPrompt).
						Class(labelClass),
				),
				Div(Label(u.GetTOTPSecret()).Class("tw-text-sm tw-font-bold")).Class("tw-my-4"),
				Form(
					Label(msgr.TOTPSetupEnterCodePrompt).Class(labelClass),
					input("otp", msgr.TOTPSetupCodePlaceholder, "").Class("mt-6"),
					formSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(b.ValidateTOTPURL),
			).Class(wrapperClass),
		)

		return
	})
}

func defaultTOTPValidatePage(b *Builder) web.PageFunc {
	return b.pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(b.cookieConfig, ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)

		injectLoginAssets(ctx)

		wrapperClass := "tw-flex tw-pt-8 tw-flex-col tw-max-w-md tw-mx-auto tw-relative tw-text-center"
		labelClass := "tw-w-80 tw-text-sm tw-mb-8 tw-font-semibold tw-text-gray-700 tw-tracking-wide"

		r.PageTitle = "TOTP Validate"
		r.Body = Div(
			errNotice(fcText),
			Div(
				Div(
					H1(msgr.TOTPValidateTitle).
						Class("tw-text-3xl tw-font-bold tw-mb-4"),
					Label(msgr.TOTPValidateEnterCodePrompt).
						Class(labelClass),
				),
				Form(
					input("otp", msgr.TOTPValidateCodePlaceholder, "").Autofocus(true).Class("mt-6"),
					formSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(b.ValidateTOTPURL),
			).Class(wrapperClass),
		)

		return
	})
}
