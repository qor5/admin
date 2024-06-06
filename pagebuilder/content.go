package pagebuilder

import (
	"net/url"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
)

const (
	pageBuilderRightContentPortal   = "pageBuilderRightContentPortal"
	pageBuilderLayerContainerPortal = "pageBuilderLayerContainerPortal"
)

func (b *Builder) previewHref(ctx *web.EventContext, pm *presets.ModelBuilder, ps string) string {
	var (
		isTpl         = ctx.R.FormValue(paramsTpl) != ""
		isLocalizable = ctx.R.Form.Has(paramLocale)
		ur            = url.Values{}
	)
	if isTpl {
		if isLocalizable && b.l10n != nil {
			ur.Add(paramsTpl, "1")
		}
	}
	ur.Add(presets.ParamID, ps)
	return b.prefix + "/" + pm.Info().URIName() + "/preview" + "?" + ur.Encode()
}
