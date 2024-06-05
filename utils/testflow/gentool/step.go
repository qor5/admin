package main

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
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

var assertIgnoreRegexp = regexp.MustCompile(`^\**\w+\.(Body|UpdatePortals\[\d+\]\.Body|UpdatePortals\[\d+\]\.AfterLoaded)$`)

func generateLines(body string) ([]string, error) {
	body = strings.TrimSpace(body)
	var resp multipartestutils.TestEventResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, errors.Wrapf(err, "UnmarshalResponseBody: %s", body)
	}

	lines := []string{}

	assertions := generateAssertions("resp", resp, func(prefix string) bool {
		return assertIgnoreRegexp.MatchString(prefix)
	})
	if len(assertions) > 0 {
		lines = append(lines, assertions...)
		lines = append(lines, "")
	}

	b := []byte(body)
	vlines := []string{}
	for _, validator := range validators {
		params, err := validator.ParseParams(b)
		if err != nil {
			continue
		}
		if _, ok := validator.Validate.(testflow.ValidatorFunc); ok {
			vlines = append(vlines, validator.Name)
			continue
		}
		if _, ok := validator.Validate.(func(t *testing.T, w *httptest.ResponseRecorder, r *http.Request)); ok {
			vlines = append(vlines, validator.Name)
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
		vlines = append(vlines, fmt.Sprintf("%s(%s)", validator.Name, strings.Join(callParams, ",")))
	}
	if len(vlines) > 0 {
		lines = append(lines, `testflow.Validate(t, w, r,`)
		lines = append(lines, strings.Join(vlines, ",\n")+",")
		lines = append(lines, ")")
	}

	return lines, nil
}
