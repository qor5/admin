package presets

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLookUpModelBuilder(t *testing.T) {
	type Order struct {
		ID      uint
		Product string
	}
	type Customer struct {
		ID   uint
		Name string
	}

	pb := New()
	mb0 := pb.Model(&Order{})
	mb1 := pb.Model(&Customer{})
	assert.Equal(t, mb0, pb.LookUpModelBuilder(mb0.Info().URIName()))
	assert.Equal(t, mb1, pb.LookUpModelBuilder(mb1.Info().URIName()))

	mb3 := pb.Model(&Customer{})
	assert.PanicsWithValue(t, `Duplicated model names registered "customers"`, func() {
		pb.LookUpModelBuilder(mb3.Info().URIName())
	})
	mb3.URIName(mb3.Info().URIName() + "-version-list-dialog")
	assert.Equal(t, mb3, pb.LookUpModelBuilder(mb3.Info().URIName()))
}

func TestCloneFieldsLayout(t *testing.T) {
	src := []any{
		"foo",
		[]string{"bar"},
		&FieldsSection{
			Title: "title",
			Rows: [][]string{
				{"a", "b"},
				{"c", "d"},
			},
		},
	}
	require.Equal(t, src, CloneFieldsLayout(src))
}
