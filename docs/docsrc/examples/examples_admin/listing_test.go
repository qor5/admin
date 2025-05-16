package examples_admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	h "github.com/theplant/htmlgo"
)

var dataSeedForListing = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."posts" ("id", "title", "body", "updated_at", "created_at", "disabled", "status", "category_id") VALUES ('2', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('3', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('4', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('5', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('6', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('7', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('8', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('9', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('10', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('11', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('12', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('13', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('14', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('15', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('16', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('17', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('18', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('19', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('20', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('21', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('22', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0'),
('23', 'post0title', 'post0body', '0001-12-31 15:54:17+00 BC', '0001-12-31 15:54:17+00 BC', 'f', 'active', '0');
`, []string{"posts"}))

var dataEmptyForListing = gofixtures.Data(gofixtures.Sql(``, []string{"posts"}))

func TestListingExample(t *testing.T) {
	dbr, _ := TestDB.DB()
	TestDB.AutoMigrate(&Post{}, &Category{})

	cases := []multipartestutils.TestCase{
		{
			Name:  "empty List",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
				})
			},
			ReqFunc: func() *http.Request {
				dataEmptyForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/posts", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"No records to show"},
		},
		{
			Name:  "not empty List",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/posts", http.NoBody)
			},
			ExpectPageBodyNotContains:     []string{"No records to show"},
			ExpectPageBodyContainsInOrder: []string{"v-pagination", ":total-visible='5'"},
		},

		{
			Name:  "show action button",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample")
					}).UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/posts", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ActionExample"},
		},

		{
			Name:  "click action not exist",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample")
					}).UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenActionDialog",
			"request": {
				"name": "ActionNotExists"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"cannot find requested action"},
		},

		{
			Name:  "click action",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample Content")
					}).UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenActionDialog",
			"request": {
				"name": "ActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"ActionExample Content"},
		},

		{
			Name:  "click action with no updateFunc",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample Content")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenActionDialog",
			"request": {
				"name": "ActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"action.updateFunc not set"},
		},

		{
			Name:  "click action with no compFunc",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenActionDialog",
			"request": {
				"name": "ActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"action.compFunc not set"},
		},

		{
			Name:  "do action not exist",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample")
					}).UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "DoAction",
			"request": {
				"name": "ActionNotExists"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"cannot find requested action"},
		},

		{
			Name:  "do action",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().Action("ActionExample").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
						return h.Div().Text("ActionExample Content")
					}).UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
						return errors.New("not implemented")
					})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "DoAction",
			"request": {
				"name": "ActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"not implemented"},
		},

		{
			Name:  "show bulk action button",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample")
						}).
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/posts", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"BulkActionExample"},
		},

		{
			Name:  "click bulk action not exist",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample")
						}).
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenBulkActionDialog",
			"request": {
				"name": "BulkActionNotExists"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"cannot find requested bulk action"},
		},

		{
			Name:  "click bulk action",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample Content")
						}).
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": ["1"],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenBulkActionDialog",
			"request": {
				"name": "BulkActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"BulkActionExample Content"},
		},

		{
			Name:  "click bulk action with no updateFunc",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*presets.ListingCompo",
	"compo": {
		"id": "posts_page",
		"popup": false,
		"long_style_search_box": false,
		"selected_ids": [],
		"keyword": "",
		"order_bys": null,
		"page": 0,
		"per_page": 0,
		"display_columns": null,
		"active_filter_tab": "",
		"filter_query": "",
		"on_mounted": ""
	},
	"injector": "posts",
	"sync_query": true,
	"method": "OpenBulkActionDialog",
	"request": {
		"name": "BulkActionExample"
	}
}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"bulk.updateFunc not set"},
		},

		{
			Name:  "click bulk action with no compFunc",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "OpenBulkActionDialog",
			"request": {
				"name": "BulkActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"bulk.compFunc not set"},
		},

		{
			Name:  "do bulk action not exist",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample")
						}).
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": [],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "DoBulkAction",
			"request": {
				"name": "BulkActionNotExists"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"cannot find requested bulk action"},
		},

		{
			Name:  "do bulk action",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().BulkAction("BulkActionExample").
						ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
							return h.Div().Text("BulkActionExample")
						}).
						UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
							return errors.New("not implemented")
						})
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*presets.ListingCompo",
			"compo": {
				"id": "posts_page",
				"popup": false,
				"long_style_search_box": false,
				"selected_ids": ["1"],
				"keyword": "",
				"order_bys": null,
				"page": 0,
				"per_page": 0,
				"display_columns": null,
				"active_filter_tab": "",
				"filter_query": "",
				"on_mounted": ""
			},
			"injector": "posts",
			"sync_query": true,
			"method": "DoBulkAction",
			"request": {
				"name": "BulkActionExample"
			}
		}`).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"not implemented"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}

func TestListingWithJoinsExample(t *testing.T) {
	dbr, _ := TestDB.DB()
	TestDB.AutoMigrate(&Post{}, &Category{})

	cases := []multipartestutils.TestCase{
		{
			Name:  "empty List",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingPostWithCategory(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
				})
			},
			ReqFunc: func() *http.Request {
				dataEmptyForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/post-with-categories", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"No records to show"},
		},
		{
			Name:  "not empty List",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingPostWithCategory(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/post-with-categories", http.NoBody)
			},
			ExpectPageBodyNotContains:     []string{"No records to show"},
			ExpectPageBodyContainsInOrder: []string{"mdi-chevron-left", "mdi-chevron-right"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}
