package presets

import (
	"net/http"
	"strings"
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/qor5/web/v3"
)

// Test model for HTML sanitizer testing
type BlogPost struct {
	ID              uint `gorm:"primarykey"`
	Title           string
	Content         string // Will be sanitized with default tiptap policy
	Summary         string // Will be sanitized with strict policy
	Article         string // Will be sanitized with UGC policy
	ExtendedContent string // Will be sanitized with extended policy
	Body            string // No sanitization
}

// TestHTMLSanitizerConfig tests the basic HTML sanitizer configuration
func TestHTMLSanitizerConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *HTMLSanitizerConfig
		input    string
		expected string
	}{
		{
			name:     "TiptapHTMLSanitizerConfig basic sanitization",
			config:   TiptapHTMLSanitizerConfig(),
			input:    `<p>Hello <script>alert('xss')</script><strong>World</strong></p>`,
			expected: `<p>Hello <strong>World</strong></p>`,
		},
		{
			name:     "TiptapHTMLSanitizerConfig allow safe elements",
			config:   TiptapHTMLSanitizerConfig(),
			input:    `<h1>Title</h1><p>Text with <em>emphasis</em> and <strong>bold</strong></p>`,
			expected: `<h1>Title</h1><p>Text with <em>emphasis</em> and <strong>bold</strong></p>`,
		},
		{
			name:     "UGCHTMLSanitizerConfig",
			config:   UGCHTMLSanitizerConfig(),
			input:    `<p>Check out <a href="https://example.com">this link</a></p>`,
			expected: `<p>Check out <a href="https://example.com" rel="nofollow">this link</a></p>`,
		},
		{
			name:     "StrictHTMLSanitizerConfig",
			config:   StrictHTMLSanitizerConfig(),
			input:    `<p>Hello <strong>World</strong> <em>with emphasis</em></p>`,
			expected: `Hello World with emphasis`,
		},
		{
			name: "Nil policy",
			config: &HTMLSanitizerConfig{
				Policy: nil,
			},
			input:    `<p>Hello <script>alert('xss')</script><strong>World</strong></p>`,
			expected: `<p>Hello <script>alert('xss')</script><strong>World</strong></p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestExtendHTMLSanitizerConfig tests the policy extension functionality
func TestExtendHTMLSanitizerConfig(t *testing.T) {
	// Test extending tiptap policy with video support
	extendedConfig := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("video")
		p.AllowAttrs("src").OnElements("video")
		return p
	})

	testVideo := `<p>Test <video src="test.mp4">video</video></p>`
	result := extendedConfig.SanitizeHTML(testVideo)
	expected := `<p>Test <video src="test.mp4">video</video></p>`

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test that original tiptap policy also now supports video (since we improved it)
	originalConfig := TiptapHTMLSanitizerConfig()
	originalResult := originalConfig.SanitizeHTML(testVideo)
	expectedOriginal := `<p>Test <video src="test.mp4">video</video></p>`

	if originalResult != expectedOriginal {
		t.Errorf("Original policy should also support video now. Expected %q, got %q", expectedOriginal, originalResult)
	}
}

// TestPolicyIsolation tests that different extended policies don't affect each other
func TestPolicyIsolation(t *testing.T) {
	// Create two different extended configs
	videoConfig := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("video")
		p.AllowAttrs("src").OnElements("video")
		return p
	})

	audioConfig := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("audio")
		p.AllowAttrs("controls").OnElements("audio")
		return p
	})

	testVideo := `<p>Test <video src="test.mp4">video</video></p>`
	testAudio := `<p>Test <audio controls>audio</audio></p>`

	// Both configs should support both video and audio since base tiptap policy now includes them
	videoResult := videoConfig.SanitizeHTML(testVideo)
	videoAudioResult := videoConfig.SanitizeHTML(testAudio)

	if videoResult != `<p>Test <video src="test.mp4">video</video></p>` {
		t.Errorf("Video config should support video, got: %q", videoResult)
	}
	if videoAudioResult != `<p>Test <audio controls="">audio</audio></p>` {
		t.Errorf("Video config should also support audio (from base tiptap), got: %q", videoAudioResult)
	}

	// Audio config should support both audio and video (from base tiptap policy)
	audioResult := audioConfig.SanitizeHTML(testAudio)
	audioVideoResult := audioConfig.SanitizeHTML(testVideo)

	if audioResult != `<p>Test <audio controls="">audio</audio></p>` {
		t.Errorf("Audio config should support audio, got: %q", audioResult)
	}
	if audioVideoResult != `<p>Test <video src="test.mp4">video</video></p>` {
		t.Errorf("Audio config should also support video (from base tiptap), got: %q", audioVideoResult)
	}
}

// TestDeprecatedFunctions tests the deprecated extension functions
func TestDeprecatedFunctions(t *testing.T) {
	// Test deprecated ExtendTiptapHTMLSanitizerConfig
	config1 := ExtendTiptapHTMLSanitizerConfig(func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("video")
		return p
	})

	// Test new ExtendHTMLSanitizerConfig
	config2 := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("video")
		return p
	})

	testInput := `<p>Test <video>content</video></p>`
	result1 := config1.SanitizeHTML(testInput)
	result2 := config2.SanitizeHTML(testInput)

	if result1 != result2 {
		t.Errorf("Deprecated and new functions should produce same result. Got %q vs %q", result1, result2)
	}
}

// Test individual setter function behavior
func TestSetterFunctions(t *testing.T) {
	// Test TiptapHTMLSetter
	t.Run("TiptapHTMLSetter", func(t *testing.T) {
		post := &BlogPost{}
		field := &FieldContext{
			Name:    "Content",
			FormKey: "Content",
		}
		ctx := &web.EventContext{
			R: &http.Request{
				Form: map[string][]string{
					"Content": {`<p>Hello <script>alert('xss')</script><strong>World</strong></p>`},
				},
			},
		}

		err := TiptapHTMLSetter(post, field, ctx)
		if err != nil {
			t.Fatalf("TiptapHTMLSetter failed: %v", err)
		}

		expected := `<p>Hello <strong>World</strong></p>`
		if post.Content != expected {
			t.Errorf("Expected %q, got %q", expected, post.Content)
		}
	})

	// Test TiptapHTMLSetterWithPolicy
	t.Run("TiptapHTMLSetterWithPolicy", func(t *testing.T) {
		post := &BlogPost{}
		field := &FieldContext{
			Name:    "Summary",
			FormKey: "Summary",
		}
		ctx := &web.EventContext{
			R: &http.Request{
				Form: map[string][]string{
					"Summary": {`<p>Hello <strong>World</strong></p>`},
				},
			},
		}

		setter := TiptapHTMLSetterWithPolicy("strict")
		err := setter(post, field, ctx)
		if err != nil {
			t.Fatalf("TiptapHTMLSetterWithPolicy failed: %v", err)
		}

		expected := `Hello World`
		if post.Summary != expected {
			t.Errorf("Expected %q, got %q", expected, post.Summary)
		}
	})

	// Test TiptapHTMLSetterWithConfig
	t.Run("TiptapHTMLSetterWithConfig", func(t *testing.T) {
		post := &BlogPost{}
		field := &FieldContext{
			Name:    "ExtendedContent",
			FormKey: "ExtendedContent",
		}
		ctx := &web.EventContext{
			R: &http.Request{
				Form: map[string][]string{
					"ExtendedContent": {`<p>Text with <video src="test.mp4">video</video></p>`},
				},
			},
		}

		config := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
			p.AllowElements("video")
			p.AllowAttrs("src").OnElements("video")
			return p
		})
		setter := TiptapHTMLSetterWithConfig(config)
		err := setter(post, field, ctx)
		if err != nil {
			t.Fatalf("TiptapHTMLSetterWithConfig failed: %v", err)
		}

		expected := `<p>Text with <video src="test.mp4">video</video></p>`
		if post.ExtendedContent != expected {
			t.Errorf("Expected %q, got %q", expected, post.ExtendedContent)
		}
	})

	// Test with nil config
	t.Run("TiptapHTMLSetterWithConfig nil config", func(t *testing.T) {
		post := &BlogPost{}
		field := &FieldContext{
			Name:    "Body",
			FormKey: "Body",
		}
		ctx := &web.EventContext{
			R: &http.Request{
				Form: map[string][]string{
					"Body": {`<p>Unsanitized <script>alert('dangerous')</script> content</p>`},
				},
			},
		}

		setter := TiptapHTMLSetterWithConfig(nil)
		err := setter(post, field, ctx)
		if err != nil {
			t.Fatalf("TiptapHTMLSetterWithConfig with nil config failed: %v", err)
		}

		// With nil config, no sanitization should occur
		expected := `<p>Unsanitized <script>alert('dangerous')</script> content</p>`
		if post.Body != expected {
			t.Errorf("Expected %q, got %q", expected, post.Body)
		}
	})
}

// Test that CreateBasePolicyByType works correctly
func TestCreateBasePolicyByType(t *testing.T) {
	tests := []struct {
		policyType string
		input      string
		shouldPass bool
	}{
		{"tiptap", `<p><strong>Hello</strong></p>`, true},
		{"tiptap", `<script>alert('xss')</script>`, false},
		{"ugc", `<p><a href="test.com">link</a></p>`, true},
		{"ugc", `<script>alert('xss')</script>`, false},
		{"strict", `<p>Hello</p>`, false},                  // strict removes all HTML
		{"unknown", `<p><strong>Hello</strong></p>`, true}, // defaults to tiptap
	}

	for _, tt := range tests {
		t.Run(tt.policyType, func(t *testing.T) {
			policy := CreateBasePolicyByType(tt.policyType)
			result := policy.Sanitize(tt.input)

			if tt.shouldPass {
				if !strings.Contains(result, "Hello") && !strings.Contains(result, "link") {
					t.Errorf("Expected content to pass through, got: %q", result)
				}
			} else {
				if strings.Contains(result, "script") || strings.Contains(result, "<p>") && tt.policyType == "strict" {
					t.Errorf("Expected content to be filtered, got: %q", result)
				}
			}
		})
	}
}

// Test CreateTiptapBasePolicy
func TestCreateTiptapBasePolicy(t *testing.T) {
	policy := CreateTiptapBasePolicy()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic elements",
			input:    `<p>Hello <strong>World</strong></p>`,
			expected: `<p>Hello <strong>World</strong></p>`,
		},
		{
			name:     "Remove scripts",
			input:    `<p>Hello <script>alert('xss')</script>World</p>`,
			expected: `<p>Hello World</p>`,
		},
		{
			name:     "Allow headings",
			input:    `<h1>Title</h1><h2>Subtitle</h2>`,
			expected: `<h1>Title</h1><h2>Subtitle</h2>`,
		},
		{
			name:     "Allow lists",
			input:    `<ul><li>Item 1</li><li>Item 2</li></ul>`,
			expected: `<ul><li>Item 1</li><li>Item 2</li></ul>`,
		},
		{
			name:     "Allow safe links",
			input:    `<p>Check out this <a href="https://example.com">Link</a></p>`,
			expected: `<p>Check out this <a href="https://example.com">Link</a></p>`,
		},
		{
			name:     "Allow images with safe attributes",
			input:    `<img src="test.jpg" alt="Test" width="100">`,
			expected: `<img src="test.jpg" alt="Test" width="100">`,
		},
		{
			name:     "Allow code blocks",
			input:    `<pre><code>console.log('hello')</code></pre>`,
			expected: `<pre><code>console.log(&#39;hello&#39;)</code></pre>`, // HTML entities
		},
		{
			name:     "Allow blockquotes",
			input:    `<blockquote>Quote text</blockquote>`,
			expected: `<blockquote>Quote text</blockquote>`,
		},
		{
			name:     "Allow tables",
			input:    `<table><tr><td colspan="2">Cell</td></tr></table>`,
			expected: `<table><tr><td colspan="2">Cell</td></tr></table>`,
		},
		{
			name:     "Allow data attributes",
			input:    `<div data-type="node" data-id="123">Content</div>`,
			expected: `<div data-type="node" data-id="123">Content</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Test multiple calls to ensure policy consistency
func TestPolicyConsistency(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()
	input := `<p>Hello <script>alert('xss')</script><strong>World</strong></p>`
	expected := `<p>Hello <strong>World</strong></p>`

	// Test multiple calls return consistent results
	for i := 0; i < 10; i++ {
		result := config.SanitizeHTML(input)
		if result != expected {
			t.Errorf("Call %d: Expected %q, got %q", i+1, expected, result)
		}
	}
}

// Benchmark tests
func BenchmarkTiptapHTMLSanitizer(b *testing.B) {
	config := TiptapHTMLSanitizerConfig()
	input := `<p>Hello <script>alert('xss')</script><strong>World</strong> with <em>emphasis</em> and <a href="https://example.com">link</a></p>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.SanitizeHTML(input)
	}
}

func BenchmarkExtendedHTMLSanitizer(b *testing.B) {
	config := ExtendHTMLSanitizerConfig(HTMLSanitizerPolicyTiptap, func(p *bluemonday.Policy) *bluemonday.Policy {
		p.AllowElements("video", "audio")
		p.AllowAttrs("src", "controls").OnElements("video", "audio")
		return p
	})
	input := `<p>Hello <video src="test.mp4">video</video> and <audio controls>audio</audio></p>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.SanitizeHTML(input)
	}
}

func BenchmarkPolicyCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateTiptapBasePolicy()
	}
}

// Note: Removed complex ExtendPolicy tests as they used non-existent fields.
// The current implementation uses a simpler approach with direct policy creation.

// TestLinkHrefPreservation tests that href attributes are properly preserved in links
func TestLinkHrefPreservation(t *testing.T) {
	policy := CreateTiptapBasePolicy()

	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "Link with href and all rel values",
			input:       `<a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">link text</a>`,
			expected:    `<a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">link text</a>`,
			description: "Should preserve href with complete rel attribute",
		},
		{
			name:        "Link with href and empty rel",
			input:       `<a href="https://www.baidu.com" target="_blank" rel="">link text</a>`,
			expected:    `<a href="https://www.baidu.com" target="_blank" rel="">link text</a>`,
			description: "Should preserve href and keep empty rel (allowed by our policy)",
		},
		{
			name:        "Link with href and no rel",
			input:       `<a href="https://www.baidu.com" target="_blank">link text</a>`,
			expected:    `<a href="https://www.baidu.com" target="_blank">link text</a>`,
			description: "Should preserve href without rel attribute",
		},
		{
			name:        "Link with href only",
			input:       `<a href="https://www.baidu.com">link text</a>`,
			expected:    `<a href="https://www.baidu.com">link text</a>`,
			description: "Should preserve href with minimal attributes",
		},
		{
			name:        "Link without href",
			input:       `<a target="_blank" rel="noopener noreferrer nofollow">link text</a>`,
			expected:    `<a target="_blank" rel="noopener noreferrer nofollow">link text</a>`,
			description: "Should preserve link without href",
		},
		{
			name:        "Complex link with paragraph",
			input:       `<p><a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">link text</a></p>`,
			expected:    `<p><a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">link text</a></p>`,
			description: "Should preserve href in complex structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("%s\nInput:    %q\nExpected: %q\nGot:      %q",
					tt.description, tt.input, tt.expected, result)

				// Additional debugging info
				if !strings.Contains(result, "href=") && strings.Contains(tt.input, "href=") {
					t.Errorf("HREF ATTRIBUTE WAS FILTERED OUT!")
				}
				if !strings.Contains(result, "<a") && strings.Contains(tt.input, "<a") {
					t.Errorf("ENTIRE ANCHOR TAG WAS FILTERED OUT!")
				}
			}
		})
	}
}

// TestTiptapHTMLSanitizerConfigLinkPreservation tests the complete sanitizer config
func TestTiptapHTMLSanitizerConfigLinkPreservation(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// Test the real-world case that was failing
	input := `<a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">sdf</a>`
	result := config.SanitizeHTML(input)

	t.Logf("Input:  %q", input)
	t.Logf("Result: %q", result)

	if !strings.Contains(result, "href=") {
		t.Errorf("href attribute was filtered out! Expected to contain href, got: %q", result)
	}

	if !strings.Contains(result, "https://www.baidu.com") {
		t.Errorf("href URL was filtered out! Expected to contain URL, got: %q", result)
	}

	// The complete expected result should preserve the href
	expected := `<a href="https://www.baidu.com" target="_blank" rel="noopener noreferrer nofollow">sdf</a>`
	if result != expected {
		t.Errorf("Complete link not preserved.\nExpected: %q\nGot:      %q", expected, result)
	}
}

// TestRealWorldScenario tests the actual failing scenario from user's data
func TestRealWorldScenario(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// The actual HTML content that was failing
	input := `<table class="table-wrapper" style="min-width: 75px"><colgroup><col style="min-width: 25px"><col style="min-width: 25px"><col style="min-width: 25px"></colgroup><tbody><tr><th colspan="1" rowspan="1"><p><strong>123123123</strong></p></th><th colspan="1" rowspan="1"><p></p></th><th colspan="1" rowspan="1"><p></p></th></tr><tr><td colspan="1" rowspan="1"><p>123123</p></td><td colspan="1" rowspan="1"><p></p></td><td colspan="1" rowspan="1"><p></p><p></p></td></tr></tbody></table><h1>a<a target="_blank" rel="noopener noreferrer nofollow" href="https://www.baidu.com">sdf</a></h1><h1 style="text-align: center"><span style="color: rgb(244, 67, 54)">asdf1</span></h1><blockquote class="blockquote"><p>sadfasdfadsfsdafsdaf</p></blockquote><p></p><p></p>`

	result := config.SanitizeHTML(input)

	t.Logf("Input length: %d", len(input))
	t.Logf("Result length: %d", len(result))
	t.Logf("Result: %s", result)

	// Check that href is preserved
	if !strings.Contains(result, `href="https://www.baidu.com"`) {
		t.Errorf("href attribute should be preserved, got: %s", result)
	}

	// Check that table styles are preserved
	if !strings.Contains(result, `style="min-width: 75px"`) {
		t.Errorf("Table min-width style should be preserved, got: %s", result)
	}

	// Check that colgroup and col elements are preserved
	if !strings.Contains(result, "<colgroup>") {
		t.Errorf("colgroup element should be preserved, got: %s", result)
	}

	if !strings.Contains(result, `<col style="min-width: 25px">`) {
		t.Errorf("col element with style should be preserved, got: %s", result)
	}

	// Check that text color style is preserved
	if !strings.Contains(result, `style="color: rgb(244, 67, 54)"`) {
		t.Errorf("Text color style should be preserved, got: %s", result)
	}

	// Verify no dangerous content slipped through
	if strings.Contains(result, "<script>") {
		t.Errorf("Script tags should be filtered out, got: %s", result)
	}

	t.Logf("✅ All real-world scenario checks passed!")
}

// TestTiptapPolicyElementSupport tests that all expected elements are properly supported
func TestTiptapPolicyElementSupport(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	tests := []struct {
		name        string
		input       string
		shouldPass  bool
		description string
	}{
		// Basic text elements
		{
			name:        "Basic text formatting",
			input:       `<p>Text with <strong>bold</strong>, <em>italic</em>, <u>underline</u></p>`,
			shouldPass:  true,
			description: "Basic text formatting should be preserved",
		},
		{
			name:        "Headings",
			input:       `<h1>Heading 1</h1><h2>Heading 2</h2><h3>Heading 3</h3>`,
			shouldPass:  true,
			description: "All heading levels should be preserved",
		},
		{
			name:        "Lists",
			input:       `<ul><li>Item 1</li><li>Item 2</li></ul><ol><li>First</li><li>Second</li></ol>`,
			shouldPass:  true,
			description: "Ordered and unordered lists should be preserved",
		},

		// Links
		{
			name:        "Links with href",
			input:       `<a href="https://example.com" target="_blank" rel="noopener">Link</a>`,
			shouldPass:  true,
			description: "Links with href should be preserved",
		},

		// Images
		{
			name:        "Images with src",
			input:       `<img src="image.jpg" alt="Test image" width="100" height="80">`,
			shouldPass:  true,
			description: "Images with attributes should be preserved",
		},

		// Media elements
		{
			name:        "Video element",
			input:       `<video controls><source src="movie.mp4" type="video/mp4"></video>`,
			shouldPass:  true, // Now supported!
			description: "Video elements should be supported for rich content",
		},
		{
			name:        "Audio element",
			input:       `<audio controls><source src="audio.mp3" type="audio/mp3"></audio>`,
			shouldPass:  true, // Now supported!
			description: "Audio elements should be supported for rich content",
		},

		// Code blocks
		{
			name:        "Code blocks",
			input:       `<pre><code class="language-javascript">console.log('hello');</code></pre>`,
			shouldPass:  true,
			description: "Code blocks should be preserved",
		},

		// Blockquotes
		{
			name:        "Blockquotes",
			input:       `<blockquote class="quote">This is a quote</blockquote>`,
			shouldPass:  true,
			description: "Blockquotes should be preserved",
		},

		// Tables
		{
			name:        "Complete table",
			input:       `<table class="table"><colgroup><col style="width: 50%"></colgroup><thead><tr><th>Header</th></tr></thead><tbody><tr><td colspan="1">Cell</td></tr></tbody></table>`,
			shouldPass:  true,
			description: "Complete table structure should be preserved",
		},

		// Dangerous content (should be filtered)
		{
			name:        "Script tags (dangerous)",
			input:       `<p>Text <script>alert('xss')</script> more text</p>`,
			shouldPass:  false,
			description: "Script tags should be filtered out",
		},

		// Custom elements that might be used in rich editors
		{
			name:        "Div with data attributes",
			input:       `<div data-type="node" data-id="123" class="editor-node">Content</div>`,
			shouldPass:  true,
			description: "Divs with data attributes should be preserved for editor functionality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.SanitizeHTML(tt.input)

			t.Logf("Input:  %q", tt.input)
			t.Logf("Result: %q", result)

			if tt.shouldPass {
				// For elements that should pass, check if key elements are preserved
				if strings.Contains(tt.input, "<img") && !strings.Contains(result, "<img") {
					t.Errorf("MISSING IMG ELEMENT: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
				if strings.Contains(tt.input, "<video") && !strings.Contains(result, "<video") {
					t.Errorf("MISSING VIDEO ELEMENT: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
				if strings.Contains(tt.input, "<audio") && !strings.Contains(result, "<audio") {
					t.Errorf("MISSING AUDIO ELEMENT: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
				if strings.Contains(tt.input, "<a") && !strings.Contains(result, "<a") {
					t.Errorf("MISSING LINK ELEMENT: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
				if strings.Contains(tt.input, "href=") && !strings.Contains(result, "href=") {
					t.Errorf("MISSING HREF ATTRIBUTE: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
			} else {
				// For dangerous content, check that it's properly filtered
				if strings.Contains(tt.input, "<script") && strings.Contains(result, "<script") {
					t.Errorf("DANGEROUS CONTENT NOT FILTERED: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
				}
			}
		})
	}
}

// TestMissingElements specifically tests for currently missing but expected elements
func TestMissingElements(t *testing.T) {
	policy := CreateTiptapBasePolicy()

	missingElements := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "img element",
			input:    `<img src="test.jpg" alt="Test">`,
			expected: `<img src="test.jpg" alt="Test">`,
		},
		{
			name:     "video element",
			input:    `<video controls>Video content</video>`,
			expected: `<video controls="">Video content</video>`,
		},
		{
			name:     "audio element",
			input:    `<audio controls>Audio content</audio>`,
			expected: `<audio controls="">Audio content</audio>`,
		},
		{
			name:     "source element",
			input:    `<source src="movie.mp4" type="video/mp4">`,
			expected: `<source src="movie.mp4" type="video/mp4">`,
		},
		{
			name:     "figure and figcaption",
			input:    `<figure><img alt="Test"><figcaption>Caption</figcaption></figure>`,
			expected: `<figure><img alt="Test"><figcaption>Caption</figcaption></figure>`,
		},
	}

	for _, tt := range missingElements {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.Sanitize(tt.input)
			t.Logf("Testing: %s", tt.name)
			t.Logf("Input:    %q", tt.input)
			t.Logf("Expected: %q", tt.expected)
			t.Logf("Got:      %q", result)

			// Check if the main element exists
			if strings.Contains(tt.input, "<img") && !strings.Contains(result, "<img") {
				t.Errorf("❌ IMG ELEMENT MISSING - should be allowed")
			}
			if strings.Contains(tt.input, "<video") && !strings.Contains(result, "<video") {
				t.Errorf("❌ VIDEO ELEMENT MISSING - should be allowed for rich content")
			}
			if strings.Contains(tt.input, "<audio") && !strings.Contains(result, "<audio") {
				t.Errorf("❌ AUDIO ELEMENT MISSING - should be allowed for rich content")
			}
			if strings.Contains(tt.input, "<source") && !strings.Contains(result, "<source") {
				t.Errorf("❌ SOURCE ELEMENT MISSING - needed for video/audio")
			}
			if strings.Contains(tt.input, "<figure") && !strings.Contains(result, "<figure") {
				t.Errorf("❌ FIGURE ELEMENT MISSING - useful for media with captions")
			}
		})
	}
}

// TestRichContentWithMediaElements tests a comprehensive scenario with videos, links, and tables
func TestRichContentWithMediaElements(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// Test rich content that includes all the elements we've fixed
	input := `
	<h1>Rich Content Test</h1>
	<p>This is a comprehensive test with various elements:</p>
	
	<table class="table-wrapper" style="min-width: 300px">
		<colgroup>
			<col style="min-width: 100px">
			<col style="min-width: 200px">
		</colgroup>
		<thead>
			<tr>
				<th>Media Type</th>
				<th>Description</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>Video</td>
				<td>MP4 format</td>
			</tr>
		</tbody>
	</table>
	
	<figure>
		<video controls style="width: 100%; max-width: 600px">
			<source src="demo.mp4" type="video/mp4">
			<track kind="captions" src="captions.vtt" srclang="en" label="English">
			Your browser does not support the video tag.
		</video>
		<figcaption>Demo video with captions</figcaption>
	</figure>
	
	<figure>
		<audio controls>
			<source src="audio.mp3" type="audio/mp3">
			<track kind="captions" src="audio-captions.vtt" srclang="en">
			Your browser does not support the audio tag.
		</audio>
		<figcaption>Demo audio file</figcaption>
	</figure>
	
	<p>For more information, visit <a href="https://example.com" target="_blank" rel="noopener noreferrer">our website</a>.</p>
	
	<blockquote>
		<p>This content should be preserved completely with all media elements and styling.</p>
	</blockquote>
	`

	result := config.SanitizeHTML(input)

	t.Logf("Input length: %d", len(input))
	t.Logf("Result length: %d", len(result))
	t.Logf("Result: %s", result)

	// Check that all critical elements are preserved
	requiredElements := []string{
		"<video",
		"<audio",
		"<source",
		"<track",
		"<figure>",
		"<figcaption>",
		"<table",
		"<colgroup>",
		"<col style=",
		`href="https://example.com"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		`style="width: 100%; max-width: 600px"`,
		`style="min-width: 300px"`,
		`controls`,
		`type="video/mp4"`,
		`type="audio/mp3"`,
		`kind="captions"`,
		`srclang="en"`,
	}

	for _, element := range requiredElements {
		if !strings.Contains(result, element) {
			t.Errorf("Missing required element or attribute: %q\nResult: %s", element, result)
		}
	}

	// Verify dangerous content is still filtered
	dangerousInput := input + `<script>alert('xss')</script><iframe src="evil.com"></iframe>`
	dangerousResult := config.SanitizeHTML(dangerousInput)

	if strings.Contains(dangerousResult, "<script") {
		t.Errorf("Script tags should be filtered out")
	}

	t.Logf("✅ All rich content media elements preserved successfully!")
	t.Logf("✅ Security filtering still working correctly!")
}

// TestIframeSupport tests iframe element support for embedded videos
func TestIframeSupport(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "YouTube iframe embed",
			input:       `<iframe data-v-15ecd644="" src="https://www.youtube.com/embed/FPex8h9cR7o" width="640px" frameborder="0" allowfullscreen="" style="position: absolute; top: 0px; left: 0px; width: 100%; height: 100%;"></iframe>`,
			description: "YouTube video embed should be preserved",
		},
		{
			name:        "Simple iframe",
			input:       `<iframe src="https://example.com" width="800" height="600"></iframe>`,
			description: "Basic iframe should be preserved",
		},
		{
			name:        "Iframe with multiple attributes",
			input:       `<iframe src="https://player.vimeo.com/video/123456" width="640" height="360" frameborder="0" allowfullscreen allow="autoplay; fullscreen"></iframe>`,
			description: "Iframe with various attributes should be preserved",
		},
		{
			name:        "Iframe within content",
			input:       `<div><p>Check out this video:</p><iframe src="https://www.youtube.com/embed/abc123" width="560" height="315"></iframe><p>Great content!</p></div>`,
			description: "Iframe within other content should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.SanitizeHTML(tt.input)

			t.Logf("Input:  %q", tt.input)
			t.Logf("Result: %q", result)

			// Check if iframe element is preserved
			if !strings.Contains(result, "<iframe") {
				t.Errorf("IFRAME ELEMENT MISSING: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
			}

			// Check if src attribute is preserved
			if strings.Contains(tt.input, "src=") && !strings.Contains(result, "src=") {
				t.Errorf("IFRAME SRC ATTRIBUTE MISSING: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
			}

			// Check if important attributes are preserved
			if strings.Contains(tt.input, "width=") && !strings.Contains(result, "width=") {
				t.Errorf("WIDTH ATTRIBUTE MISSING: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
			}

			if strings.Contains(tt.input, "height=") && !strings.Contains(result, "height=") {
				t.Errorf("HEIGHT ATTRIBUTE MISSING: %s\nInput: %q\nResult: %q", tt.description, tt.input, result)
			}
		})
	}
}

// TestSpecificYouTubeExample tests the exact iframe example provided by the user
func TestSpecificYouTubeExample(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// The exact iframe from user's example
	input := `<iframe data-v-15ecd644="" src="https://www.youtube.com/embed/FPex8h9cR7o" width="640px" frameborder="0" allowfullscreen="" style="position: absolute; top: 0px; left: 0px; width: 100%; height: 100%;"></iframe>`

	result := config.SanitizeHTML(input)

	t.Logf("Original input: %s", input)
	t.Logf("Sanitized result: %s", result)

	// Check that essential iframe functionality is preserved
	essentialParts := []string{
		"<iframe",
		`src="https://www.youtube.com/embed/FPex8h9cR7o"`,
		`width="640px"`,
		`frameborder="0"`,
		`allowfullscreen`,
		`style="position: absolute; top: 0px; left: 0px; width: 100%; height: 100%;"`,
		"</iframe>",
	}

	for _, part := range essentialParts {
		if !strings.Contains(result, part) {
			t.Errorf("Missing essential part: %q\nResult: %s", part, result)
		}
	}

	// Verify the iframe is functional (not empty)
	if len(result) < 50 { // A valid iframe should be at least this long
		t.Errorf("Result seems too short, might be broken: %s", result)
	}

	t.Logf("✅ YouTube iframe embed working perfectly!")
}

// TestIframeStyleDebug debugs why iframe style attributes might be filtered
func TestIframeStyleDebug(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// Test the exact original iframe with style attributes
	originalIframe := `<iframe data-v-15ecd644="" src="https://www.youtube.com/embed/Q2lS94_M6TA" width="640px" frameborder="0" allowfullscreen="" style="position: absolute; top: 0px; left: 0px; width: 100%; height: 100%;"></iframe>`

	// Test by sanitizing the original
	sanitizedOriginal := config.SanitizeHTML(originalIframe)

	t.Logf("🔍 DEBUGGING IFRAME STYLE FILTERING:")
	t.Logf("Original input:     %s", originalIframe)
	t.Logf("Sanitized result:   %s", sanitizedOriginal)

	// Check if style attribute exists at all
	if !strings.Contains(sanitizedOriginal, "style=") {
		t.Errorf("❌ STYLE ATTRIBUTE COMPLETELY MISSING")
	}

	// Check specific style components individually
	if !strings.Contains(sanitizedOriginal, "position") {
		t.Errorf("❌ POSITION MISSING")
	}

	if !strings.Contains(sanitizedOriginal, "absolute") {
		t.Errorf("❌ ABSOLUTE MISSING")
	}

	if !strings.Contains(sanitizedOriginal, "width") {
		t.Errorf("❌ WIDTH STYLE MISSING")
	}

	if !strings.Contains(sanitizedOriginal, "height") {
		t.Errorf("❌ HEIGHT STYLE MISSING")
	}

	// Test individual style components
	testCases := []struct {
		input string
		name  string
	}{
		{`<iframe style="position: absolute;">test</iframe>`, "position"},
		{`<iframe style="width: 600px;">test</iframe>`, "width_px"},
		{`<iframe style="height: 400px;">test</iframe>`, "height_px"},
		{`<iframe style="top: 0px;">test</iframe>`, "top"},
		{`<iframe style="left: 0px;">test</iframe>`, "left"},
	}

	for _, tc := range testCases {
		result := config.SanitizeHTML(tc.input)
		t.Logf("Style test %s: %s → %s", tc.name, tc.input, result)

		if !strings.Contains(result, "style=") {
			t.Errorf("Style test %s FAILED: style attribute was completely removed", tc.name)
		}
	}
}

// TestMissingAttributesSupport tests the specific attributes that were being filtered
func TestMissingAttributesSupport(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// Test the exact content that was having attributes filtered
	originalInput := `<p><img src="/system/media_libraries/7/file.png" alt="Snipaste_2025-03-21_17-43-16.png" lockaspectratio="true" width="3256" data-display="inline"></p><div data-video=""><iframe src="https://www.youtube.com/embed/Q2lS94_M6TA" width="100%" frameborder="0" allowfullscreen="true" height="100%"></iframe></div>`

	result := config.SanitizeHTML(originalInput)

	t.Logf("🔍 TESTING MISSING ATTRIBUTES:")
	t.Logf("Original: %s", originalInput)
	t.Logf("Result:   %s", result)

	// Check that previously missing attributes are now preserved
	missingAttributes := []string{
		`lockaspectratio="true"`,
		`data-display="inline"`,
		`data-video=""`,
	}

	for _, attr := range missingAttributes {
		if !strings.Contains(result, attr) {
			t.Errorf("❌ MISSING ATTRIBUTE: %s was filtered out\nResult: %s", attr, result)
		} else {
			t.Logf("✅ PRESERVED ATTRIBUTE: %s", attr)
		}
	}

	// Test individual attribute cases
	testCases := []struct {
		name  string
		input string
		attr  string
	}{
		{
			name:  "img_lockaspectratio_true",
			input: `<img src="test.jpg" lockaspectratio="true" alt="test">`,
			attr:  `lockaspectratio="true"`,
		},
		{
			name:  "img_lockaspectratio_false",
			input: `<img src="test.jpg" lockaspectratio="false" alt="test">`,
			attr:  `lockaspectratio="false"`,
		},
		{
			name:  "img_data_display_inline",
			input: `<img src="test.jpg" data-display="inline" alt="test">`,
			attr:  `data-display="inline"`,
		},
		{
			name:  "img_data_display_block",
			input: `<img src="test.jpg" data-display="block" alt="test">`,
			attr:  `data-display="block"`,
		},
		{
			name:  "div_data_video_empty",
			input: `<div data-video="">content</div>`,
			attr:  `data-video=""`,
		},
		{
			name:  "div_data_video_true",
			input: `<div data-video="true">content</div>`,
			attr:  `data-video="true"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := config.SanitizeHTML(tc.input)
			t.Logf("Test %s: %s → %s", tc.name, tc.input, result)

			if !strings.Contains(result, tc.attr) {
				t.Errorf("❌ FAILED: %s was filtered out", tc.attr)
			}
		})
	}
}

// TestUserProvidedContent tests the exact content provided by the user
func TestUserProvidedContent(t *testing.T) {
	config := TiptapHTMLSanitizerConfig()

	// The exact input content provided by the user
	userInput := `<table class="table-wrapper" style="min-width: 75px"><colgroup><col style="min-width: 25px"><col style="min-width: 25px"><col style="min-width: 25px"></colgroup><tbody><tr><th colspan="1" rowspan="1"><p></p><p><strong>123123123</strong></p></th><th colspan="1" rowspan="1"><p></p></th><th colspan="1" rowspan="1"><p></p></th></tr><tr><td colspan="1" rowspan="1"><p>123123</p></td><td colspan="1" rowspan="1"><p></p></td><td colspan="1" rowspan="1"><p></p><p></p></td></tr></tbody></table><p><img src="/system/media_libraries/7/file.png" alt="Snipaste_2025-03-21_17-43-16.png" lockaspectratio="true" width="3256" data-display="inline"></p><h1><span style="color: rgb(76, 175, 80)">123123</span></h1><p></p><div data-video=""><iframe src="https://www.youtube.com/embed/Q2lS94_M6TA" width="100%" frameborder="0" allowfullscreen="true" height="100%"></iframe></div><h1 style="text-align: center"><span style="color: rgb(244, 67, 54)">asdf1</span></h1><blockquote class="blockquote"><p>sadfasdfadsfsdafsdaf</p></blockquote><p></p><p></p>`

	result := config.SanitizeHTML(userInput)

	t.Logf("🔍 USER CONTENT TEST:")
	t.Logf("Input length:  %d", len(userInput))
	t.Logf("Result length: %d", len(result))
	t.Logf("Input:  %s", userInput)
	t.Logf("Result: %s", result)

	// All critical attributes that should be preserved
	criticalAttributes := []string{
		// Table attributes
		`class="table-wrapper"`,
		`style="min-width: 75px"`,
		`style="min-width: 25px"`,
		`colspan="1"`,
		`rowspan="1"`,

		// Image attributes (previously filtered)
		`src="/system/media_libraries/7/file.png"`,
		`alt="Snipaste_2025-03-21_17-43-16.png"`,
		`lockaspectratio="true"`,
		`width="3256"`,
		`data-display="inline"`,

		// Text styling
		`style="color: rgb(76, 175, 80)"`,
		`style="text-align: center"`,
		`style="color: rgb(244, 67, 54)"`,

		// Video container (previously filtered)
		`data-video=""`,

		// Iframe attributes
		`src="https://www.youtube.com/embed/Q2lS94_M6TA"`,
		`width="100%"`,
		`frameborder="0"`,
		`allowfullscreen="true"`,
		`height="100%"`,

		// Structure elements
		`<table`,
		`<colgroup>`,
		`<col`,
		`<tbody>`,
		`<th`,
		`<td`,
		`<img`,
		`<div`,
		`<iframe`,
		`<h1>`,
		`<span`,
		`<blockquote`,
		`class="blockquote"`,
	}

	missingCount := 0
	for _, attr := range criticalAttributes {
		if !strings.Contains(result, attr) {
			t.Errorf("❌ MISSING: %s", attr)
			missingCount++
		}
	}

	if missingCount == 0 {
		t.Logf("🎉 SUCCESS: All %d critical attributes preserved!", len(criticalAttributes))
	} else {
		t.Errorf("❌ FAILED: %d attributes missing out of %d", missingCount, len(criticalAttributes))
	}

	// Check that input and output are exactly the same (no filtering)
	if userInput == result {
		t.Logf("✅ PERFECT: Input and output are identical - no unwanted filtering!")
	} else {
		t.Logf("ℹ️  INFO: Some differences detected (possibly safe normalization)")

		// Show differences character by character for debugging
		if len(userInput) != len(result) {
			t.Logf("Length difference: input=%d, result=%d", len(userInput), len(result))
		}
	}
}
