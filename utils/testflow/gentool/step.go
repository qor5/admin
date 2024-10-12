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

func QuoteBackticks(s string) string {
	var sb strings.Builder
	sb.WriteString("`")
	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			sb.WriteString("` + \"")
			for i < len(s) && s[i] == '`' {
				sb.WriteString("`")
				i++
			}
			sb.WriteString("\" + `")
			i--
		} else {
			sb.WriteByte(s[i])
		}
	}
	sb.WriteString("`")
	result := sb.String()
	result = strings.TrimPrefix(result, "`` + ")
	result = strings.TrimSuffix(result, " + ``")
	return result
}

func generateStep(rr RequestResponse) (*Step, error) {
	u, err := url.Parse(fmt.Sprintf("%s://%s:%d%s?%s", rr.Scheme, rr.Host, rr.Port, rr.Path, rr.Query))
	if err != nil {
		return nil, errors.Wrap(err, "url.Parse")
	}

	var eventFunc string
	queries := make(map[string]string)
	for key, values := range u.Query() {
		if key == "__execute_event__" {
			eventFunc = values[len(values)-1]
			continue
		}
		queries[strconv.Quote(key)] = strconv.Quote(values[len(values)-1])
	}

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
			if key == "__action__" {
				if eventFunc != "__dispatch_stateful_action__" {
					continue
				}
				formFields[strconv.Quote(key)] = QuoteBackticks("\n" + values[0])
				continue
			}
			formFields[strconv.Quote(key)] = strconv.Quote(values[0])
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

var assertIgnoreRegexp = regexp.MustCompile(`^\**\w+\.(Body|RunScript|UpdatePortals\[\d+\]\.Body|UpdatePortals\[\d+\]\.AfterLoaded)$`)

func generateLines(body string) ([]string, error) {
	body = strings.TrimSpace(body)
	var resp multipartestutils.TestEventResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, errors.Wrapf(err, "UnmarshalResponseBody: %s", body)
	}

	var lines []string

	assertions := generateAssertions("resp", resp, func(prefix string) bool {
		return assertIgnoreRegexp.MatchString(prefix)
	})
	if len(assertions) > 0 {
		lines = append(lines, assertions...)
		if resp.RunScript != "" {
			lines = append(lines, fmt.Sprintf("assert.Equal(t, testflow.RemoveTime(%s), testflow.RemoveTime(resp.RunScript))", QuoteBackticks(resp.RunScript)))
		} else {
			lines = append(lines, "assert.Empty(t, resp.RunScript)")
		}
		lines = append(lines, "")
	}

	b := []byte(body)
	var vlines []string
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
		lines = append(lines,
			`testflow.Validate(t, w, r,`,
			strings.Join(vlines, ",\n")+",",
			")",
		)
	}

	return lines, nil
}
