package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/url"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/pquerna/otp"
	. "github.com/theplant/htmlgo"
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
	case FailCodeIncorrectTOTP:
		return msgr.ErrorIncorrectTOTP
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

	return Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center").
		Role("alert").
		Children(
			Span(msg).Class("block sm:inline"),
		)
}

func warnNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return Div().Class("bg-orange-100 border border-orange-400 text-orange-700 px-4 py-3 rounded relative text-center").
		Role("alert").
		Children(
			Span(msg).Class("block sm:inline"),
		)
}

func infoNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return Div().Class("bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded relative text-center").
		Role("alert").
		Children(
			Span(msg).Class("block sm:inline"),
		)
}

func errorBody(msg string) HTMLComponent {
	return Div(
		Text(msg),
	)
}

func defaultLoginPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		// i18n start
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)
		var languages []HTMLComponent
		{
			i18nBuilder := b.i18nBuilder
			currentLang := i18nBuilder.GetCurrentLangFromCookie(ctx.R)
			ls := i18nBuilder.GetSupportLanguages()
			for _, l := range ls {
				elem := Option(display.Self.Name(l)).Value(l.String())
				if currentLang == l.String() {
					elem.Attr("selected", "selected")
				}
				languages = append(languages, elem)
			}
		}
		// i18n end

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)
		wcFlash := GetWarnCodeFlash(ctx.W, ctx.R)
		wcText := getWarnCodeText(msgr, wcFlash)
		icFlash := GetInfoCodeFlash(ctx.W, ctx.R)
		icText := getInfoCodeText(msgr, icFlash)
		wlFlash := GetWrongLoginInputFlash(ctx.W, ctx.R)

		if icFlash == InfoCodePasswordSuccessfullyChanged {
			wcText = ""
		}

		wrapperClass := "flex pt-8 flex-col max-w-md mx-auto"

		var oauthHTML HTMLComponent
		if b.oauthEnabled {
			ul := Div().Class("flex flex-col justify-center mt-8 text-center")
			for _, provider := range b.providers {
				ul.AppendChildren(
					A().
						Href("/auth/begin?provider="+provider.Key).
						Class("px-6 py-3 mt-4 font-semibold text-gray-900 bg-white border-2 border-gray-500 rounded-md shadow outline-none hover:bg-yellow-50 hover:border-yellow-400 focus:outline-none").
						Children(
							provider.Logo,
							Text(provider.Text),
						),
				)
			}

			oauthHTML = Div(
				ul,
			)
		}

		var userPassHTML HTMLComponent
		if b.userPassEnabled {
			wrapperClass += " pt-16"
			userPassHTML = Div(
				Form(
					Div(
						Label(msgr.AccountLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("account"),
						Input("account").Placeholder(msgr.AccountPlaceholder).Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
							Value(wlFlash.Ia),
					),
					Div(
						Label(msgr.PasswordLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
						Input("password").Placeholder(msgr.PasswordPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
							Value(wlFlash.Ip),
					).Class("mt-6"),
					Div(
						Button(msgr.SignInBtn).Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
					).Class("mt-6"),
				).Method("post").Action("/auth/userpass/login"),
				If(!b.noForgetPasswordLink,
					Div(
						A(Text(msgr.ForgetPasswordLink)).Href("/auth/forget-password").
							Class("text-gray-500"),
					).Class("text-right mt-2"),
				),
			)
		}

		r.PageTitle = "Sign In"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			warnNotice(wcText),
			infoNotice(icText),
			Div(
				userPassHTML,
				oauthHTML,
				Select(
					languages...,
				).Id("lang").Attr("onchange", `
const lang = document.getElementById("lang");
document.cookie="lang=" + lang.value + "; path=/";
location.reload();
`).Class("mt-12 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500").
					ClassIf("hidden", len(languages) < 2),
			).Class(wrapperClass),
		)

		return
	}
}

func defaultForgetPasswordPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)
		inputFlash := GetWrongForgetPasswordInputFlash(ctx.W, ctx.R)
		secondsToResend := GetSecondsToRedoFlash(ctx.W, ctx.R)
		activeBtnText := msgr.SendResetPasswordEmailBtn
		activeBtnClass := "w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"
		inactiveBtnText := msgr.ResendResetPasswordEmailBtn
		inactiveBtnClass := "w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-gray-500 rounded-md"
		inactiveBtnTextWithInitSeconds := fmt.Sprintf("%s (%d)", inactiveBtnText, secondsToResend)

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			If(secondsToResend > 0,
				warnNotice(msgr.SendEmailTooFrequentlyNotice),
			),
			Div(
				H1(msgr.ForgotMyPasswordTitle).Class("leading-tight text-3xl mt-0 mb-6"),
				Form(
					Div(
						Label(msgr.ForgetPasswordEmailLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("account"),
						Input("account").Placeholder(msgr.ForgetPasswordEmailPlaceholder).Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(inputFlash.Account),
					),
					Div(
						If(secondsToResend > 0,
							Button(inactiveBtnTextWithInitSeconds).Id("submitBtn").Class(inactiveBtnClass).Disabled(true),
						).Else(
							Button(activeBtnText).Class(activeBtnClass),
						),
					).Class("mt-6"),
				).Method("post").Action("/auth/send-reset-password-link"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)

		if secondsToResend > 0 {
			ctx.Injector.TailHTML(fmt.Sprintf(`
<script>
var secondsToResend = %d;
var btnText = "%s";
var submitBtn = document.getElementById("submitBtn");
var interv = setInterval(function(){
    secondsToResend--;
    if (secondsToResend === 0) {
        clearInterval(interv);
        submitBtn.innerText = btnText;
        submitBtn.className = "%s";
        submitBtn.disabled = false;
        return;
    }
    submitBtn.innerText = btnText + " (" + secondsToResend + ")" ;
}, 1000);
</script>
        `, secondsToResend, inactiveBtnText, activeBtnClass))
		}
		return
	}
}

func defaultResetPasswordLinkSentPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		a := ctx.R.URL.Query().Get("a")

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Style(StyleCSS),
			Div(
				H1(fmt.Sprintf("%s %s.", msgr.ResetPasswordLinkWasSentTo, a)).Class("leading-tight text-2xl mt-0 mb-4"),
				H2(msgr.ResetPasswordLinkSentPrompt).Class("leading-tight text-1xl mt-0"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}

func defaultResetPasswordPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		errMsg := getFailCodeText(msgr, fcFlash)
		if errMsg == "" {
			errMsg = GetCustomErrorMessageFlash(ctx.W, ctx.R)
		}
		wrpiFlash := GetWrongResetPasswordInputFlash(ctx.W, ctx.R)

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

		r.Body = Div(
			Style(StyleCSS),
			errNotice(errMsg),
			Div(
				H1(msgr.ResetYourPasswordTitle).Class("leading-tight text-3xl mt-0 mb-6"),
				Form(
					Input("user_id").Type("hidden").Value(id),
					Input("token").Type("hidden").Value(token),
					Div(
						Label(msgr.ResetPasswordLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
						Input("password").Placeholder(msgr.ResetPasswordPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(wrpiFlash.Password),
					),
					Div(
						Label(msgr.ResetPasswordConfirmLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("confirm_password"),
						Input("confirm_password").Placeholder(msgr.ResetPasswordConfirmPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(wrpiFlash.ConfirmPassword),
					).Class("mt-6"),
					Div(
						Button(msgr.Confirm).Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
					).Class("mt-6"),
				).Method("post").Action("/auth/do-reset-password"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}

func defaultChangePasswordPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		errMsg := getFailCodeText(msgr, fcFlash)
		if errMsg == "" {
			errMsg = GetCustomErrorMessageFlash(ctx.W, ctx.R)
		}
		input := GetWrongChangePasswordInputFlash(ctx.W, ctx.R)

		r.PageTitle = "Change Password"

		r.Body = Div(
			Style(StyleCSS),
			errNotice(errMsg),
			Div(
				H1(msgr.ChangePasswordTitle).Class("leading-tight text-3xl mt-0 mb-6"),
				Form(
					Div(
						Label(msgr.ChangePasswordOldLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("old_password"),
						Input("old_password").Placeholder(msgr.ChangePasswordOldPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(input.OldPassword),
					),
					Div(
						Label(msgr.ChangePasswordNewLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
						Input("password").Placeholder(msgr.ChangePasswordNewPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(input.NewPassword),
					).Class("mt-6"),
					Div(
						Label(msgr.ChangePasswordNewConfirmLabel).Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("confirm_password"),
						Input("confirm_password").Placeholder(msgr.ChangePasswordNewConfirmPlaceholder).Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(input.ConfirmPassword),
					).Class("mt-6"),
					Div(
						Button(msgr.Confirm).Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
					).Class("mt-6"),
				).Method("post").Action("/auth/do-change-password"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}

func defaultTOTPSetupPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
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

		wrapperClass := "flex pt-8 flex-col max-w-md mx-auto relative text-center"
		labelClass := "w-80 text-sm mb-8 font-semibold text-gray-700 tracking-wide"

		r.PageTitle = "TOTP Setup"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			Div(
				Div(
					H1(msgr.TOTPSetupTitle).
						Class("text-3xl font-bold mb-4"),
					Label(msgr.TOTPSetupScanPrompt).
						Class(labelClass),
				),
				Div(
					Img(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(QRCode.Bytes()))),
				).Class("my-2 flex items-center justify-center"),
				Div(
					Label(msgr.TOTPSetupSecretPrompt).
						Class(labelClass),
				),
				Div(Label(u.GetTOTPSecret()).Class("text-sm font-bold")).Class("my-4"),
				Form(
					Label(msgr.TOTPSetupEnterCodePrompt).Class(labelClass),
					Input("otp").Placeholder(msgr.TOTPSetupCodePlaceholder).
						Class("my-6 block w-full px-4 py-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40"),
					Button(msgr.Verify).
						Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
				).Method("POST").Action(totpDoURL),
			).Class(wrapperClass),
		)

		return
	}
}

func defaultTOTPValidatePage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nLoginKey, Messages_en_US).(*Messages)

		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := getFailCodeText(msgr, fcFlash)

		wrapperClass := "flex pt-8 flex-col max-w-md mx-auto relative text-center"
		labelClass := "w-80 text-sm mb-8 font-semibold text-gray-700 tracking-wide"

		r.PageTitle = "TOTP Validate"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			Div(
				Div(
					H1(msgr.TOTPValidateTitle).
						Class("text-3xl font-bold mb-4"),
					Label(msgr.TOTPValidateEnterCodePrompt).
						Class(labelClass),
				),
				Form(
					Input("otp").Placeholder(msgr.TOTPValidateCodePlaceholder).
						Class("my-6 block w-full px-4 py-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40"),
					Button(msgr.Verify).
						Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
				).Method("POST").Action(totpDoURL),
			).Class(wrapperClass),
		)

		return
	}
}
