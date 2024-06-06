package containers

import (
	"fmt"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type WebFooter struct {
	ID          uint
	EnglishUrl  string
	JapaneseUrl string
}

func (*WebFooter) TableName() string {
	return "container_footers"
}

func RegisterFooter(pb *pagebuilder.Builder) {
	footer := pb.RegisterContainer("Footer").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			footer := obj.(*WebFooter)
			return FooterTemplate(footer, input)
		})

	footer.Model(&WebFooter{}).Editing("EnglishUrl", "JapaneseUrl")
}

func FooterTemplate(data *WebFooter, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper("", "container-footer", "", "", "",
		"", false, false, "",
		Div(RawHTML(fmt.Sprintf(`
<div class='container-footer-main'>
<div class='container-footer-primary'>
<div class='container-footer-links'>
<div class='container-footer-links-group'>
<div class='container-footer-links-title'>
<a href='/what-we-do/'>What we do</a>
</div>

<ul data-list-unset='true' class='container-footer-links-list'>
<li class='container-footer-links-item'>
<a href='/commerce/'>Commerce</a>
</li>

<li class='container-footer-links-item'>
<a href='/content/'>Content</a>
</li>

<li class='container-footer-links-item'>
<a href='/consulting/'>Consulting</a>
</li>

<li class='container-footer-links-item'>
<a href='/personalization/'>Personalization</a>
</li>
</ul>
</div>

<div class='container-footer-links-group'>
<div class='container-footer-links-title'>
<a href='/projects/'>Projects</a>
</div>

<ul data-list-unset='true' class='container-footer-links-list'>
<li class='container-footer-links-item'>
<a href='/projects/?filter=Commerce'>Commerce</a>
</li>

<li class='container-footer-links-item'>
<a href='/projects/?filter=Content'>Content</a>
</li>

<li class='container-footer-links-item'>
<a href='/projects/?filter=Consulting'>Consulting</a>
</li>

<li class='container-footer-links-item'>
<a href='/projects/?filter=Mobile%%20Apps'>Mobile Apps</a>
</li>

<li class='container-footer-links-item'>
<a href='/projects/?filter=Personalization'>Personalization</a>
</li>
</ul>
</div>

<div class='container-footer-links-group'>
<div class='container-footer-links-title'>
<a href='/why-clients-choose-us/'>Why clients choose us</a>
</div>

<ul data-list-unset='true' class='container-footer-links-list'>
<li class='container-footer-links-item'>
<a href='	
/why-clients-choose-us/#speed'>Speed</a>
</li>

<li class='container-footer-links-item'>
<a href='/why-clients-choose-us/#reliability'>Reliability</a>
</li>

<li class='container-footer-links-item'>
<a href='/why-clients-choose-us/#expertise'>Expertise</a>
</li>

<li class='container-footer-links-item'>
<a href='/why-clients-choose-us/#pricing'>Pricing</a>
</li>

<li class='container-footer-links-item'>
<a href='/why-clients-choose-us/#workflow'>Workflow</a>
</li>
</ul>
</div>

<div class='container-footer-links-group'>
<div class='container-footer-links-title'>
<a href='/our-company/'>Our company</a>
</div>

<ul data-list-unset='true' class='container-footer-links-list'>
<li class='container-footer-links-item'>
<a href='/our-company/culture/'>Our culture</a>
</li>

<li class='container-footer-links-item'>
<a href='/join-our-team/'>Join our team</a>
</li>

<li class='container-footer-links-item'>
<a href='/news-and-events/'>News &amp; Events</a>
</li>

<li class='container-footer-links-item'>
<a href='/articles/'>Articles</a>
</li>

<li class='container-footer-links-item'>
<a href='/privacy-policy/'>Privacy policy</a>
</li>

<li class='container-footer-links-item'>
<a href='/contact/'>Contact us</a>
</li>
</ul>
</div>
</div>

<div class='container-footer-follow'>
<div class='container-footer-links-title'>Follow us</div>

<div class='container-footer-follow-link'>
<a href='https://www.facebook.com/theplant/' target='_blank' class='link-facebook'><svg width="40" height="40" viewBox="0 0 40 40" xmlns="http://www.w3.org/2000/svg"><path d="M8.29522 5.99976C7.44176 5.99976 6.75 6.6915 6.75 7.54511V32.4544C6.75 33.3077 7.44176 33.9998 8.29522 33.9998H21.7056V23.1567H18.0565V18.931H21.7056V15.8147C21.7056 12.1979 23.9145 10.2288 27.1406 10.2288C28.6861 10.2288 30.0142 10.3437 30.4012 10.3954V14.1748L28.1637 14.1759C26.409 14.1759 26.0695 15.0094 26.0695 16.233V18.931H30.2538L29.7092 23.1567H26.0695V33.9998H33.2045C34.058 33.9998 34.75 33.3077 34.75 32.4544V7.54511C34.75 6.6915 34.058 5.99976 33.2045 5.99976H8.29522Z" fill="currentColor"/></svg></a>

<a href='https://twitter.com/_theplant' target='_blank' class='link-twitter'><svg width="40" height="40" viewBox="0 0 40 40" xmlns="http://www.w3.org/2000/svg"><path d="M14.8126 34C26.8873 34 33.4916 23.9962 33.4916 15.321C33.4916 15.0369 33.4916 14.754 33.4724 14.4725C34.7572 13.5431 35.8663 12.3924 36.7477 11.0743C35.5495 11.6052 34.2785 11.9534 32.9771 12.1072C34.3475 11.2867 35.3732 9.99632 35.8633 8.47609C34.5746 9.24077 33.1648 9.77969 31.6946 10.0696C29.6597 7.90575 26.4262 7.37615 23.8072 8.77773C21.1883 10.1793 19.8353 13.1635 20.5069 16.057C15.2285 15.7924 10.3105 13.2992 6.97704 9.19795C5.2346 12.1976 6.12461 16.035 9.00953 17.9615C7.9648 17.9305 6.94284 17.6487 6.02991 17.1398V17.223C6.03076 20.348 8.23359 23.0396 11.2967 23.6583C10.3302 23.9219 9.31617 23.9605 8.33246 23.771C9.19249 26.4452 11.6571 28.2773 14.4658 28.33C12.1411 30.157 9.26943 31.1488 6.31277 31.1458C5.79044 31.1448 5.26862 31.1132 4.75 31.0511C7.7522 32.9777 11.2454 33.9996 14.8126 33.9949" fill="currentColor"/></svg></a>

<a href='https://linkedin.com/company/the-plant' target='_blank' class='link-linkedin'><svg width="40" height="40" viewBox="0 0 40 40" xmlns="http://www.w3.org/2000/svg"><path d="M34.8015 34H28.9896V24.9041C28.9896 22.7356 28.9523 19.9458 25.9697 19.9458C22.9443 19.9458 22.4829 22.3093 22.4829 24.7496V34H16.6776V15.2964H22.2495V17.854H22.3306C23.1053 16.3834 25.0021 14.8329 27.8292 14.8329C33.7156 14.8329 34.8026 18.7054 34.8026 23.7426V34H34.8015ZM10.1228 12.7422C8.25558 12.7422 6.75 11.2323 6.75 9.37056C6.75 7.50886 8.25558 6 10.1228 6C11.9823 6 13.4911 7.50996 13.4911 9.37056C13.4911 11.2312 11.9823 12.7422 10.1228 12.7422ZM7.21132 34H13.0298V15.2964H7.21132V34Z" fill="currentColor"/></svg></a>
</div>
</div>
</div>

<div class='container-footer-secondary'>
<ul data-list-unset='true' class='container-footer-language'>
<li>
<a href='%s'>English</a>
</li>

<li>
<a href='%s'>日本語</a>
</li>
</ul>
<svg class="container-footer-logo" viewBox="0 0 151 29" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M137.448 9.664V0l-13.557 9.664v18.911H151V0l-13.552 9.664z" fill="#fff"/><path d="M5.654 13.053H0V9.63h15.254v3.423H9.599v15.522H5.654V13.053zM17.413 9.63h3.682v6.846c.544-.743 1.175-1.292 1.894-1.645a5.059 5.059 0 012.262-.531c1.56 0 2.717.416 3.471 1.247.754.814 1.131 2.105 1.131 3.874v9.154h-3.682v-8.65c0-.973-.184-1.636-.552-1.99-.368-.354-.86-.53-1.473-.53-.473 0-.885.07-1.236.212a3.457 3.457 0 00-.947.61 3.436 3.436 0 00-.631.902 2.59 2.59 0 00-.237 1.115v8.331h-3.682V9.63zM45.549 24.25c-.246 1.557-.868 2.742-1.868 3.556-.981.796-2.42 1.194-4.313 1.194-2.28 0-3.997-.637-5.154-1.91-1.14-1.274-1.71-3.078-1.71-5.413 0-1.168.167-2.203.5-3.105.333-.92.797-1.698 1.394-2.335a5.882 5.882 0 012.182-1.433c.842-.336 1.771-.504 2.788-.504 2.104 0 3.682.61 4.734 1.83 1.07 1.221 1.604 2.867 1.604 4.936v1.486h-9.52c.035 1.115.316 1.99.842 2.627.526.62 1.315.929 2.367.929 1.455 0 2.314-.62 2.577-1.858h3.577zm-3.393-4.139c0-.955-.237-1.698-.71-2.229-.456-.53-1.175-.796-2.157-.796-.49 0-.92.08-1.288.239-.369.16-.684.38-.947.663a2.896 2.896 0 00-.579.956c-.14.353-.219.742-.236 1.167h5.917zM54.61 9.63h6.706c1.438 0 2.621.168 3.55.504.947.319 1.692.752 2.236 1.3.56.53.947 1.159 1.157 1.884.228.708.342 1.45.342 2.23 0 .83-.114 1.626-.342 2.387a4.668 4.668 0 01-1.157 2.017c-.544.566-1.28 1.017-2.21 1.353-.911.336-2.068.504-3.471.504h-2.946v6.766H54.61V9.63zm6.68 8.889c.683 0 1.244-.07 1.683-.212.456-.142.815-.336 1.078-.584.263-.248.438-.549.526-.902.105-.372.158-.77.158-1.194 0-.443-.053-.832-.158-1.168a1.834 1.834 0 00-.552-.849c-.263-.23-.623-.407-1.079-.53-.438-.124-.999-.186-1.683-.186h-2.788v5.625h2.814zM71.169 9.63h3.682v18.945h-3.682V9.63zM86.281 26.824a7.142 7.142 0 01-1.84 1.513c-.667.389-1.535.583-2.604.583-.614 0-1.201-.08-1.762-.238a4.17 4.17 0 01-1.447-.743 3.722 3.722 0 01-.973-1.274c-.245-.53-.368-1.159-.368-1.884 0-.955.21-1.733.631-2.335.42-.601.973-1.07 1.657-1.406a8.266 8.266 0 012.288-.717c.86-.141 1.727-.23 2.604-.265l1.762-.08v-.69c0-.849-.237-1.432-.71-1.75-.456-.32-1-.479-1.631-.479-1.455 0-2.288.558-2.498 1.672l-3.367-.318c.246-1.45.877-2.494 1.894-3.131 1.017-.655 2.384-.982 4.102-.982 1.052 0 1.947.133 2.683.398.737.248 1.324.61 1.762 1.088.456.478.78 1.061.973 1.751.21.672.316 1.433.316 2.282v8.756H86.28v-1.75zm-.079-4.431l-1.63.08c-.772.035-1.394.114-1.868.238-.473.124-.841.283-1.104.478a1.33 1.33 0 00-.5.637 2.45 2.45 0 00-.131.822c0 .46.158.823.473 1.088.316.266.754.398 1.315.398.947 0 1.718-.22 2.314-.663.334-.248.605-.557.816-.929.21-.389.315-.867.315-1.433v-.716zM92.87 14.725h3.577v1.91c.544-.814 1.184-1.406 1.92-1.778a5.235 5.235 0 012.341-.557c1.56 0 2.717.416 3.471 1.247.754.814 1.131 2.105 1.131 3.874v9.154h-3.682v-8.65c0-.973-.184-1.636-.552-1.99-.368-.354-.859-.53-1.473-.53-.473 0-.885.07-1.236.212a3.455 3.455 0 00-.947.61 3.431 3.431 0 00-.63.902 2.59 2.59 0 00-.238 1.115v8.331h-3.681v-13.85zM109.276 17.617h-2.156v-2.893h2.156v-3.688h3.682v3.689h3.156v2.892h-3.156v6.395c0 .725.149 1.22.447 1.485.298.248.684.372 1.157.372.246 0 .491-.009.737-.027.263-.035.517-.088.762-.159l.526 2.733a7.825 7.825 0 01-1.499.319c-.473.07-.929.106-1.367.106-1.508 0-2.63-.363-3.367-1.088-.718-.725-1.078-1.946-1.078-3.662v-6.474z" fill="#fff"/></svg>
</div>
</div>
`, data.EnglishUrl, data.JapaneseUrl))).Class("container-wrapper"))
	return
}
