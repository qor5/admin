package redirection

import (
	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/presets"
	v "github.com/qor5/x/v3/ui/vuetify"
)

const (
	UploadFileEvent = "redirection_UploadFileEvent"
)

func (b *Builder) uploadFile(ctx *web.EventContext) (r web.EventResponse, err error) {

	if err = b.importRecords(ctx, &r); err != nil {
		return
	}
	presets.ShowMessage(&r, "success", v.ColorSuccess)
	return
}
