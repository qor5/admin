package pagebuilder

import (
	"net/url"
	"path"

	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/presets"
)

const (
	pageBuilderRightContentPortal   = "pageBuilderRightContentPortal"
	pageBuilderLayerContainerPortal = "pageBuilderLayerContainerPortal"
	pageBuilderAddContainersPortal  = "pageBuilderAddContainersPortal"
)

func (b *ModelBuilder) PreviewHref(_ *web.EventContext, ps string) string {
	ur := url.Values{}
	ur.Add(presets.ParamID, ps)
	return path.Join(b.builder.pb.GetURIPrefix(), b.builder.prefix, b.mb.Info().URIName(),
		"preview") + "?" + ur.Encode()
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

const previewTemplateEmptySvg = `<svg width="195" height="89" viewBox="0 0 195 89" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M81.6866 32.5202V1.5H153.758V56.4479C153.758 62.1076 149.171 66.6944 143.511 66.6944C137.851 66.6944 133.265 62.1076 133.265 56.4479L133.279 42.8027C133.279 37.143 128.692 32.5562 123.033 32.5562C117.373 32.5562 112.786 37.143 112.786 42.8027L112.8 71.0939C112.8 76.7536 108.214 81.3404 102.554 81.3404C96.8943 81.3404 92.3075 76.7536 92.3075 71.0939L92.3219 61.4379C92.3219 55.7782 87.7351 51.1914 82.0754 51.1914H82.0394C76.3797 51.1914 71.7929 55.7782 71.7929 61.4379L71.8074 74.4782C71.8074 80.1379 67.2206 84.7247 61.5609 84.7247H2.625" fill="white"/>
<rect x="87.125" y="18.75" width="21.5" height="13.75" rx="1" fill="#F5F5F5"/>
<rect x="119.375" y="18.75" width="28.5" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="89.375" y="22" width="16.75" height="1.5" rx="0.5" fill="#D3D3D3"/>
<rect x="89.375" y="24.75" width="16.75" height="1.5" rx="0.5" fill="#D3D3D3"/>
<rect x="89.375" y="27.75" width="10.25" height="1.5" rx="0.5" fill="#D3D3D3"/>
<rect x="119.375" y="29.5" width="28.5" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="132.125" y="38.75" width="15.75" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="132.125" y="48.25" width="15.75" height="3" rx="0.5" fill="#EBEBEB"/>
<path d="M0.125 84.7178C5.78468 84.7178 10.3715 80.131 10.3715 74.4713L10.3571 61.431C10.3571 55.7714 14.9439 51.1846 20.6035 51.1846H20.6396H82.0536" fill="white"/>
<path d="M41.7385 51.1851C41.7329 47.9007 41.7241 42.7748 41.7241 42.7748C41.7241 37.1151 46.3109 32.5283 51.9706 32.5283H122.681" fill="white"/>
<path d="M102.625 81.25H69.875H69.125C70.7865 79.44 71.5466 77.0484 71.5466 74.417L71.532 61.4487C71.532 55.8151 76.1725 51.25 81.889 51.25H81.9256C87.6494 51.25 92.2826 55.8151 92.2826 61.4487L92.268 71.0513C92.268 76.6849 96.9085 81.25 102.625 81.25Z" fill="#838383"/>
<path d="M144.125 66.75H112.745V66.25C112.745 63 112.745 62.0021 112.745 59L112.457 44.1404C112.457 37.7104 117.185 32.5 123.008 32.5H123.045C128.876 32.5 133.596 37.7104 133.596 44.1404L133.581 55.1096C133.574 61.5396 138.294 66.75 144.125 66.75Z" fill="#838383"/>
<path d="M81.625 2.25C81.625 1.14543 82.5204 0.25 83.625 0.25H151.875C152.98 0.25 153.875 1.14543 153.875 2.25V9.25H81.625V2.25Z" fill="#3E63DD"/>
<circle cx="88" cy="4.625" r="1.125" fill="white"/>
<circle cx="92.5" cy="4.625" r="1.125" fill="white"/>
<circle cx="97" cy="4.625" r="1.125" fill="white"/>
<path d="M24.375 61.875C24.375 52.075 30.375 51.25 33.625 51.25H66C57.3 51.85 56.5 58.125 56.5 63V66.875C56.5 67.4273 56.0523 67.875 55.5 67.875H26.375C25.2704 67.875 24.375 67.9579 24.375 66.8534L24.375 61.875Z" fill="#F5F5F5"/>
<rect x="41.875" y="47.25" width="18.75" height="4" fill="white"/>
<rect x="48.375" y="38.5" width="31" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="48.375" y="45" width="31" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="86.875" y="38.5" width="21.75" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="86.875" y="45" width="21.75" height="3" rx="0.5" fill="#EBEBEB"/>
<rect x="29.625" y="56" width="23" height="1.5" rx="0.5" fill="#E8E8E8"/>
<rect x="29.625" y="61.75" width="14" height="1.5" rx="0.5" fill="#E8E8E8"/>
<path fill-rule="evenodd" clip-rule="evenodd" d="M147.897 86.9998C146.054 87.6446 144.077 87.9946 142.019 87.9946C132.037 87.9946 123.945 79.7596 123.945 69.6012C123.945 59.6597 131.695 51.5604 141.382 51.2189C145.126 44.3778 152.304 39.75 160.545 39.75C170.975 39.75 179.704 47.1642 181.92 57.0964C189.269 57.9452 194.979 64.294 194.979 72C194.979 80.2843 188.38 87 180.239 87C180.179 87 180.118 86.9996 180.057 86.9989V86.9998H147.897Z" fill="#3E63DD"/>
<path fill-rule="evenodd" clip-rule="evenodd" d="M138.221 88.2701C137.554 88.3427 136.876 88.38 136.19 88.38C125.937 88.38 117.625 80.0681 117.625 69.815C117.625 59.5618 125.937 51.25 136.19 51.25C136.641 51.25 137.088 51.2661 137.531 51.2977C141.607 45.5221 148.333 41.75 155.939 41.75C166.292 41.75 175.013 48.7375 177.643 58.2536C185.779 58.431 192.319 65.0817 192.319 73.26C192.319 81.5498 185.599 88.27 177.309 88.27C176.911 88.27 176.516 88.2544 176.125 88.2239V88.2701H138.221Z" fill="#8DA4EF"/>
<path d="M154.295 54.6913C154.684 54.0504 155.614 54.0504 156.004 54.6913L164.488 68.652C164.893 69.3184 164.414 70.1714 163.634 70.1714H146.665C145.885 70.1714 145.405 69.3185 145.81 68.652L154.295 54.6913Z" fill="white"/>
<rect x="151.832" y="64.8193" width="6.63599" height="15.168" rx="1" fill="white"/>
</svg>`
