package examples_admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLinkageSelectFilterItemRemoteServer(t *testing.T) {
	recorder := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", http.NoBody)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
		return
	}

	handler := &LinkageSelectFilterItemRemoteServer{}

	handler.ServeHTTP(recorder, req)
	pr := PaginatedResponse{}

	if err = json.Unmarshal(recorder.Body.Bytes(), &pr); err != nil {
		t.Fatalf("Could not unmarshal JSON: %v", err)
		return
	}
	if len(pr.Data) == 0 {
		t.Fatalf("Could not find any items")
		return
	}
}

func checkLevelData(t *testing.T, v TestLinkageSelectFilterItemCase) {
	r := citiesResponse(v.Page, v.PageSize, v.Level, v.Search, v.ParentID)
	if v.ExceptLength < 0 && len(r.Data) == 0 {
		t.Fatalf("find level %v error", v.Level)
		return
	}
	if v.ExceptLength >= 0 && len(r.Data) != v.ExceptLength {
		t.Fatalf("level %v error, except length %v, got %v", v.Level, v.ExceptLength, len(r.Data))
		return
	}
	if v.ExceptTotal >= 0 && r.Total != v.ExceptTotal {
		t.Fatalf("level %v error, except total %v, got %v", v.Level, v.ExceptTotal, r.Total)
		return
	}
	for _, m := range r.Data {
		if m.Level != v.Level {
			t.Fatalf("find level %v error,%#+v", v.Level, m)
			return
		}
		if v.Search != "" {
			if !strings.Contains(m.Name, v.Search) {
				t.Fatalf("search %v error,%#+v", v.Search, m)
				return
			}
		}
		if v.ParentID != "" {
			if !checkIsParent(&m, v.ParentID) {
				t.Fatalf("parentID %v error,%#+v", v.ParentID, m)
				return
			}
		}
		if v.Level == 2 && m.Parent == nil {
			t.Fatalf("city parent error,%#+v", m)
			return
		}
		if v.Level == 3 && (m.Parent == nil || m.Parent.Parent == nil) {
			t.Fatalf("city parent error,%#+v", m)
			return
		}
	}
}

type (
	TestLinkageSelectFilterItemCase struct {
		Name         string
		Level        int
		ExceptLength int
		ExceptTotal  int
		Page         int
		PageSize     int
		Search       string
		ParentID     string
	}
)

func TestLinkageSelectFilterItem(t *testing.T) {
	cases := []TestLinkageSelectFilterItemCase{
		{
			Name:         "Test Level 1 no parameter",
			Level:        1,
			ExceptTotal:  -1,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
		},
		{
			Name:         "Test Level 1 no parameter",
			Level:        2,
			ExceptTotal:  -1,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
		},
		{
			Name:         "Test Level 1 no parameter",
			Level:        3,
			ExceptTotal:  -1,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
		},
		{
			Name:         "Test Level 1 Search",
			Level:        1,
			ExceptTotal:  2,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			Search:       "江",
		},
		{
			Name:         "Test Level 2 Search",
			Level:        2,
			ExceptTotal:  2,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			Search:       "州",
		},
		{
			Name:         "Test Level 3 Search",
			Level:        3,
			ExceptTotal:  8,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			Search:       "区",
		},
		{
			Name:         "Test Level 2 Last Level Parent",
			Level:        2,
			ExceptTotal:  2,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			ParentID:     "1",
		},
		{
			Name:         "Test Level 3 Last Level Parent",
			Level:        2,
			ExceptTotal:  1,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			ParentID:     "3",
		},
		{
			Name:         "Test Level 3 Find Level 1 Parent",
			Level:        3,
			ExceptTotal:  4,
			ExceptLength: -1,
			Page:         1,
			PageSize:     2,
			ParentID:     "1",
		},

		{
			Name:         "Test Level 1 Page Over Pages",
			Level:        1,
			ExceptTotal:  2,
			ExceptLength: 0,
			Page:         3,
			PageSize:     2,
			ParentID:     "",
		},
	}

	for _, v := range cases {
		t.Run(v.Name, func(t *testing.T) {
			checkLevelData(t, v)
		})
	}
}
