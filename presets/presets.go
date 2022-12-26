package presets

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/presets/actions"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/perm"
	h "github.com/theplant/htmlgo"
	"go.uber.org/zap"
	goji "goji.io"
	"goji.io/middleware"
	"goji.io/pat"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type Builder struct {
	prefix                                string
	models                                []*ModelBuilder
	mux                                   *goji.Mux
	builder                               *web.Builder
	i18nBuilder                           *i18n.Builder
	logger                                *zap.Logger
	permissionBuilder                     *perm.Builder
	verifier                              *perm.Verifier
	layoutFunc                            func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc)
	dataOperator                          DataOperator
	messagesFunc                          MessagesFunc
	homePageFunc                          web.PageFunc
	notFoundFunc                          web.PageFunc
	homePageLayoutConfig                  *LayoutConfig
	notFoundPageLayoutConfig              *LayoutConfig
	brandFunc                             ComponentFunc
	profileFunc                           ComponentFunc
	switchLanguageFunc                    ComponentFunc
	brandProfileSwitchLanguageDisplayFunc func(brand, profile, switchLanguage h.HTMLComponent) h.HTMLComponent
	menuTopItems                          []ComponentFunc
	notificationCountFunc                 func(ctx *web.EventContext) int
	notificationContentFunc               ComponentFunc
	brandTitle                            string
	vuetifyOptions                        string
	progressBarColor                      string
	rightDrawerWidth                      string
	writeFieldDefaults                    *FieldDefaults
	listFieldDefaults                     *FieldDefaults
	detailFieldDefaults                   *FieldDefaults
	extraAssets                           []*extraAsset
	assetFunc                             AssetFunc
	menuGroups                            MenuGroups
	menuOrder                             []interface{}
	wrapHandlers                          []func(in http.Handler) (out http.Handler)
}

type AssetFunc func(ctx *web.EventContext)

type extraAsset struct {
	path        string
	contentType string
	body        web.ComponentsPack
	refTag      string
}

const (
	CoreI18nModuleKey   i18n.ModuleKey = "CoreI18nModuleKey"
	ModelsI18nModuleKey i18n.ModuleKey = "ModelsI18nModuleKey"
)

const (
	OpenConfirmationDialogEvent = "presets_ConfirmationDialogEvent"
)

func New() *Builder {
	l, _ := zap.NewDevelopment()
	r := &Builder{
		logger:  l,
		builder: web.New(),
		i18nBuilder: i18n.New().
			RegisterForModule(language.English, CoreI18nModuleKey, Messages_en_US).
			RegisterForModule(language.SimplifiedChinese, CoreI18nModuleKey, Messages_zh_CN),
		writeFieldDefaults:   NewFieldDefaults(WRITE),
		listFieldDefaults:    NewFieldDefaults(LIST),
		detailFieldDefaults:  NewFieldDefaults(DETAIL),
		progressBarColor:     "amber",
		brandTitle:           "Admin",
		rightDrawerWidth:     "600",
		verifier:             perm.NewVerifier(PermModule, nil),
		homePageLayoutConfig: &LayoutConfig{SearchBoxInvisible: true},
		notFoundPageLayoutConfig: &LayoutConfig{
			SearchBoxInvisible:          true,
			NotificationCenterInvisible: true,
		},
	}

	r.GetWebBuilder().RegisterEventFunc(OpenConfirmationDialogEvent, r.openConfirmationDialog)
	r.layoutFunc = r.defaultLayout
	return r
}

func (b *Builder) I18n() (r *i18n.Builder) {
	return b.i18nBuilder
}

func (b *Builder) SetI18n(v *i18n.Builder) (r *Builder) {
	b.i18nBuilder = v
	return b
}

func (b *Builder) Permission(v *perm.Builder) (r *Builder) {
	b.permissionBuilder = v
	b.verifier = perm.NewVerifier(PermModule, v)
	return b
}

func (b *Builder) GetPermission() (r *perm.Builder) {
	return b.permissionBuilder
}

func (b *Builder) URIPrefix(v string) (r *Builder) {
	b.prefix = strings.TrimRight(v, "/")
	return b
}

func (b *Builder) GetURIPrefix() string {
	return b.prefix
}

func (b *Builder) LayoutFunc(v func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc)) (r *Builder) {
	b.layoutFunc = v
	return b
}

func (b *Builder) HomePageLayoutConfig(v *LayoutConfig) (r *Builder) {
	b.homePageLayoutConfig = v
	return b
}

func (b *Builder) NotFoundPageLayoutConfig(v *LayoutConfig) (r *Builder) {
	b.notFoundPageLayoutConfig = v
	return b
}

func (b *Builder) Builder(v *web.Builder) (r *Builder) {
	b.builder = v
	return b
}

func (b *Builder) GetWebBuilder() (r *web.Builder) {
	return b.builder
}

func (b *Builder) Logger(v *zap.Logger) (r *Builder) {
	b.logger = v
	return b
}

func (b *Builder) MessagesFunc(v MessagesFunc) (r *Builder) {
	b.messagesFunc = v
	return b
}

func (b *Builder) HomePageFunc(v web.PageFunc) (r *Builder) {
	b.homePageFunc = v
	return b
}

func (b *Builder) NotFoundFunc(v web.PageFunc) (r *Builder) {
	b.notFoundFunc = v
	return b
}

func (b *Builder) BrandFunc(v ComponentFunc) (r *Builder) {
	b.brandFunc = v
	return b
}

func (b *Builder) ProfileFunc(v ComponentFunc) (r *Builder) {
	b.profileFunc = v
	return b
}

func (b *Builder) SwitchLanguageFunc(v ComponentFunc) (r *Builder) {
	b.switchLanguageFunc = v
	return b
}

func (b *Builder) BrandProfileSwitchLanguageDisplayFuncFunc(f func(brand, profile, switchLanguage h.HTMLComponent) h.HTMLComponent) (r *Builder) {
	b.brandProfileSwitchLanguageDisplayFunc = f
	return b
}

func (b *Builder) NotificationFunc(contentFunc ComponentFunc, countFunc func(ctx *web.EventContext) int) (r *Builder) {
	b.notificationCountFunc = countFunc
	b.notificationContentFunc = contentFunc
	b.GetWebBuilder().RegisterEventFunc(actions.NotificationCenter, b.notificationCenter)
	return b
}

func (b *Builder) BrandTitle(v string) (r *Builder) {
	b.brandTitle = v
	return b
}

func (b *Builder) VuetifyOptions(v string) (r *Builder) {
	b.vuetifyOptions = v
	return b
}

func (b *Builder) RightDrawerWidth(v string) (r *Builder) {
	b.rightDrawerWidth = v
	return b
}

func (b *Builder) ProgressBarColor(v string) (r *Builder) {
	b.progressBarColor = v
	return b
}

func (b *Builder) AssetFunc(v AssetFunc) (r *Builder) {
	b.assetFunc = v
	return b
}

func (b *Builder) ExtraAsset(path string, contentType string, body web.ComponentsPack, refTag ...string) (r *Builder) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var theOne *extraAsset
	for _, ea := range b.extraAssets {
		if ea.path == path {
			theOne = ea
			break
		}
	}

	if theOne == nil {
		theOne = &extraAsset{path: path, contentType: contentType, body: body}
		b.extraAssets = append(b.extraAssets, theOne)
	} else {
		theOne.contentType = contentType
		theOne.body = body
	}

	if len(refTag) > 0 {
		theOne.refTag = refTag[0]
	}

	return b
}

func (b *Builder) FieldDefaults(v FieldMode) (r *FieldDefaults) {
	if v == WRITE {
		return b.writeFieldDefaults
	}

	if v == LIST {
		return b.listFieldDefaults
	}

	if v == DETAIL {
		return b.detailFieldDefaults
	}

	return r
}

func (b *Builder) NewFieldsBuilder(v FieldMode) (r *FieldsBuilder) {
	r = NewFieldsBuilder().Defaults(b.FieldDefaults(v))
	return
}

func (b *Builder) Model(v interface{}) (r *ModelBuilder) {
	r = NewModelBuilder(b, v)
	b.models = append(b.models, r)
	return r
}

func (b *Builder) DataOperator(v DataOperator) (r *Builder) {
	b.dataOperator = v
	return b
}

func modelNames(ms []*ModelBuilder) (r []string) {
	for _, m := range ms {
		r = append(r, m.uriName)
	}
	return
}

func (b *Builder) defaultBrandFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	return
}

func (b *Builder) MenuGroup(name string) *MenuGroupBuilder {
	mgb := b.menuGroups.MenuGroup(name)
	if !b.isMenuGroupInOrder(mgb) {
		b.menuOrder = append(b.menuOrder, mgb)
	}
	return mgb
}

func (b *Builder) isMenuGroupInOrder(mgb *MenuGroupBuilder) bool {
	for _, v := range b.menuOrder {
		if v == mgb {
			return true
		}
	}
	return false
}

func (b *Builder) removeMenuGroupInOrder(mgb *MenuGroupBuilder) {
	for i, om := range b.menuOrder {
		if om == mgb {
			b.menuOrder = append(b.menuOrder[:i], b.menuOrder[i+1:]...)
			break
		}
	}

}

// item can be Slug name, model name, *MenuGroupBuilder
// the underlying logic is using Slug name,
// so if the Slug name is customized, item must be the Slug name
// example:
// b.MenuOrder(
//
//	b.MenuGroup("Product Management").SubItems(
//		"products",
//		"Variant",
//	),
//	"customized-uri",
//
// )
func (b *Builder) MenuOrder(items ...interface{}) {
	for _, item := range items {
		switch v := item.(type) {
		case string:
			b.menuOrder = append(b.menuOrder, v)
		case *MenuGroupBuilder:
			if b.isMenuGroupInOrder(v) {
				b.removeMenuGroupInOrder(v)
			}
			b.menuOrder = append(b.menuOrder, v)
		default:
			panic(fmt.Sprintf("unknown menu order item type: %T\n", item))
		}
	}
}

type defaultMenuIconRE struct {
	re   *regexp.Regexp
	icon string
}

var defaultMenuIconREs = []defaultMenuIconRE{
	// user
	{re: regexp.MustCompile(`\busers?|members?\b`), icon: "person"},
	// store
	{re: regexp.MustCompile(`\bstores?\b`), icon: "store"},
	// order
	{re: regexp.MustCompile(`\borders?\b`), icon: "shopping_cart"},
	// product
	{re: regexp.MustCompile(`\bproducts?\b`), icon: "format_list_bulleted"},
	// post
	{re: regexp.MustCompile(`\bposts?|articles?\b`), icon: "article"},
	// web
	{re: regexp.MustCompile(`\bweb|site\b`), icon: "web"},
	// seo
	{re: regexp.MustCompile(`\bseo\b`), icon: "travel_explore"},
	// i18n
	{re: regexp.MustCompile(`\bi18n|translations?\b`), icon: "language"},
	// chart
	{re: regexp.MustCompile(`\banalytics?|charts?|statistics?\b`), icon: "analytics"},
	// dashboard
	{re: regexp.MustCompile(`\bdashboard\b`), icon: "dashboard"},
	// setting
	{re: regexp.MustCompile(`\bsettings?\b`), icon: "settings"},
}

func defaultMenuIcon(mLabel string) string {
	ws := strings.Join(strings.Split(strcase.ToSnake(mLabel), "_"), " ")
	for _, v := range defaultMenuIconREs {
		if v.re.MatchString(ws) {
			return v.icon
		}
	}

	return "widgets"
}

const menuFontWeight = "500"
const subMenuFontWeight = "400"

func (b *Builder) menuItem(ctx *web.EventContext, m *ModelBuilder, isSub bool) (r h.HTMLComponent) {
	menuIcon := m.menuIcon
	fontWeight := subMenuFontWeight
	if isSub {
		// menuIcon = ""
	} else {
		fontWeight = menuFontWeight
		if menuIcon == "" {
			menuIcon = defaultMenuIcon(m.label)
		}
	}

	href := m.Info().ListingHref()
	if m.link != "" {
		href = m.link
	}
	if m.defaultURLQueryFunc != nil {
		href = fmt.Sprintf("%s?%s", href, m.defaultURLQueryFunc(ctx.R).Encode())
	}
	item := VListItem(
		VListItemAction(
			VIcon(menuIcon),
		).Attr("style", "margin-right: 16px"),
		VListItemContent(
			VListItemTitle(
				h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, m.label)),
			).Attr("style", fmt.Sprintf("white-space: normal; font-weight: %s;font-size: 14px;", fontWeight)),
		),
	)
	if strings.HasPrefix(href, "/") {
		item.Attr("@click", web.Plaid().PushStateURL(href).Go())
	} else {
		item.Href(href)
	}
	if b.isMenuItemActive(ctx, m) {
		item = item.Class("v-list-item--active")
	}
	return item
}

func (b *Builder) isMenuItemActive(ctx *web.EventContext, m *ModelBuilder) bool {
	href := m.Info().ListingHref()
	if m.link != "" {
		href = m.link
	}
	path := strings.TrimSuffix(ctx.R.URL.Path, "/")
	if path == href {
		return true
	}
	if href == b.prefix {
		return false
	}
	if strings.HasPrefix(path, href) {
		return true
	}

	return false
}

func (b *Builder) createMenus(ctx *web.EventContext) (r h.HTMLComponent) {
	mMap := make(map[string]*ModelBuilder)
	for _, m := range b.models {
		mMap[m.uriName] = m
	}

	inOrderMap := make(map[string]struct{})
	var menus []h.HTMLComponent
	for _, om := range b.menuOrder {
		switch v := om.(type) {
		case *MenuGroupBuilder:
			ver := b.verifier.Do(PermList).On("menu:groups").SnakeOn(v.name).WithReq(ctx.R)
			if ver.IsAllowed() != nil {
				continue
			}
			groupIcon := v.icon
			if groupIcon == "" {
				groupIcon = defaultMenuIcon(v.name)
			}
			var subMenus = []h.HTMLComponent{
				VListItem(
					VListItemAction(
						VIcon(groupIcon),
					).Attr("style", "margin-right: 16px;"),
					VListItemContent(
						VListItemTitle(h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, v.name))).
							Attr("style", fmt.Sprintf("white-space: normal; font-weight: %s;font-size: 14px;", menuFontWeight)),
					),
				).Slot("activator").Class("pa-0"),
			}
			subCount := 0
			hasActiveMenuItem := false
			for _, subOm := range v.subMenuItems {
				m, ok := mMap[subOm]
				if !ok {
					m = mMap[inflection.Plural(strcase.ToKebab(subOm))]
				}
				if m == nil {
					continue
				}
				m.menuGroupName = v.name
				if m.notInMenu {
					continue
				}
				if ver.SnakeOn(m.uriName).IsAllowed() != nil {
					ver.RemoveOn(1)
					continue
				}
				ver.RemoveOn(1)
				subMenus = append(subMenus, b.menuItem(ctx, m, true))
				subCount++
				inOrderMap[m.uriName] = struct{}{}
				if b.isMenuItemActive(ctx, m) {
					hasActiveMenuItem = true
				}
			}
			if subCount == 0 {
				continue
			}

			menus = append(menus, VListGroup(
				subMenus...).
				Value(hasActiveMenuItem),
			)
		case string:
			m, ok := mMap[v]
			if !ok {
				m = mMap[inflection.Plural(strcase.ToKebab(v))]
			}
			if m == nil {
				continue
			}
			if b.verifier.Do(PermList).On("menu").SnakeOn(m.uriName).WithReq(ctx.R).IsAllowed() != nil {
				continue
			}

			if m.notInMenu {
				continue
			}

			menus = append(menus, b.menuItem(ctx, m, false))
			inOrderMap[m.uriName] = struct{}{}
		}
	}

	for _, m := range b.models {
		_, ok := inOrderMap[m.uriName]
		if ok {
			continue
		}

		if b.verifier.Do(PermList).On("menu").SnakeOn(m.uriName).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		if m.notInMenu {
			continue
		}
		menus = append(menus, b.menuItem(ctx, m, false))
	}

	r = h.Div(
		VList(menus...).Class("primary--text").Dense(true),
	)
	return
}

func (b *Builder) runBrandFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.brandFunc != nil {
		return b.brandFunc(ctx)
	}

	return VRow(h.H1(i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle))).Class("text-button")
}

func (b *Builder) runSwitchLanguageFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.switchLanguageFunc != nil {
		return b.switchLanguageFunc(ctx)
	}

	var supportLanguages = b.I18n().GetSupportLanguagesFromRequest(ctx.R)

	if len(b.I18n().GetSupportLanguages()) <= 1 || len(supportLanguages) == 0 {
		return nil
	}

	queryName := b.I18n().GetQueryName()
	msgr := MustGetMessages(ctx.R)

	if len(supportLanguages) == 1 {
		return h.Template().Children(
			h.Div(
				VList(
					VListItem(
						VListItemIcon(
							VIcon("translate").Small(true).Class("ml-1"),
						).Attr("style", "margin-right: 16px"),
						VListItemContent(
							VListItemTitle(
								h.Div(h.Text(fmt.Sprintf("%s%s %s", msgr.Language, msgr.Colon, display.Self.Name(supportLanguages[0])))).Role("button"),
							),
						),
					).Class("pa-0").Dense(true),
				).Class("pa-0 ma-n4 mt-n6"),
			).Attr("@click", web.Plaid().Query(queryName, supportLanguages[0].String()).Go()),
		)
	}

	var matcher = language.NewMatcher(supportLanguages)

	lang := ctx.R.FormValue(queryName)
	if lang == "" {
		lang = b.i18nBuilder.GetCurrentLangFromCookie(ctx.R)
	}

	accept := ctx.R.Header.Get("Accept-Language")

	var displayLanguage language.Tag
	_, i := language.MatchStrings(matcher, lang, accept)
	displayLanguage = supportLanguages[i]

	var languages []h.HTMLComponent
	for _, tag := range supportLanguages {
		languages = append(languages,
			h.Div(
				VListItem(
					VListItemContent(
						VListItemTitle(
							h.Div(h.Text(display.Self.Name(tag))),
						),
					),
				).Attr("@click", web.Plaid().Query(queryName, tag.String()).Go()),
			),
		)
	}

	return VMenu().OffsetY(true).Children(
		h.Template().Attr("v-slot:activator", "{on, attrs}").Children(
			h.Div(
				VList(
					VListItem(
						VListItemIcon(
							VIcon("translate").Small(true).Class("ml-1"),
						).Attr("style", "margin-right: 16px"),
						VListItemContent(
							VListItemTitle(
								h.Text(fmt.Sprintf("%s%s %s", msgr.Language, msgr.Colon, display.Self.Name(displayLanguage))),
							),
						),
						VListItemIcon(
							VIcon("arrow_drop_down").Small(false).Class("mr-1"),
						),
					).Class("pa-0").Dense(true),
				).Class("pa-0 ma-n4 mt-n6"),
			).Attr("v-bind", "attrs").Attr("v-on", "on"),
		),

		VList(
			languages...,
		).Dense(true),
	)
}

func (b *Builder) AddMenuTopItemFunc(v ComponentFunc) (r *Builder) {
	b.menuTopItems = append(b.menuTopItems, v)
	return b
}

func (b *Builder) runBrandProfileSwitchLanguageDisplayFunc(brand, profile, switchLanguage h.HTMLComponent, ctx *web.EventContext) (r h.HTMLComponent) {
	if b.brandProfileSwitchLanguageDisplayFunc != nil {
		return b.brandProfileSwitchLanguageDisplayFunc(brand, profile, switchLanguage)
	}

	var items []h.HTMLComponent
	items = append(items,
		h.If(brand != nil,
			VListItem(
				VCardText(brand),
			),
		),
		h.If(profile != nil,
			VListItem(
				VCardText(profile),
			),
		),
		h.If(switchLanguage != nil,
			VListItem(
				VCardText(switchLanguage),
			).Dense(true),
		),
	)
	for _, v := range b.menuTopItems {
		items = append(items,
			h.If(v(ctx) != nil,
				VListItem(
					VCardText(v(ctx)),
				),
			))
	}

	return h.Div(
		items...,
	)
}

func MustGetMessages(r *http.Request) *Messages {
	return i18n.MustGetModuleMessages(r, CoreI18nModuleKey, Messages_en_US).(*Messages)
}

const RightDrawerPortalName = "presets_RightDrawerPortalName"
const rightDrawerContentPortalName = "presets_RightDrawerContentPortalName"
const DialogPortalName = "presets_DialogPortalName"
const dialogContentPortalName = "presets_DialogContentPortalName"
const NotificationCenterPortalName = "notification-center"
const defaultConfirmationDialogPortalName = "presets_confirmationDialogPortalName"
const listingDialogPortalName = "presets_listingDialogPortalName"

const closeRightDrawerVarScript = "vars.presetsRightDrawer = false"
const closeDialogVarScript = "vars.presetsDialog = false"
const CloseListingDialogVarScript = "vars.presetsListingDialog = false"

func (b *Builder) overlay(overlayType string, r *web.EventResponse, comp h.HTMLComponent, width string) {
	if overlayType == actions.Dialog {
		b.dialog(r, comp, width)
		return
	}
	b.rightDrawer(r, comp, width)
}

func (b *Builder) rightDrawer(r *web.EventResponse, comp h.HTMLComponent, width string) {
	if width == "" {
		width = b.rightDrawerWidth
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: RightDrawerPortalName,
		Body: VNavigationDrawer(
			web.GlobalEvents().Attr("@keyup.esc", "vars.presetsRightDrawer = false"),
			web.Portal(comp).Name(rightDrawerContentPortalName),
		).
			// Attr("@input", "plaidForm.dirty && vars.presetsRightDrawer == false && !confirm('You have unsaved changes on this form. If you close it, you will lose all unsaved changes. Are you sure you want to close it?') ? vars.presetsRightDrawer = true: vars.presetsRightDrawer = $event"). // remove because drawer plaidForm has to be reset when UpdateOverlayContent
			Class("v-navigation-drawer--temporary").
			Attr("v-model", "vars.presetsRightDrawer").
			Right(true).
			Fixed(true).
			Attr("width", width).
			Bottom(false).
			Attr(":height", `"100%"`),
		// Temporary(true),
		// HideOverlay(true).
		// Floating(true).

	})
	r.VarsScript = "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)"
}

// 				Attr("@input", "alert(plaidForm.dirty) && !confirm('You have unsaved changes on this form. If you close it, you will lose all unsaved changes. Are you sure you want to close it?') ? vars.presetsDialog = true : vars.presetsDialog = $event").

func (b *Builder) dialog(r *web.EventResponse, comp h.HTMLComponent, width string) {
	if width == "" {
		width = b.rightDrawerWidth
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DialogPortalName,
		Body: web.Scope(
			VDialog(
				web.Portal(comp).Name(dialogContentPortalName),
			).
				Attr("v-model", "vars.presetsDialog").
				Width(width),
		).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){ vars.presetsDialog = true }, 100)"
}

type LayoutConfig struct {
	SearchBoxInvisible          bool
	NotificationCenterInvisible bool
}

func (b *Builder) notificationCenter(ctx *web.EventContext) (er web.EventResponse, err error) {
	total := b.notificationCountFunc(ctx)
	content := b.notificationContentFunc(ctx)

	icon := VIcon("notifications").Color("white")
	er.Body = VMenu().OffsetY(true).Children(
		h.Template().Attr("v-slot:activator", "{on, attrs}").Children(
			VBtn("").Icon(true).Children(
				h.If(total > 0,
					VBadge(
						icon,
					).Content(total).Overlap(true).Color("red"),
				).Else(icon),
			).Attr("v-bind", "attrs").Attr("v-on", "on").Class("ml-1"),
		),
		VCard(content))

	return
}

const (
	ConfirmationDialogConfirmEventKey = "presets_ConfirmationDialogConfirmEventKey"
	ConfirmationDialogTextKey         = "presets_ConfirmationDialogTextKey"
	ConfirmationDialogPortalNameKey   = "presets_ConfirmationDialogPortalNameKey"
)

func (b *Builder) openConfirmationDialog(ctx *web.EventContext) (er web.EventResponse, err error) {
	confirmEvent := ctx.R.FormValue(ConfirmationDialogConfirmEventKey)
	if confirmEvent == "" {
		ShowMessage(&er, "confirm event is empty", "error")
		return
	}

	msgr := MustGetMessages(ctx.R)
	text := msgr.ConfirmationDialogText
	if v := ctx.R.FormValue(ConfirmationDialogTextKey); v != "" {
		text = v
	}

	portal := defaultConfirmationDialogPortalName
	if v := ctx.R.FormValue(ConfirmationDialogPortalNameKey); v != "" {
		portal = v
	}
	showVar := fmt.Sprintf("show_%s", portal)

	er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
		Name: portal,
		Body: VDialog(
			VCard(
				VCardTitle(VIcon("warning").Class("red--text mr-4"), h.Text(text)),
				VCardActions(
					VSpacer(),
					VBtn(msgr.Cancel).
						Depressed(true).
						Class("ml-2").
						On("click", fmt.Sprintf("vars.%s = false", showVar)),

					VBtn(msgr.OK).
						Color("primary").
						Depressed(true).
						Dark(true).
						Attr("@click", fmt.Sprintf("%s;vars.%s = false", confirmEvent, showVar)),
				),
			),
		).MaxWidth("600px").
			Attr("v-model", fmt.Sprintf("vars.%s", showVar)).
			Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, showVar)),
	})

	er.VarsScript = fmt.Sprintf("setTimeout(function(){ vars.%s = true }, 100)", showVar)
	return
}

func (b *Builder) defaultLayout(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		b.InjectAssets(ctx)

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err == perm.PermissionDenied {
			pr.Body = h.Text(perm.PermissionDenied.Error())
			return pr, nil
		}
		if err != nil {
			panic(err)
		}

		var profile h.HTMLComponent
		if b.profileFunc != nil {
			profile = b.profileFunc(ctx)
		}

		showNotificationCenter := cfg == nil || !cfg.NotificationCenterInvisible
		var notifier h.HTMLComponent
		if b.notificationCountFunc != nil && b.notificationContentFunc != nil {
			notifier = web.Portal().Name(NotificationCenterPortalName).Loader(web.GET().EventFunc(actions.NotificationCenter))
		}

		showSearchBox := cfg == nil || !cfg.SearchBoxInvisible

		msgr := i18n.MustGetModuleMessages(ctx.R, CoreI18nModuleKey, Messages_en_US).(*Messages)

		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle))
		pr.Body = VApp(
			VNavigationDrawer(
				b.runBrandProfileSwitchLanguageDisplayFunc(b.runBrandFunc(ctx), profile, b.runSwitchLanguageFunc(ctx), ctx),
				VDivider(),
				b.createMenus(ctx),
			).App(true).
				// Clipped(true).
				Fixed(true).
				Value(true).
				Attr("v-model", "vars.navDrawer").
				Attr(web.InitContextVars, `{navDrawer: null}`),

			VAppBar(
				VAppBarNavIcon().On("click.stop", "vars.navDrawer = !vars.navDrawer"),
				h.Span(innerPr.PageTitle).Class("text-h6 font-weight-regular"),
				VSpacer(),
				h.If(showSearchBox,
					VLayout(
						// h.Form(
						VTextField().
							SoloInverted(true).
							PrependIcon("search").
							Label(msgr.Search).
							Flat(true).
							Clearable(true).
							HideDetails(true).
							Value(ctx.R.URL.Query().Get("keyword")).
							Attr("@keyup.enter", web.Plaid().
								Query("keyword", web.Var("[$event.target.value]")).
								MergeQuery(true).
								PushState(true).
								Go()).
							Attr("@click:clear", web.Plaid().
								Query("keyword", "").
								PushState(true).
								Go()),
						// ).Method("GET"),
					).AlignCenter(true).Attr("style", "max-width: 650px"),
				),
				h.If(showNotificationCenter,
					notifier,
				),
			).Dark(true).
				Color(ColorPrimary).
				App(true).
				Fixed(true),
			// ClippedLeft(true),

			web.Portal().Name(RightDrawerPortalName),
			web.Portal().Name(DialogPortalName),
			web.Portal().Name(DeleteConfirmPortalName),
			web.Portal().Name(defaultConfirmationDialogPortalName),
			web.Portal().Name(listingDialogPortalName),

			VProgressLinear().
				Attr(":active", "isFetching").
				Attr("style", "position: fixed; z-index: 99").
				Indeterminate(true).
				Height(2).
				Color(b.progressBarColor),
			h.Template(
				VSnackbar(h.Text("{{vars.presetsMessage.message}}")).
					Attr("v-model", "vars.presetsMessage.show").
					Attr(":color", "vars.presetsMessage.color").
					Timeout(2000).
					Top(true),
			).Attr("v-if", "vars.presetsMessage"),
			VMain(
				innerPr.Body.(h.HTMLComponent),
			),
		).Id("vt-app").
			Attr(web.InitContextVars, `{presetsRightDrawer: false, presetsDialog: false, presetsListingDialog: false, presetsMessage: {show: false, color: "success", message: ""}}`)

		return
	}
}

// for pages outside the default presets layout
func (b *Builder) PlainLayout(in web.PageFunc) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		b.InjectAssets(ctx)

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err == perm.PermissionDenied {
			pr.Body = h.Text(perm.PermissionDenied.Error())
			return pr, nil
		}
		if err != nil {
			panic(err)
		}

		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle))
		pr.Body = VApp(
			web.Portal().Name(DialogPortalName),
			web.Portal().Name(DeleteConfirmPortalName),
			web.Portal().Name(defaultConfirmationDialogPortalName),

			VProgressLinear().
				Attr(":active", "isFetching").
				Attr("style", "position: fixed; z-index: 99").
				Indeterminate(true).
				Height(2).
				Color(b.progressBarColor),
			h.Template(
				VSnackbar(h.Text("{{vars.presetsMessage.message}}")).
					Attr("v-model", "vars.presetsMessage.show").
					Attr(":color", "vars.presetsMessage.color").
					Timeout(2000).
					Top(true),
			).Attr("v-if", "vars.presetsMessage"),
			VMain(
				innerPr.Body.(h.HTMLComponent),
			),
		).Id("vt-app").
			Attr(web.InitContextVars, `{presetsDialog: false, presetsMessage: {show: false, color: "success", message: ""}}`)

		return
	}
}

func (b *Builder) InjectAssets(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(strings.Replace(`
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto+Mono">
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500">
			<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
			<link rel="stylesheet" href="{{prefix}}/assets/main.css">
			<script src='{{prefix}}/assets/vue.js'></script>
			<style>
				[v-cloak] {
					display: none;
				}
				.vx-list-item--active {
					position: relative;
				}
				.vx-list-item--active:after {
					opacity: .12;
					background-color: currentColor;
					bottom: 0;
					content: "";
					left: 0;
					pointer-events: none;
					position: absolute;
					right: 0;
					top: 0;
					transition: .3s cubic-bezier(.25,.8,.5,1);
					line-height: 0;
				}
				.vx-list-item--active:hover {
					background-color: inherit!important;
				}
			</style>
		`, "{{prefix}}", b.prefix, -1))

	b.InjectExtraAssets(ctx)

	if len(os.Getenv("DEV_PRESETS")) > 0 {
		ctx.Injector.TailHTML(`
<script src='http://localhost:3080/js/chunk-vendors.js'></script>
<script src='http://localhost:3080/js/app.js'></script>
<script src='http://localhost:3100/js/chunk-vendors.js'></script>
<script src='http://localhost:3100/js/app.js'></script>
			`)

	} else {
		ctx.Injector.TailHTML(strings.Replace(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix, -1))
	}

	if b.assetFunc != nil {
		b.assetFunc(ctx)
	}
}

func (b *Builder) InjectExtraAssets(ctx *web.EventContext) {
	for _, ea := range b.extraAssets {
		if len(ea.refTag) > 0 {
			ctx.Injector.HeadHTML(ea.refTag)
			continue
		}

		if strings.HasSuffix(ea.path, "css") {
			ctx.Injector.HeadHTML(fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">", b.extraFullPath(ea)))
			continue
		}

		ctx.Injector.HeadHTML(fmt.Sprintf("<script src=\"%s\"></script>", b.extraFullPath(ea)))
	}
}

func (b *Builder) defaultHomePageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	r.Body = h.Div().Text("home")
	return
}

func (b *Builder) getHomePageFunc() web.PageFunc {
	if b.homePageFunc != nil {
		return b.homePageFunc
	}
	return b.defaultHomePageFunc
}

func (b *Builder) defaultNotFoundPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	r.Body = h.Div(
		h.H1("404"),
		h.Text("Sorry, we couldn't find this page."),
	).Class("text-center mt-8")
	return
}

func (b *Builder) getNotFoundPageFunc() web.PageFunc {
	if b.notFoundFunc != nil {
		return b.notFoundFunc
	}
	return b.defaultNotFoundPageFunc
}

func (b *Builder) extraFullPath(ea *extraAsset) string {
	return b.prefix + "/extra" + ea.path
}

func (b *Builder) initMux() {
	b.logger.Info("initializing mux for", zap.Reflect("models", modelNames(b.models)), zap.String("prefix", b.prefix))
	mux := goji.NewMux()
	ub := b.builder

	mainJSPath := b.prefix + "/assets/main.js"
	mux.Handle(pat.Get(mainJSPath),
		ub.PacksHandler("text/javascript",
			Vuetify(b.vuetifyOptions),
			JSComponentsPack(),
			vuetifyx.JSComponentsPack(),
			web.JSComponentsPack(),
		),
	)
	log.Println("mounted url", mainJSPath)

	vueJSPath := b.prefix + "/assets/vue.js"
	mux.Handle(pat.Get(vueJSPath),
		ub.PacksHandler("text/javascript",
			web.JSVueComponentsPack(),
		),
	)
	log.Println("mounted url", vueJSPath)

	mainCSSPath := b.prefix + "/assets/main.css"
	mux.Handle(pat.Get(mainCSSPath),
		ub.PacksHandler("text/css",
			CSSComponentsPack(),
		),
	)
	log.Println("mounted url", mainCSSPath)

	for _, ea := range b.extraAssets {
		fullPath := b.extraFullPath(ea)
		mux.Handle(pat.Get(fullPath), ub.PacksHandler(
			ea.contentType,
			ea.body,
		))
		log.Println("mounted url", fullPath)
	}

	mux.Handle(
		pat.New(b.prefix),
		b.wrap(nil, b.layoutFunc(b.getHomePageFunc(), b.homePageLayoutConfig)),
	)

	for _, m := range b.models {
		pluralUri := inflection.Plural(m.uriName)
		info := m.Info()
		routePath := info.ListingHref()
		inPageFunc := m.listing.GetPageFunc()
		if m.singleton {
			inPageFunc = m.editing.singletonPageFunc
			if m.layoutConfig == nil {
				m.layoutConfig = &LayoutConfig{}
			}
			m.layoutConfig.SearchBoxInvisible = true
		}
		mux.Handle(
			pat.New(routePath),
			b.wrap(m, b.layoutFunc(inPageFunc, m.layoutConfig)),
		)
		log.Println("mounted url", routePath)
		if m.hasDetailing {
			routePath = fmt.Sprintf("%s/%s/:id", b.prefix, pluralUri)
			mux.Handle(
				pat.New(routePath),
				b.wrap(m, b.layoutFunc(m.detailing.GetPageFunc(), m.layoutConfig)),
			)
			log.Println("mounted url", routePath)
		}
	}

	// Handle 404
	mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.RequestURI, b.prefix) || middleware.Handler(r.Context()) != nil {
				handler.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusNotFound)
			b.wrap(
				nil,
				b.layoutFunc(b.getNotFoundPageFunc(), b.notFoundPageLayoutConfig),
			).ServeHTTP(w, r)
			return
		})
	})

	b.mux = mux
}
func (b *Builder) AddWrapHandler(f func(in http.Handler) (out http.Handler)) {
	b.wrapHandlers = append(b.wrapHandlers, f)
}

func (b *Builder) wrap(m *ModelBuilder, pf web.PageFunc) http.Handler {
	p := b.builder.Page(pf)
	if m != nil {
		m.registerDefaultEventFuncs()
		p.MergeHub(&m.EventsHub)
	}

	handlers := b.I18n().EnsureLanguage(
		p,
	)
	for _, wrapHandler := range b.wrapHandlers {
		handlers = wrapHandler(handlers)
	}

	return handlers
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.mux == nil {
		b.initMux()
	}
	RedirectSlashes(b.mux).ServeHTTP(w, r)
}

func RedirectSlashes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path[:len(path)-1], r.URL.RawQuery)
			} else {
				path = path[:len(path)-1]
			}
			redirectURL := fmt.Sprintf("//%s%s", r.Host, path)
			http.Redirect(w, r, redirectURL, 301)
			return
		}
		next.ServeHTTP(w, r)
	})
}
