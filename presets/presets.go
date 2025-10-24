package presets

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/hook"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"go.uber.org/zap"
	"golang.org/x/text/language"

	"github.com/qor5/admin/v3/presets/actions"
)

type Builder struct {
	containerClassName                    string
	prefix                                string
	models                                []*ModelBuilder
	handler                               http.Handler
	warmupOnce                            sync.Once
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
	switchLocaleFunc                      ComponentFunc
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
	menuOrder                             *MenuOrderBuilder
	wrapHandlers                          map[string]func(in http.Handler) (out http.Handler)
	handlerHook                           hook.Hook[http.Handler]
	plugins                               []Plugin
	notFoundHandler                       http.Handler
	customBuilders                        []*CustomBuilder
	toolbarFunc                           func(ctx *web.EventContext) h.HTMLComponent
}

type AssetFunc func(ctx *web.EventContext)
type (
	BreadcrumbItemsFuncKey struct{}
	BreadcrumbItemsFunc    func(ctx *web.EventContext, disableLast bool) (r []h.HTMLComponent)
)

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

var staticFileRe = regexp.MustCompile(`\.(css|js|gif|jpg|jpeg|png|ico|svg|ttf|eot|woff|woff2|js\.map)$`)

func New() *Builder {
	l, _ := zap.NewDevelopment()
	b := &Builder{
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
	b.menuOrder = NewMenuOrderBuilder(b)
	b.GetWebBuilder().RegisterEventFunc(OpenConfirmDialog, b.openConfirmDialog)
	b.layoutFunc = b.defaultLayout
	b.detailLayoutFunc = b.defaultLayout
	b.notFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if staticFileRe.MatchString(strings.ToLower(r.URL.Path)) {
			http.NotFound(w, r)
			return
		}
		b.wrap(nil, b.layoutFunc(b.getNotFoundPageFunc(), b.notFoundPageLayoutConfig)).ServeHTTP(w, r)
	})
	b.toolbarFunc = b.defaultToolBar

	stateful.Install(b.builder, b.dc)
	return b
}

func (b *Builder) ContainerClass(name string) (r *Builder) {
	b.containerClassName = name
	return b
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

func (b *Builder) GetVerifier() (r *perm.Verifier) {
	return b.verifier
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

func (b *Builder) SwitchLocaleFunc(v ComponentFunc) (r *Builder) {
	b.switchLocaleFunc = v
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

func (b *Builder) ToolBarFunc(f ComponentFunc) (r *Builder) {
	b.toolbarFunc = f
	return b
}

func (b *Builder) WarpToolBarFunc(w func(ComponentFunc) ComponentFunc) (r *Builder) {
	b.toolbarFunc = w(b.toolbarFunc)
	return b
}

func (b *Builder) ExtraAsset(path string, contentType string, body web.ComponentsPack, refTag ...string) (r *Builder) {
	path = strings.TrimLeft(path, "/")
	path = "/" + path

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

func (b *Builder) HandleCustomPage(pattern string, cb *CustomBuilder) *Builder {
	cb.pattern = pattern
	b.customBuilders = append(b.customBuilders, cb)
	return b
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
	if !b.menuOrder.isMenuGroupInOrder(mgb) {
		b.menuOrder.Append(mgb)
	}
	return mgb
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
	b.menuOrder.Append(items...)
}

type defaultMenuIconRE struct {
	re   *regexp.Regexp
	icon string
}

var defaultMenuIconREs = []defaultMenuIconRE{
	// announce
	{re: regexp.MustCompile(`\bannouncements?\b`), icon: "mdi-bullhorn-variant-outline"},
	// Campaign
	{re: regexp.MustCompile(`\bcampaigns?\b`), icon: "mdi-clock-time-three-outline"},
	// categories
	{re: regexp.MustCompile(`\bcategories?\b`), icon: "mdi-shape-plus"},
	// coupon
	{re: regexp.MustCompile(`\bcoupons?\b`), icon: "mdi-cash-multiple"},
	// user
	{re: regexp.MustCompile(`\busers?|members?\b`), icon: "mdi-account"},
	// store
	{re: regexp.MustCompile(`\bstores?\b`), icon: "mdi-store"},
	// order
	{re: regexp.MustCompile(`\borders?\b`), icon: "mdi-cart"},
	// featured product
	{re: regexp.MustCompile(`\bfeatured?\b`), icon: "mdi-creation"},
	// product
	{re: regexp.MustCompile(`\bproducts?\b`), icon: "mdi-format-list-bulleted"},
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
	{re: regexp.MustCompile(`\bsettings?|config?\b`), icon: "mdi-cog"},
	// email
	{re: regexp.MustCompile(`\bemail?\b`), icon: "mdi-email-outline"},
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
		} else if menuIcon == "" {
			// fontWeight = menuFontWeight
			menuIcon = defaultMenuIcon(m.label)
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

func (b *Builder) RunBrandFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.brandFunc != nil {
		return b.brandFunc(ctx)
	}

	return h.H1(i18n.T(ctx.R, ModelsI18nModuleKey, b.brandTitle)).Class("text-h6")
}

func (b *Builder) AddMenuTopItemFunc(key string, v ComponentFunc) (r *Builder) {
	b.menuTopItems[key] = v
	return b
}

func (b *Builder) MenuComponentFunc(fn func(menus []h.HTMLComponent, menuGroupSelected, menuItemSelected string) h.HTMLComponent) (r *Builder) {
	b.menuOrder.MenuComponentFunc(fn)
	return b
}

func (b *Builder) RunSwitchLocalCodeFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	if b.switchLocaleFunc != nil {
		return b.switchLocaleFunc(ctx)
	}
	return nil
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

var CloseRightDrawerVarConfirmScript = ConfirmLeaveScript("vars.confirmDrawerLeave=true;", "vars.presetsRightDrawer = false;")

const (
	CloseRightDrawerVarScript   = "vars.presetsRightDrawer = false"
	CloseDialogVarScript        = "vars.presetsDialog = false"
	CloseListingDialogVarScript = "vars.presetsListingDialog = false"
)

func ConfirmLeaveScript(confirmEvent, leaveEvent string) string {
	return fmt.Sprintf("if(Object.values(vars.%s).some(value => value === true)){%s}else{%s};", VarsPresetsDataChanged, confirmEvent, leaveEvent)
}

func (b *Builder) overlay(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, width string) {
	overlayType := ctx.Param(ParamOverlay)
	if overlayType == actions.Dialog {
		b.dialog(ctx, r, comp, width)
		return
	} else if overlayType == actions.Content {
		b.contentDrawer(ctx, r, comp, width)
		return
	}
	b.rightDrawer(ctx, r, comp, width)
}

type (
	IdCurrentActiveProcessor       func(ctx *web.EventContext, current string) (string, error)
	ctxKeyIdCurrentActiveProcessor struct{}
)

func newActiveWatcher(ctx *web.EventContext, varToWatch string) (h.HTMLComponent, error) {
	varCurrentActive := ctx.R.FormValue(ParamVarCurrentActive)
	idCurrentActive := ctx.R.FormValue(ParamID)
	idCurrentActiveProcessor, ok := ctx.R.Context().Value(ctxKeyIdCurrentActiveProcessor{}).(IdCurrentActiveProcessor)
	if ok {
		var err error
		idCurrentActive, err = idCurrentActiveProcessor(ctx, idCurrentActive)
		if err != nil {
			return nil, err
		}
	}
	if varCurrentActive == "" || idCurrentActive == "" {
		return nil, nil
	}
	return h.Div().Style("display: none;").Attr("v-on-mounted", fmt.Sprintf(`({watch}) => {
			vars.%s = %q;
			watch(() => %s, (value) => {
				if (!value) {
					vars.%s = '';
				}
			})
		}`,
		varCurrentActive, idCurrentActive,
		varToWatch,
		varCurrentActive,
	)).Attr("v-on-unmounted", fmt.Sprintf(`() => {
		vars.%s = '';
	}`, varCurrentActive)), nil
}

func (b *Builder) rightDrawer(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, width string) {
	if width == "" {
		width = b.rightDrawerWidth
	}
	msgr := MustGetMessages(ctx.R)
	listenChangeEvent := fmt.Sprintf("if(!$event && Object.values(vars.%s).some(value => value === true)) {vars.presetsRightDrawer=true};", VarsPresetsDataChanged)

	activeWatcher, err := newActiveWatcher(ctx, "vars.presetsRightDrawer")
	if err != nil {
		panic(err)
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: RightDrawerPortalName,
		Body: VNavigationDrawer(
			vuetifyx.VXDialog().Persistent(true).
				Title(msgr.DialogTitleDefault).
				Text(msgr.LeaveBeforeUnsubmit).
				OkText(msgr.OK).
				CancelText(msgr.Cancel).
				Attr("@click:ok", "vars.confirmDrawerLeave=false;vars.presetsRightDrawer = false").
				Attr("v-model", "vars.confirmDrawerLeave"),
			activeWatcher,
			web.GlobalEvents().Attr("@keyup.esc", fmt.Sprintf(" if (!Object.values(vars.%s).some(value => value === true)) { vars.presetsRightDrawer = false} else {vars.confirmDrawerLeave=true};", VarsPresetsDataChanged)),
			web.Portal(comp).Name(RightDrawerContentPortalName),
		).
			// Attr("@input", "plaidForm.dirty && vars.presetsRightDrawer == false && !confirm('You have unsaved changes on this form. If you close it, you will lose all unsaved changes. Are you sure you want to close it?') ? vars.presetsRightDrawer = true: vars.presetsRightDrawer = $event"). // remove because drawer plaidForm has to be reset when UpdateOverlayContent
			Class("v-navigation-drawer--temporary").
			Attr("v-model", "vars.presetsRightDrawer").
			Attr("@update:model-value", listenChangeEvent).
			Location(LocationRight).
			Temporary(true).
			// Fixed(true).
			Width(width).
			Attr(":height", `"100%"`),
		// Temporary(true),
		// HideOverlay(true).
		// Floating(true).

	})
	r.RunScript = fmt.Sprintf(`setTimeout(function(){ vars.presetsRightDrawer = true,vars.confirmDrawerLeave=false,vars.%s = {} }, 100)`, VarsPresetsDataChanged)
}

func (b *Builder) contentDrawer(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, _ string) {
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

func (b *Builder) dialog(ctx *web.EventContext, r *web.EventResponse, comp h.HTMLComponent, width string) {
	if width == "" {
		width = b.rightDrawerWidth
	}

	activeWatcher, err := newActiveWatcher(ctx, "vars.presetsDialog")
	if err != nil {
		panic(err)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DialogPortalName,
		Body: web.Scope(
			activeWatcher,
			vuetifyx.VXDialog(
				h.Div().Class("overflow-y-auto").Children(
					web.Portal(comp).Name(dialogContentPortalName),
				),
			).
				ContentOnlyMode(true).
				ContentPadding("0 2px 9px").
				Attr("v-model", "vars.presetsDialog").
				Attr(":width", width),
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
	ConfirmDialogTitleText        = "presets_ConfirmDialog_TitleText"
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
	titleText := msgr.ConfirmDialogTitleText
	if v := ctx.R.FormValue(ConfirmDialogTitleText); v != "" {
		titleText = v
	}
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
		Body: web.Scope(
			vuetifyx.VXDialog().
				Size(vuetifyx.DialogSizeDefault).
				Title(titleText).
				Text(promptText).
				CancelText(cancelText).
				OkText(okText).
				Attr("@click:ok", fmt.Sprintf("%s; locals.show = false", confirmEvent)).
				Attr("v-model", "locals.show"),
		).VSlot("{ locals }").Init("{show: true}"),
	})

	return
}

func (b *Builder) defaultToolBar(ctx *web.EventContext) h.HTMLComponent {
	return VContainer(
		VRow(
			VCol(b.RunBrandFunc(ctx)).Cols(5),
			VCol(
				b.RunSwitchLocalCodeFunc(ctx),
				// VBtn("").Children(
				//	languageSwitchIcon,
				//	VIcon("mdi-menu-down"),
				// ).Attr("variant", "plain").
				//	Attr("icon", ""),
			).Cols(5).Class("py-0 d-flex justify-end pl-0 pr-2"),
			VDivider().Attr("vertical", true).Class("i18n-divider"),
			VCol(
				b.AppBarNav(),
			).Cols(2).Class("position-relative"),
		).Attr("align", "center").Attr("justify", "center"),
	)
}

func (b *Builder) AppBarNav() h.HTMLComponent {
	return VAppBarNavIcon().Attr("icon", "mdi-menu").
		Class("text-grey-darken-1 menu-control-icon").
		Attr("@click", "vars.navDrawer = !vars.navDrawer").Density(DensityCompact)
}

func (b *Builder) defaultLeftMenuComp(ctx *web.EventContext) h.HTMLComponent {
	// call CreateMenus before in(ctx) to fill the menuGroupName for modelBuilders first
	menu := b.menuOrder.CreateMenus(ctx)
	var profile h.HTMLComponent
	if b.profileFunc != nil {
		profile = VAppBar(
			b.profileFunc(ctx),
		).Location("bottom").Class("border-t-sm border-b-0").Elevation(0)
	}
	return VNavigationDrawer(
		// b.RunBrandProfileSwitchLanguageDisplayFunc(b.RunBrandFunc(ctx), profile, b.RunSwitchLanguageFunc(ctx), ctx),
		// b.RunBrandFunc(ctx),
		// profile,
		VLayout(
			VMain(
				b.toolbarFunc(ctx),
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
		Elevation(0)
}

func (b *Builder) defaultLayoutCompo(_ *web.EventContext, menu, body h.HTMLComponent) h.HTMLComponent {
	return VCard(
		VProgressLinear().
			Attr(":active", "vars.globalProgressBar.show").
			Attr(":model-value", "vars.globalProgressBar.value").
			Attr("style", "position: fixed; z-index: 2000;").
			Height(2).
			Color(b.progressBarColor),
		h.Template(
			VSnackbar(
				h.Div().Style("white-space: pre-wrap").Text("{{vars.presetsMessage.message}}"),
			).
				Attr("v-model", "vars.presetsMessage.show").
				Attr(":color", "vars.presetsMessage.color").
				Attr("style", "bottom: 48px;").
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
			menu,
			VMain(
				body,
			).
				Class("overflow-y-auto main-container").
				Attr("style", "height:100vh; padding-left: calc(var(--v-layout-left) + 16px); --v-layout-right: 16px"),
		),
	).Attr("id", "vt-app").Elevation(0).
		Attr(web.VAssign("vars", fmt.Sprintf(`{presetsRightDrawer: false, presetsDialog: false, presetsListingDialog: false, 
navDrawer: true,%s:{},presetsMessage: {show: false, color: "", message: ""}
}`, VarsPresetsDataChanged))...).Class(b.containerClassName)
}

func (b *Builder) defaultLayout(in web.PageFunc, cfg *LayoutConfig) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		b.InjectAssets(ctx)
		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if errors.Is(err, perm.PermissionDenied) {
			pr.Body = h.Text(perm.PermissionDenied.Error())
			return pr, nil
		}
		if err != nil {
			panic(err)
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
			breadcrumbFunc, ok := ctx.ContextValue(BreadcrumbItemsFuncKey{}).(BreadcrumbItemsFunc)
			if ok {
				innerPageTitleCompo = CreateVXBreadcrumbs(ctx, breadcrumbFunc)
			}
		} else {
			ctx.WithContextValue(CtxPageTitleComponent, nil)
		}
		afterTitleComponent, haveAfterTitleComponent := GetComponentFromContext(ctx, ctxDetailingAfterTitleComponent)
		actionsComponentTeleportToID := GetActionsComponentTeleportToID(ctx)
		pageTitleComp = h.Div(
			VAppBarNavIcon().
				Density("compact").
				Class("mr-2").
				Attr("v-if", "!vars.navDrawer").
				On("click.stop", "vars.navDrawer = !vars.navDrawer"),
			innerPageTitleCompo,
			h.If(haveAfterTitleComponent, afterTitleComponent),
			h.Iff(actionsComponentTeleportToID != "", func() h.HTMLComponent {
				return h.Components(
					VSpacer(),
					h.Div().Id(actionsComponentTeleportToID),
				)
			}),
		).Class("d-flex align-center mx-6 border-b w-100").Style("padding-bottom:24px")
		pr.Body = b.defaultLayoutCompo(ctx, b.defaultLeftMenuComp(ctx), h.Components(
			VAppBar(
				pageTitleComp,
			).Elevation(0).Attr("height", 100),
			innerPr.Body,
		),
		)
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
				VSnackbar(
					h.Div().Style("white-space: pre-wrap").Text("{{vars.presetsMessage.message}}"),
				).
					Attr("v-model", "vars.presetsMessage.show").
					Attr(":color", "vars.presetsMessage.color").
					Timeout(2000).
					Location(LocationTop),
			).Attr("v-if", "vars.presetsMessage"),
			VMain(
				innerPr.Body,
			),
		).
			Attr("id", "vt-app").
			Attr(web.VAssign("vars", `{presetsDialog: false, presetsMessage: {show: false, color: "success", 
message: ""}}`)...)

		return
	}
}

func (b *Builder) InjectAssets(ctx *web.EventContext) {
	ctx.Injector.HeadHTML(strings.ReplaceAll(`
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
		`, "{{prefix}}", b.prefix))

	b.InjectExtraAssets(ctx)

	ctx.Injector.TailHTML(strings.ReplaceAll(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix))

	if b.assetFunc != nil {
		b.assetFunc(ctx)
	}
}

func (b *Builder) InjectExtraAssets(ctx *web.EventContext) {
	for _, ea := range b.extraAssets {
		if ea.refTag != "" {
			ctx.Injector.HeadHTML(ea.refTag)
			continue
		}

		if strings.HasSuffix(ea.path, "css") {
			ctx.Injector.HeadHTML(fmt.Sprintf("<link rel=\"stylesheet\" href=%q>", b.extraFullPath(ea)))
			continue
		}

		ctx.Injector.HeadHTML(fmt.Sprintf("<script src=%q></script>", b.extraFullPath(ea)))
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

var DefaultNotFoundPageFunc = func(ctx *web.EventContext) (r web.PageResponse, err error) {
	msgr := MustGetMessages(ctx.R)
	r.Body = h.Div(
		h.H1("404").Class("mb-2"),
		h.Text(msgr.NotFoundPageNotice),
	).Class("text-center mt-8")
	return
}

func (*Builder) DefaultNotFoundPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	return DefaultNotFoundPageFunc(ctx)
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

	vuetifyx.HandleMaterialDesignIcons(b.prefix, mux)

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
	for _, cb := range b.customBuilders {
		routePath := fmt.Sprintf("%s/%s", b.prefix, cb.pattern)
		log.Printf("mounted url: %s", routePath)
		mux.Handle(
			routePath,
			b.wrapInner(func(p *web.PageBuilder) {
				p.MergeHub(&cb.EventsHub)
			}, cb.defaultLayout),
		)
	}
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
	if b.handlerHook != nil {
		b.handler = b.handlerHook(b.handler)
	}
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedResponse := &responseWriterWrapper{w, http.StatusOK}
		handler.ServeHTTP(capturedResponse, r)
		if capturedResponse.statusCode == http.StatusNotFound {
			// If no other handler wrote to the response, assume 404 and write our custom response.
			b.notFoundHandler.ServeHTTP(w, r)
		}
	})
}

func (b *Builder) WrapNotFoundHandler(w func(in http.Handler) (out http.Handler)) {
	b.notFoundHandler = w(b.notFoundHandler)
}

func (b *Builder) NotFoundHandler() http.Handler {
	return b.notFoundHandler
}

func (b *Builder) WithHandlerHook(hooks ...hook.Hook[http.Handler]) *Builder {
	b.handlerHook = hook.Prepend(b.handlerHook, hooks...)
	return b
}

// NewMuxHook creates a handler wrapper for sub-modules that have their own mux and prefix
func (b *Builder) NewMuxHook(mux *http.ServeMux) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, pattern := mux.Handler(r); pattern != "" {
				mux.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (b *Builder) AddWrapHandler(key string, f func(in http.Handler) (out http.Handler)) {
	b.wrapHandlers[key] = f
}

func (b *Builder) wrap(m *ModelBuilder, pf web.PageFunc) http.Handler {
	return b.wrapInner(func(p *web.PageBuilder) {
		if m != nil {
			m.registerDefaultEventFuncs()
			p.MergeHub(&m.EventsHub)
		}
	}, pf)
}

func (b *Builder) wrapInner(f func(p *web.PageBuilder), pf web.PageFunc) http.Handler {
	p := b.builder.Page(pf)
	if f != nil {
		f(p)
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
	p.Wrap(func(in web.PageFunc) web.PageFunc {
		return func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r, err = in(ctx)
			if err == nil && r.Body != nil {
				currentVuetifyLocale := strings.ReplaceAll(
					i18n.LanguageTagFromContext(ctx.R.Context(), language.English).String(),
					"-", "",
				)
				r.Body = h.Div(
					VProgressLinear().
						Attr(":active", "vars.globalProgressBar.show").
						Attr(":model-value", "vars.globalProgressBar.value").
						Attr("style", "position: fixed; z-index: 2000;").
						Height(2).
						Color(b.progressBarColor),
					VLocaleProvider().Locale(currentVuetifyLocale).FallbackLocale("en").Children(r.Body),
				)
			}
			return r, err
		}
	})

	p.Wrap(func(in web.PageFunc) web.PageFunc {
		return func(ctx *web.EventContext) (r web.PageResponse, err error) {
			defer func() {
				if v := recover(); v != nil {
					if render, ok := v.(PageRenderIface); ok {
						if rerr, ok := v.(error); ok {
							log.Printf("catch render err: %+v", rerr)
						}
						r, err = render.Render(ctx)
						return
					}
					panic(v)
				}
			}()
			return in(ctx)
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
	if len(lo.Uniq(mns)) != len(mns) {
		panic(fmt.Sprintf("Duplicated model names registered %v", mns))
	}
	b.initMux()
}

func (b *Builder) LookUpModelBuilder(uriName string) *ModelBuilder {
	mbs := lo.Filter(b.models, func(mb *ModelBuilder, _ int) bool {
		return mb.uriName == uriName
	})
	if len(mbs) > 1 {
		panic(fmt.Sprintf("Duplicated model names registered %q", uriName))
	}
	if len(mbs) == 0 {
		return nil
	}
	return mbs[0]
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.warmupOnce.Do(func() {
		if b.handler == nil {
			b.Build()
		}
	})
	if b.handler == nil {
		log.Printf("presets: Builder.handler is nil after Build; cannot serve request for %s", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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
			http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CreateVXBreadcrumbs(ctx *web.EventContext, f BreadcrumbItemsFunc) h.HTMLComponent {
	items := f(ctx, true)

	joinedItems := make([]h.HTMLComponent, 0, len(items)*2-1)
	for i, item := range items {
		joinedItems = append(joinedItems, item)
		if i < len(items)-1 {
			joinedItems = append(joinedItems, VBreadcrumbsDivider(h.Text("»")))
		}
	}

	return vuetifyx.VXBreadcrumbs(joinedItems...).Class("pa-0", W100)
}
