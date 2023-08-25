package containers

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/web"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ContactForm struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Heading            string
	Text               string
	SendButtonText     string
	FormButtonText     string
	MessagePlaceholder string
	NamePlaceholder    string
	EmailPlaceholder   string
	ThankyouMessage    string
	ActionUrl          string
	PrivacyPolicy      string
}

func (*ContactForm) TableName() string {
	return "container_contact_form"
}

func RegisterContactFormContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("ContactForm").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ContactForm)
			return ContactFormBody(v, input)
		})
	vb.Model(&ContactForm{}).Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Heading", "Text", "SendButtonText", "FormButtonText", "MessagePlaceholder", "NamePlaceholder", "EmailPlaceholder", "ThankyouMessage", "ActionUrl", "PrivacyPolicy")

}

const (
	CONTACT_FORM_ICON = RawHTML(`<lottie-player class="container-contact_form-icon" data-event-type="mouseover" data-event-target=".container-contact_form-inner" src='{"v":"5.5.7","meta":{"g":"LottieFiles AE 0.1.21","a":"","k":"","d":"","tc":""},"fr":24,"ip":0,"op":29,"w":256,"h":356,"nm":"contact","ddd":0,"assets":[],"layers":[{"ddd":0,"ind":1,"ty":4,"nm":"Shape Layer 3","sr":1,"ks":{"o":{"a":0,"k":100,"ix":11},"r":{"a":0,"k":0,"ix":10},"p":{"a":0,"k":[128,178,0],"ix":2},"a":{"a":0,"k":[0,0,0],"ix":1},"s":{"a":0,"k":[100,100,100],"ix":6}},"ao":0,"shapes":[{"ty":"gr","it":[{"ty":"st","c":{"a":0,"k":[0.295881981943,0.644987996419,0.918154009651,1],"ix":3},"o":{"a":0,"k":100,"ix":4},"w":{"a":0,"k":0,"ix":5},"lc":1,"lj":1,"ml":4,"bm":0,"nm":"Stroke 1","mn":"ADBE Vector Graphic - Stroke","hd":false},{"ty":"fl","c":{"a":0,"k":[1,0.396078431373,0,1],"ix":4},"o":{"a":0,"k":100,"ix":5},"r":1,"bm":0,"nm":"Fill 1","mn":"ADBE Vector Graphic - Fill","hd":false},{"ty":"tr","p":{"a":0,"k":[0,0],"ix":2},"a":{"a":0,"k":[0,0],"ix":1},"s":{"a":0,"k":[100,100],"ix":3},"r":{"a":0,"k":0,"ix":6},"o":{"a":0,"k":100,"ix":7},"sk":{"a":0,"k":0,"ix":4},"sa":{"a":0,"k":0,"ix":5},"nm":"Transform"}],"nm":"Shape 2","np":2,"cix":2,"bm":0,"ix":1,"mn":"ADBE Vector Group","hd":false},{"ty":"gr","it":[{"ind":0,"ty":"sh","ix":1,"ks":{"a":1,"k":[{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":0,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-43],[-58,-28],[-58.5,43.5],[57,29]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":1,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-46.5],[-57,-32],[-57.5,43.5],[58,29]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":2,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-54],[-57,-40.5],[-57.5,43.5],[58,29]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":3,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-64],[-57.5,-49.5],[-57.5,43.5],[58,29]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":4,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-72],[-57,-58],[-57.5,41.75],[58,27.25]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":5,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-75.5],[-57,-61],[-57.5,37.75],[58,23.25]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":6,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-68.5],[-57,-54.5],[-57.25,33.5],[58.25,19]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":7,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-53],[-56.5,-38],[-57,29],[58.5,14.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":8,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-37.5],[-57,-23],[-57,27.25],[58.5,12.75]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":9,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-30],[-57,-15.5],[-57,30.75],[58.5,16.25]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":10,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-33],[-57,-18.5],[-57,38.25],[58.5,23.75]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":11,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-40.5],[-56.5,-26],[-57.5,47],[58,32.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":12,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-49],[-56.5,-34.5],[-57.5,54.5],[58,40]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":13,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-57.5],[-56.5,-43],[-57.5,58],[58,43.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":14,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-60.5],[-56.5,-46],[-58,55],[57.5,40.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":15,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-58.5],[-56.5,-44],[-57.5,49],[58,34.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":16,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58,-54],[-57,-39.5],[-57.5,42.5],[58,28]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":17,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-49.5],[-56.5,-35],[-57.5,36.5],[58,22]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":18,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-43],[-56.5,-28.5],[-57.5,33.5],[58,19]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":19,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-38],[-56.5,-23.5],[-57.5,36],[58,21.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":20,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-37],[-56.5,-22.5],[-57,42],[58.5,27.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":21,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-35.5],[-56.5,-21],[-57,48],[58.5,33.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":22,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-39],[-56.5,-24.5],[-57,50.5],[58.5,36]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":23,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-42],[-56.5,-27.5],[-57.5,51.5],[58,37]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":24,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-43],[-56.5,-28.5],[-58,50],[57.5,35.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":25,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-43],[-56.5,-28.5],[-58,47],[57.5,32.5]],"c":true}]},{"i":{"x":0.833,"y":0.833},"o":{"x":0.167,"y":0.167},"t":26,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-43],[-56.5,-28.5],[-58,44],[57.5,29.5]],"c":true}]},{"t":27,"s":[{"i":[[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0]],"v":[[58.5,-43],[-56.5,-28.5],[-58,43.5],[57.5,29]],"c":true}]}],"ix":2},"nm":"Path 1","mn":"ADBE Vector Shape - Group","hd":false},{"ty":"st","c":{"a":0,"k":[0.295881981943,0.644987996419,0.918154009651,1],"ix":3},"o":{"a":0,"k":100,"ix":4},"w":{"a":0,"k":0,"ix":5},"lc":1,"lj":1,"ml":4,"bm":0,"nm":"Stroke 1","mn":"ADBE Vector Graphic - Stroke","hd":false},{"ty":"fl","c":{"a":0,"k":[1,0.396078431373,0,1],"ix":4},"o":{"a":0,"k":100,"ix":5},"r":1,"bm":0,"nm":"Fill 1","mn":"ADBE Vector Graphic - Fill","hd":false},{"ty":"tr","p":{"a":0,"k":[0,0],"ix":2},"a":{"a":0,"k":[0,0],"ix":1},"s":{"a":0,"k":[100,100],"ix":3},"r":{"a":0,"k":0,"ix":6},"o":{"a":0,"k":100,"ix":7},"sk":{"a":0,"k":0,"ix":4},"sa":{"a":0,"k":0,"ix":5},"nm":"Transform"}],"nm":"Shape 1","np":3,"cix":2,"bm":0,"ix":2,"mn":"ADBE Vector Group","hd":false}],"ip":0,"op":72,"st":0,"bm":0},{"ddd":0,"ind":2,"ty":4,"nm":"Shape Layer 2","sr":1,"ks":{"o":{"a":0,"k":99,"ix":11},"r":{"a":0,"k":0,"ix":10},"p":{"a":1,"k":[{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":0,"s":[127.5,178.5,0],"to":[0,-5.387,0],"ti":[0,-2.254,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":5,"s":[127.5,146.177,0],"to":[0,2.254,0],"ti":[0,-2.564,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":9,"s":[127.5,192.022,0],"to":[0,2.564,0],"ti":[0,4,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":14,"s":[127.5,161.564,0],"to":[0,-2.283,0],"ti":[0,-6.508,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":20,"s":[127.5,185.507,0],"to":[0,4.893,0],"ti":[0,-1.574,0]},{"t":24,"s":[127.5,178.5,0]}],"ix":2},"a":{"a":0,"k":[0,0,0],"ix":1},"s":{"a":0,"k":[100,100,100],"ix":6}},"ao":0,"shapes":[{"ty":"gr","it":[{"ind":0,"ty":"sh","ix":1,"ks":{"a":0,"k":{"i":[[0,0],[0,0],[0,0],[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0],[0,0],[0,0],[0,0]],"v":[[128,-52.5],[-57.5,-29.5],[-59,111],[59.5,98],[103.5,159],[104,92],[127.5,89.5]],"c":true},"ix":2},"nm":"Path 1","mn":"ADBE Vector Shape - Group","hd":false},{"ty":"st","c":{"a":0,"k":[0.295881981943,0.644987996419,0.918154009651,1],"ix":3},"o":{"a":0,"k":100,"ix":4},"w":{"a":0,"k":0,"ix":5},"lc":1,"lj":1,"ml":4,"bm":0,"nm":"Stroke 1","mn":"ADBE Vector Graphic - Stroke","hd":false},{"ty":"fl","c":{"a":0,"k":[1,0.549019607843,0.250980392157,1],"ix":4},"o":{"a":0,"k":100,"ix":5},"r":1,"bm":0,"nm":"Fill 1","mn":"ADBE Vector Graphic - Fill","hd":false},{"ty":"tr","p":{"a":0,"k":[0,0],"ix":2},"a":{"a":0,"k":[0,0],"ix":1},"s":{"a":0,"k":[100,100],"ix":3},"r":{"a":0,"k":0,"ix":6},"o":{"a":0,"k":100,"ix":7},"sk":{"a":0,"k":0,"ix":4},"sa":{"a":0,"k":0,"ix":5},"nm":"Transform"}],"nm":"Shape 1","np":3,"cix":2,"bm":0,"ix":1,"mn":"ADBE Vector Group","hd":false}],"ip":-4,"op":72,"st":0,"bm":0},{"ddd":0,"ind":3,"ty":4,"nm":"Shape Layer 1","sr":1,"ks":{"o":{"a":0,"k":100,"ix":11},"r":{"a":0,"k":0,"ix":10},"p":{"a":1,"k":[{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":3,"s":[128,177,0],"to":[0,-2.742,0],"ti":[0,-2.254,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":8,"s":[128,160.551,0],"to":[0,2.254,0],"ti":[0,-1.059,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":13,"s":[128,190.522,0],"to":[0,1.059,0],"ti":[0,4,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":18,"s":[128,166.903,0],"to":[0,-2.283,0],"ti":[0,-6.508,0]},{"i":{"x":0.667,"y":1},"o":{"x":0.333,"y":0},"t":22,"s":[128,184.213,0],"to":[0,4.893,0],"ti":[0,-1.574,0]},{"t":27,"s":[128,177,0]}],"ix":2},"a":{"a":0,"k":[0,0,0],"ix":1},"s":{"a":0,"k":[100,100,100],"ix":6}},"ao":0,"shapes":[{"ty":"gr","it":[{"ind":0,"ty":"sh","ix":1,"ks":{"a":0,"k":{"i":[[0,0],[0,0],[0,0],[0,0],[0,0],[0,0],[0,0]],"o":[[0,0],[0,0],[0,0],[0,0],[0,0],[0,0],[0,0]],"v":[[-104,-158],[-104.5,-90.5],[-127,-87.5],[-128,53.5],[58,30.5],[57,-111],[-59.5,-97]],"c":true},"ix":2},"nm":"Path 1","mn":"ADBE Vector Shape - Group","hd":false},{"ty":"st","c":{"a":0,"k":[0.295881981943,0.644987996419,0.918154009651,1],"ix":3},"o":{"a":0,"k":100,"ix":4},"w":{"a":0,"k":0,"ix":5},"lc":1,"lj":1,"ml":4,"bm":0,"nm":"Stroke 1","mn":"ADBE Vector Graphic - Stroke","hd":false},{"ty":"fl","c":{"a":0,"k":[0.321568627451,0.81568627451,0.945098039216,1],"ix":4},"o":{"a":0,"k":100,"ix":5},"r":1,"bm":0,"nm":"Fill 1","mn":"ADBE Vector Graphic - Fill","hd":false},{"ty":"tr","p":{"a":0,"k":[0,0],"ix":2},"a":{"a":0,"k":[0,0],"ix":1},"s":{"a":0,"k":[100,100],"ix":3},"r":{"a":0,"k":0,"ix":6},"o":{"a":0,"k":100,"ix":7},"sk":{"a":0,"k":0,"ix":4},"sa":{"a":0,"k":0,"ix":5},"nm":"Transform"}],"nm":"Shape 1","np":3,"cix":2,"bm":0,"ix":1,"mn":"ADBE Vector Group","hd":false}],"ip":0,"op":75,"st":3,"bm":0}],"markers":[]}'></lottie-player>`)
)

func ContactFormBody(data *ContactForm, input *pagebuilder.RenderInput) (body HTMLComponent) {
	var n = Div(
		CONTACT_FORM_ICON,
		Div(
			H2(data.Heading),
		).Class("container-contact_form-title"),
		Div(
			P(
				Text(data.Text),
			).Class("p-large"),
			A(
				Span(data.FormButtonText),
			).Class("container-contact_form-link button").Href("#"),
		).Class("container-contact_form-brief"),
		Form(
			Div(
				Div(Textarea("").Class("textarea").Name("message").Placeholder(data.MessagePlaceholder).Required(true)).Class("container-contact_form-message"),
				Div(
					Input("").Class("input").Type("text").Name("name").Placeholder(data.NamePlaceholder).Required(true),
					Input("").Class("input").Type("email").Name("email").Placeholder(data.EmailPlaceholder).Required(true),
				).Class("container-contact_form-contact"),
			).Class("container-contact_form-filled"),
			Div(
				Div(
					Label("").
						Children(
							Input("").Class("container-contact_form-policy-checkbox").Type("checkbox").Name("privacy_policy"),
							Div(RawHTML(data.PrivacyPolicy)).Class("container-contact_form-policy-text"),
						).Class("container-contact_form-policy"),
					Div(
						Div(Text(data.ThankyouMessage)).Class("container-contact_form-response-text"),
					).Class("container-contact_form-response"),
				).Class("container-contact_form-append"),
				Div(
					Button("").Class("button").Children(Span(data.SendButtonText)),
				).Class("container-contact_form-button"),
			).Class("container-contact_form-submit"),
		).Class("container-contact_form-form").Action(data.ActionUrl).Method("POST"),
	).Class("container-contact_form-inner")

	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("ContactForm"))+"_%v", data.ID), data.AnchorID, "container-contact_form container-lottie",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, input.IsEditor, input.IsReadonly, "",
		Div(
			n,
		).Class("container-wrapper"),
	)
	return
}
