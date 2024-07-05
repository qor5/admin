package main

const flowTemplate = `package integration_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/utils/testflow"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Flow{{.FlowName}} struct {
	*Flow
}

func TestFlow{{.FlowName}}(t *testing.T) {
	flow{{.FlowName}}(t, &Flow{{.FlowName}}{
		Flow: &Flow{
			db: DB, h: PresetsBuilder,
		},
	})
}

func flow{{.FlowName}}(t *testing.T, f *Flow{{.FlowName}}) {
	{{range .Steps}}
	{{.FuncName}}(t, f)
	{{end}}
}

{{range .Steps}}
func {{.FuncName}}(t *testing.T, f *Flow{{.FlowName}}) *testflow.Then {
	r := multipartestutils.NewMultipartBuilder().
		PageURL("{{.URL}}").
		EventFunc({{.EventFunc}}).
		{{range $key, $value := .Queries}}Query({{$key}}, {{$value}}).
		{{end}}
		{{range $key, $value := .FormFields}}AddField({{$key}}, {{$value}}).
		{{end}}BuildEventFuncRequest()

	w := httptest.NewRecorder()
	f.h.ServeHTTP(w, r)

	var resp multipartestutils.TestEventResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	{{.Extra}}

	return testflow.NewThen(t, w, r)
}
{{end}}
`
