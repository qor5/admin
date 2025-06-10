package docsrc

import (
	"github.com/theplant/docgo"

	"github.com/qor5/admin/v3/docs/docsrc/content"
	advanced_functions "github.com/qor5/admin/v3/docs/docsrc/content/advanced-functions"
	"github.com/qor5/admin/v3/docs/docsrc/content/basics"
	digging_deeper "github.com/qor5/admin/v3/docs/docsrc/content/digging-deeper"
	getting_started "github.com/qor5/admin/v3/docs/docsrc/content/getting-started"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
)

var DocTree = []interface{}{
	content.Home,
	&docgo.DocsGroup{
		Title: "Getting Started",
		Docs: []*docgo.DocBuilder{
			getting_started.OneMinuteQuickStart,
		},
	},
	&docgo.DocsGroup{
		Title: "Building Admin",
		Docs: []*docgo.DocBuilder{
			basics.PresetsInstantCRUD,
			// listing
			basics.Listing,
			basics.ListingCustomizations,
			basics.Filter,
			// editing
			basics.EditingCustomizations,
			// brand
			basics.Brand,
			// menu
			basics.ManageMenu,
			advanced_functions.DetailPageForComplexObject,
			basics.Layout,
			basics.Login,
			// permission
			basics.Permissions,
			basics.Role,
			// other basics
			// basics.NotificationCenter, // 历史遗产，先去除掉
			basics.ShortCut,
			basics.ConfirmDialog,
			basics.Slug,
			basics.SEO,
			basics.Activity,
			basics.Worker,
			basics.Publish,
			basics.I18n,
			basics.L10n,
			basics.Redirection,
			basics.CustomPage,
		},
	},

	&docgo.DocsGroup{
		Title: "Web Application",
		Docs: []*docgo.DocBuilder{
			advanced_functions.TheGoHTMLBuilder,
			advanced_functions.PageFuncAndEventFunc,
			advanced_functions.LayoutFunctionAndPageInjector,
			advanced_functions.LazyPortalsAndReload,
			advanced_functions.SwitchPagesWithPushState,
			advanced_functions.ReloadPageWithAFlash,
			advanced_functions.PartialRefreshWithPortal,
			advanced_functions.ManipulatePageURLInEventFunc,
			advanced_functions.SummaryOfEventResponse,
			advanced_functions.WebScope,
			advanced_functions.EventHandling,
			basics.FormHandling,
		},
	},

	&docgo.DocsGroup{
		Title: "UI Components",
		Docs: []*docgo.DocBuilder{
			// TODO: move BasicInputs to ATasteOfUsingVuetifyInGo
			basics.BasicInputs,
			advanced_functions.ATasteOfUsingVuetifyInGo,
			// vuetifyx
			basics.LinkageSelect,
			// build ui component
			digging_deeper.CompositeNewComponentWithGo,
			digging_deeper.IntegrateAHeavyVueComponent,
		},
	},

	&docgo.DocsGroup{
		Title: "Appendix",
		Docs: []*docgo.DocBuilder{
			docgo.Doc(utils.ExamplesDoc()).
				Title("All Demo Examples").
				Slug("appendix/all-demo-examples"),
		},
	},
}
