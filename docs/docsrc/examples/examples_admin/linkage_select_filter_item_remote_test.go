package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/multipartestutils"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var linkageSelectFilterData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.addresses (id, province, city, district) VALUES (1, '浙江', '杭州', '拱墅区');
INSERT INTO public.addresses (id, province, city, district) VALUES (2, '浙江', '杭州', '西湖区');

`, []string{"addresses"}))

func TestLinkageSelectFilter(t *testing.T) {
	pb := presets.New()
	b := LinkageSelectFilterItemRemoteExample(pb, http.NewServeMux(), TestDB)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index LinkageSelectFilterItemRemote",
			Debug: true,
			ReqFunc: func() *http.Request {
				linkageSelectFilterData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/addresses", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"西湖区", "拱墅区"},
		},
		{
			Name:  "Index LinkageSelectFilterItemRemote",
			Debug: true,
			ReqFunc: func() *http.Request {
				linkageSelectFilterData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/addresses").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"vx-linkageselect-remote", "浙江", "杭州", "拱墅区"},
		},
		{
			Name:  "Index LinkageSelectFilterItemRemote",
			Debug: true,
			ReqFunc: func() *http.Request {
				linkageSelectFilterData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/addresses").
					EventFunc(actions.Update).
					AddField("ProvinceCityDistrict[0].Name", "浙江").
					AddField("ProvinceCityDistrict[1].Name", "宁波").
					AddField("ProvinceCityDistrict[2].Name", "镇海区").
					BuildEventFuncRequest()
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				m := examples_presets.Address{}
				TestDB.Order("id desc ").First(&m)
				if m.Province != "浙江" || m.City != "宁波" || m.District != "镇海区" {
					t.Fatalf("create address error %#+v", m)
					return
				}
			},
		},
		{
			Name:  "Index LinkageSelectFilterItemRemote Filter",
			Debug: true,
			ReqFunc: func() *http.Request {
				linkageSelectFilterData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/addresses?f_province_city_district=浙江__1,杭州__3,拱墅区__7", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"拱墅区"},
			ExpectPageBodyNotContains:     []string{"西湖区"},
		},

		{
			Name:  "Index LinkageSelectFilterItemRemote No Selected",
			Debug: true,
			ReqFunc: func() *http.Request {
				linkageSelectFilterData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/addresses?f_province_city_district=", http.NoBody)
			},
			ExpectPageBodyNotContains: []string{`"selected":true`},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
