package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"net/url"

	"github.com/pquerna/otp"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	. "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

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
					VBtn("").
						Block(true).
						Size(SizeLarge).
						Class("mt-4").
						Variant(VariantOutlined).
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

		wIn := vh.GetWrongLoginInputFlash(ctx.W, ctx.R)
		isRecaptchaEnabled := vh.RecaptchaEnabled()
		if isRecaptchaEnabled {
			DefaultViewCommon.InjectRecaptchaAssets(ctx, "login-form", "token")
		}

		var userPassHTML HTMLComponent
		if vh.UserPassEnabled() {
			userPassHTML = Div(
				Form(
					Div(
						Label(msgr.AccountLabel).Class(DefaultViewCommon.LabelClass).For("account"),
						DefaultViewCommon.Input("account", msgr.AccountPlaceholder, wIn.Account),
					),
					Div(
						Label(msgr.PasswordLabel).Class(DefaultViewCommon.LabelClass).For("password"),
						DefaultViewCommon.PasswordInput("password", msgr.PasswordPlaceholder, wIn.Password, true),
					).Class("mt-6"),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					DefaultViewCommon.FormSubmitBtn(msgr.SignInBtn).
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

		r.PageTitle = msgr.LoginPageTitle
		var bodyForm HTMLComponent
		bodyForm = Div(
			userPassHTML,
			oauthHTML,
			If(len(langs) > 0,
				web.Scope(
					VSelect().
						Items(langs).
						ItemTitle("Label").
						ItemValue("Value").
						Attr("v-model", `selectLocals.currLangVal`).
						Attr("@update:model-value", web.Plaid().MergeQuery(true).Query(qn, web.Var("selectLocals.currLangVal")).PushState(true).Go()).
						Variant(VariantOutlined).
						Density(DensityCompact).
						Class("mt-12").
						HideDetails(true),
				).VSlot(" { locals : selectLocals } ").Init(fmt.Sprintf(`{currLangVal: '%s'}`, currLangVal)),
			),
		).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle)

		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			bodyForm,
		)

		return
	})
}

func defaultForgetPasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

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
			DefaultViewCommon.InjectRecaptchaAssets(ctx, "forget-form", "token")
		}

		r.PageTitle = msgr.ForgetPasswordPageTitle
		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			If(secondsToResend > 0,
				DefaultViewCommon.WarnNotice(msgr.SendEmailTooFrequentlyNotice),
			),
			Div(
				H1(msgr.ForgotMyPasswordTitle).Class(DefaultViewCommon.TitleClass),
				Form(
					Div(
						Label(msgr.ForgetPasswordEmailLabel).Class(DefaultViewCommon.LabelClass).For("account"),
						DefaultViewCommon.Input("account", msgr.ForgetPasswordEmailPlaceholder, wIn.Account),
					),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(DefaultViewCommon.LabelClass).For("otp"),
							DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					DefaultViewCommon.FormSubmitBtn(inactiveBtnTextWithInitSeconds).
						Attr("id", "disabledBtn").
						ClassIf("d-none", secondsToResend <= 0),
					DefaultViewCommon.FormSubmitBtn(activeBtnText).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", vh.RecaptchaSiteKey(), isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled).
						Attr("id", "submitBtn").
						ClassIf("d-none", secondsToResend > 0),
				).Id("forget-form").Method(http.MethodPost).Action(actionURL),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle),
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

		r.PageTitle = msgr.ResetPasswordLinkSentPageTitle
		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			Div(
				H1(fmt.Sprintf("%s %s.", msgr.ResetPasswordLinkWasSentTo, a)).Class("text-h5"),
				H2(msgr.ResetPasswordLinkSentPrompt).Class("text-body-1 mt-2"),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle),
		)
		return
	})
}

func defaultResetPasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		wIn := vh.GetWrongResetPasswordInputFlash(ctx.W, ctx.R)

		doTOTP := ctx.R.URL.Query().Get("totp") == "1"
		actionURL := vh.ResetPasswordURL()
		if doTOTP {
			actionURL = login.MustSetQuery(actionURL, "totp", "1")
		}

		var user interface{}

		r.PageTitle = msgr.ResetPasswordPageTitle

		query := ctx.R.URL.Query()
		id := query.Get("id")
		if id == "" {
			r.Body = Div(Text("user not found"))
			return r, nil
		} else {
			user, err = vh.FindUserByID(id)
			if err != nil {
				if err == login.ErrUserNotFound {
					r.Body = Div(Text("user not found"))
					return r, nil
				}
				panic(err)
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

		DefaultViewCommon.InjectZxcvbn(ctx)

		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			Div(
				H1(msgr.ResetYourPasswordTitle).Class(DefaultViewCommon.TitleClass),
				Form(
					Input("user_id").Type("hidden").Value(id),
					Input("token").Type("hidden").Value(token),
					Div(
						Label(msgr.ResetPasswordLabel).Class(DefaultViewCommon.LabelClass).For("password"),
						DefaultViewCommon.PasswordInputWithStrengthMeter(DefaultViewCommon.PasswordInput("password", msgr.ResetPasswordLabel, wIn.Password, true), "password", wIn.Password),
					),
					Div(
						Label(msgr.ResetPasswordConfirmLabel).Class(DefaultViewCommon.LabelClass).For("confirm_password"),
						DefaultViewCommon.PasswordInput("confirm_password", msgr.ResetPasswordConfirmPlaceholder, wIn.ConfirmPassword, true),
					).Class("mt-6"),
					If(doTOTP,
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(DefaultViewCommon.LabelClass).For("otp"),
							DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					DefaultViewCommon.FormSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(actionURL),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle),
		)
		return
	})
}

func defaultChangePasswordPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		wIn := vh.GetWrongChangePasswordInputFlash(ctx.W, ctx.R)

		DefaultViewCommon.InjectZxcvbn(ctx)

		r.PageTitle = msgr.ChangePasswordPageTitle

		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			Div(
				H1(msgr.ChangePasswordTitle).Class(DefaultViewCommon.TitleClass),
				Form(
					Div(
						Label(msgr.ChangePasswordOldLabel).Class(DefaultViewCommon.LabelClass).For("old_password"),
						DefaultViewCommon.PasswordInput("old_password", msgr.ChangePasswordOldPlaceholder, wIn.OldPassword, true),
					),
					Div(
						Label(msgr.ChangePasswordNewLabel).Class(DefaultViewCommon.LabelClass).For("password"),
						DefaultViewCommon.PasswordInputWithStrengthMeter(DefaultViewCommon.PasswordInput("password", msgr.ChangePasswordNewPlaceholder, wIn.NewPassword, true), "password", wIn.NewPassword),
					).Class("mt-6"),
					Div(
						Label(msgr.ChangePasswordNewConfirmLabel).Class(DefaultViewCommon.LabelClass).For("confirm_password"),
						DefaultViewCommon.PasswordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, wIn.ConfirmPassword, true),
					).Class("mt-6"),
					If(vh.TOTPEnabled(),
						Div(
							Label(msgr.TOTPValidateCodeLabel).Class(DefaultViewCommon.LabelClass).For("otp"),
							DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, wIn.TOTP),
						).Class("mt-6"),
					),
					DefaultViewCommon.FormSubmitBtn(msgr.Confirm),
				).Method(http.MethodPost).Action(vh.ChangePasswordURL()),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle),
		)
		return
	})
}

func changePasswordDialog(_ *login.ViewHelper, ctx *web.EventContext, showVar string, content HTMLComponent) HTMLComponent {
	pmsgr := presets.MustGetMessages(ctx.R)
	return web.Scope(VDialog(
		VCard(
			content,
			VCardActions(
				VSpacer(),
				VBtn(pmsgr.Cancel).
					Variant(VariantFlat).
					Class("ml-2").
					On("click", fmt.Sprintf("dialogLocals.%s = false", showVar)),

				VBtn(pmsgr.OK).
					Color("primary").
					Variant(VariantFlat).
					Theme(ThemeDark).
					Attr("@click", web.Plaid().EventFunc("login_changePassword").Go()),
			),
		),
	).MaxWidth("600px").
		Attr("v-model", fmt.Sprintf("dialogLocals.%s", showVar)),
	).VSlot(" { locals : dialogLocals}").Init(fmt.Sprintf(`{%s: true}`, showVar))
}

func defaultChangePasswordDialogContent(vh *login.ViewHelper, _ *presets.Builder) func(ctx *web.EventContext) HTMLComponent {
	return func(ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		return Div(
			VCardTitle(Text(msgr.ChangePasswordTitle)),
			VCardText(
				Div(
					DefaultViewCommon.PasswordInput("old_password", msgr.ChangePasswordOldPlaceholder, "", true).
						Label(msgr.ChangePasswordOldLabel).
						Attr(web.VField("old_password", "")...),
				),
				Div(
					DefaultViewCommon.PasswordInputWithStrengthMeter(
						DefaultViewCommon.PasswordInput("password", msgr.ChangePasswordNewPlaceholder, "", true).
							Label(msgr.ChangePasswordNewLabel).
							Attr(web.VField("password", "")...),
						"password", ""),
				).Class("mt-12"),
				Div(
					DefaultViewCommon.PasswordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, "", true).
						Label(msgr.ChangePasswordNewConfirmLabel).
						Attr(web.VField("confirm_password", "")...),
				).Class("mt-12"),
				If(vh.TOTPEnabled(),
					Div(
						DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, "").
							Label(msgr.TOTPValidateCodeLabel).
							Attr(web.VField("otp", "")...),
					).Class("mt-12"),
				),
			),
		)
	}
}

func defaultTOTPSetupPage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		user := login.GetCurrentUser(ctx.R)
		u := user.(login.UserPasser)

		var QRCode bytes.Buffer

		// Generate key from TOTPSecret
		var key *otp.Key
		totpSecret := u.GetTOTPSecret()
		if len(totpSecret) == 0 {
			r.Body = DefaultViewCommon.ErrorBody("need setup totp")
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
			r.Body = DefaultViewCommon.ErrorBody(err.Error())
			return
		}

		err = png.Encode(&QRCode, img)
		if err != nil {
			r.Body = DefaultViewCommon.ErrorBody(err.Error())
			return
		}

		r.PageTitle = msgr.TOTPSetupPageTitle
		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			Div(
				Div(
					H1(msgr.TOTPSetupTitle).
						Class(DefaultViewCommon.TitleClass),
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
					DefaultViewCommon.Input("otp", msgr.TOTPSetupCodePlaceholder, "").Class("mt-6"),
					DefaultViewCommon.FormSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(vh.ValidateTOTPURL()),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle).Class("text-center"),
		)

		return
	})
}

func defaultTOTPValidatePage(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)

		r.PageTitle = msgr.TOTPValidatePageTitle
		r.Body = Div(
			DefaultViewCommon.Notice(vh, msgr, ctx.W, ctx.R),
			Div(
				Div(
					H1(msgr.TOTPValidateTitle).
						Class(DefaultViewCommon.TitleClass),
					Label(msgr.TOTPValidateEnterCodePrompt),
				),
				Form(
					DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, "").Autofocus(true).Class("mt-6"),
					DefaultViewCommon.FormSubmitBtn(msgr.Verify),
				).Method(http.MethodPost).Action(vh.ValidateTOTPURL()),
			).Class(DefaultViewCommon.WrapperClass).Style(DefaultViewCommon.WrapperStyle).Class("text-center"),
		)

		return
	})
}
