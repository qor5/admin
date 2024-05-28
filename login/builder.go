package login

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
)

const (
	OpenChangePasswordDialogEvent = "login_openChangePasswordDialog"
)

func New(pb *presets.Builder) *login.Builder {
	r := login.New()
	r.I18n(pb.I18n())

	vh := r.ViewHelper()
	r.LoginPageFunc(defaultLoginPage(vh, pb))
	r.ForgetPasswordPageFunc(defaultForgetPasswordPage(vh, pb))
	r.ResetPasswordLinkSentPageFunc(defaultResetPasswordLinkSentPage(vh, pb))
	r.ResetPasswordPageFunc(defaultResetPasswordPage(vh, pb))
	r.ChangePasswordPageFunc(defaultChangePasswordPage(vh, pb))
	r.TOTPSetupPageFunc(defaultTOTPSetupPage(vh, pb))
	r.TOTPValidatePageFunc(defaultTOTPValidatePage(vh, pb))

	registerChangePasswordEvents(r, pb)

	return r
}

func registerChangePasswordEvents(b *login.Builder, pb *presets.Builder) {
	vh := b.ViewHelper()

	showVar := "showChangePasswordDialog"
	pb.GetWebBuilder().RegisterEventFunc(OpenChangePasswordDialogEvent, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DialogPortalName,
			Body: changePasswordDialog(vh, ctx, showVar, defaultChangePasswordDialogContent(vh, pb)(ctx)),
		})

		web.AppendRunScripts(&r, fmt.Sprintf(`
(function(){
var tag = document.createElement("script");
tag.src = "%s";
tag.onload= function(){
	vars.meter_score = function(x){return zxcvbn(x).score};
}
document.getElementsByTagName("head")[0].appendChild(tag);
})()
        `, login.ZxcvbnJSURL))
		return
	})

	pb.GetWebBuilder().RegisterEventFunc("login_changePassword", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		oldPassword := ctx.R.FormValue("old_password")
		password := ctx.R.FormValue("password")
		confirmPassword := ctx.R.FormValue("confirm_password")
		otp := ctx.R.FormValue("otp")

		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		err = b.ChangePassword(ctx.R, oldPassword, password, confirmPassword, otp)
		if err != nil {
			msg := msgr.ErrorSystemError
			var color string
			if ne, ok := err.(*login.NoticeError); ok {
				msg = ne.Message
				switch ne.Level {
				case login.NoticeLevel_Info:
					color = "info"
				case login.NoticeLevel_Warn:
					color = "warning"
				case login.NoticeLevel_Error:
					color = "error"
				}
			} else {
				switch err {
				case login.ErrWrongPassword:
					msg = msgr.ErrorIncorrectPassword
				case login.ErrEmptyPassword:
					msg = msgr.ErrorPasswordCannotBeEmpty
				case login.ErrPasswordNotMatch:
					msg = msgr.ErrorPasswordNotMatch
				case login.ErrWrongTOTPCode:
					msg = msgr.ErrorIncorrectTOTPCode
				case login.ErrTOTPCodeHasBeenUsed:
					msg = msgr.ErrorTOTPCodeReused
				}
				color = "error"
			}

			presets.ShowMessage(&r, msg, color)
			return r, nil
		}

		presets.ShowMessage(&r, msgr.InfoPasswordSuccessfullyChanged, "info")
		web.AppendRunScripts(&r, fmt.Sprintf("vars.%s = false", showVar))
		return r, nil
	})
}
