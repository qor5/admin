package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
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
				return httptest.NewRequest("GET", "/posts", nil)
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
				return httptest.NewRequest("GET", "/posts", nil)
			},
			ExpectPageBodyNotContains:     []string{"No records to show"},
			ExpectPageBodyContainsInOrder: []string{"v-pagination", ":total-visible='5'"},
		},
		{
			Name:  "PaginationTotalVisible 2",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return listingExample(presets.New(), TestDB, func(mb *presets.ModelBuilder) {
					mb.Listing().PaginationTotalVisible(2)
				})
			},
			ReqFunc: func() *http.Request {
				dataSeedForListing.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/posts", nil)
			},
			ExpectPageBodyNotContains:     []string{"No records to show"},
			ExpectPageBodyContainsInOrder: []string{"v-pagination", ":total-visible='2'"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}
