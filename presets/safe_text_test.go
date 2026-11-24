package presets

import (
	"context"
	"strings"
	"testing"

	h "github.com/theplant/htmlgo"
)

const cstiPayload = `{{ ''.constructor.constructor('alert(1)')() }}`

type cstiObj struct {
	Name string
}

func renderHTML(t *testing.T, c h.HTMLComponent) string {
	t.Helper()
	b, err := c.MarshalHTML(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Every default renderer that shows object data must put it inside v-pre,
// otherwise the browser-side Vue compiler executes "{{ }}" in user data.
func TestDefaultRenderersDoNotCompileObjectData(t *testing.T) {
	obj := &cstiObj{Name: cstiPayload}
	field := &FieldContext{Name: "Name", Label: "Name"}
	wantSafe := `<span v-pre>{{ &#39;&#39;.constructor.constructor(&#39;alert(1)&#39;)() }}</span>`

	cases := map[string]h.HTMLComponent{
		"listing cell":      cfTextTd(obj, field, nil),
		"readonly field":    cfReadonlyText(obj, field, nil),
		"safeText directly": safeText(cstiPayload),
	}
	for name, comp := range cases {
		t.Run(name, func(t *testing.T) {
			got := renderHTML(t, comp)
			if !strings.Contains(got, wantSafe) {
				t.Fatalf("object data must be wrapped in v-pre, got:\n%s", got)
			}
			if strings.Count(got, "{{") != 1 {
				t.Fatalf("payload must appear exactly once, got:\n%s", got)
			}
		})
	}
}

// v-pre must stay on an inner element so the cell wrapper's @click on the
// <td> is still compiled by Vue.
func TestListingCellKeepsTdCompilable(t *testing.T) {
	got := renderHTML(t, cfTextTd(&cstiObj{Name: "x"}, &FieldContext{Name: "Name"}, nil))
	if strings.HasPrefix(strings.TrimSpace(got), "<td v-pre") {
		t.Fatalf("v-pre must not be on <td>, got:\n%s", got)
	}
}
