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

const previewIframeEmptySvg = `<svg width="169" height="103" viewBox="0 0 169 103" fill="none" xmlns="http://www.w3.org/2000/svg">
<g filter="url(#filter0_b_1965_73205)">
<rect x="17.75" y="9.80273" width="103.365" height="61.1753" rx="2.10949" fill="url(#paint0_linear_1965_73205)"/>
<rect x="18.2774" y="10.3301" width="102.31" height="60.1206" rx="1.58212" stroke="url(#paint1_linear_1965_73205)" stroke-width="1.05475"/>
<path d="M26.188 19.2955C26.188 18.7129 26.6602 18.2407 27.2427 18.2407H55.1332C55.7158 18.2407 56.188 18.7129 56.188 19.2955V26.936C56.188 27.5185 55.7158 27.9907 55.1332 27.9907H27.2427C26.6602 27.9907 26.188 27.5185 26.188 26.936V19.2955Z" fill="#E4E5E9"/>
<path d="M26.5005 34.0547C26.5005 33.4722 26.9727 33 27.5552 33H82.4457C83.0283 33 83.5005 33.4722 83.5005 34.0547V37.2153C83.5005 37.7978 83.0283 38.27 82.4457 38.27H27.5552C26.9727 38.27 26.5005 37.7978 26.5005 37.2153V34.0547Z" fill="#E4E5E9"/>
<path d="M26.5005 45.0547C26.5005 44.4722 26.9727 44 27.5552 44H111.935C112.518 44 112.99 44.4722 112.99 45.0547V48.219C112.99 48.8015 112.518 49.2737 111.935 49.2737H27.5552C26.9727 49.2737 26.5005 48.8015 26.5005 48.219V45.0547Z" fill="#E4E5E9"/>
<path d="M26.5 57.0547C26.5 56.4722 26.9722 56 27.5547 56H43.4453C44.0278 56 44.5 56.4722 44.5 57.0547V59.9453C44.5 60.5278 44.0278 61 43.4453 61H27.5547C26.9722 61 26.5 60.5278 26.5 59.9453V57.0547Z" fill="#E4E5E9"/>
</g>
<g filter="url(#filter1_b_1965_73205)">
<rect x="55.7207" y="40.3896" width="103.365" height="49.5731" rx="2.10949" fill="url(#paint2_linear_1965_73205)"/>
<rect x="56.2481" y="40.917" width="102.31" height="48.5184" rx="1.58212" stroke="url(#paint3_linear_1965_73205)" stroke-width="1.05475"/>
<rect x="64.1587" y="48.8276" width="43.2446" height="32.6972" rx="2.10949" fill="url(#paint4_linear_1965_73205)"/>
<path d="M113.732 65.7037C113.732 65.1212 114.204 64.6489 114.786 64.6489H149.593C150.176 64.6489 150.648 65.1212 150.648 65.7037V68.8679C150.648 69.4504 150.176 69.9227 149.593 69.9227H114.786C114.204 69.9227 113.732 69.4504 113.732 68.8679V65.7037Z" fill="#3E63DD"/>
<path d="M113.732 77.306C113.732 76.7234 114.204 76.2512 114.786 76.2512H132.717C133.3 76.2512 133.772 76.7234 133.772 77.306V80.4702C133.772 81.0527 133.3 81.525 132.717 81.525H114.786C114.204 81.525 113.732 81.0527 113.732 80.4702V77.306Z" fill="#3E63DD"/>
</g>
<path fill-rule="evenodd" clip-rule="evenodd" d="M148.832 82.5271C148.245 82.3306 147.686 82.8895 147.883 83.4765L151.702 94.8823C151.879 95.4105 152.55 95.5684 152.944 95.1744L155.567 92.5509L162.68 99.6639L164.979 97.3655L157.866 90.2525L160.53 87.588C160.924 87.194 160.766 86.5233 160.238 86.3464L148.832 82.5271Z" fill="#8DA4EF"/>
<g filter="url(#filter2_b_1965_73205)">
<rect x="107.75" width="35.25" height="18" rx="9" fill="#8DA4EF" fill-opacity="0.15"/>
<rect x="108" y="0.25" width="34.75" height="17.5" rx="8.75" stroke="#8DA4EF" stroke-opacity="0.4" stroke-width="0.5"/>
</g>
<g filter="url(#filter3_b_1965_73205)">
<rect x="0.5" y="59.25" width="48" height="19.5" rx="9.75" fill="#8DA4EF" fill-opacity="0.15"/>
<rect x="0.75" y="59.5" width="47.5" height="19" rx="9.5" stroke="#8DA4EF" stroke-opacity="0.4" stroke-width="0.5"/>
</g>
<defs>
<filter id="filter0_b_1965_73205" x="14.75" y="6.80273" width="109.365" height="67.1753" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
<feFlood flood-opacity="0" result="BackgroundImageFix"/>
<feGaussianBlur in="BackgroundImageFix" stdDeviation="1.5"/>
<feComposite in2="SourceAlpha" operator="in" result="effect1_backgroundBlur_1965_73205"/>
<feBlend mode="normal" in="SourceGraphic" in2="effect1_backgroundBlur_1965_73205" result="shape"/>
</filter>
<filter id="filter1_b_1965_73205" x="52.7207" y="37.3896" width="109.365" height="55.573" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
<feFlood flood-opacity="0" result="BackgroundImageFix"/>
<feGaussianBlur in="BackgroundImageFix" stdDeviation="1.5"/>
<feComposite in2="SourceAlpha" operator="in" result="effect1_backgroundBlur_1965_73205"/>
<feBlend mode="normal" in="SourceGraphic" in2="effect1_backgroundBlur_1965_73205" result="shape"/>
</filter>
<filter id="filter2_b_1965_73205" x="103.25" y="-4.5" width="44.25" height="27" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
<feFlood flood-opacity="0" result="BackgroundImageFix"/>
<feGaussianBlur in="BackgroundImageFix" stdDeviation="2.25"/>
<feComposite in2="SourceAlpha" operator="in" result="effect1_backgroundBlur_1965_73205"/>
<feBlend mode="normal" in="SourceGraphic" in2="effect1_backgroundBlur_1965_73205" result="shape"/>
</filter>
<filter id="filter3_b_1965_73205" x="-4" y="54.75" width="57" height="28.5" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
<feFlood flood-opacity="0" result="BackgroundImageFix"/>
<feGaussianBlur in="BackgroundImageFix" stdDeviation="2.25"/>
<feComposite in2="SourceAlpha" operator="in" result="effect1_backgroundBlur_1965_73205"/>
<feBlend mode="normal" in="SourceGraphic" in2="effect1_backgroundBlur_1965_73205" result="shape"/>
</filter>
<linearGradient id="paint0_linear_1965_73205" x1="69.4326" y1="9.80273" x2="69.4326" y2="70.9781" gradientUnits="userSpaceOnUse">
<stop stop-color="#F6F8FE"/>
<stop offset="1" stop-color="white" stop-opacity="0"/>
</linearGradient>
<linearGradient id="paint1_linear_1965_73205" x1="69.4326" y1="9.80273" x2="69.4326" y2="117.836" gradientUnits="userSpaceOnUse">
<stop stop-color="#3E63DD"/>
<stop offset="1" stop-color="white" stop-opacity="0"/>
</linearGradient>
<linearGradient id="paint2_linear_1965_73205" x1="107.403" y1="40.3896" x2="107.403" y2="89.9628" gradientUnits="userSpaceOnUse">
<stop stop-color="#F6F8FE"/>
<stop offset="1" stop-color="white" stop-opacity="0"/>
</linearGradient>
<linearGradient id="paint3_linear_1965_73205" x1="107.403" y1="40.3896" x2="107.403" y2="127.934" gradientUnits="userSpaceOnUse">
<stop stop-color="#3E63DD"/>
<stop offset="1" stop-color="white" stop-opacity="0"/>
</linearGradient>
<linearGradient id="paint4_linear_1965_73205" x1="135.5" y1="75" x2="31.4277" y2="70.0211" gradientUnits="userSpaceOnUse">
<stop stop-color="white" stop-opacity="0"/>
<stop offset="1" stop-color="#3E63DD"/>
</linearGradient>
</defs>
</svg>
`
