package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
)

// Only the public asset paths can be accessed without login.
func TestLoginPublicPaths(t *testing.T) {
	mux, c := admin.TestL18nHandler(TestDB)
	h := c.GetLoginSessionBuilder().Middleware()(mux)

	for _, c := range []struct {
		path   string
		expect int
	}{
		{path: "/orders/1.js", expect: http.StatusFound},
		{path: "/assets/main.js", expect: http.StatusOK},
		{path: "/assets/vue.js", expect: http.StatusOK},
		{path: "/vuetify/assets/index.css", expect: http.StatusOK},
		{path: "/assets/materialdesignicons-webfont.woff2", expect: http.StatusOK},
		{path: "/auth/assets/style.css", expect: http.StatusOK},
		// passed login, not mounted in this mux
		{path: "/favicon.ico", expect: http.StatusNotFound},
	} {
		t.Run(c.path, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, c.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != c.expect {
				t.Errorf("GET %s: expect %d, got %d", c.path, c.expect, w.Code)
			}
		})
	}
}
