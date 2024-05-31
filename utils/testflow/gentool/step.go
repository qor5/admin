package main

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
)

type Step struct {
	FlowName   string
	FuncName   string
	URL        string
	EventFunc  string
	Queries    map[string]string
	FormFields map[string]string
	Extra      string
}

func generateStep(rr RequestResponse) (*Step, error) {
	u, err := url.Parse(fmt.Sprintf("%s://%s:%d%s?%s", rr.Scheme, rr.Host, rr.Port, rr.Path, rr.Query))
	if err != nil {
		return nil, errors.Wrap(err, "url.Parse")
	}

	queries := make(map[string]string)
	for key, values := range u.Query() {
		queries[key] = values[0]
	}

	eventFunc := queries["__execute_event__"]
	delete(queries, "__execute_event__")

	formFields := make(map[string]string)
	if rr.Request.MimeType == "multipart/form-data" {
		boundary := getBoundary(rr.Request.Header.Headers)
		if boundary == "" {
			return nil, errors.New("no boundary found in Content-Type header")
		}
		reader := multipart.NewReader(strings.NewReader(rr.Request.Body.Text), boundary)
		form, err := reader.ReadForm(65535)
		if err != nil {
			return nil, errors.Wrap(err, "ReadForm")
		}
		for key, values := range form.Value {
			formFields[key] = values[0]
		}
	}

	lines, err := generateLines(rr.Response.Body.Text)
	if err != nil {
		return nil, err
	}

	return &Step{
		FuncName:   "Event_" + eventFunc,
		URL:        u.Path,
		EventFunc:  strconv.Quote(eventFunc),
		Queries:    queries,
		FormFields: formFields,
		Extra:      strings.Join(lines, "\n"),
	}, nil
}

func getBoundary(headers []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
},
) string {
	for _, header := range headers {
		if header.Name == "Content-Type" {
			parts := strings.Split(header.Value, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "boundary=") {
					return strings.TrimPrefix(part, "boundary=")
				}
			}
		}
	}
	return ""
}

type NamedValidator struct {
	Name        string
	ParseParams func(body []byte) ([]any, error)
	Validate    any
}

var validators = []NamedValidator{
	{
		Name:        "testflow.OpenRightDrawer",
		ParseParams: testflow.ParseOpenRightDrawerParams,
		Validate:    testflow.OpenRightDrawer,
	},
}

func generateLines(body string) ([]string, error) {
	body = strings.TrimSpace(body)
	var resp multipartestutils.TestEventResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, errors.Wrapf(err, "UnmarshalResponseBody: %s", body)
	}

	assertion := func(action, actual string, expected any) string {
		if expected == nil {
			return fmt.Sprintf("assert.%s(t, resp.%s)", action, actual)
		}
		expectedExp := fmt.Sprint(expected)
		if s, ok := expected.(string); ok {
			expectedExp = strconv.Quote(s)
		}
		if action == "Len" {
			return fmt.Sprintf("assert.%s(t, resp.%s,%s, )", action, actual, expectedExp)
		}
		return fmt.Sprintf("assert.%s(t, %s, resp.%s)", action, expectedExp, actual)
	}

	lines := []string{}

	// PageTitle
	if resp.PageTitle != "" {
		lines = append(lines, assertion("Equal", "PageTitle", resp.PageTitle))
	} else {
		lines = append(lines, assertion("Empty", "PageTitle", nil))
	}
	// Reload
	if resp.Reload {
		lines = append(lines, assertion("True", "Reload", nil))
	} else {
		lines = append(lines, assertion("False", "Reload", nil))
	}
	// PushState
	if resp.PushState != nil {
		lines = append(lines, assertion("NotNil", "PushState", nil))
		if resp.PushState.MyMergeQuery {
			lines = append(lines, assertion("True", "PushState.MyMergeQuery", nil))
		} else {
			lines = append(lines, assertion("False", "PushState.MyMergeQuery", nil))
		}
		if resp.PushState.MyURL != "" {
			lines = append(lines, assertion("Equal", "PushState.MyURL", resp.PushState.MyURL))
		} else {
			lines = append(lines, assertion("Empty", "PushState.MyURL", nil))
		}
		if resp.PushState.MyStringQuery != "" {
			lines = append(lines, assertion("Equal", "PushState.MyStringQuery", resp.PushState.MyStringQuery))
		} else {
			lines = append(lines, assertion("Empty", "PushState.MyStringQuery", nil))
		}
		if len(resp.PushState.MyClearMergeQueryKeys) > 0 {
			lines = append(lines, assertion("Equal", "PushState.MyClearMergeQueryKeys", resp.PushState.MyClearMergeQueryKeys))
		} else {
			lines = append(lines, assertion("Empty", "PushState.MyClearMergeQueryKeys", nil))
		}
	} else {
		lines = append(lines, assertion("Nil", "PushState", nil))
	}
	// RedirectURL
	if resp.RedirectURL != "" {
		lines = append(lines, assertion("Equal", "RedirectURL", resp.RedirectURL))
	} else {
		lines = append(lines, assertion("Empty", "RedirectURL", nil))
	}
	// ReloadPortals
	if len(resp.ReloadPortals) > 0 {
		lines = append(lines, assertion("Equal", "ReloadPortals", resp.ReloadPortals))
	} else {
		lines = append(lines, assertion("Empty", "ReloadPortals", nil))
	}
	// UpdatePortals
	if len(resp.UpdatePortals) > 0 {
		lines = append(lines, assertion("Len", "UpdatePortals", len(resp.UpdatePortals)))
		for i, portal := range resp.UpdatePortals {
			if portal.Name != "" {
				lines = append(lines, assertion("Equal", fmt.Sprintf("UpdatePortals[%d].Name", i), portal.Name))
			} else {
				lines = append(lines, assertion("Empty", fmt.Sprintf("UpdatePortals[%d].Name", i), nil))
			}
		}
	} else {
		lines = append(lines, assertion("Empty", "UpdatePortals", nil))
	}
	// Data
	if resp.Data != nil {
		lines = append(lines, assertion("NotNil", "Data", nil))
	} else {
		lines = append(lines, assertion("Nil", "Data", nil))
	}
	// RunScript
	if resp.RunScript != "" {
		lines = append(lines, assertion("Equal", "RunScript", resp.RunScript))
	} else {
		lines = append(lines, assertion("Empty", "RunScript", nil))
	}

	lines = append(lines, "")

	vLines := []string{}
	b := []byte(body)
	for _, validator := range validators {
		params, err := validator.ParseParams(b)
		if err != nil || len(params) < 1 {
			continue
		}
		if _, ok := validator.Validate.(testflow.ValidatorFunc); ok {
			vLines = append(vLines, validator.Name)
			continue
		}
		if _, ok := validator.Validate.(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request)); ok {
			vLines = append(vLines, validator.Name)
			continue
		}
		var callParams []string
		for _, v := range params {
			switch vv := v.(type) {
			case string:
				callParams = append(callParams, strconv.Quote(vv))
			default:
				callParams = append(callParams, fmt.Sprint(vv))
			}
		}
		vLines = append(vLines, fmt.Sprintf("%s(%s)", validator.Name, strings.Join(callParams, ",")))
	}
	if len(vLines) > 0 {
		lines = append(lines, `testflow.Validate(t, w, r,`)
		lines = append(lines, strings.Join(vLines, ",\n")+",")
		lines = append(lines, ")")
	}

	return lines, nil
}
