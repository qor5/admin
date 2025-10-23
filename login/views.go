package login

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"net/url"

	_ "embed"

	"github.com/pquerna/otp"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
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
		bodyForm := Div(
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

type OAuthProviderDisplay struct {
	Logo HTMLComponent
	Text string
}

type AdvancedLoginPageConfig struct {
	WelcomeLabel         string
	TitleLabel           string
	AccountLabel         string
	AccountPlaceholder   string
	PasswordLabel        string
	PasswordPlaceholder  string
	SignInButtonLabel    string
	ForgetPasswordLabel  string
	OAuthProviderDisplay func(provider *login.Provider) OAuthProviderDisplay
	BrandLogo            HTMLComponent
	LeftImage            HTMLComponent
	Qor5DescriptionLabel string
	RightMaxWidth        string
	PageTitle            string
}

//go:embed embed/defaultLeftImage.jpg
var defaultLeftImage []byte

func NewAdvancedLoginPage(customize func(ctx *web.EventContext, config *AdvancedLoginPageConfig) (*AdvancedLoginPageConfig, error)) func(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
	return func(vh *login.ViewHelper, pb *presets.Builder) web.PageFunc {
		return pb.PlainLayout(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nAdminLoginKey, Messages_en_US).(*Messages)

			config := &AdvancedLoginPageConfig{
				WelcomeLabel:         msgr.LoginWelcomeLabel,
				TitleLabel:           msgr.LoginTitleLabel,
				AccountLabel:         msgr.LoginAccountLabel,
				AccountPlaceholder:   msgr.LoginAccountPlaceholder,
				PasswordLabel:        msgr.LoginPasswordLabel,
				PasswordPlaceholder:  msgr.LoginPasswordPlaceholder,
				SignInButtonLabel:    msgr.LoginSignInButtonLabel,
				ForgetPasswordLabel:  msgr.LoginForgetPasswordLabel,
				Qor5DescriptionLabel: msgr.LoginQor5DescriptionLabel,
				OAuthProviderDisplay: func(provider *login.Provider) OAuthProviderDisplay {
					return OAuthProviderDisplay{
						Logo: provider.Logo,
						Text: i18n.T(ctx.R, I18nAdminLoginKey, provider.Text),
					}
				},
				BrandLogo: nil,
				LeftImage: VImg().Class("fill-height").Cover(true).Src("data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(defaultLeftImage)),
				// LeftImage: VImg().Class("fill-height").Cover(true).Src("https://cdn.vuetifyjs.com/images/parallax/material2.jpg"),
				// LeftImage: VImg().Attr("style", "height: 100vh").Cover(true).Src("https://cdn.vuetifyjs.com/images/parallax/material2.jpg"),
				RightMaxWidth: "455px",
				PageTitle:     msgr.LoginTitleLabel,
			}
			if customize != nil {
				config, err = customize(ctx, config)
				if err != nil {
					return r, err
				}
			}

			// i18n start
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

			leftCompo := VCol().Cols(0).Md(6).Class("hidden-md-and-down").Children(
				config.LeftImage,
				Div().Class("position-absolute").Style("top: 50px; left: 50px;").Children(
					RawHTML(
						`<svg width="61" height="27" viewBox="0 0 61 27" fill="none" xmlns="http://www.w3.org/2000/svg">
							<path fill-rule="evenodd" clip-rule="evenodd" d="M40.6667 0H61V20.25H40.6667V0ZM47.4445 6.75H54.2222V13.5H47.4445V6.75ZM47.4445 13.5V20.25H54.2222L47.4445 13.5Z" fill="#17A2F5"/>
							<path d="M33.889 6.75041H27.1112V13.5004H33.889V6.75041Z" fill="#17A2F5"/>
							<path fill-rule="evenodd" clip-rule="evenodd" d="M0 0H20.3332V20.25L0 20.25V0ZM6.77777 6.75H13.5555V13.5H6.77777V6.75ZM20.3333 27L20.3332 20.25L13.5555 20.25L20.3333 27Z" fill="#17A2F5"/>
						</svg>
						`,
					),
				),
			)

			var oauthCompo HTMLComponent
			if vh.OAuthEnabled() {
				var buttons []HTMLComponent
				for _, provider := range vh.OAuthProviders() {
					display := config.OAuthProviderDisplay(provider)
					buttons = append(buttons,
						VBtn("").Class("bg-grey-lighten-4").
							Block(true).
							Size(SizeLarge).
							Variant(VariantFlat).
							Href(fmt.Sprintf("%s?provider=%s", vh.OAuthBeginURL(), provider.Key)).
							Children(
								Div().Class("d-flex flex-row ga-2 text-body-1").Children(
									display.Logo,
									Div().Class("text-body1").Text(display.Text),
								),
							),
					)
				}
				if len(buttons) > 0 {
					oauthCompo = Div().Class("d-flex flex-column ga-6").Children(buttons...)
				}
			}

			wIn := vh.GetWrongLoginInputFlash(ctx.W, ctx.R)
			isRecaptchaEnabled := vh.RecaptchaEnabled()
			if isRecaptchaEnabled {
				DefaultViewCommon.InjectRecaptchaAssets(ctx, "login-form", "token")
			}

			var userPassCompo HTMLComponent
			if vh.UserPassEnabled() {
				compo := Form().Class("login-form d-flex flex-column").Id("login-form").Method(http.MethodPost).Action(vh.PasswordLoginURL())
				compo.AppendChildren(
					DefaultViewCommon.Input("account", config.AccountPlaceholder, wIn.Account).Class("mb-5").Label(config.AccountLabel),
					DefaultViewCommon.PasswordInput("password", config.PasswordPlaceholder, wIn.Password, true).Class("mb-5").Label(config.PasswordLabel),
					If(isRecaptchaEnabled,
						// recaptcha response token
						Input("token").Id("token").Type("hidden"),
					),
					DefaultViewCommon.FormSubmitBtn(config.SignInButtonLabel).
						ClassIf("g-recaptcha", isRecaptchaEnabled).
						AttrIf("data-sitekey", vh.RecaptchaSiteKey(), isRecaptchaEnabled).
						AttrIf("data-callback", "onSubmit", isRecaptchaEnabled),
				)

				userPassCompo = Div().Class("d-flex flex-column").Children(
					compo,
					If(!vh.NoForgetPasswordLink(),
						A(Text(config.ForgetPasswordLabel)).Href(vh.ForgetPasswordPageURL()).Class("align-self-end mt-2 mb-7 grey--text text-subtitle-2 text--darken-1"),
					),
					VDivider().Color(ColorGreyDarken3).Class("mt-2 mb-10"),
				)
			}

			var langCompo HTMLComponent
			if len(langs) > 0 {
				ctx.Injector.HeadHTML(`
			<style>
				.transparent-language-select.vx-select-wrap .v-input .v-field {
					background-color: transparent;
				}
				.transparent-language-select.vx-select-wrap .v-input .v-field .v-select__selection-text {
					font-size: 14px !important;
					font-weight: 400;
				}
				.transparent-language-select.vx-select-wrap .v-input .v-field .v-field__outline {
					display: none;
				}
			</style>
			`)
				langCompo = web.Scope().VSlot(" { locals : selectLocals } ").Init(fmt.Sprintf(`{currLangVal: '%s'}`, currLangVal)).Children(
					vx.VXSelect().Class("transparent-language-select").
						Items(langs).
						ItemTitle("Label").
						ItemValue("Value").
						Attr("v-model", `selectLocals.currLangVal`).
						Attr("@update:model-value", web.Plaid().MergeQuery(true).Query(qn, web.Var("selectLocals.currLangVal")).PushState(true).Go()),
				)
			}

			rightCompo := VCol().Cols(12).Md(6).Class("d-flex flex-column justify-center align-center").Children(
				Div().Class("d-flex flex-column pa-4").Style(fmt.Sprintf("max-width: %s; width: 100%%", config.RightMaxWidth)).Children(
					Div().Class("d-flex flex-row align-center ga-2 mb-5").Children(
						config.BrandLogo,
						VSpacer(),
						langCompo,
					),
					Div().Text(config.WelcomeLabel).Class("mb-4 text-h4"),
					Div().Text(config.TitleLabel).Class("mb-10 text-h2").Style("font-size: 42px;"),
					userPassCompo,
					oauthCompo,
				),
			)

			ctx.Injector.HeadHTML(`
			<style>
				.login-form input {
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}
				.login-form input:focus::placeholder {
					color: transparent;
				}
			</style>
			`)

			r.PageTitle = config.PageTitle
			r.Body = Components(
				VRow().NoGutters(true).
					Attr(":class", "$vuetify.display.mdAndDown ? 'fill-height justify-center bg-grey-lighten-5':'fill-height justify-center'").Children(
					leftCompo,
					rightCompo,
				),
				Div().Class("position-absolute").Style("top: 0px; left: 0px; right: 0px;").Children(
					DefaultViewCommon.Notice(vh, i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages), ctx.W, ctx.R),
				),
			)

			return
		})
	}
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

func ZxcvbnLogicComponent() HTMLComponent {
	return Div().Style("display:none").Attr("v-on-mounted", fmt.Sprintf(`({window})=>{
		var tag = window.document.createElement("script");
		tag.src = %q;
		tag.onload= function(){
			vars.meter_score = function(x){
				return x ? window.zxcvbn(x).score+1 : 0;
			};
		}
		window.document.getElementsByTagName("head")[0].appendChild(tag);
	}`, login.ZxcvbnJSURL))
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
		}
		user, err = vh.FindUserByID(id)
		if err != nil {
			if err == login.ErrUserNotFound {
				r.Body = Div(Text("user not found"))
				return r, nil
			}
			panic(err)
		}
		token := query.Get("token")
		if token == "" {
			r.Body = Div(Text("invalid token"))
			return r, nil
		}
		storedToken, _, expired := user.(login.UserPasser).GetResetPasswordToken()
		if expired {
			r.Body = Div(Text("token expired"))
			return r, nil
		}
		if token != storedToken {
			r.Body = Div(Text("invalid token"))
			return r, nil
		}

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
						ZxcvbnLogicComponent(),
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
						ZxcvbnLogicComponent(),
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
	msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
	return web.Scope(vx.VXDialog(
		content,
	).OkText(pmsgr.OK).
		Title(msgr.ChangePasswordTitle).
		HideClose(true).
		Persistent(true).
		NoClickAnimation(true).
		CancelText(pmsgr.Cancel).
		Width(400).
		Attr("@click:ok", web.Plaid().EventFunc("login_changePassword").Go()).
		Attr("v-model", fmt.Sprintf("dialogLocals.%s", showVar)).
		Attr("@click:outside", presets.ShowSnackbarScript(pmsgr.LeaveBeforeUnsubmit, ColorWarning)),
	).VSlot(" { locals : dialogLocals}").Init(fmt.Sprintf(`{%s: true}`, showVar))
}

func defaultChangePasswordDialogContent(vh *login.ViewHelper, _ *presets.Builder) func(ctx *web.EventContext) HTMLComponent {
	return func(ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, login.I18nLoginKey, login.Messages_en_US).(*login.Messages)
		return Div(
			// VCardTitle(Text(msgr.ChangePasswordTitle)).Class("pa-6"),
			VCardText(
				Form().Children( // just used to prevent 1password auto submit
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
					),
					Div(
						DefaultViewCommon.PasswordInput("confirm_password", msgr.ChangePasswordNewConfirmPlaceholder, "", true).
							Label(msgr.ChangePasswordNewConfirmLabel).
							Attr(web.VField("confirm_password", "")...),
					),
					If(vh.TOTPEnabled(),
						Div(
							DefaultViewCommon.Input("otp", msgr.TOTPValidateCodePlaceholder, "").
								Label(msgr.TOTPValidateCodeLabel).
								Attr(web.VField("otp", "")...),
						),
					),
				),
			).Class("pa-0"),
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
		totpSecret := u.GetTOTPSecret()
		if totpSecret == "" {
			r.Body = DefaultViewCommon.ErrorBody("need setup totp")
			return
		}
		key, err := otp.NewKeyFromURL(
			fmt.Sprintf("otpauth://totp/%s:%s?issuer=%s&secret=%s",
				url.PathEscape(vh.TOTPIssuer()),
				url.PathEscape(u.GetAccountName()),
				url.QueryEscape(vh.TOTPIssuer()),
				url.QueryEscape(totpSecret),
			),
		)
		if err != nil {
			r.Body = DefaultViewCommon.ErrorBody(err.Error())
			return
		}

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
