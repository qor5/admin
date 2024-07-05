package examples_vuetify

// @snippet_begin(VuetifyVariantSubForm)

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type myFormValue1 struct {
	Type  string
	Form1 struct {
		Gender string
	}
	Form2 struct {
		Feature1 bool
		Slider1  int
	}
}

func VuetifyVariantSubForm(ctx *web.EventContext) (pr web.PageResponse, err error) {
	var fv myFormValue1
	ctx.MustUnmarshalForm(&fv)
	if fv.Type == "" {
		fv.Type = "Type1"
	}
	var verr web.ValidationErrors

	pr.Body = VContainer(
		examples.PrettyFormAsJSON(ctx),
		web.Scope(
			VSelect().
				Items([]string{
					"Type1",
					"Type2",
				}).
				Attr("v-model", "form.Type").
				Attr("@update:menu", web.POST().
					EventFunc("switchForm").
					Go()),

			web.Portal(
				h.If(fv.Type == "Type1",
					form1(ctx, &fv),
				).Else(
					form2(ctx, &fv, &verr),
				),
			).Name("subform"),

			VBtn("Submit").OnClick("submit"),
		).VSlot("{ locals, form }").FormInit(h.JSONString(fv)),
	)
	return
}

func form1(ctx *web.EventContext, fv *myFormValue1) h.HTMLComponent {
	return VContainer(
		h.H1("Form1"),
		VRadioGroup(
			VRadio().Value("F").Label("Female"),
			VRadio().Value("M").Label("Male"),
		).
			Attr("v-model", "form.Form1.Gender").
			Label("Gender"),
	)
}

func form2(ctx *web.EventContext, fv *myFormValue1, verr *web.ValidationErrors) h.HTMLComponent {
	return VContainer(
		h.H1("Form2"),

		VSwitch().
			Color("red").
			Attr("v-model", "form.Form2.Feature1").
			Label("Feature1"),

		VSlider().
			Step(1).
			Attr("v-model", "form.Form2.Slider1").
			ErrorMessages(verr.GetFieldErrors("Slider1")...).
			Label("Slider1"),
	)
}

func submit1(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.Reload = true
	return
}

func switchForm(ctx *web.EventContext) (r web.EventResponse, err error) {
	var verr web.ValidationErrors

	var fv myFormValue1
	ctx.MustUnmarshalForm(&fv)
	form := form1(ctx, &fv)
	if fv.Type == "Type2" {
		form = form2(ctx, &fv, &verr)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "subform",
		Body: form,
	})

	return
}

var VuetifyVariantSubFormPB = web.Page(VuetifyVariantSubForm).
	EventFunc("switchForm", switchForm).
	EventFunc("submit", submit1)

var VuetifyVariantSubFormPath = examples.URLPathByFunc(VuetifyVariantSubForm)

// @snippet_end
