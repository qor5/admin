package examples_presets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestPresetsEditingSingletonNested(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSingletonNested(pb, TestDB)

	// Ensure clean state for singleton table
	if err := TestDB.Exec("DELETE FROM singleton_with_nesteds").Error; err != nil {
		t.Fatalf("failed to cleanup singleton_with_nesteds: %+v", err)
	}

	getID := func(t *testing.T) string {
		var rec SingletonWithNested
		if err := TestDB.First(&rec).Error; err != nil {
			t.Fatalf("failed to fetch singleton: %+v", err)
		}
		return fmt.Sprintf("%d", rec.ID)
	}

	// 1) Update title successfully
	t.Run("update title success", func(t *testing.T) {
		c := multipartestutils.TestCase{
			Name: "singleton update title",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
					AddField("Title", "hello").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusOK {
					t.Fatalf("expected 200, got %d", w.Code)
				}
				var rec SingletonWithNested
				if err := TestDB.First(&rec).Error; err != nil {
					t.Fatalf("failed to fetch singleton after update: %+v", err)
				}
				if rec.Title != "hello" {
					t.Fatalf("expected title 'hello', got %q", rec.Title)
				}
				if len(rec.Items) != 0 {
					t.Fatalf("expected no items, got %d", len(rec.Items))
				}
			},
		}
		multipartestutils.RunCase(t, c, pb)
	})

	// 2) Overlong title -> validation error
	t.Run("title too long shows validation error", func(t *testing.T) {
		id := getID(t)
		c := multipartestutils.TestCase{
			Name: "singleton title validation",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
					AddField(presets.ParamID, id).
					AddField("Title", "01234567890-exceed").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"title must not be longer than 10 characters"},
		}
		multipartestutils.RunCase(t, c, pb)
	})

	// 3) Add items then save
	t.Run("add items then save", func(t *testing.T) {
		id := getID(t)
		c := multipartestutils.TestCase{
			Name: "singleton add items",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
					AddField(presets.ParamID, id).
					AddField("Title", "hello").
					AddField("Items[0].Name", "A").
					AddField("Items[0].Value", "1").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var rec SingletonWithNested
				if err := TestDB.First(&rec).Error; err != nil {
					t.Fatalf("failed to fetch singleton after add: %+v", err)
				}
				if len(rec.Items) != 1 {
					t.Fatalf("expected 1 item, got %d", len(rec.Items))
				}
				if rec.Items[0] == nil || rec.Items[0].Name != "A" || rec.Items[0].Value != "1" {
					t.Fatalf("expected item {A,1}, got %+v", rec.Items[0])
				}
			},
		}
		multipartestutils.RunCase(t, c, pb)
	})

	// 4) Modify the item then save
	t.Run("modify items then save", func(t *testing.T) {
		id := getID(t)
		c := multipartestutils.TestCase{
			Name: "singleton modify items",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
					AddField(presets.ParamID, id).
					AddField("Title", "hello").
					AddField("Items[0].Name", "AA").
					AddField("Items[0].Value", "11").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var rec SingletonWithNested
				if err := TestDB.First(&rec).Error; err != nil {
					t.Fatalf("failed to fetch singleton after modify: %+v", err)
				}
				if len(rec.Items) != 1 {
					t.Fatalf("expected 1 item, got %d", len(rec.Items))
				}
				if rec.Items[0] == nil || rec.Items[0].Name != "AA" || rec.Items[0].Value != "11" {
					t.Fatalf("expected item {AA,11}, got %+v", rec.Items[0])
				}
			},
		}
		multipartestutils.RunCase(t, c, pb)
	})

	// 5) Delete the item then save
	t.Run("delete items then save", func(t *testing.T) {
		id := getID(t)
		c := multipartestutils.TestCase{
			Name: "singleton delete items",
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
					AddField(presets.ParamID, id).
					AddField("Title", "hello").
					// mark index 0 of Items as deleted
					AddField("__Deleted.Items", "0").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var rec SingletonWithNested
				if err := TestDB.First(&rec).Error; err != nil {
					t.Fatalf("failed to fetch singleton after delete: %+v", err)
				}
				if len(rec.Items) != 0 {
					t.Fatalf("expected 0 item, got %d", len(rec.Items))
				}
			},
		}
		multipartestutils.RunCase(t, c, pb)
	})

	// 6) Verify custom title is displayed
	t.Run("custom title is displayed", func(t *testing.T) {
		// Clean up and create fresh record
		if err := TestDB.Exec("DELETE FROM singleton_with_nesteds").Error; err != nil {
			t.Fatalf("failed to cleanup: %+v", err)
		}

		// First update the title to a known value
		updateReq := multipartestutils.NewMultipartBuilder().
			PageURL("/singleton-with-nesteds?__execute_event__=presets_Update").
			AddField("Title", "MyTitle").
			BuildEventFuncRequest()
		w := httptest.NewRecorder()
		pb.ServeHTTP(w, updateReq)

		// Then fetch the page and verify the custom title is rendered
		c := multipartestutils.TestCase{
			Name: "singleton custom title",
			ReqFunc: func() *http.Request {
				req := httptest.NewRequest("GET", "/singleton-with-nesteds", nil)
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				body := w.Body.String()
				// Check that the custom title format is present
				// The title should appear in the VToolbarTitle component
				if !strings.Contains(body, "Custom Title: MyTitle") {
					t.Errorf("Expected custom title 'Custom Title: MyTitle' in response, but not found. Body length: %d", len(body))
				}
			},
		}
		multipartestutils.RunCase(t, c, pb)
	})
}
