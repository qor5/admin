package examples_vuetify

// @snippet_begin(VuetifyBasicInputsSample)
import (
	"mime/multipart"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type myFormValue struct {
	MyValue          string
	TextareaValue    string
	Gender           string
	Agreed           bool
	Feature1         bool
	Slider1          int
	PortalAddedValue string
	Files1           []*multipart.FileHeader
	Files2           []*multipart.FileHeader
	Files3           []*multipart.FileHeader
}

var s = &myFormValue{
	MyValue:       "123",
	TextareaValue: "This is textarea value",
	Gender:        "M",
	Agreed:        false,
	Feature1:      true,
	Slider1:       60,
}

func VuetifyBasicInputs(ctx *web.EventContext) (pr web.PageResponse, err error) {
	var verr web.ValidationErrors
	if ve, ok := ctx.Flash.(web.ValidationErrors); ok {
		verr = ve
	}

	pr.Body = VContainer(
		examples.PrettyFormAsJSON(ctx),
		web.Scope(
			VTextField().
				Label("Form ValueIs").
				Variant("solo").
				Clearable(true).
				Attr("v-model", "form.MyValue").
				ErrorMessages(verr.GetFieldErrors("MyValue")...),
			VTextarea().
				Attr("v-model", "form.TextareaValue").
				ErrorMessages(verr.GetFieldErrors("TextareaValue")...).
				Variant("solo"),
			VRadioGroup(
				VRadio().Value("F").Label("Female"),
				VRadio().Value("M").Label("Male"),
			).
				Attr("v-model", "form.Gender"),
			VCheckbox().
				Attr("v-model", "form.Agreed").
				ErrorMessages(verr.GetFieldErrors("Agreed")...).
				Label("Agree"),
			VSwitch().
				Color("primary").
				Attr("v-model", "form.Feature1"),

			VSlider().
				Step(1).
				Attr("v-model", "form.Slider1").
				ErrorMessages(verr.GetFieldErrors("Slider1")...),

			web.Portal().Name("Portal1"),

			VFileInput().
				Attr("v-model", "form.Files1"),

			VFileInput().Label("Auto post to server after select file").Multiple(true).
				Attr("@change", web.POST().
					EventFunc("update").
					FieldValue("Files2", web.Var("$event")).
					Go()),

			h.Div(
				h.Input("Files3").Type("file").
					Attr("@input", web.POST().
						EventFunc("update").
						FieldValue("Files3", web.Var("$event")).
						Go()),
			).Class("mb-4"),

			VBtn("Update").OnClick("update").Color("primary"),
			h.P().Text("The following button will update a portal with a hidden field, if you click this button, and then click the above update button, you will find additional value posted to server"),
			VBtn("Add Portal Hidden Value").OnClick("addPortal"),
		).VSlot("{ locals, form }").FormInit(h.JSONString(s)),
	)

	return
}

func addPortal(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "Portal1",
		Body: h.Input("").Type("hidden").
			Attr(":value", "form.PortalAddedValue = 'this is my portal added hidden value'"),
	})
	return
}

func update(ctx *web.EventContext) (r web.EventResponse, err error) {
	s = &myFormValue{}
	ctx.MustUnmarshalForm(s)
	verr := web.ValidationErrors{}
	if len(s.MyValue) < 10 {
		verr.FieldError("MyValue", "my value is too small")
	}

	if len(s.TextareaValue) > 5 {
		verr.FieldError("TextareaValue", "textarea value is too large")
	}

	if !s.Agreed {
		verr.FieldError("Agreed", "You must agree the terms")
	}

	if s.Slider1 > 50 {
		verr.FieldError("Slider1", "You slide too much")
	}

	ctx.Flash = verr
	r.Reload = true

	return
}

var VuetifyBasicInputsPB = web.Page(VuetifyBasicInputs).
	EventFunc("update", update).
	EventFunc("addPortal", addPortal)

// @snippet_end

var VuetifyBasicInputsPath = examples.URLPathByFunc(VuetifyBasicInputs)
