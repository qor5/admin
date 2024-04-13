package presets

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/qor5/web/v3"
)

func TestIsMenuItemActive(t *testing.T) {
	cases := []struct {
		// path means current url path
		path string
		// link means menu item link
		link string

		excepted bool
	}{
		{"", "/", true},
		{"/", "/", true},
		{"/", "/order", false},
		{"/order", "/order", true},
		{"/order/1", "/order", true},
		{"/order#", "/order", true},
		{"/product", "/order", false},
		{"/product", "/", false},
	}

	type io struct {
		ctx      *web.EventContext
		m        *ModelBuilder
		excepted bool
	}

	var toIO []io
	b := New()
	for _, c := range cases {
		toIO = append(toIO, io{
			ctx: &web.EventContext{
				R: &http.Request{
					URL: &url.URL{
						Path: c.path,
					},
				},
			},
			m: &ModelBuilder{
				link:      c.link,
				modelInfo: &ModelInfo{mb: NewModelBuilder(b, &struct{}{})},
			},
			excepted: c.excepted,
		})
	}

	for i, io := range toIO {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if b.isMenuItemActive(io.ctx, io.m) != io.excepted {
				t.Errorf("isMenuItemActive() = %v, excepted %v", b.isMenuItemActive(io.ctx, io.m), io.excepted)
			}
		})
	}
}
