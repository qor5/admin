package login

import (
	"fmt"
	"gorm.io/gorm"

	"github.com/goplaid/web"
	. "github.com/theplant/htmlgo"
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
}

var noticeCodeTexts = map[NoticeCode]string{
	NoticeCodePasswordSuccessfullyReset: "Password successfully reset",
}

func defaultLoginPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		fcText := failCodeTexts[fcFlash]
		ncFlash := GetNoticeCodeFlash(ctx.W, ctx.R)
		ncText := noticeCodeTexts[ncFlash]
		wlFlash := GetWrongLoginInputFlash(ctx.W, ctx.R)

		wrapperClass := "flex pt-8 h-screen flex-col max-w-md mx-auto"

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
			If(fcText != "",
				Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span(fcText).Class("block sm:inline"),
					),
			),
			If(ncText != "",
				Div().Class("bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span(ncText).Class("block sm:inline"),
					),
			),
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
			If(fcText != "",
				Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span(fcText).Class("block sm:inline"),
					),
			),
			If(secondsToResend > 0,
				Div().Class("bg-orange-100 border border-orange-400 text-orange-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span("Sending emails too frequently, please try again later").Class("block sm:inline"),
					),
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
			).Class("flex pt-8 h-screen flex-col max-w-md mx-auto pt-16"),
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
			).Class("flex pt-8 h-screen flex-col max-w-md mx-auto pt-16"),
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
			If(errMsg != "",
				Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span(errMsg).Class("block sm:inline"),
					),
			),
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
			).Class("flex pt-8 h-screen flex-col max-w-md mx-auto pt-16"),
		)
		return
	}
}
