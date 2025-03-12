package emailbuilder

import (
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets"
)

var (
	cardHeight        = 146
	cardContentHeight = 56
)

func DefaultMailTemplate(pb *presets.Builder) *presets.ModelBuilder {
	mb := pb.Model(&EmailTemplate{}).Label("Email Templates")
	dp := mb.Detailing(EmailEditorField)
	mb.Editing("Subject", "JSONBody", "HTMLBody")

	dp.Title(func(evCtx *web.EventContext, obj any, style presets.DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		titleCompo = h.Div(
			v.VToolbarTitle("Inbox"),
			v.VSpacer(),
			vx.VXBtn("Save").Variant("elevated").Attr("@click", web.Emit("save_mail")).Color("primary").Attr(":disabled", "vars.$EmailEditorLoading"),
			vx.VXBtn("Send Email").
				Variant("elevated").
				Color("secondary").
				Attr("@click", web.Emit("open_send_mail_dialog")).
				Color("secondary").
				Class("ml-2"),
		).Class("d-flex align-center w-100")

		return
	})

	return mb
}
