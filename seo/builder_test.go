package seo

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	_ "github.com/lib/pq"
	"github.com/qor5/admin/v3/l10n"
	"github.com/theplant/testingutils"
)

func TestBuilder_Render(t *testing.T) {
	u, _ := url.Parse("http://dev.qor5.com/product/1")
	defaultRequest := &http.Request{
		Method: "GET",
		URL:    u,
	}

	globalSeoSetting := []*QorSEOSetting{
		{
			Name: defaultGlobalSEOName,
			Setting: Setting{
				Title: "global | {{SiteName}}",
			},
			Variables: map[string]string{"SiteName": "Qor5 dev"},
			Locale:    l10n.Locale{LocaleCode: "en"},
		},
		{
			Name: defaultGlobalSEOName,
			Setting: Setting{
				Title: "全局 | {{SiteName}}",
			},
			Variables: map[string]string{"SiteName": "Qor5 开发"},
			Locale:    l10n.Locale{LocaleCode: "zh"},
		},
	}

	tests := []struct {
		name      string
		prepareDB func()
		builder   *Builder
		obj       interface{}
		want      string
	}{
		{
			name:      "Render_non-model_seo_with_setting_variables_and_default_context_variables",
			prepareDB: func() { dbForTest.Save(&globalSeoSetting) },
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				return builder
			}(),
			obj: NewNonModelSEO(defaultGlobalSEOName, "en"),
			want: `
			<title>global | Qor5 dev</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
			`,
		},
		{
			name: "Render_model_seo_with_locale",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | {{SiteName}}",
					},
					Variables: map[string]string{"SiteName": "Qor5 开发"},
					Locale:    l10n.Locale{LocaleCode: "zh"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().AppendChildren(
					builder.RegisterSEO("Product", &Product{}),
				)
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1",
					EnabledCustomize: true,
				},
				Locale: l10n.Locale{LocaleCode: "zh"},
			},
			want: `<title>product1</title>`,
		},
		{
			name: "Render_model_seo_without_locale",
			prepareDB: func() {
				settings := []*QorSEOSetting{
					{
						Name: defaultGlobalSEOName,
						Setting: Setting{
							Title: "global | {{SiteName}}",
						},
						Variables: map[string]string{"SiteName": "Qor5 dev"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
						Variables: map[string]string{"SiteName": "Qor5 开发"},
					},
				}
				dbForTest.Save(settings)
			},
			builder: func() *Builder {
				builder := New(dbForTest).AutoMigrate()
				builder.GetGlobalSEO().AppendChildren(
					builder.RegisterSEO("Product", &Product{}),
				)
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1",
					EnabledCustomize: true,
				},
			},
			want: `<title>product1</title>`,
		},
		{
			name: "Render_non-model_seo_with_global_setting_variables",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				var product []*QorSEOSetting
				product = append(product, &QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | {{SiteName}}",
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				})
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().AppendChildren(
					builder.RegisterSEO("Product"),
				)
				return builder
			}(),
			obj:  NewNonModelSEO("Product", "en"),
			want: `<title>product | Qor5 dev</title>`,
		},

		{
			name: "Render_non-model_seo_with_locale",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := []*QorSEOSetting{
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "产品 | {{SiteName}}",
						},
						Locale: l10n.Locale{LocaleCode: "zh"},
					},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en", "zh")).AutoMigrate()
				builder.GetGlobalSEO().AppendChildren(
					builder.RegisterSEO("Product"),
				)
				return builder
			}(),
			obj:  NewNonModelSEO("Product", "zh"),
			want: `<title>产品 | Qor5 开发</title>`,
		},

		{
			name: "Render_non-model_seo_without_locale",
			prepareDB: func() {
				settings := []*QorSEOSetting{
					{
						Name: defaultGlobalSEOName,
						Setting: Setting{
							Title: "global | {{SiteName}}",
						},
						Variables: map[string]string{"SiteName": "Qor5 dev"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
					},
				}
				dbForTest.Save(&settings)
			},
			builder: func() *Builder {
				builder := New(dbForTest).AutoMigrate()
				builder.GetGlobalSEO().AppendChildren(
					builder.RegisterSEO("Product"),
				)
				return builder
			}(),
			obj:  NewNonModelSEO("Product"),
			want: `<title>product | Qor5 dev</title>`,
		},
		{
			name: "Render_seo_with_setting_and_opengraph_prop_and_without_locale",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product {{ProductTag}} | {{SiteName}}",
					},
					Variables: map[string]string{"ProductTag": "Men"},
					Locale:    l10n.Locale{LocaleCode: "en"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product", &Product{}).
					RegisterSettingVariables("ProductTag").
					RegisterMetaProperty("og:image",
						func(i interface{}, setting *Setting, request *http.Request) string {
							return "http://dev.qor5.com/images/logo.png"
						},
					).SetParent(builder.GetGlobalSEO())
				return builder
			}(),
			obj: &Product{
				Name:   "product",
				Locale: l10n.Locale{LocaleCode: "en"},
			},
			want: `
			<title>product Men | Qor5 dev</title>
			<meta property='og:image' name='og:image' content='http://dev.qor5.com/images/logo.png'>`,
		},

		{
			name: "Render_model_setting_with_global_and_SEO_setting_variables",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name:      "Product",
					Variables: map[string]string{"ProductTag": "Men"},
					Locale:    l10n.Locale{LocaleCode: "en"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product", &Product{}).SetParent(builder.GetGlobalSEO())
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1 | {{ProductTag}} | {{SiteName}}",
					EnabledCustomize: true,
				},
				Locale: l10n.Locale{LocaleCode: "en"},
			},
			want: `<title>product1 | Men | Qor5 dev</title>`,
		},

		{
			name: "Render_model_setting_with_default_SEO_setting",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | Qor5 dev",
					},
					Variables: map[string]string{"ProductTag": "Men"},
					Locale:    l10n.Locale{LocaleCode: "en"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product", &Product{}).SetParent(builder.GetGlobalSEO())
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1 | {{ProductTag}} | {{SiteName}}",
					EnabledCustomize: false,
				},
				Locale: l10n.Locale{LocaleCode: "en"},
			},
			want: `<title>product | Qor5 dev</title>`,
		},

		{
			name: "Render_model_setting_with_inherit_global_and_SEO_setting",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Description: "product description",
					},
					Variables: map[string]string{"ProductTag": "Men"},
					Locale:    l10n.Locale{LocaleCode: "en"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product", &Product{}).SetParent(builder.GetGlobalSEO())
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Keywords:         "shoes, {{ProductTag}}",
					EnabledCustomize: true,
				},
				Locale: l10n.Locale{LocaleCode: "en"},
			},
			want: `
			<title>global | Qor5 dev</title>
			<meta name='description' content='product description'>
			<meta name='keywords' content='shoes, Men'>
			`,
		},

		{
			name: "Render_model_setting_without_inherit_global_and_SEO_setting",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Description: "product description",
					},
					Variables: map[string]string{"ProductTag": "Men"},
					Locale:    l10n.Locale{LocaleCode: "en"},
				}
				dbForTest.Save(&product)
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en"), WithInherit(false)).AutoMigrate()
				builder.RegisterSEO("Product", &Product{})
				return builder
			}(),
			obj: &Product{
				Name: "product 1",
				SEO: Setting{
					Keywords:         "shoes, {{ProductTag}}",
					EnabledCustomize: true,
				},
				Locale: l10n.Locale{LocaleCode: "en"},
			},
			want: `
			<title></title>
			<meta name='description'>
			<meta name='keywords' content='shoes, Men'>
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetDB()
			tt.prepareDB()
			if got, _ := tt.builder.Render(tt.obj, defaultRequest).MarshalHTML(context.TODO()); !metaEqual(string(got), tt.want) {
				t.Errorf("Render = %v\nExpected = %v", string(got), tt.want)
			}
		})
	}
}

func TestBuilder_GetSEOPriority(t *testing.T) {
	cases := []struct {
		name     string
		builder  *Builder
		expected map[string]int
	}{
		{
			name: "with global seo",
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("PLP").AppendChildren(
					builder.RegisterSEO("Region"),
					builder.RegisterSEO("City"),
					builder.RegisterSEO("Prefecture"),
				).AppendChildren(
					builder.RegisterSEO("Post"),
					builder.RegisterSEO("Product"),
				)
				return builder
			}(),
			expected: map[string]int{
				defaultGlobalSEOName: 1,
				"PLP":                2,
				"Post":               3,
				"Product":            3,
				"Region":             3,
				"City":               3,
				"Prefecture":         3,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for seoName, priority := range c.expected {
				actualPriority := c.builder.GetSEOPriority(seoName)
				if actualPriority != priority {
					t.Errorf("GetPriorities = %v, want %v", actualPriority, priority)
				}
			}
		})
	}
}

func TestBuilder_RemoveSEO(t *testing.T) {
	cases := []struct {
		name     string
		builder  *Builder
		expected *Builder
	}{{
		name: "test remove SEO",
		builder: func() *Builder {
			builder := New(dbForTest, WithLocales("en")).AutoMigrate()
			builder.RegisterSEO("Parent1").AppendChildren(
				builder.RegisterSEO("Son1"),
				builder.RegisterSEO("Son2"),
			)
			builder.RemoveSEO("Parent1")
			return builder
		}(),
		expected: func() *Builder {
			builder := New(dbForTest, WithLocales("en")).AutoMigrate()
			builder.RegisterSEO("Son1")
			builder.RegisterSEO("Son2")
			return builder
		}(),
	}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.builder
			expected := c.expected
			if len(actual.registeredSEO) != len(expected.registeredSEO) {
				t.Errorf("The length is not equal")
			}
			for _, desired := range expected.registeredSEO {
				if seo := actual.GetSEO(desired.name); seo == nil {
					t.Errorf("not found SEO %v in actual", desired.name)
				} else {
					if seo.parent == nil {
						if desired.parent != nil {
							t.Errorf("actual parent is nil, expected: %s", desired.parent.name)
						}
					} else {
						if seo.parent.name != desired.parent.name {
							t.Errorf("actual parent is %s, expected: %s", seo.parent.name, desired.parent.name)
						}
					}
				}
			}
		})
	}
}

func TestBuilder_SortSEOs(t *testing.T) {
	cases := []struct {
		name     string
		builder  *Builder
		data     []*QorSEOSetting
		expected []*QorSEOSetting
	}{
		{
			name: "with global seo",
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("PLP").AppendChildren(
					builder.RegisterSEO("Region"),
					builder.RegisterSEO("City"),
					builder.RegisterSEO("Prefecture"),
				)
				builder.RegisterSEO("Post")
				builder.RegisterSEO("Product")
				return builder
			}(),
			data: []*QorSEOSetting{
				{Name: "Post"},
				{Name: "Region"},
				{Name: "PLP"},
				{Name: defaultGlobalSEOName},
				{Name: "City"},
				{Name: "Prefecture"},
				{Name: "Product"},
			},
			expected: []*QorSEOSetting{
				{Name: defaultGlobalSEOName},
				{Name: "PLP"},
				{Name: "Region"},
				{Name: "City"},
				{Name: "Prefecture"},
				{Name: "Post"},
				{Name: "Product"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.builder.SortSEOs(c.data)
			r := testingutils.PrettyJsonDiff(c.expected, c.data)
			if r != "" {
				t.Error(r)
			}
		})
	}
}

func TestBuilder_BatchRender(t *testing.T) {
	u, _ := url.Parse("http://dev.qor5.com/product/1")
	defaultRequest := &http.Request{
		Method: "GET",
		URL:    u,
	}
	globalSeoSetting := []*QorSEOSetting{
		{
			Name: defaultGlobalSEOName,
			Setting: Setting{
				Title: "global | {{SiteName}}",
			},
			Variables: map[string]string{"SiteName": "Qor5 dev"},
			Locale:    l10n.Locale{LocaleCode: "en"},
		},
		{
			Name: defaultGlobalSEOName,
			Setting: Setting{
				Title: "全局 | {{SiteName}}",
			},
			Variables: map[string]string{"SiteName": "Qor5 开发"},
			Locale:    l10n.Locale{LocaleCode: "zh"},
		},
	}

	cases := []struct {
		name      string
		prepareDB func()
		builder   *Builder
		objs      interface{}
		wants     []string
	}{
		{
			name: "render_non-model_seo_with_setting_vars_and_default_context_vars",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | {{SiteName}}",
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				}
				if err := dbForTest.Save(&product).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", &Product{})
				return builder
			}(),
			objs: NewNonModelSEOSlice("Product"),
			wants: []string{
				`
			<title>product | Qor5 dev</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_non-model_seo_without_locale",
			prepareDB: func() {
				settings := []*QorSEOSetting{
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
					},
					{
						Name: defaultGlobalSEOName,
						Setting: Setting{
							Title: "global | {{SiteName}}",
						},
						Variables: map[string]string{"SiteName": "Qor5 dev"},
					},
				}
				if err := dbForTest.Save(&settings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", &Product{})
				return builder
			}(),
			objs: NewNonModelSEOSlice("Product"),
			wants: []string{
				`
			<title>product | Qor5 dev</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_multiple_seos_with_slice_of_values",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | {{SiteName}}",
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				}
				if err := dbForTest.Save(&product).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", &Product{}).RegisterContextVariable(
					"ProductName",
					func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.(*Product).Name
					},
				)
				return builder
			}(),
			objs: []Product{ // slice of values
				{
					Name: "productA",
					SEO: Setting{
						Title:            "productA",
						Description:      "{{SiteName}}",
						EnabledCustomize: true,
					},
				},
				{
					Name: "productB",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: true,
					},
				},
			},
			wants: []string{
				`
			<title>productA</title>
			<meta name='description' content='Qor5 dev'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>productB</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_multiple_seos_with_global_setting_variables_and_context_variables",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				product := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "product | {{SiteName}}",
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				}
				if err := dbForTest.Save(&product).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", &Product{}).RegisterContextVariable(
					"ProductName",
					func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.(*Product).Name
					},
				)
				return builder
			}(),
			objs: []*Product{
				{
					Name: "productA",
					SEO: Setting{
						Title:            "productA",
						Description:      "{{SiteName}}",
						EnabledCustomize: true,
					},
				},
				{
					Name: "productB",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: true,
					},
				},
			},
			wants: []string{
				`
			<title>productA</title>
			<meta name='description' content='Qor5 dev'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>productB</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_multiple_seos_with_three_levels_of_inheritance",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				settings := []*QorSEOSetting{
					{
						Name: "Default PLP",
						Setting: Setting{
							Title: "plp | {{SiteName}}",
						},
						Variables: map[string]string{
							// override SiteName var inherited from global seo
							"SiteName": "Qor5-PLP",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
				}
				if err := dbForTest.Save(&settings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Default PLP").RegisterSettingVariables("SiteName")
				builder.RegisterSEO("Product", Product{}).RegisterContextVariable(
					"ProductName",
					func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.(*Product).Name
					},
				).SetParent(builder.GetSEO("Default PLP"))
				return builder
			}(),
			objs: []*Product{
				{
					Name: "productA",
					SEO: Setting{
						Title:            "productA",
						Description:      "{{SiteName}}",
						EnabledCustomize: true,
					},
				},
				{
					Name: "productB",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: true,
					},
				},
				{
					Name: "productC",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: false,
					},
				},
			},
			wants: []string{
				`
			<title>productA</title>
			<meta name='description' content='Qor5-PLP'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>productB</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>product | Qor5-PLP</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_multiple_seos_with_different_locale",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				settings := []*QorSEOSetting{
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{ProductName}}",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "产品 | {{ProductName}}",
						},
						Locale: l10n.Locale{LocaleCode: "zh"},
					},
				}
				if err := dbForTest.Save(&settings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en", "zh")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", Product{}).RegisterContextVariable(
					"ProductName",
					func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.(*Product).Name
					},
				)
				return builder
			}(),
			objs: []*Product{
				{
					Name: "productA",
					SEO: Setting{
						Title:            "productA",
						Description:      "{{SiteName}}",
						EnabledCustomize: true,
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				},
				{
					Name: "产品A",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: true,
					},
					Locale: l10n.Locale{LocaleCode: "zh"},
				},
				{
					Name: "productB",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: false,
					},
					Locale: l10n.Locale{LocaleCode: "en"},
				},
				{
					Name: "产品B",
					SEO: Setting{
						Title:            "{{ProductName}}",
						EnabledCustomize: false,
					},
					Locale: l10n.Locale{LocaleCode: "zh"},
				},
			},
			wants: []string{
				`
			<title>productA</title>
			<meta name='description' content='Qor5 dev'>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>产品A</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>product | productB</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
				`
			<title>产品 | 产品B</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
		{
			name: "render_multiple_no_seos_with_different_locale",
			prepareDB: func() {
				if err := dbForTest.Save(&globalSeoSetting).Error; err != nil {
					panic(err)
				}
				settings := []*QorSEOSetting{
					{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{ProductName}}",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Product",
						Setting: Setting{
							Title: "产品 | {{ProductName}}",
						},
						Locale: l10n.Locale{LocaleCode: "zh"},
					},
				}
				if err := dbForTest.Save(&settings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en", "zh")).AutoMigrate()
				builder.GetGlobalSEO().RegisterMetaProperty(
					"og:url",
					func(_ interface{}, _ *Setting, req *http.Request) string {
						return req.URL.String()
					})
				builder.RegisterSEO("Product", Product{}).RegisterContextVariable(
					"ProductName",
					func(obj interface{}, _ *Setting, _ *http.Request) string {
						if product, ok := obj.(*Product); ok {
							return product.Name
						}
						return "ProductName"
					},
				)
				return builder
			}(),
			objs: NewNonModelSEOSlice("Product", "en", "zh"),
			wants: []string{
				`
			<title>product | ProductName</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`, `
			<title>产品 | ProductName</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetDB()
			c.prepareDB()
			comps := c.builder.BatchRender(c.objs, defaultRequest)
			for i, comp := range comps {
				got, _ := comp.MarshalHTML(context.TODO())
				if !metaEqual(string(got), c.wants[i]) {
					t.Errorf("Render = %v\nExpected = %v", string(got), c.wants[i])
				}
			}
		})
	}
}
