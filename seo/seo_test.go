package seo

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/l10n"
	"github.com/theplant/testingutils"
)

func TestSEO_AddChildren(t *testing.T) {
	cases := []struct {
		name        string
		getSeoRoot  func() *SEO
		expected    [][]string
		shouldPanic bool
	}{
		{
			name: "add_itself_as_child",
			getSeoRoot: func() *SEO {
				defer func() {
					if err := recover(); err == nil {
						panic("The program show panic, but it doesn't")
					}
				}()
				seoRoot := &SEO{name: "Root"}
				node1 := &SEO{name: "Node1"}
				node2 := &SEO{name: "Node2"}
				node3 := &SEO{name: "Node3"}
				seoRoot.AppendChildren(node1, node2)
				node2.AppendChildren(node3)
				// add itself as child, this will cause program panic
				node2.AppendChildren(node2)
				return seoRoot
			},
			shouldPanic: true,
		},
		{
			name: "add_children",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				node1 := &SEO{name: "Node1"}
				node2 := &SEO{name: "Node2"}
				node3 := &SEO{name: "Node3"}
				seoRoot.AppendChildren(node1, node2)
				node2.AppendChildren(node3)
				return seoRoot
			},
			expected: [][]string{
				{"Root"},
				{"Node1", "Node2"},
				{"nil", "Node3"},
				{"nil"},
			},
		},
		{
			name: "add_nil_child",
			getSeoRoot: func() *SEO {
				rootSeo := &SEO{name: "Root"}
				var nilSeo *SEO
				node1 := &SEO{name: "Node1"}
				node2 := &SEO{name: "Node2"}
				// add nil SEO as child, this will cause program to panic
				rootSeo.AppendChildren(node1.AppendChildren(node2), nilSeo)
				return rootSeo
			},
			shouldPanic: true,
		},
		{
			name: "add_child_that_cause_conflicts",
			getSeoRoot: func() *SEO {
				rootSeo := &SEO{name: "Root"}
				rootSeo.RegisterContextVariable(
					"ctx1",
					func(_ interface{}, _ *Setting, _ *http.Request) string {
						return "ctx1"
					},
				)
				child := &SEO{name: "Child"}
				child.RegisterSettingVariables("ctx1")
				rootSeo.AppendChildren(child)
				return rootSeo
			},
			shouldPanic: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err != nil) != c.shouldPanic {
					panic(err)
				}
			}()
			check(c.getSeoRoot(), c.expected, t)
		})
	}
}

func TestSEO_removeSelf(t *testing.T) {
	cases := []struct {
		name     string
		seoRoot  *SEO
		expected [][]string
	}{
		{
			name: "test_remove_itself",
			seoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				node1 := &SEO{name: "level1-1"}
				node2 := &SEO{name: "level1-2"}
				node3 := &SEO{name: "level2-1"}
				node4 := &SEO{name: "level2-2"}
				seoRoot.AppendChildren(node1.AppendChildren(node3, node4), node2)
				node1.removeSelf()
				return seoRoot
			}(),
			expected: [][]string{
				{"Root"},
				{"level1-2", "level2-1", "level2-2"},
				{"nil", "nil", "nil"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			check(c.seoRoot, c.expected, t)
		})
	}
}

func TestSEO_RegisterContextVariable(t *testing.T) {
	ctxFunc1 := func(i interface{}, setting *Setting, request *http.Request) string {
		return "contextFunc1"
	}
	ctxFunc2 := func(i interface{}, setting *Setting, request *http.Request) string {
		return "contextFunc2"
	}
	cases := []struct {
		name        string
		getSeoRoot  func() *SEO
		shouldPanic bool
		expected    map[string]map[string]ContextVarFunc
	}{
		{
			name: "register_context_variables",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterContextVariable("ctxFunc1", ctxFunc1)
				return seoRoot
			},
			expected: map[string]map[string]ContextVarFunc{
				"Root": {"ctxFunc1": ctxFunc1},
			},
		},
		{
			name: "register_context_var_that_exists_in_parent",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterContextVariable("c1", ctxFunc1)
				child := &SEO{name: "Child"}
				// If the context var "c1" is already exist in parent,
				// it should use the ctxFunc2 func to replace the ctxFunc1 func from parent
				child.SetParent(seoRoot).RegisterContextVariable("c1", ctxFunc2)
				return seoRoot
			},
			expected: map[string]map[string]ContextVarFunc{
				"Root":  {"c1": ctxFunc1},
				"Child": {"c1": ctxFunc2},
			},
		},
		{
			name: "register_context_var_that_conflicts_with_setting_var",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterSettingVariables("ctxFunc1")
				child := &SEO{name: "Child"}
				child.SetParent(seoRoot).RegisterContextVariable("ctxFunc1", ctxFunc1)
				return seoRoot
			},
			shouldPanic: true,
		},
		{
			name: "register_context_var_with_nil_func",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterContextVariable("aa", nil)
				return seoRoot
			},
			shouldPanic: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err != nil) != c.shouldPanic {
					t.Errorf("%v", err)
				}
			}()
			seoRoot := c.getSeoRoot()
			if c.shouldPanic {
				t.Errorf("The program should")
			}
			var que []*SEO
			que = append(que, seoRoot)
			cnt := 0
			for len(que) > 0 {
				cur := que[0]
				cnt++
				que = que[1:]
				if _, isExist := c.expected[cur.name]; !isExist {
					t.Errorf("The %v seo should not exist", cur.name)
				}
				if len(cur.contextVars) != len(c.expected[cur.name]) {
					t.Errorf("The length of expected context vars is not equal actual length")
				}
				for varName, ctxFunc := range cur.contextVars {
					if _, isExist := c.expected[cur.name][varName]; !isExist {
						t.Errorf("The context var %v should not exist", varName)
					}
					if c.expected[cur.name][varName](nil, nil, nil) != ctxFunc(nil, nil, nil) {
						t.Errorf("The context func for %v is different from what was expected", varName)
					}
				}
				que = append(que, cur.children...)
			}
			if cnt != len(c.expected) {
				t.Errorf("The number of seo nodes does not match the expectation")
			}
		})
	}
}

func TestSEO_RegisterSettingVariables(t *testing.T) {
	ctxFunc1 := func(i interface{}, setting *Setting, request *http.Request) string {
		return "contextFunc1"
	}
	cases := []struct {
		name        string
		getSeoRoot  func() *SEO
		shouldPanic bool
		expected    map[string]map[string]struct{}
	}{
		{
			name: "register_setting_var",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterSettingVariables("Var1")
				return seoRoot
			},
			expected: map[string]map[string]struct{}{
				"Root": {"Var1": {}},
			},
		},
		{
			name: "register_setting_var_that_conflicts_with_context_var",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterContextVariable("c1", ctxFunc1)
				child := &SEO{name: "Child"}
				child.SetParent(seoRoot).RegisterSettingVariables("c1")
				return seoRoot
			},
			shouldPanic: true,
		},
		{
			name: "chain_call",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterContextVariable("ctx1", ctxFunc1).
					AppendChildren(
						(&SEO{name: "Child1"}).RegisterSettingVariables("ctx1"),
					)
				return seoRoot
			},
			shouldPanic: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err != nil) != c.shouldPanic {
					t.Errorf("%v", err)
				}
			}()
			seoRoot := c.getSeoRoot()
			if c.shouldPanic {
				t.Errorf("The program should panic")
			}
			var que []*SEO
			que = append(que, seoRoot)
			cnt := 0
			for len(que) > 0 {
				cur := que[0]
				cnt++
				que = que[1:]
				if _, isExist := c.expected[cur.name]; !isExist {
					t.Errorf("The %v seo should not exist", cur.name)
				}
				if len(cur.settingVars) != len(c.expected[cur.name]) {
					t.Errorf("The length of expected setting vars is not equal actual length")
				}
				for varName := range cur.settingVars {
					if _, isExist := c.expected[cur.name][varName]; !isExist {
						t.Errorf("The setting var %v should not exist", varName)
					}
				}
				que = append(que, cur.children...)
			}
			if cnt != len(c.expected) {
				t.Errorf("The number of seo nodes does not match the expectation")
			}
		})
	}
}

func TestSEO_RegisterMetaProperty(t *testing.T) {
	cases := []struct {
		name        string
		getSeoRoot  func() *SEO
		shouldPanic bool
	}{
		{
			name: "malformed_prop_name",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterMetaProperty(
					"ogaudio",
					func(_ interface{}, _ *Setting, _ *http.Request) string {
						return "ogaudio"
					},
				)
				return seoRoot
			},
			shouldPanic: true,
		},
		{
			name: "prop_func_is_nil",
			getSeoRoot: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterMetaProperty("og:audio", nil)
				return seoRoot
			},
			shouldPanic: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err != nil) != c.shouldPanic {
					t.Errorf("%v", err)
				}
			}()
			_ = c.getSeoRoot()
			if c.shouldPanic {
				t.Errorf("The program should panic")
			}
		})
	}
}

func TestSEO_getLocaleFinalQorSEOSetting(t *testing.T) {
	cases := []struct {
		name      string
		prepareDB func()
		seo       *SEO
		expected  *QorSEOSetting
	}{
		{
			name: "override_setting_var_from_parent",
			prepareDB: func() {
				resetDB()
				seoSettings := []*QorSEOSetting{
					{
						Name:    "nodeA",
						Setting: Setting{Title: "{{varA}}"},
						Variables: map[string]string{
							"varA": "1",
						},
					},
					{
						Name:    "nodeB",
						Setting: Setting{Title: "{{varB}}"},
						Variables: map[string]string{
							"varB": "2",
						},
					},
					{
						Name:    "nodeC",
						Setting: Setting{Title: ""},
						Variables: map[string]string{
							"varB": "3",
						},
					},
				}
				if err := dbForTest.Create(seoSettings).Error; err != nil {
					panic(err)
				}
			},
			seo: func() *SEO {
				// seoRoot --> nodeA --> nodeB --> nodeC
				seoRoot := &SEO{}
				nodeA := &SEO{name: "nodeA"}
				nodeA.RegisterSettingVariables("varA")
				nodeB := &SEO{name: "nodeB"}
				nodeB.RegisterSettingVariables("varB")
				nodeC := &SEO{name: "nodeC"}
				// Override the `varB` from the nodeB
				nodeC.RegisterSettingVariables("varB")
				seoRoot.AppendChildren(nodeA.AppendChildren(nodeB.AppendChildren(nodeC)))
				return nodeC
			}(),
			expected: &QorSEOSetting{
				Name:    "nodeC",
				Setting: Setting{Title: "{{varB}}"},
				Variables: map[string]string{
					"varA": "1",
					"varB": "3",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.prepareDB != nil {
				c.prepareDB()
			}
			actual := &QorSEOSetting{}
			seoSetting := c.seo.getLocaleFinalQorSEOSetting("", dbForTest)
			actual.Name = seoSetting.Name
			actual.Setting = seoSetting.Setting
			actual.Variables = seoSetting.Variables
			r := testingutils.PrettyJsonDiff(c.expected, actual)
			if r != "" {
				t.Error(r)
			}
		})
	}
}

func TestSEO_getFinalQorSEOSetting(t *testing.T) {
	cases := []struct {
		name      string
		prepareDB func()
		seo       *SEO
		expected  map[string]*QorSEOSetting
	}{
		{
			name: "override_setting_var_from_parent",
			prepareDB: func() {
				resetDB()
				seoSets := []*QorSEOSetting{
					{
						Name:    "nodeA",
						Setting: Setting{Title: "{{Greeting}}"},
						Variables: map[string]string{
							"Greeting": "Hello",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name:    "nodeA",
						Setting: Setting{Title: "{{Greeting}}"},
						Variables: map[string]string{
							"Greeting": "你好",
						},
						Locale: l10n.Locale{LocaleCode: "zh"},
					},
					{
						Name:    "nodeB",
						Setting: Setting{Title: ""}, // The title filed will inherit from its parent
						Variables: map[string]string{
							"Greeting": "Hello, SEO", // override value from parent
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name:    "nodeB",
						Setting: Setting{Title: "nodeB"}, // The title filed will override its parent
						Variables: map[string]string{
							"Greeting": "你好, SEO", // override value from parent
						},
						Locale: l10n.Locale{LocaleCode: "zh"},
					},
				}
				if err := dbForTest.Create(seoSets).Error; err != nil {
					panic(err)
				}
			},
			seo: func() *SEO {
				nodeA := &SEO{name: "nodeA"}
				nodeB := &SEO{name: "nodeB"}
				nodeB.SetParent(nodeA)
				return nodeB
			}(),
			expected: map[string]*QorSEOSetting{
				"en": {
					Name:    "nodeB",
					Setting: Setting{Title: "{{Greeting}}"},
					Variables: map[string]string{
						"Greeting": "Hello, SEO",
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				},
				"zh": {
					Name:    "nodeB",
					Setting: Setting{Title: "nodeB"},
					Variables: map[string]string{
						"Greeting": "你好, SEO",
					},
					Locale: l10n.Locale{LocaleCode: "zh"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.prepareDB != nil {
				c.prepareDB()
			}
			seoSets := c.seo.getFinalQorSEOSetting(dbForTest)
			if len(seoSets) != len(c.expected) {
				t.Errorf("The number of configuration does not match expetations")
			}
			for locale, actualSets := range seoSets {
				if expectedSets, isExist := c.expected[locale]; !isExist {
					t.Errorf("There is no SEO configuration available for %v", locale)
				} else {
					actual := &QorSEOSetting{}
					actual.Setting = actualSets.Setting
					actual.Variables = actualSets.Variables
					actual.Name = actualSets.Name
					actual.LocaleCode = actualSets.LocaleCode
					r := testingutils.PrettyJsonDiff(expectedSets, actual)
					if r != "" {
						t.Error(r)
					}
				}
			}
		})
	}
}

func TestSEO_getFinalContextVars(t *testing.T) {
	ctxFunc1 := func(_ interface{}, _ *Setting, _ *http.Request) string {
		return "ctxFunc1"
	}
	ctxFunc2 := func(_ interface{}, _ *Setting, _ *http.Request) string {
		return "ctxFunc2"
	}
	cases := []struct {
		name     string
		seo      *SEO
		expected map[string]string
	}{
		{
			name: "override_context_var_from_parent",
			seo: func() *SEO {
				nodeA := &SEO{name: "nodeA"}
				nodeA.
					RegisterContextVariable("ctxVarA", ctxFunc1).
					RegisterContextVariable("ctxVarB", ctxFunc2)
				nodeB := &SEO{name: "nodeB"}
				nodeB.RegisterContextVariable("ctxVarA", ctxFunc2)
				nodeA.AppendChildren(nodeB)
				return nodeB
			}(),
			expected: map[string]string{
				"ctxVarA": "ctxFunc2",
				"ctxVarB": "ctxFunc2",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			contextVars := c.seo.getFinalContextVars()
			if len(contextVars) != len(c.expected) {
				t.Errorf("The actual number of context vars is different from what was expected")
			}
			for varName, varFunc := range contextVars {
				res := varFunc(nil, nil, nil)
				if res != c.expected[varName] {
					t.Errorf("The actual value %v is not equal to %v", res, c.expected[varName])
				}
			}
		})
	}
}

func TestSEO_getFinalPropFuncForOG(t *testing.T) {
	cases := []struct {
		name     string
		getSEO   func() *SEO
		expected map[string]string
	}{
		{
			name: "override_config_inherited_from_upper_level",
			getSEO: func() *SEO {
				seoRoot := &SEO{name: "Root"}
				seoRoot.RegisterMetaProperty(
					"og:audio",
					func(_ interface{}, _ *Setting, _ *http.Request) string {
						return "https://example.com/bond/root.mp3"
					},
				)
				child := &SEO{name: "Child"}
				child.RegisterMetaProperty(
					"og:audio",
					func(_ interface{}, _ *Setting, _ *http.Request) string {
						return "https://example.com/bond/child.mp3"
					},
				).SetParent(seoRoot)
				return child
			},
			expected: map[string]string{
				"og:audio": "https://example.com/bond/child.mp3",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seo := c.getSEO()
			finalMetaProps := seo.getFinalMetaProps()
			if len(finalMetaProps) != len(c.expected) {
				t.Errorf("The number of og property is not equal to expectation")
			}
			for prop, propFunc := range finalMetaProps {
				actualVal := propFunc(nil, nil, nil)
				if c.expected[prop] != actualVal {
					t.Errorf("The %v property's actual value: %v, but %v is expected",
						prop, actualVal, c.expected[prop])
				}
			}
		})
	}
}

// check function checks if the structure of seoRoot conforms to the expected shape
// after level-order traversal
func check(seoRoot *SEO, expected [][]string, t *testing.T) {
	var que []*SEO
	que = append(que, seoRoot)
	level := 0
	for len(que) > 0 {
		curLen := len(que)
		if curLen != len(expected[level]) {
			t.Errorf("The number of nodes in the current level does not meet the expected quantity.")
		}
		i := 0
		for i < curLen {
			cur := que[0]
			expectedName := expected[level][i]
			if cur == nil {
				if expectedName != "nil" {
					t.Errorf("actual: %v, expected: %v", "nil", expectedName)
				}
			} else {
				if expectedName != cur.name {
					t.Errorf("actual: %v, expected: %v", cur.name, expectedName)
				}
				if cur.children == nil {
					que = append(que, nil)
				} else {
					que = append(que, cur.children...)
				}
			}
			que = que[1:]
			i++
		}
		level++
	}
}
