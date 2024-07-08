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

const previewEmptySvg = `<svg width="120" height="73" viewBox="0 0 120 73" fill="none" xmlns="http://www.w3.org/2000/svg">
<rect x="119.5" y="72.125" width="29" height="8" transform="rotate(180 119.5 72.125)" fill="url(#paint0_linear_1436_33079)"/>
<path d="M56.5 2.12525C56.5 1.02068 57.3954 0.125244 58.5 0.125244H93.125C94.2296 0.125244 95.125 1.02067 95.125 2.12524V70.1252C95.125 71.2298 94.2296 72.1252 93.125 72.1252H58.5C57.3954 72.1252 56.5 71.2298 56.5 70.1252V2.12525Z" fill="#3E63DD"/>
<path d="M51.5 2.125C51.5 1.02043 52.3954 0.125 53.5 0.125H88.5C89.6046 0.125 90.5 1.02043 90.5 2.125V70.125C90.5 71.2296 89.6046 72.125 88.5 72.125H53.5C52.3954 72.125 51.5 71.2296 51.5 70.125V2.125Z" fill="#F6F8FE"/>
<circle cx="71.75" cy="67.6252" r="1.5" fill="#3E63DD"/>
<circle cx="76.25" cy="4.25024" r="0.375" fill="#8DA4EF"/>
<path d="M55.4375 9.12524C55.4375 8.81458 55.6893 8.56274 56 8.56274H85.625C85.9357 8.56274 86.1875 8.81458 86.1875 9.12524V60.8752C86.1875 61.1859 85.9357 61.4377 85.625 61.4377H56C55.6893 61.4377 55.4375 61.1859 55.4375 60.8752V9.12524Z" fill="#F5F5F5" stroke="#8DA4EF" stroke-width="0.375"/>
<rect x="67.625" y="3.87524" width="7.5" height="0.75" rx="0.375" fill="#8DA4EF"/>
<rect x="59.75" y="32.0002" width="4.5" height="3" fill="#3E63DD"/>
<path opacity="0.2" d="M55.625 9.12524C55.625 8.91814 55.7929 8.75024 56 8.75024H61.25V61.2502H56C55.7929 61.2502 55.625 61.0824 55.625 60.8752V9.12524Z" fill="#8DA4EF"/>
<path opacity="0.2" d="M85.625 8.75024C85.8321 8.75024 86 8.91814 86 9.12524L86 13.2502L61.25 13.2502L61.25 9.12524C61.25 8.91814 61.4179 8.75024 61.625 8.75024L85.625 8.75024Z" fill="#8DA4EF"/>
<path d="M6.125 25.3752C6.125 24.2707 7.02043 23.3752 8.125 23.3752H62.25C63.3546 23.3752 64.25 24.2707 64.25 25.3752V70.8752C64.25 71.9798 63.3546 72.8752 62.25 72.8752H8.125C7.02043 72.8752 6.125 71.9798 6.125 70.8752V25.3752Z" fill="#3E63DD"/>
<path d="M6.125 25.3752C6.125 24.2707 7.02043 23.3752 8.125 23.3752H62.25C63.3546 23.3752 64.25 24.2707 64.25 25.3752V32.3752H6.125V25.3752Z" fill="#3451B2"/>
<path d="M0.5 25.3752C0.5 24.2707 1.39543 23.3752 2.5 23.3752H56.625C57.7296 23.3752 58.625 24.2707 58.625 25.3752V70.8752C58.625 71.9798 57.7296 72.8752 56.625 72.8752H2.5C1.39543 72.8752 0.5 71.9798 0.5 70.8752V25.3752Z" fill="#F6F8FE"/>
<path d="M0.5 25.3752C0.5 24.2707 1.39543 23.3752 2.5 23.3752H56.625C57.7296 23.3752 58.625 24.2707 58.625 25.3752V32.3752H0.5V25.3752Z" fill="#8DA4EF"/>
<circle cx="6.125" cy="27.8752" r="1.125" fill="#3E63DD"/>
<circle cx="10.625" cy="27.8752" r="1.125" fill="#3E63DD"/>
<circle cx="15.125" cy="27.8752" r="1.125" fill="#3E63DD"/>
<rect x="5" y="39.1252" width="48.75" height="3.75" fill="#E8EAED"/>
<rect x="5" y="52.6252" width="48.75" height="6" fill="#E8EAED"/>
<mask id="path-20-inside-1_1436_33079" fill="white">
<path d="M5 44.375H53.75V49.625H5V44.375Z"/>
</mask>
<path d="M5 44.375H53.75V49.625H5V44.375Z" fill="#E8EAED"/>
<path d="M53.75 49.25H5V50H53.75V49.25Z" fill="#3E63DD" mask="url(#path-20-inside-1_1436_33079)"/>
<rect x="26.75" y="46.6252" width="6" height="6" fill="white"/>
<path d="M29.6875 45.5002C27.445 45.5002 25.625 47.3202 25.625 49.5627C25.625 51.8052 27.445 53.6252 29.6875 53.6252C31.93 53.6252 33.75 51.8052 33.75 49.5627C33.75 47.3202 31.93 45.5002 29.6875 45.5002ZM31.7187 49.969H30.0938V51.594H29.2812V49.969H27.6562V49.1565H29.2812V47.5315H30.0938V49.1565H31.7187V49.969Z" fill="#3E63DD"/>
<defs>
<linearGradient id="paint0_linear_1436_33079" x1="119.5" y1="76.125" x2="160.929" y2="76.2329" gradientUnits="userSpaceOnUse">
<stop stop-color="white" stop-opacity="0"/>
<stop offset="1" stop-color="#3E63DD"/>
</linearGradient>
</defs>
</svg>`
