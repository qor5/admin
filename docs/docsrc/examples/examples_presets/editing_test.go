package examples_presets

import (
	"net/http"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var companyData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.companies (id, name) VALUES (12, 'terry_company');
`, []string{"companies"}))

func TestPresetsEditingValidate(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingValidate(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "editing create",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"timeout='2000'", "name must not be empty"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingSetter(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSetter(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "default field setterFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name must not be empty"},
		},
		{
			Name:  "wrap setter",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "system").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`You can not use \"system\" as name`},
		},
		{
			Name:  "setter return global error",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "global").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"You can not use global as name"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingCustomizationDescription(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingCustomizationDescription(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Customer Name`, `Customer Email`, `Description`, `vx-tiptap-editor`},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("Body", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					AddField("Body_richeditor_medialibrary.Values", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`},
		},
		{
			Name:  "do new without avatar",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
				mediaBuilder := media.New(TestDB).AutoMigrate()
				defer pb.Use(mediaBuilder)

				mb, _, _, _ := PresetsEditingCustomizationDescription(pb, TestDB)
				mb.Editing().Field("Avatar").WithContextValue(media.MediaBoxConfig, &media_library.MediaBoxConfig{})
				mb.Editing().WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
					return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
						err = in(obj, ctx)
						customer := obj.(*Customer)
						if customer.Avatar.FileName == "" || customer.Avatar.Url == "" {
							err.FieldError("Avatar.Values", "avatar must not be empty")
						}
						return
					}
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("Body", "").
					AddField("Avatar.Values", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					AddField("Body_richeditor_medialibrary.Values", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`, `avatar must not be empty`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingTiptap(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingTiptap(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Customer Name`, `Customer Email`, `Description`, `tiptap`},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("CompanyID", "0").
					AddField("Description", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`name must not be empty`, `email must not be empty`, `description must not be empty`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingSection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSection(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`section1`, "Name", "Create"},
			ExpectPortalUpdate0NotContains:     []string{"Save", "Edit"},
		},
		{
			Name:  "do new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "Terry").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Terry"},
		},
		{
			Name:  "detailing section use editing validator",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=section_save_section1").
					AddField("Name", "terryterryterry").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`too long name`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingHTMLSanitizer(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingCustomizationDescription(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "HTML sanitizer tiptap policy - allow safe tags",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "John Doe").
					AddField("Email", "john@example.com").
					AddField("Description", "Normal description").
					AddField("HTMLSanitizerPolicyTiptapInput", "<p>Hello <strong>world</strong>!</p><script>alert('xss')</script>").
					AddField("HTMLSanitizerPolicyUGCInput", "<p>Test content</p>").
					AddField("HTMLSanitizerPolicyStrictInput", "<p>Test content</p>").
					AddField("HTMLSanitizerPolicyCustomInput", "<video controls><source src='test.mp4'></video>").
					AddField("CompanyID", "0").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"John Doe"},
		},
		{
			Name:  "HTML sanitizer UGC policy - filter script tags",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "Jane Doe").
					AddField("Email", "jane@example.com").
					AddField("Description", "Normal description").
					AddField("HTMLSanitizerPolicyTiptapInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyUGCInput", "<p>User content with <script>alert('hack')</script> embedded</p>").
					AddField("HTMLSanitizerPolicyStrictInput", "<p>Test content</p>").
					AddField("HTMLSanitizerPolicyCustomInput", "<video controls><source src='test.mp4'></video>").
					AddField("CompanyID", "0").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Jane Doe"},
		},
		{
			Name:  "HTML sanitizer strict policy - very restrictive",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "Bob Smith").
					AddField("Email", "bob@example.com").
					AddField("Description", "Normal description").
					AddField("HTMLSanitizerPolicyTiptapInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyUGCInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyStrictInput", "<p>Only text allowed</p><img src='test.jpg'>").
					AddField("HTMLSanitizerPolicyCustomInput", "<video controls><source src='test.mp4'></video>").
					AddField("CompanyID", "0").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Bob Smith"},
		},
		{
			Name:  "HTML sanitizer custom policy - allow custom video tags",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "Alice Johnson").
					AddField("Email", "alice@example.com").
					AddField("Description", "Normal description").
					AddField("HTMLSanitizerPolicyTiptapInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyUGCInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyStrictInput", "<p>Content</p>").
					AddField("HTMLSanitizerPolicyCustomInput", "<video controls src='video.mp4'>Your browser does not support video.</video><script>alert('bad')</script>").
					AddField("CompanyID", "0").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Alice Johnson"},
		},
		{
			Name:  "HTML sanitizer validation - empty fields",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "").
					AddField("Email", "").
					AddField("Description", "").
					AddField("HTMLSanitizerPolicyTiptapInput", "").
					AddField("HTMLSanitizerPolicyUGCInput", "").
					AddField("HTMLSanitizerPolicyStrictInput", "").
					AddField("HTMLSanitizerPolicyCustomInput", "").
					AddField("CompanyID", "0").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"name must not be empty",
				"email must not be empty",
				"description must not be empty",
				"HTMLSanitizerPolicyTiptapInput must not be empty",
				"HTMLSanitizerPolicyUGCInput must not be empty",
				"HTMLSanitizerPolicyStrictInput must not be empty",
				"HTMLSanitizerPolicyCustomInput must not be empty",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestHTMLSanitizerPolicyTypes(t *testing.T) {
	testCases := []struct {
		name          string
		policyType    presets.HTMLSanitizerPolicyType
		input         string
		expectAllowed []string
		expectBlocked []string
	}{
		{
			name:          "Tiptap policy allows rich content",
			policyType:    presets.HTMLSanitizerPolicyTiptap,
			input:         "<p>Hello <strong>world</strong>!</p><h1>Title</h1><a href='http://example.com'>Link</a><script>alert('xss')</script>",
			expectAllowed: []string{"<p>", "<strong>", "<h1>", "<a"},
			expectBlocked: []string{"<script>"},
		},
		{
			name:          "UGC policy allows user content",
			policyType:    presets.HTMLSanitizerPolicyUGC,
			input:         "<p>User content</p><strong>Bold</strong><script>alert('hack')</script><iframe src='evil.com'></iframe>",
			expectAllowed: []string{"<p>", "<strong>"},
			expectBlocked: []string{"<script>", "<iframe"},
		},
		{
			name:          "Strict policy is very restrictive",
			policyType:    presets.HTMLSanitizerPolicyStrict,
			input:         "<p>Text only</p><strong>Bold</strong><img src='image.jpg'><script>evil()</script>",
			expectAllowed: []string{},
			expectBlocked: []string{"<p>", "<strong>", "<img>", "<script>"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy := presets.CreateHTMLSanitizerPolicy(tc.policyType)
			result := policy.Sanitize(tc.input)

			for _, allowed := range tc.expectAllowed {
				if !strings.Contains(result, allowed) {
					t.Errorf("Expected %q to be allowed in result, but got: %s", allowed, result)
				}
			}

			for _, blocked := range tc.expectBlocked {
				if strings.Contains(result, blocked) {
					t.Errorf("Expected %q to be blocked, but found in result: %s", blocked, result)
				}
			}
		})
	}
}

func TestCreateDefaultTiptapSanitizerPolicy(t *testing.T) {
	policy := presets.CreateDefaultTiptapSanitizerPolicy()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Allow basic formatting",
			input:    "<p>Hello <strong>world</strong>!</p>",
			expected: "<p>Hello <strong>world</strong>!</p>",
		},
		{
			name:     "Allow headings",
			input:    "<h1>Title</h1><h2>Subtitle</h2>",
			expected: "<h1>Title</h1><h2>Subtitle</h2>",
		},
		{
			name:     "Allow lists",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "<ul><li>Item 1</li><li>Item 2</li></ul>",
		},
		{
			name:     "Allow safe links",
			input:    "<a href='https://example.com' target='_blank'>Link</a>",
			expected: "<a href=\"https://example.com\" target=\"_blank\">Link</a>",
		},
		{
			name:     "Allow images with safe attributes",
			input:    "<img src='image.jpg' alt='Description' width='100'>",
			expected: "<img src=\"image.jpg\" alt=\"Description\" width=\"100\">",
		},
		{
			name:     "Allow video and audio",
			input:    "<video controls><source src='video.mp4' type='video/mp4'></video>",
			expected: "<video controls=\"\"><source src=\"video.mp4\" type=\"video/mp4\"></video>",
		},
		{
			name:     "Block dangerous scripts",
			input:    "<p>Safe content</p><script>alert('xss')</script>",
			expected: "<p>Safe content</p>",
		},
		{
			name:     "Block dangerous events",
			input:    "<p onclick='alert(1)'>Click me</p>",
			expected: "<p>Click me</p>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := policy.Sanitize(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCreateHTMLSanitizerSetterFunc(t *testing.T) {
	// Test the CreateHTMLSanitizer function returns a proper setter function
	policy := presets.CreateHTMLSanitizerPolicy(presets.HTMLSanitizerPolicyTiptap)
	setter := presets.CreateHTMLSanitizer(&presets.HTMLSanitizerConfig{
		Policy: policy,
	})

	// Create a mock object and context
	customer := &Customer{}
	field := &presets.FieldContext{
		Name:    "HTMLSanitizerPolicyTiptapInput",
		FormKey: "HTMLSanitizerPolicyTiptapInput",
	}

	// Create a mock HTTP request with form data
	req := multipartestutils.NewMultipartBuilder().
		AddField("HTMLSanitizerPolicyTiptapInput", "<p>Hello <strong>world</strong>!</p><script>alert('xss')</script>").
		BuildEventFuncRequest()

	// Parse the form to ensure form values are available
	req.ParseForm()
	req.ParseMultipartForm(32 << 20) // 32 MB

	ctx := &web.EventContext{R: req}

	// Call the setter function
	err := setter(customer, field, ctx)
	if err != nil {
		t.Errorf("Setter function returned error: %v", err)
	}

	// Check that the field was set with sanitized content
	expected := "<p>Hello <strong>world</strong>!</p>"
	if customer.HTMLSanitizerPolicyTiptapInput != expected {
		t.Errorf("Expected %q, got %q", expected, customer.HTMLSanitizerPolicyTiptapInput)
	}
}

func TestPresetsEditingSaver(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSaverValidation(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name: "saver return error",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "system").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`You can not use system as name`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
