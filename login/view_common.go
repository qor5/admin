package login

import (
	"fmt"
	"net/http"

	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/login"
	. "github.com/theplant/htmlgo"
)

var DefaultViewCommon = &ViewCommon{
	WrapperClass: "d-flex pt-16 flex-column mx-auto",
	WrapperStyle: "max-width: 28rem;",
	TitleClass:   "text-h5 mb-6 font-weight-bold",
	LabelClass:   "d-block mb-1 grey--text text--darken-2 text-sm-body-2",
}

type ViewCommon struct {
	WrapperClass string
	WrapperStyle string
	TitleClass   string
	LabelClass   string
}

func (vc *ViewCommon) Notice(vh *login.ViewHelper, msgr *login.Messages, w http.ResponseWriter, r *http.Request) HTMLComponent {
	var nn HTMLComponent
	if n := vh.GetNoticeFlash(w, r); n != nil && n.Message != "" {
		switch n.Level {
		case login.NoticeLevel_Info:
			nn = vc.InfoNotice(n.Message)
		case login.NoticeLevel_Warn:
			nn = vc.WarnNotice(n.Message)
		case login.NoticeLevel_Error:
			nn = vc.ErrNotice(n.Message)
		}
	}
	return Components(
		vc.ErrNotice(vh.GetFailFlashMessage(msgr, w, r)),
		vc.WarnNotice(vh.GetWarnFlashMessage(msgr, w, r)),
		vc.InfoNotice(vh.GetInfoFlashMessage(msgr, w, r)),
		nn,
	)
}

func (vc *ViewCommon) ErrNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return VAlert(Text(msg)).
		Density(DensityCompact).
		Class("text-center").
		Icon(false).
		Type("error")
}

func (vc *ViewCommon) WarnNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return VAlert(Text(msg)).
		Density(DensityCompact).
		Class("text-center").
		Icon(false).
		Type("warning")
}

func (vc *ViewCommon) InfoNotice(msg string) HTMLComponent {
	if msg == "" {
		return nil
	}

	return VAlert(Text(msg)).
		Density(DensityCompact).
		Class("text-center").
		Icon(false).
		Type("info")
}

func (vc *ViewCommon) ErrorBody(msg string) HTMLComponent {
	return Div(
		Text(msg),
	)
}

func (vc *ViewCommon) Input(
	id string,
	placeholder string,
	val string,
) *VTextFieldBuilder {
	return VTextField().
		Attr("name", id).
		Id(id).
		Placeholder(placeholder).
		ModelValue(val).
		Variant(VariantOutlined).
		HideDetails(true).
		Density(DensityCompact)
}

func (vc *ViewCommon) PasswordInput(
	id string,
	placeholder string,
	val string,
	canReveal bool,
) *VTextFieldBuilder {
	in := vc.Input(id, placeholder, val)
	if canReveal {
		varName := fmt.Sprintf(`show_%s`, id)
		in.Attr(":append-inner-icon", fmt.Sprintf(`vars.%s ? "mdi-eye-off" : "mdi-eye"`, varName)).
			Attr(":type", fmt.Sprintf(`vars.%s ? "text" : "password"`, varName)).
			Attr("@click:append-inner", fmt.Sprintf(`vars.%s = !vars.%s`, varName, varName)).
			Attr(web.ObjectAssign("vars", fmt.Sprintf(`{%s: false}`, varName))...)
	}

	return in
}

// need to import zxcvbn js
// func (vc *ViewCommon) PasswordInputWithStrengthMeter(in *v.VTextFieldBuilder, id string, val string) HTMLComponent {
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
func (vc *ViewCommon) PasswordInputWithStrengthMeter(in *VTextFieldBuilder, id string, val string) HTMLComponent {
	passVar := fmt.Sprintf(`password_%s`, id)
	meterScoreVar := fmt.Sprintf(`meter_score_%s`, id)
	in.Attr("v-model", fmt.Sprintf(`progressLocals.%s`, passVar)).
		On("input", fmt.Sprintf(`progressLocals.%s = progressLocals.%s ? zxcvbn(progressLocals.%s).score + 1 : 0`, meterScoreVar, passVar, passVar))
	return web.Scope(
		in,
		VProgressLinear().
			Class("mt-2").
			Attr(":value", fmt.Sprintf(`progressLocals.%s * 20`, meterScoreVar)).
			Attr(":color", fmt.Sprintf(`["grey", "red", "deep-orange", "amber", "yellow", "light-green"][progressLocals.%s]`, meterScoreVar)).
			Attr("v-show", fmt.Sprintf(`!!progressLocals.%s`, passVar)),
	).VSlot(" { locals : progressLocals } ").Init(fmt.Sprintf(`{%s: "%s", %s: "%s" ? zxcvbn("%s").score + 1 : 0}`, passVar, val, meterScoreVar, val, val))
}

func (vc *ViewCommon) FormSubmitBtn(
	label string,
) *VBtnBuilder {
	return VBtn(label).
		Color("primary").
		Block(true).
		Size(SizeLarge).
		Attr("type", "submit").
		Class("mt-6")
}

// requirements:
// - submit button
//   - add class `g-recaptcha`
//   - add attr `data-sitekey=<key>`
//   - add attr `data-callback=onSubmit`
//
// - add token field like `Input("token").Id("token").Type("hidden")`
func (vc *ViewCommon) InjectRecaptchaAssets(ctx *web.EventContext, formID string, tokenFieldID string) {
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

func (vc *ViewCommon) InjectZxcvbn(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(fmt.Sprintf(`
<script src="%s"></script>
    `, login.ZxcvbnJSURL))
}
