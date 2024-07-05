package pregexp_test

import (
	"regexp"
	"sync"
	"testing"

	"github.com/qor5/admin/v3/utils/pregexp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexp(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{"simple digits", `\d+`, `\d+`},
		{"word characters", `\w+`, `\w+`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			re := pregexp.MustCompile(test.pattern)
			require.NotNil(t, re)
			assert.Equal(t, test.expected, re.String())
		})
	}
}

func TestRegexp_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	pattern := `(\d+)`
	n := 100 // Number of goroutines

	// To store the result regexp pointers
	resultRegexps := make([]*regexp.Regexp, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resultRegexps[index] = pregexp.MustCompile(pattern)
		}(i)
	}

	wg.Wait()

	// Check all returned regexps are the same
	for i := 1; i < n; i++ {
		assert.Equal(t, resultRegexps[0], resultRegexps[i], "Expected all regexps to be the same instance")
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected [][]string
		errMsg   string
	}{
		{"match digits", `(\d+)`, "123 abc 456", [][]string{{"123", "123"}, {"456", "456"}}, ""},
		{"no match", `(\d+)`, "abc", nil, "MatchFailed: (\\d+)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pregexp.Match(test.pattern, test.text)
			if test.errMsg != "" {
				require.Error(t, err)
				assert.Equal(t, test.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestMatchOne(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected []string
		errMsg   string
	}{
		{"one match", `(\d+)`, "123", []string{"123", "123"}, ""},
		{"no match", `(\d+)`, "abc", nil, "MatchFailed: (\\d+)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pregexp.MatchOne(test.pattern, test.text)
			if test.errMsg != "" {
				require.Error(t, err)
				assert.Equal(t, test.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestMatchOneThen(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		subIdx   int
		expected string
		errMsg   string
	}{
		{"match and index", `(\d+)`, "xx123xx", 1, "123", ""},
		{"index out of range", `(\d+)`, "xx123xxx", 2, "", "MatchOneThenFailed(2): (\\d+)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pregexp.MatchOneThen(test.pattern, test.text, test.subIdx)
			if test.errMsg != "" {
				require.Error(t, err)
				assert.Equal(t, test.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestNamedMatchOne(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected map[string]string
		errMsg   string
	}{
		{"named match", `(?P<digit>\d+)`, "zxc123cccc", map[string]string{"": "123", "digit": "123"}, ""},
		{"no match", `(?P<digit>\d+)`, "abc", nil, "NamedMatchOneFailed((?P<digit>\\d+)): matched 0 subs"},
		{
			"return non-empty one if same name",
			`<v-navigation-drawer v-model='vars.presetsRightDrawer'[\s\S]+?(<v-app-bar-title[^>]+>\s*<div[^>]+>(?P<title>.+?)\s*<\/div>\s*<\/v-app-bar-title>|<v-toolbar-title[^>]+>(?P<title>.+?)<\/v-toolbar-title>)[\s\S]+?<\/v-navigation-drawer>`,
			`
		<v-navigation-drawer v-model='vars.presetsRightDrawer' :location='"right"' :temporary='true' :width='"600"' :height='"100%"' class='v-navigation-drawer--temporary'>
	<global-events @keyup.esc='vars.presetsRightDrawer = false'></global-events>
	
	<go-plaid-portal :visible='true' :form='form' :locals='locals' portal-name='presets_RightDrawerContentPortalName'>
	<go-plaid-scope v-slot='{ form }'>
	<v-layout>
	<v-app-bar color='white' :elevation='0'>
	<v-app-bar-title class='pl-2'>
	<div class='d-flex'>WithPublishProduct 7_2024-05-23-v01</div>
	</v-app-bar-title>
	
	<v-btn :icon='"mdi-close"' @click.stop='vars.presetsRightDrawer = false'></v-btn>
	</v-app-bar></v-navigation-drawer>`,
			map[string]string{"": "<v-navigation-drawer v-model='vars.presetsRightDrawer' :location='\"right\"' :temporary='true' :width='\"600\"' :height='\"100%\"' class='v-navigation-drawer--temporary'>\n\t<global-events @keyup.esc='vars.presetsRightDrawer = false'></global-events>\n\t\n\t<go-plaid-portal :visible='true' :form='form' :locals='locals' portal-name='presets_RightDrawerContentPortalName'>\n\t<go-plaid-scope v-slot='{ form }'>\n\t<v-layout>\n\t<v-app-bar color='white' :elevation='0'>\n\t<v-app-bar-title class='pl-2'>\n\t<div class='d-flex'>WithPublishProduct 7_2024-05-23-v01</div>\n\t</v-app-bar-title>\n\t\n\t<v-btn :icon='\"mdi-close\"' @click.stop='vars.presetsRightDrawer = false'></v-btn>\n\t</v-app-bar></v-navigation-drawer>", "title": "WithPublishProduct 7_2024-05-23-v01"},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := pregexp.NamedMatchOne(test.pattern, test.text)
			if test.errMsg != "" {
				require.Error(t, err)
				assert.Equal(t, test.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestReplaceAllSubmatchFunc(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		replFunc func(string, [][]int) string
		expected string
	}{
		{
			"replace digits",
			`(\d+)`,
			"abc123def456",
			func(match string, groups [][]int) string {
				return "[" + match + "]"
			},
			"abc[123]def[456]",
		},
		{
			"replace one group in sub",
			`([a-z]+):([a-z]+)`,
			"abc foo:bar def baz:qux ghi",
			func(match string, groups [][]int) string {
				idx := 1
				repl := "xxx"
				return match[0:groups[idx][0]] + repl + match[groups[idx][1]:]
			},
			"abc xxx:bar def xxx:qux ghi",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := pregexp.ReplaceAllSubmatchFunc(test.pattern, test.text, test.replFunc)
			assert.Equal(t, test.expected, result)
		})
	}
}
