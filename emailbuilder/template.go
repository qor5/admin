package emailbuilder

import (
	"github.com/qor5/admin/v3/presets"
)

var (
	cardHeight        = 146
	cardContentHeight = 56
)

func DefaultMailTemplate(pb *presets.Builder) *presets.ModelBuilder {
	mb := pb.Model(&EmailTemplate{}).Label("Email Templates")
	mb.Detailing(EmailEditorField)
	return mb
}
