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

func (b *ModelBuilder) PreviewHref(ctx *web.EventContext, ps string) string {
	var (
		isTpl         = ctx.R.FormValue(paramsTpl) != ""
		isLocalizable = ctx.R.Form.Has(paramLocale)
		ur            = url.Values{}
	)
	if isTpl {
		if isLocalizable && b.builder.l10n != nil {
			ur.Add(paramsTpl, "1")
		}
	}
	ur.Add(presets.ParamID, ps)
	return b.builder.prefix + "/" + b.name + "/preview" + "?" + ur.Encode()
}

const defaultContainerEmptyIcon = `<svg width="129" height="129" viewBox="0 0 129 129" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M107.4 39.1819L63.7961 63.0939L21.5984 39.1819L65.2026 18.7864L107.4 39.1819Z" fill="#5D81F7"/>
<path d="M107.401 39.1814L63.7971 63.0934V113.731L107.401 87.7087V39.1814Z" fill="#2346BC"/>
<path d="M128.5 52.5442L84.8953 76.4562V127.093L128.5 101.071V52.5442Z" fill="url(#paint0_linear_494_20724)"/>
<path d="M21.5984 87.7087V39.1814L63.7961 63.0934V113.731L21.5984 87.7087Z" fill="#3E63DD"/>
<path d="M0.500244 102.48V53.9524L42.6979 77.8644V128.502L0.500244 102.48Z" fill="url(#paint1_linear_494_20724)"/>
<path d="M77.8623 56.0842L64.2806 63.5322L51.1371 56.0842L64.7187 49.7314L77.8623 56.0842Z" fill="white"/>
<path d="M77.8622 56.0842L64.2805 63.5322V79.3045L77.8622 71.1993V56.0842Z" fill="#D6DAE8"/>
<path d="M51.1371 71.1993V56.0842L64.2806 63.5322V79.3045L51.1371 71.1993Z" fill="#E7EAF1"/>
<path d="M107.4 20.8946L63.7962 44.8066L21.5986 20.8946L65.2028 0.499023L107.4 20.8946Z" fill="url(#paint2_linear_494_20724)"/>
<defs>
<linearGradient id="paint0_linear_494_20724" x1="128.5" y1="52.5442" x2="151.668" y2="100.699" gradientUnits="userSpaceOnUse">
<stop stop-color="white" stop-opacity="0"/>
<stop offset="0.432618" stop-color="#5D81F7" stop-opacity="0.2"/>
<stop offset="1" stop-color="white"/>
</linearGradient>
<linearGradient id="paint1_linear_494_20724" x1="-13.5656" y1="21.6009" x2="60.2303" y2="32.8145" gradientUnits="userSpaceOnUse">
<stop stop-color="white" stop-opacity="0"/>
<stop offset="0.432618" stop-color="#5D81F7" stop-opacity="0.2"/>
<stop offset="1" stop-color="white"/>
</linearGradient>
<linearGradient id="paint2_linear_494_20724" x1="30.7414" y1="-10.7537" x2="90.5214" y2="51.1362" gradientUnits="userSpaceOnUse">
<stop stop-color="white" stop-opacity="0"/>
<stop offset="0.432618" stop-color="#5D81F7" stop-opacity="0.2"/>
<stop offset="1" stop-color="white"/>
</linearGradient>
</defs>
</svg>`
