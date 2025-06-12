package examples_admin

import (
	"context"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/tiptap"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestTiptapEditorBuilder_ClassList tests the classList functionality
func TestTiptapEditorBuilder_ClassList(t *testing.T) {
	var db *gorm.DB

	t.Run("Attr method with class", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test single class
		builder.Attr("class", "single-class")
		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "single-class")

		// Test multiple classes in one string
		builder.Attr("class", "class1 class2 class3")
		html, err = builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr = string(html)
		assert.Contains(t, htmlStr, "class1")
		assert.Contains(t, htmlStr, "class2")
		assert.Contains(t, htmlStr, "class3")
	})

	t.Run("SetAttr method with class", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test SetAttr with string class
		builder.SetAttr("class", "setattr-class")
		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "setattr-class")

		// Test SetAttr with multiple classes
		builder.SetAttr("class", "setattr1 setattr2")
		html, err = builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr = string(html)
		assert.Contains(t, htmlStr, "setattr1")
		assert.Contains(t, htmlStr, "setattr2")
	})

	t.Run("Combined Attr and SetAttr class usage", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Use both methods to add classes
		builder.Attr("class", "attr-class1 attr-class2")
		builder.SetAttr("class", "setattr-class1 setattr-class2")

		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "attr-class1")
		assert.Contains(t, htmlStr, "attr-class2")
		assert.Contains(t, htmlStr, "setattr-class1")
		assert.Contains(t, htmlStr, "setattr-class2")
	})

	t.Run("Class with mixed attributes", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test class mixed with other attributes
		builder.Attr("id", "test-id", "class", "mixed-class", "data-test", "value")
		builder.SetAttr("style", "color: red;")
		builder.SetAttr("class", "additional-class")

		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "mixed-class")
		assert.Contains(t, htmlStr, "additional-class")
		assert.Contains(t, htmlStr, "test-id")
		assert.Contains(t, htmlStr, "value")
		assert.Contains(t, htmlStr, "color: red")
	})

	t.Run("Duplicate and repeated classes", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test duplicate classes
		builder.Attr("class", "duplicate-class")
		builder.Attr("class", "duplicate-class another-class")
		builder.SetAttr("class", "duplicate-class third-class")

		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "duplicate-class")
		assert.Contains(t, htmlStr, "another-class")
		assert.Contains(t, htmlStr, "third-class")
	})

	t.Run("Class parsing with whitespace", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test with extra spaces and tabs
		builder.Attr("class", "  spaced-class1   spaced-class2  ")
		builder.SetAttr("class", "\ttab-class\nnewline-class\n")

		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "spaced-class1")
		assert.Contains(t, htmlStr, "spaced-class2")
		assert.Contains(t, htmlStr, "tab-class")
		assert.Contains(t, htmlStr, "newline-class")
	})

	t.Run("SetAttr with non-string class values", func(t *testing.T) {
		builder := tiptap.TiptapEditor(db, "test-key")

		// Test type checking - non-string values should be ignored
		builder.SetAttr("class", "valid-class")
		builder.SetAttr("class", 123)               // Should be ignored
		builder.SetAttr("class", []string{"array"}) // Should be ignored

		html, err := builder.MarshalHTML(context.Background())
		assert.NoError(t, err)
		htmlStr := string(html)
		assert.Contains(t, htmlStr, "valid-class")
		// Non-string values should not appear
	})
}

// TestStringsFieldsBehavior tests the strings.Fields function behavior
func TestStringsFieldsBehavior(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"single", []string{"single"}},
		{"class1 class2", []string{"class1", "class2"}},
		{"  spaced  classes  ", []string{"spaced", "classes"}},
		{"tab\tclass", []string{"tab", "class"}},
		{"", []string{}},
	}

	for _, tc := range testCases {
		result := strings.Fields(tc.input)
		assert.Equal(t, tc.expected, result,
			"strings.Fields(%q) should return %v", tc.input, tc.expected)
	}
}
