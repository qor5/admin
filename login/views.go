package login

import (
	"github.com/goplaid/web"

	. "github.com/theplant/htmlgo"
)

var loginFailTexts = map[FailCode]string{
	FailCodeSystemError:                 "System Error",
	FailCodeCompleteUserAuthFailed:      "Complete User Auth Failed",
	FailCodeUserNotFound:                "User Not Found",
	FailCodeIncorrectUsernameOrPassword: "Incorrect username or password",
}

func defaultLoginPage(b *Builder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		fcFlash := GetFailCodeFlash(ctx.W, ctx.R)
		wlFlash := GetWrongLoginInputFlash(ctx.W, ctx.R)
		loginFailText := loginFailTexts[fcFlash]

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
						Label("Username").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("username"),
						Input("username").Placeholder("Username").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
							Value(wlFlash.Iu),
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
			)
		}

		r.PageTitle = "Sign In"
		r.Body = Div(
			Style(StyleCSS),
			If(loginFailText != "",
				Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center -mb-8").
					Role("alert").
					Children(
						Span(loginFailText).Class("block sm:inline"),
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
