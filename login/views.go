package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/url"

	"github.com/goplaid/web"
	"github.com/pquerna/otp"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var failCodeTexts = map[FailCode]string{
	FailCodeSystemError:                    "System Error",
	FailCodeCompleteUserAuthFailed:         "Complete User Auth Failed",
	FailCodeUserNotFound:                   "User Not Found",
	FailCodeIncorrectAccountNameOrPassword: "Incorrect email or password",
	FailCodeUserLocked:                     "User Locked",
	FailCodeAccountIsRequired:              "Email is required",
	FailCodePasswordCannotBeEmpty:          "Password cannot be empty",
	FailCodePasswordNotMatch:               "Password do not match",
	FailCodeInvalidToken:                   "Invalid token",
	FailCodeTokenExpired:                   "Token expired",
	FailCodeIncorrectTOTP:                  "Incorrect passcode",
}

var warnCodeTexts = map[WarnCode]string{
	WarnCodePasswordHasBeenChanged: "Password has been changed",
}

var infoCodeTexts = map[InfoCode]string{
	InfoCodePasswordSuccessfullyReset: "Password successfully reset",
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

func defaultLoginPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := failCodeTexts[fcFlash]
		wcFlash := GetWarnCodeFlash(ctx.W, ctx.R)
		wcText := warnCodeTexts[wcFlash]
		ncFlash := GetInfoCodeFlash(ctx.W, ctx.R)
		ncText := infoCodeTexts[ncFlash]
		wlFlash := GetWrongLoginInputFlash(ctx.W, ctx.R)

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
					Input("login_type").Type("hidden").Value("1"),
					Div(
						Label("Email").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("account"),
						Input("account").Placeholder("Email").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
							Value(wlFlash.Ia),
					),
					Div(
						Label("Password").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
						Input("password").Placeholder("Password").Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
							Value(wlFlash.Ip),
					).Class("mt-6"),
					Div(
						Button("Sign In").Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
					).Class("mt-6"),
				).Method("post").Action("/auth/userpass/login"),
				If(!b.noForgetPasswordLink,
					Div(
						A(Text("Forget your password?")).Href("/auth/forget-password").
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
			infoNotice(ncText),
			Div(
				userPassHTML,
				oauthHTML,
			).Class(wrapperClass),
		)
		return
	}
}

func defaultForgetPasswordPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := failCodeTexts[fcFlash]
		inputFlash := GetWrongForgetPasswordInputFlash(ctx.W, ctx.R)
		secondsToResend := GetSecondsToRedoFlash(ctx.W, ctx.R)
		activeBtnText := "Send reset password email"
		activeBtnClass := "w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"
		inactiveBtnText := "Resend reset password email"
		inactiveBtnClass := "w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-gray-500 rounded-md"
		inactiveBtnTextWithInitSeconds := fmt.Sprintf("%s (%d)", inactiveBtnText, secondsToResend)

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			If(secondsToResend > 0,
				warnNotice("Sending emails too frequently, please try again later"),
			),
			Div(
				H1("I forgot my password").Class("leading-tight text-3xl mt-0 mb-6"),
				Form(
					Div(
						Label("Enter your email").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("account"),
						Input("account").Placeholder("email").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(inputFlash.Account),
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
		a := ctx.R.URL.Query().Get("a")

		r.PageTitle = "Forget Your Password?"
		r.Body = Div(
			Style(StyleCSS),
			Div(
				H1(fmt.Sprintf("A reset password link was sent to %s.", a)).Class("leading-tight text-2xl mt-0 mb-4"),
				H2("You can close this page and reset your password from this link.").Class("leading-tight text-1xl mt-0"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}

func defaultResetPasswordPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		errMsg := failCodeTexts[fcFlash]
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
				H1("Reset your password").Class("leading-tight text-3xl mt-0 mb-6"),
				Form(
					Input("user_id").Type("hidden").Value(id),
					Input("token").Type("hidden").Value(token),
					Div(
						Label("Change password").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
						Input("password").Placeholder("Password").Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(wrpiFlash.Password),
					),
					Div(
						Label("Re-enter new password").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("confirm_password"),
						Input("confirm_password").Placeholder("Password").Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").Value(wrpiFlash.ConfirmPassword),
					).Class("mt-6"),
					Div(
						Button("Confirm").Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
					).Class("mt-6"),
				).Method("post").Action("/auth/do-reset-password"),
			).Class("flex pt-8 flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}

func defaultTOTPSetupPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := failCodeTexts[fcFlash]

		user := GetCurrentUser(ctx.R)
		u := user.(UserPasser)

		var QRCode bytes.Buffer

		// Generate key from TOTPSecret
		var key *otp.Key
		totpSecret := u.GetTOTPSecret()
		if len(totpSecret) == 0 {
			r.Body = errorBody(errNeedTOTPSetup.Error())
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
					H1("Two Factor Authentication").
						Class("text-3xl font-bold mb-4"),
					Label("Scan this QR code with Google Authenticator (or similar) app").
						Class(labelClass),
				),
				Div(
					Img(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(QRCode.Bytes()))),
				).Class("my-2 flex items-center justify-center"),
				Div(
					Label("Or manually enter the following code into your preferred authenticator app").
						Class(labelClass),
				),
				Div(Label(u.GetTOTPSecret()).Class("text-sm font-bold")).Class("my-4"),
				Form(
					Label("Then enter the provided one-time code below").Class(labelClass),
					Input("otp").Placeholder("Enter your passcode here").
						Class("my-6 block w-full px-4 py-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40"),
					Button("Verify").
						Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
				).Method("POST").Action(pathTOTPDo),
			).Class(wrapperClass),
		)

		return
	}
}

func defaultTOTPValidatePage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := failCodeTexts[fcFlash]

		wrapperClass := "flex pt-8 flex-col max-w-md mx-auto relative text-center"
		labelClass := "w-80 text-sm mb-8 font-semibold text-gray-700 tracking-wide"

		r.PageTitle = "TOTP Validate"
		r.Body = Div(
			Style(StyleCSS),
			errNotice(fcText),
			Div(
				Div(
					H1("Two Factor Authentication").
						Class("text-3xl font-bold mb-4"),
					Label("Enter the provided one-time code below").
						Class(labelClass),
				),
				Form(
					Input("otp").Placeholder("Enter your passcode here").
						Class("my-6 block w-full px-4 py-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40"),
					Button("Verify").
						Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
				).Method("POST").Action(pathTOTPDo),
			).Class(wrapperClass),
		)

		return
	}
}

func errorBody(msg string) HTMLComponent {
	return Div(
		Text(msg),
	)
}
