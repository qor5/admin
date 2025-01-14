package redirection

import (
	"github.com/qor5/web/v3"
)

const (
	UploadFileEvent = "redirection_UploadFileEvent"
)

func (r *Builder) uploadFile(ctx *web.EventContext) (res web.EventResponse, err error) {

	if err = r.importRecords(ctx); err != nil {
		return
	}
	return
}
