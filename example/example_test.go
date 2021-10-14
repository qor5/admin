package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor/qor5/example/admin"
	"github.com/stretchr/testify/assert"
)

func TestRedirectTrailingSlash(t *testing.T) {
	mux := admin.Router()

	r := httptest.NewRequest("GET", "/admin/?auth=token", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusFound, w.Code)

	p := w.Result().Header.Get("Location")
	if !strings.Contains(p, "/admin?auth=token") {
		t.Errorf("didn't get correct url: %#+v", p)
	}
}
