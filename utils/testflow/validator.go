package testflow

import (
	"cmp"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ValidatorFunc func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request)

func Validate(t *testing.T, w *httptest.ResponseRecorder, r *http.Request, fs ...ValidatorFunc) (*httptest.ResponseRecorder, *http.Request) {
	for _, f := range fs {
		f(t, w, r)
	}
	return w, r
}

type Then struct {
	t *testing.T
	w *httptest.ResponseRecorder
	r *http.Request
}

func NewThen(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) *Then {
	return &Then{t, w, r}
}

func (v *Then) Then(f ValidatorFunc) {
	f(v.t, v.w, v.r)
}

func ParseOpenRightDrawerParams(body []byte) ([]any, error) {
	var resp multipartestutils.TestEventResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if len(resp.UpdatePortals) != 1 {
		return nil, errors.New("UpdatePortals !=1")
	}
	if resp.UpdatePortals[0].Name != "presets_RightDrawerPortalName" {
		return nil, errors.Errorf("invalid portal name %q", resp.UpdatePortals[0].Name)
	}
	const pattern = `<v-navigation-drawer v-model='vars.presetsRightDrawer'[\s\S]+?(<v-app-bar-title[^>]+>\s*<div[^>]+>(?P<title>.+?)\s*<\/div>\s*<\/v-app-bar-title>|<v-toolbar-title[^>]+>(?P<title>.+?)<\/v-toolbar-title>)[\s\S]+?<\/v-navigation-drawer>`
	groups, err := MatchOne(pattern, resp.UpdatePortals[0].Body)
	if err != nil {
		return nil, err
	}
	return []any{
		cmp.Or(groups[2], groups[3]), // title
	}, nil
}

func OpenRightDrawer(title string) ValidatorFunc {
	return func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
		params, err := ParseOpenRightDrawerParams(w.Body.Bytes())
		require.NoError(t, err)
		assert.Equal(t, title, params[0])
	}
}

func ContainsInOrder(candidates ...string) ValidatorFunc {
	return func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
		assert.Truef(t, containsInOrder(w.Body.String(), candidates), "should contains in correct order: %v", candidates)
	}
}

func ContainsInOrderAtUpdatePortal(idx int, candidates ...string) ValidatorFunc {
	return func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request) {
		var resp multipartestutils.TestEventResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Truef(t, len(resp.UpdatePortals) > idx, "UpdatePortal %d not found", idx)
		assert.Truef(t, containsInOrder(resp.UpdatePortals[idx].Body, candidates), "should contains in correct order: %v", candidates)
	}
}

func containsInOrder(body string, candidates []string) bool {
	var previousIndex int
	for _, candidate := range candidates {
		i := strings.Index(body, candidate)

		if i < previousIndex {
			return false
		}
		previousIndex = i
	}
	return true
}
