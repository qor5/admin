package presets

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

type Builder struct {
	prefix                                string
	models                                []*ModelBuilder
	handler                               http.Handler
	builder                               *web.Builder
	dc                                    *stateful.DependencyCenter
	i18nBuilder                           *i18n.Builder
	logger                                *zap.Logger
	permissionBuilder                     *perm.Builder
	verifier                              *perm.Verifier
	layoutFunc                            func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc)
	detailLayoutFunc                      func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc)
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
	menuTopItems                          map[string]ComponentFunc
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
	wrapHandlers                          map[string]func(in http.Handler) (out http.Handler)
	plugins                               []Plugin
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
	OpenConfirmDialog = "presets_ConfirmDialog"
)

func New() *Builder {
	l, _ := zap.NewDevelopment()
	r := &Builder{
		logger:  l,
		builder: web.New(),
		dc:      stateful.NewDependencyCenter(),
		i18nBuilder: i18n.New().
			RegisterForModule(language.English, CoreI18nModuleKey, Messages_en_US).
			RegisterForModule(language.SimplifiedChinese, CoreI18nModuleKey, Messages_zh_CN).
			RegisterForModule(language.Japanese, CoreI18nModuleKey, Messages_ja_JP),
		writeFieldDefaults:   NewFieldDefaults(WRITE),
		listFieldDefaults:    NewFieldDefaults(LIST),
		detailFieldDefaults:  NewFieldDefaults(DETAIL),
		progressBarColor:     "amber",
		menuTopItems:         make(map[string]ComponentFunc),
		brandTitle:           "Admin",
		rightDrawerWidth:     "600",
		verifier:             perm.NewVerifier(PermModule, nil),
		homePageLayoutConfig: &LayoutConfig{},
		notFoundPageLayoutConfig: &LayoutConfig{
			NotificationCenterInvisible: true,
		},
		wrapHandlers: make(map[string]func(in http.Handler) (out http.Handler)),
	}
	r.GetWebBuilder().RegisterEventFunc(OpenConfirmDialog, r.openConfirmDialog)
	r.layoutFunc = r.defaultLayout
	r.detailLayoutFunc = r.defaultLayout
	stateful.Install(r.builder, r.dc)
	return r
}

func (b *Builder) GetDependencyCenter() *stateful.DependencyCenter {
	return b.dc
}

func (b *Builder) GetI18n() (r *i18n.Builder) {
	return b.i18nBuilder
}

func (b *Builder) I18n(v *i18n.Builder) (r *Builder) {
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

func (b *Builder) GetLayoutFunc() func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc) {
	return b.layoutFunc
}

func (b *Builder) DetailLayoutFunc(v func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc)) (r *Builder) {
	b.detailLayoutFunc = v
	return b
}

func (b *Builder) GetDetailLayoutFunc() func(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc) {
	return b.detailLayoutFunc
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

func (b *Builder) GetProfileFunc() ComponentFunc {
	return b.profileFunc
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

func (b *Builder) GetBrandTitle() string {
	return b.brandTitle
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

func (b *Builder) GetProgressBarColor() string {
	return b.progressBarColor
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
	{re: regexp.MustCompile(`\busers?|members?\b`), icon: "mdi-account"},
	// store
	{re: regexp.MustCompile(`\bstores?\b`), icon: "mdi-store"},
	// order
	{re: regexp.MustCompile(`\borders?\b`), icon: "mdi-cart"},
	// product
	{re: regexp.MustCompile(`\bproducts?\b`), icon: "mdi-format-list-bulleted"},
	// post
	{re: regexp.MustCompile(`\bposts?|articles?\b`), icon: "mdi-note"},
	// web
	{re: regexp.MustCompile(`\bweb|site\b`), icon: "mdi-web"},
	// seo
	{re: regexp.MustCompile(`\bseo\b`), icon: "mdi-search-web"},
	// i18n
	{re: regexp.MustCompile(`\bi18n|translations?\b`), icon: "mdi-translate"},
	// chart
	{re: regexp.MustCompile(`\banalytics?|charts?|statistics?\b`), icon: "mdi-google-analytics"},
	// dashboard
	{re: regexp.MustCompile(`\bdashboard\b`), icon: "mdi-view-dashboard"},
	// setting
	{re: regexp.MustCompile(`\bsettings?\b`), icon: "mdi-cog"},
}

func defaultMenuIcon(mLabel string) string {
	ws := strings.Join(strings.Split(strcase.ToSnake(mLabel), "_"), " ")
	for _, v := range defaultMenuIconREs {
		if v.re.MatchString(ws) {
			return v.icon
		}
	}

	return "mdi-alert-octagon-outline"
}

const (
	menuFontWeight    = "500"
	subMenuFontWeight = "400"
)

func (m *ModelBuilder) DefaultMenuItem(
	customizeChildren func(evCtx *web.EventContext, isSub bool, menuIcon string, children ...h.HTMLComponent) ([]h.HTMLComponent, error),
) func(evCtx *web.EventContext, isSub bool) (h.HTMLComponent, error) {
	return func(evCtx *web.EventContext, isSub bool) (h.HTMLComponent, error) {
		menuIcon := m.menuIcon
		// fontWeight := subMenuFontWeight
		if isSub {
			// menuIcon = ""
		} else {
			// fontWeight = menuFontWeight
			if menuIcon == "" {
				menuIcon = defaultMenuIcon(m.label)
			}
		}
		href := m.Info().ListingHref()
		if m.link != "" {
			href = m.link
		}
		if m.defaultURLQueryFunc != nil {
			href = fmt.Sprintf("%s?%s", href, m.defaultURLQueryFunc(evCtx.R).Encode())
		}

		children := []h.HTMLComponent{
			h.Iff(menuIcon != "", func() h.HTMLComponent {
				return web.Slot(VIcon(menuIcon)).Name(VSlotPrepend)
			}),
			VListItemTitle(
				h.Text(m.Info().LabelName(evCtx, false)),
			),
		}
		if customizeChildren != nil {
			var err error
			children, err = customizeChildren(evCtx, isSub, menuIcon, children...)
			if err != nil {
				return nil, err
			}
		}
		item := VListItem().Rounded(true).Value(m.label).Children(children...)

		item.Href(href)
		if strings.HasPrefix(href, "/") {
			funcStr := fmt.Sprintf(`(e) => {
	if (e.metaKey || e.ctrlKey) { return; }
	e.stopPropagation();
	e.preventDefault();
	%s;
}
`, web.Plaid().PushStateURL(href).Go())
			item.Attr("@click", funcStr)
		}
		// if b.isMenuItemActive(ctx, m) {
		//	item = item.Class("v-list-item--active text-primary")
		// }
		return item, nil
	}
}

func (b *Builder) isMenuItemActive(ctx *web.EventContext, m *ModelBuilder) bool {
	href := m.Info().ListingHref()
	if m.link != "" {
		href = m.link
	}
	path := strings.TrimSuffix(ctx.R.URL.Path, "/")
	if path == "" && href == "/" {
		return true
	}
	if path == href {
		return true
	}
	if href == b.prefix {
		return false
	}
	if href != "/" && strings.HasPrefix(path, href) {
		return true
	}

	return false
}

func (b *Builder) CreateMenus(ctx *web.EventContext) (r h.HTMLComponent) {
	mMap := make(map[string]*ModelBuilder)
	for _, m := range b.models {
		mMap[m.uriName] = m
	}
	var (
		activeMenuItem string
		selection      string
	)
	inOrderMap := make(map[string]struct{})
	var menus []h.HTMLComponent
	for _, om := range b.menuOrder {
		switch v := om.(type) {
		case *MenuGroupBuilder:
			disabled := false
			if b.verifier.Do(PermList).SnakeOn("mg_"+v.name).WithReq(ctx.R).IsAllowed() != nil {
				disabled = true
			}
			groupIcon := v.icon
			if groupIcon == "" {
				groupIcon = defaultMenuIcon(v.name)
			}
			subMenus := []h.HTMLComponent{
				h.Template(
					VListItem(
						web.Slot(
							VIcon(groupIcon),
						).Name("prepend"),
						VListItemTitle().Attr("style", fmt.Sprintf("white-space: normal; font-weight: %s;font-size: 14px;", menuFontWeight)),
						// VListItemTitle(h.Text(i18n.T(ctx.R, ModelsI18nModuleKey, v.name))).
					).Attr("v-bind", "props").
						Title(i18n.T(ctx.R, ModelsI18nModuleKey, v.name)).
						Class("rounded-lg"),
					// Value(i18n.T(ctx.R, ModelsI18nModuleKey, v.name)),
				).Attr("v-slot:activator", "{ props }"),
			}
			subCount := 0
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
				if m.Info().Verifier().Do(PermList).WithReq(ctx.R).IsAllowed() != nil {
					continue
				}
				menuItem, err := m.menuItem(ctx, true)
				if err != nil {
					panic(err)
				}
				subMenus = append(subMenus, menuItem)
				subCount++
				inOrderMap[m.uriName] = struct{}{}
				if b.isMenuItemActive(ctx, m) {
					// activeMenuItem = m.label
					activeMenuItem = v.name
					selection = m.label
				}
			}
			if subCount == 0 {
				continue
			}
			if disabled {
				continue
			}

			menus = append(menus,
				VListGroup(subMenus...).Value(v.name),
			)
		case string:
			m, ok := mMap[v]
			if !ok {
				m = mMap[inflection.Plural(strcase.ToKebab(v))]
			}
			if m == nil {
				continue
			}
			if m.Info().Verifier().Do(PermList).WithReq(ctx.R).IsAllowed() != nil {
				continue
			}

			if m.notInMenu {
				continue
			}
			menuItem, err := m.menuItem(ctx, false)
			if err != nil {
				panic(err)
			}
			menus = append(menus, menuItem)
			inOrderMap[m.uriName] = struct{}{}

			if b.isMenuItemActive(ctx, m) {
				selection = m.label
			}
		}
	}

	for _, m := range b.models {
		_, ok := inOrderMap[m.uriName]
		if ok {
			continue
		}

		if m.Info().Verifier().Do(PermList).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		if m.notInMenu {
			continue
		}

		if b.isMenuItemActive(ctx, m) {
			selection = m.label
		}
		menuItem, err := m.menuItem(ctx, false)
		if err != nil {
			panic(err)
		}
		menus = append(menus, menuItem)
	}

	r = h.Div(
		web.Scope(
			VList(menus...).
				OpenStrategy("single").
				Class("primary--text").
				Density(DensityCompact).
				Attr("v-model:opened", "locals.menuOpened").
				Attr("v-model:selected", "locals.selection").
				Attr("color", "transparent"),
			// .Attr("v-model:selected", h.JSONString([]string{"Pages"})),
		).VSlot("{ locals }").Init(
			fmt.Sprintf(`{ menuOpened:  ["%s"]}`, activeMenuItem),
			fmt.Sprintf(`{ selection:  ["%s"]}`, selection),
		))
	return
}

func (b *Builder) RunBrandFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.brandFunc != nil {
		return b.brandFunc(ctx)
	}

	return h.H1(i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle)).Class("text-h6")
}

func (b *Builder) RunSwitchLanguageFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.switchLanguageFunc != nil {
		return b.switchLanguageFunc(ctx)
	}

	supportLanguages := b.GetI18n().GetSupportLanguagesFromRequest(ctx.R)

	if len(b.GetI18n().GetSupportLanguages()) <= 1 || len(supportLanguages) == 0 {
		return nil
	}
	queryName := b.GetI18n().GetQueryName()
	msgr := MustGetMessages(ctx.R)
	if len(supportLanguages) == 1 {
		return h.Template().Children(
			h.Div(
				VList(
					VListItem(
						web.Slot(
							VIcon("mdi-widget-translate").Size(SizeSmall).Class("mr-4 ml-1"),
						).Name("prepend"),
						VListItemTitle(
							h.Div(h.Text(fmt.Sprintf("%s%s %s", msgr.Language, msgr.Colon, display.Self.Name(supportLanguages[0])))).Role("button"),
						),
					).Class("pa-0").Density(DensityCompact),
				).Class("pa-0 ma-n4 mt-n6"),
			).Attr("@click", web.Plaid().MergeQuery(true).Query(queryName, supportLanguages[0].String()).Go()),
		)
	}
	languageIcon := EnLanguageIcon
	lang := ctx.R.FormValue(queryName)
	if lang == "" {
		lang = b.i18nBuilder.GetCurrentLangFromCookie(ctx.R)
	}
	switch lang {
	case language.SimplifiedChinese.String():
		languageIcon = ZhLanguageIcon
	case language.Japanese.String():
		languageIcon = JPIcon
	}
	var languages []h.HTMLComponent
	for _, tag := range supportLanguages {
		languages = append(languages,
			h.Div(
				VListItem(
					VListItemTitle(
						h.Div(h.Text(display.Self.Name(tag))),
					),
				).Attr("@click", web.Plaid().MergeQuery(true).Query(queryName, tag.String()).Go()),
			),
		)
	}

	return VMenu().Children(
		h.Template().Attr("v-slot:activator", "{isActive, props}").Children(
			h.Div(
				VBtn("").Children(
					h.RawHTML(languageIcon),
					// VIcon("mdi-menu-down"),
				).Attr("variant", "text").
					Attr("icon", "").
					Class("i18n-switcher-btn"),
			).Attr("v-bind", "props").Style("display: inline-block;"),
		),
		VList(
			languages...,
		).Density(DensityCompact),
	)
}

func (b *Builder) AddMenuTopItemFunc(key string, v ComponentFunc) (r *Builder) {
	b.menuTopItems[key] = v
	return b
}

func (b *Builder) RunBrandProfileSwitchLanguageDisplayFunc(brand, profile, switchLanguage h.HTMLComponent, ctx *web.EventContext) (r h.HTMLComponent) {
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
			).Density(DensityCompact),
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

const (
	RightDrawerPortalName          = "presets_RightDrawerPortalName"
	RightDrawerContentPortalName   = "presets_RightDrawerContentPortalName"
	DialogPortalName               = "presets_DialogPortalName"
	dialogContentPortalName        = "presets_DialogContentPortalName"
	NotificationCenterPortalName   = "notification-center"
	DefaultConfirmDialogPortalName = "presets_confirmDialogPortalName"
	ListingDialogPortalName        = "presets_listingDialogPortalName"
	singletonEditingPortalName     = "presets_SingletonEditingPortalName"
	DeleteConfirmPortalName        = "deleteConfirm"
)

const (
	CloseRightDrawerVarScript   = "vars.presetsRightDrawer = false"
	CloseDialogVarScript        = "vars.presetsDialog = false"
	CloseListingDialogVarScript = "vars.presetsListingDialog = false"
)

func (b *Builder) overlay(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, width string) {
	overlayType := ctx.Param(ParamOverlay)
	if overlayType == actions.Dialog {
		b.dialog(r, comp, width)
		return
	} else if overlayType == actions.Content {
		b.contentDrawer(ctx, r, comp, width)
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
			web.Portal(comp).Name(RightDrawerContentPortalName),
		).
			// Attr("@input", "plaidForm.dirty && vars.presetsRightDrawer == false && !confirm('You have unsaved changes on this form. If you close it, you will lose all unsaved changes. Are you sure you want to close it?') ? vars.presetsRightDrawer = true: vars.presetsRightDrawer = $event"). // remove because drawer plaidForm has to be reset when UpdateOverlayContent
			Class("v-navigation-drawer--temporary").
			Attr("v-model", "vars.presetsRightDrawer").
			Location(LocationRight).
			Temporary(true).
			Persistent(true).
			// Fixed(true).
			Width(width).
			Attr(":height", `"100%"`),
		// Temporary(true),
		// HideOverlay(true).
		// Floating(true).

	})
	r.RunScript = "setTimeout(function(){ vars.presetsRightDrawer = true }, 100)"
}

func (b *Builder) contentDrawer(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, width string) {
	if width == "" {
		width = b.rightDrawerWidth
	}
	portalName := ctx.Param(ParamPortalName)
	p := RightDrawerContentPortalName
	if portalName != "" {
		p = portalName
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: p,
		Body: comp,
	})
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
		).VSlot("{ form }"),
	})
	r.RunScript = "setTimeout(function(){ vars.presetsDialog = true }, 100)"
}

type LayoutConfig struct {
	NotificationCenterInvisible bool
}

func (b *Builder) notificationCenter(ctx *web.EventContext) (er web.EventResponse, err error) {
	total := b.notificationCountFunc(ctx)
	content := b.notificationContentFunc(ctx)
	icon := VIcon("mdi-bell-outline").Size(20).Color("grey-darken-1")
	er.Body = VMenu().Children(
		h.Template().Attr("v-slot:activator", "{ props }").Children(
			VBtn("").Icon(true).Children(
				h.If(total > 0,
					VBadge(
						icon,
					).Content(total).Floating(true).Color("red"),
				).Else(icon),
			).Attr("v-bind", "props").
				Density(DensityCompact).
				Variant(VariantText),
			// .Class("ml-1")
		),
		VCard(content),
	)
	return
}

const (
	ConfirmDialogConfirmEvent     = "presets_ConfirmDialog_ConfirmEvent"
	ConfirmDialogPromptText       = "presets_ConfirmDialog_PromptText"
	ConfirmDialogOKText           = "presets_ConfirmDialog_OKText"
	ConfirmDialogCancelText       = "presets_ConfirmDialog_CancelText"
	ConfirmDialogDialogPortalName = "presets_ConfirmDialog_DialogPortalName"
)

func (b *Builder) openConfirmDialog(ctx *web.EventContext) (er web.EventResponse, err error) {
	confirmEvent := ctx.R.FormValue(ConfirmDialogConfirmEvent)
	if confirmEvent == "" {
		ShowMessage(&er, "confirm event is empty", "error")
		return
	}

	msgr := MustGetMessages(ctx.R)
	promptText := msgr.ConfirmDialogPromptText
	if v := ctx.R.FormValue(ConfirmDialogPromptText); v != "" {
		promptText = v
	}
	okText := msgr.OK
	if v := ctx.R.FormValue(ConfirmDialogOKText); v != "" {
		okText = v
	}
	cancelText := msgr.Cancel
	if v := ctx.R.FormValue(ConfirmDialogCancelText); v != "" {
		cancelText = v
	}

	portal := DefaultConfirmDialogPortalName
	if v := ctx.R.FormValue(ConfirmDialogDialogPortalName); v != "" {
		portal = v
	}

	er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
		Name: portal,
		Body: web.Scope(VDialog(
			VCard(
				VCardTitle(VIcon("warning").Class("red--text mr-4"), h.Text(promptText)),
				VCardActions(
					VSpacer(),
					VBtn(cancelText).
						Variant(VariantFlat).
						Class("ml-2").
						On("click", "locals.show = false"),

					VBtn(okText).
						Color("primary").
						Variant(VariantFlat).
						Theme(ThemeDark).
						Attr("@click", fmt.Sprintf("%s; locals.show = false", confirmEvent)),
				),
			),
		).MaxWidth("600px").
			Attr("v-model", "locals.show"),
		).VSlot("{ locals }").Init("{show: true}"),
	})

	return
}

func (b *Builder) defaultLayout(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		b.InjectAssets(ctx)

		// call CreateMenus before in(ctx) to fill the menuGroupName for modelBuilders first
		menu := b.CreateMenus(ctx)

		toolbar := VContainer(
			VRow(
				VCol(b.RunBrandFunc(ctx)).Cols(7),
				VCol(
					b.RunSwitchLanguageFunc(ctx),
					// VBtn("").Children(
					//	languageSwitchIcon,
					//	VIcon("mdi-menu-down"),
					// ).Attr("variant", "plain").
					//	Attr("icon", ""),
				).Cols(3).Class("text-right"),
				VDivider().Attr("vertical", true).Class("i18n-divider"),
				VCol(
					VAppBarNavIcon().Attr("icon", "mdi-menu").
						Class("text-grey-darken-1 menu-control-icon").
						Attr("@click", "vars.navDrawer = !vars.navDrawer").Density(DensityCompact),
				).Cols(2).Class("position-relative"),
			).Attr("align", "center").Attr("justify", "center"),
		)

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if errors.Is(err, perm.PermissionDenied) {
			pr.Body = h.Text(perm.PermissionDenied.Error())
			return pr, nil
		}
		if err != nil {
			panic(err)
		}

		var profile h.HTMLComponent
		if b.profileFunc != nil {
			profile = VAppBar(
				b.profileFunc(ctx),
			).Location("bottom").Class("border-t-sm border-b-0").Elevation(0)
		}

		// showNotificationCenter := cfg == nil || !cfg.NotificationCenterInvisible
		// var notifier h.HTMLComponent
		// if b.notificationCountFunc != nil && b.notificationContentFunc != nil {
		//	notifier = web.Portal().Name(NotificationCenterPortalName).Loader(web.GET().EventFunc(actions.NotificationCenter))
		// }
		// ctx.R = ctx.R.WithContext(context.WithValue(ctx.R.Context(), ctxNotifyCenter, notifier))

		// showSearchBox := cfg == nil || !cfg.SearchBoxInvisible

		// _ := i18n.MustGetModuleMessages(ctx.R, CoreI18nModuleKey, Messages_en_US).(*Messages)

		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle))
		var pageTitleComp h.HTMLComponent
		innerPageTitleCompo, ok := ctx.ContextValue(CtxPageTitleComponent).(h.HTMLComponent)
		if !ok {
			innerPageTitleCompo = VToolbarTitle(innerPr.PageTitle) // Class("text-h6 font-weight-regular"),
		} else {
			ctx.WithContextValue(CtxPageTitleComponent, nil)
		}
		actionsComponentTeleportToID := GetActionsComponentTeleportToID(ctx)
		pageTitleComp = h.Div(
			VAppBarNavIcon().
				Density("compact").
				Class("mr-2").
				Attr("v-if", "!vars.navDrawer").
				On("click.stop", "vars.navDrawer = !vars.navDrawer"),
			innerPageTitleCompo,
			VSpacer(),
			h.Iff(actionsComponentTeleportToID != "", func() h.HTMLComponent {
				return h.Div().Id(actionsComponentTeleportToID)
			}),
		).Class("d-flex align-center mx-6 border-b w-100").Style("padding-bottom:24px")
		pr.Body = VCard(
			h.Template(
				VSnackbar(h.Text("{{vars.presetsMessage.message}}")).
					Attr("v-model", "vars.presetsMessage.show").
					Attr(":color", "vars.presetsMessage.color").
					Timeout(2000).
					Location(LocationTop),
			).Attr("v-if", "vars.presetsMessage"),
			VLayout(
				web.Portal().Name(RightDrawerPortalName),

				// App(true).
				// Fixed(true),
				// ClippedLeft(true),
				web.Portal().Name(DialogPortalName),
				web.Portal().Name(DeleteConfirmPortalName),
				web.Portal().Name(DefaultConfirmDialogPortalName),
				web.Portal().Name(ListingDialogPortalName),

				VNavigationDrawer(
					// b.RunBrandProfileSwitchLanguageDisplayFunc(b.RunBrandFunc(ctx), profile, b.RunSwitchLanguageFunc(ctx), ctx),
					// b.RunBrandFunc(ctx),
					// profile,
					VLayout(
						VMain(
							toolbar,
							VCard(
								menu,
							).Class("menu-content mt-2 mb-4 ml-4 pr-4").Variant(VariantText),
						).Class("menu-wrap"),
						// VDivider(),
						profile,
					).Class("ma-2 border-sm rounded elevation-0").Attr("style",
						"height: calc(100% - 16px);"),
					// ).Class("ma-2").
					// 	Style("height: calc(100% - 20px); border: 1px solid grey"),
				).
					// 256px is directly The measured size in figma
					// in actual use, need plus 8px for padding left and right
					// plus border 2px
					Width(256+8+8+2).
					// App(true).
					// Clipped(true).
					// Fixed(true).
					Attr("v-model", "vars.navDrawer").
					// Attr("style", "border-right: 1px solid grey ").
					Permanent(true).
					Floating(true).
					Elevation(0),

				VMain(
					VProgressLinear().
						Attr(":active", "isFetching").
						Class("ml-4").
						Attr("style", "position: fixed; z-index: 99;").
						Indeterminate(true).
						Height(2).
						Color(b.progressBarColor),
					VAppBar(
						pageTitleComp,
					).Elevation(0).Attr("height", 100),
					innerPr.Body,
				).
					Class("overflow-y-auto main-container").
					Attr("style", "height:100vh; padding-left: calc(var(--v-layout-left) + 16px); --v-layout-right: 16px"),
			),
		).Attr("id", "vt-app").Elevation(0).
			Attr(web.VAssign("vars", `{presetsRightDrawer: false, presetsDialog: false, presetsListingDialog: false, 
navDrawer: true
}`)...)

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
			web.Portal().Name(DefaultConfirmDialogPortalName),

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
					Location(LocationTop),
			).Attr("v-if", "vars.presetsMessage"),
			VMain(
				innerPr.Body.(h.HTMLComponent),
			),
		).
			Attr("id", "vt-app").
			Attr(web.VAssign("vars", `{presetsDialog: false, presetsMessage: {show: false, color: "success", 
message: ""}}`)...)

		return
	}
}

func (b *Builder) InjectAssets(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(strings.Replace(`
			<link rel="stylesheet" href="{{prefix}}/vuetify/assets/index.css" async>
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

	ctx.Injector.TailHTML(strings.Replace(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix, -1))

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

func (b *Builder) DefaultNotFoundPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	msgr := MustGetMessages(ctx.R)
	r.Body = h.Div(
		h.H1("404").Class("mb-2"),
		h.Text(msgr.NotFoundPageNotice),
	).Class("text-center mt-8")
	return
}

func (b *Builder) getNotFoundPageFunc() web.PageFunc {
	pf := b.DefaultNotFoundPageFunc
	if b.notFoundFunc != nil {
		pf = b.notFoundFunc
	}
	return pf
}

func (b *Builder) extraFullPath(ea *extraAsset) string {
	return b.prefix + "/extra" + ea.path
}

func (b *Builder) initMux() {
	b.logger.Info("initializing mux for", zap.Reflect("models", modelNames(b.models)), zap.String("prefix", b.prefix))
	mux := http.NewServeMux()
	ub := b.builder

	mainJSPath := b.prefix + "/assets/main.js"
	mux.Handle("GET "+mainJSPath,
		ub.PacksHandler("text/javascript",
			Vuetify(),
			JSComponentsPack(),
			vuetifyx.JSComponentsPack(),
			web.JSComponentsPack(),
		),
	)
	log.Println("mounted url:", mainJSPath)

	vueJSPath := b.prefix + "/assets/vue.js"
	mux.Handle("GET "+vueJSPath,
		ub.PacksHandler("text/javascript",
			web.JSVueComponentsPack(),
		),
	)

	HandleMaterialDesignIcons(b.prefix, mux)

	log.Println("mounted url:", vueJSPath)

	for _, ea := range b.extraAssets {
		fullPath := b.extraFullPath(ea)
		mux.Handle("GET "+fullPath, ub.PacksHandler(
			ea.contentType,
			ea.body,
		))
		log.Println("mounted url:", fullPath)
	}

	homeURL := b.prefix
	if homeURL == "" {
		homeURL = "/{$}"
	}
	mux.Handle(
		homeURL,
		b.wrap(nil, b.layoutFunc(b.getHomePageFunc(), b.homePageLayoutConfig)),
	)

	for _, m := range b.models {
		m.listing.setup()

		pluralUri := inflection.Plural(m.uriName)
		info := m.Info()
		routePath := info.ListingHref()
		inPageFunc := m.listing.GetPageFunc()
		if m.singleton {
			inPageFunc = m.editing.singletonPageFunc
			if m.layoutConfig == nil {
				m.layoutConfig = &LayoutConfig{}
			}
		}
		mux.Handle(
			routePath,
			b.wrap(m, b.layoutFunc(inPageFunc, m.layoutConfig)),
		)
		log.Printf("mounted url: %s\n", routePath)
		if m.hasDetailing {
			routePath = fmt.Sprintf("%s/%s/{id}", b.prefix, pluralUri)
			mux.Handle(
				routePath,
				b.wrap(m, b.detailLayoutFunc(m.detailing.GetPageFunc(), m.layoutConfig)),
			)
			log.Printf("mounted url: %s", routePath)
		}
	}

	// b.handler = mux
	// Handle 404
	b.handler = b.notFound(mux)
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	if code == http.StatusNotFound {
		// default 404 will use http.Error to set Content-Type to text/plain,
		// So we have to set it to html before WriteHeader
		rw.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	// don't write content, because we use customized page body
	if rw.statusCode == http.StatusNotFound {
		return 0, nil
	}
	return rw.ResponseWriter.Write(b)
}

func (b *Builder) notFound(handler http.Handler) http.Handler {
	notFoundHandler := b.wrap(
		nil,
		b.layoutFunc(b.getNotFoundPageFunc(), b.notFoundPageLayoutConfig),
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedResponse := &responseWriterWrapper{w, http.StatusOK}
		handler.ServeHTTP(capturedResponse, r)
		if capturedResponse.statusCode == http.StatusNotFound {
			// If no other handler wrote to the response, assume 404 and write our custom response.
			notFoundHandler.ServeHTTP(w, r)
		}
		return
	})
}

func (b *Builder) AddWrapHandler(key string, f func(in http.Handler) (out http.Handler)) {
	b.wrapHandlers[key] = f
}

func (b *Builder) wrap(m *ModelBuilder, pf web.PageFunc) http.Handler {
	p := b.builder.Page(pf)
	if m != nil {
		m.registerDefaultEventFuncs()
		p.MergeHub(&m.EventsHub)
	}
	p.WrapEventFunc(func(in web.EventFunc) web.EventFunc {
		return func(ctx *web.EventContext) (r web.EventResponse, err error) {
			r, err = in(ctx)
			if err != nil {
				return
			}
			wrapper, ok := getEventFuncAddonWrapper(ctx)
			if !ok {
				return
			}
			err = wrapper(func(ctx *web.EventContext, r *web.EventResponse) (err error) {
				return nil
			})(ctx, &r)
			return
		}
	})

	handlers := b.GetI18n().EnsureLanguage(
		p,
	)
	for _, wrapHandler := range b.wrapHandlers {
		handlers = wrapHandler(handlers)
	}

	return handlers
}

func (b *Builder) Build() {
	mns := modelNames(b.models)
	if len(slices.Compact(mns)) != len(mns) {
		panic(fmt.Sprintf("Duplicated model names registered %v", mns))
	}
	b.initMux()
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.handler == nil {
		b.Build()
	}
	redirectSlashes(b.handler).ServeHTTP(w, r)
}

func redirectSlashes(next http.Handler) http.Handler {
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
