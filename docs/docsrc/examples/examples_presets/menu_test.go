package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestPresetsCustomizeMenu(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsCustomizeMenu(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "",
			Debug: true,
			ReqFunc: func() *http.Request {
				// detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/presets-customize-menu").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"/books?extra=abc", "Books", "<v-badge :content='1' :dot='true' color='error' class='pe-1'>", "mdi-bell-outline", "Media", "Videos", "Musics"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsCustomizeMenuWithPermission(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsGroupMenuWithPermission(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "allow group item",
			Debug: true,
			ReqFunc: func() *http.Request {
				// detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/presets-group-menu-with-permission/videos").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"Books", "Media", "Videos"},
			ExpectPageBodyNotContains:     []string{"Musics"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsPresetsMenuComponent(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsMenuComponent(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "multiple openStrategy",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/presets-menu-component/books").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`:open-strategy='"multiple"'`},
			ExpectPageBodyNotContains:     []string{`:open-strategy='"single"'`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
