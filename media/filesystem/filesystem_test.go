package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/media/base"
)

func TestGetFullPath(t *testing.T) {
	fs := FileSystem{}

	t.Run("default base path", func(t *testing.T) {
		// Test with default base path (nil option)
		path, err := fs.GetFullPath("test.jpg", nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// The path should be absolute and contain "public"
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got: %s", path)
		}
		if !strings.Contains(path, "public") {
			t.Errorf("Expected path to contain 'public', got: %s", path)
		}
		if !strings.HasSuffix(path, "test.jpg") {
			t.Errorf("Expected path to end with 'test.jpg', got: %s", path)
		}
	})

	t.Run("custom base path", func(t *testing.T) {
		// Test with custom base path
		customOption := &base.Option{"PATH": "./custom"}
		path, err := fs.GetFullPath("file.txt", customOption)
		if err != nil {
			t.Fatalf("Unexpected error with custom base: %v", err)
		}

		// Path should be absolute and end with custom/file.txt
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got: %s", path)
		}
		if !strings.HasSuffix(path, "custom/file.txt") {
			t.Errorf("Expected path to end with custom/file.txt, got: %s", path)
		}
	})

	t.Run("nested paths", func(t *testing.T) {
		path, err := fs.GetFullPath("images/thumbnails/photo.jpg", nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got: %s", path)
		}
		if !strings.HasSuffix(path, "public/images/thumbnails/photo.jpg") {
			t.Errorf("Expected path to end with 'public/images/thumbnails/photo.jpg', got: %s", path)
		}
	})

	t.Run("empty option with path key", func(t *testing.T) {
		// Test with option that has empty path
		emptyOption := &base.Option{"PATH": ""}
		path, err := fs.GetFullPath("test.jpg", emptyOption)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should use default path when option path is empty
		if !strings.Contains(path, "public") {
			t.Errorf("Expected to use default 'public' path when option path is empty, got: %s", path)
		}
	})
}

func TestGetFullPath_SecurityValidation(t *testing.T) {
	fs := FileSystem{}

	tests := []struct {
		name      string
		url       string
		option    *base.Option
		shouldErr bool
		desc      string
	}{
		{
			name:      "path_traversal_single_dotdot",
			url:       "../config.txt",
			option:    nil,
			shouldErr: true,
			desc:      "Single ../ should be blocked",
		},
		{
			name:      "path_traversal_multiple_dotdot",
			url:       "../../../etc/passwd",
			option:    nil,
			shouldErr: true,
			desc:      "Multiple ../ should be blocked",
		},
		{
			name:      "path_traversal_mixed",
			url:       "images/../../etc/passwd",
			option:    nil,
			shouldErr: true,
			desc:      "Mixed path with traversal should be blocked",
		},
		{
			name:      "absolute_path",
			url:       "/etc/passwd",
			option:    nil,
			shouldErr: false, // filepath.Join treats this as relative to base
			desc:      "Absolute path will be treated as relative to base",
		},
		{
			name:      "custom_base_normal",
			url:       "uploads/file.txt",
			option:    &base.Option{"PATH": "./custom"},
			shouldErr: false,
			desc:      "Normal path with custom base should work",
		},
		{
			name:      "custom_base_traversal",
			url:       "../../etc/passwd",
			option:    &base.Option{"PATH": "./custom"},
			shouldErr: true,
			desc:      "Path traversal with custom base should be blocked",
		},
		{
			name:      "encoded_traversal",
			url:       "..%2F..%2Fetc%2Fpasswd",
			option:    nil,
			shouldErr: true,
			desc:      "URL encoded traversal should be blocked",
		},
		{
			name:      "windows_style_traversal",
			url:       "..\\..\\windows\\system32\\config\\sam",
			option:    nil,
			shouldErr: true,
			desc:      "Windows style path traversal should be blocked",
		},
		{
			name:      "current_dir_reference",
			url:       "./config/app.conf",
			option:    nil,
			shouldErr: false,
			desc:      "Current directory reference should be allowed",
		},
		{
			name:      "just_dotdot",
			url:       "..",
			option:    nil,
			shouldErr: true,
			desc:      "Just '..' should be blocked",
		},
		{
			name:      "dotdot_with_slash",
			url:       "../",
			option:    nil,
			shouldErr: true,
			desc:      "'../' should be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := fs.GetFullPath(tt.url, tt.option)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("%s: expected error but got none. Path: %s", tt.desc, path)
				} else {
					// Check that error message contains relevant information
					errMsg := err.Error()
					if !strings.Contains(errMsg, "illegal file path") {
						t.Errorf("%s: expected security error message, got: %s", tt.desc, errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.desc, err)
				} else {
					// Verify the path is absolute
					if !filepath.IsAbs(path) {
						t.Errorf("%s: expected absolute path but got: %s", tt.desc, path)
					}
				}
			}
		})
	}
}

func TestGetFullPath_DirectoryCreation(t *testing.T) {
	// Test the directory auto-creation feature (main modification)
	fs := FileSystem{}

	t.Run("nested directory creation", func(t *testing.T) {
		tempDir := t.TempDir()
		option := &base.Option{"PATH": tempDir}

		// Test path that requires creating nested directories
		testPath := "level1/level2/level3/test.txt"
		fullPath, err := fs.GetFullPath(testPath, option)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify the directory structure was created
		expectedDir := filepath.Join(tempDir, "level1/level2/level3")
		if _, err = os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to be created", expectedDir)
		}

		// Verify the full path is correct
		expectedPath := filepath.Join(tempDir, testPath)
		if fullPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, fullPath)
		}
	})
}

func TestGetFullPath_EdgeCases(t *testing.T) {
	fs := FileSystem{}

	tests := []struct {
		name        string
		url         string
		option      *base.Option
		shouldError bool
		description string
	}{
		{
			name:        "empty_url",
			url:         "",
			option:      nil,
			shouldError: false,
			description: "Empty URL should work and point to base directory",
		},
		{
			name:        "url_with_spaces",
			url:         "my file with spaces.txt",
			option:      nil,
			shouldError: false,
			description: "URL with spaces should work",
		},
		{
			name:        "url_with_unicode",
			url:         "测试文件.txt",
			option:      nil,
			shouldError: false,
			description: "URL with unicode characters should work",
		},
		{
			name:        "very_long_path",
			url:         strings.Repeat("a/", 100) + "file.txt",
			option:      nil,
			shouldError: false,
			description: "Very long path should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := fs.GetFullPath(tt.url, tt.option)

			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got none. Path: %s", tt.description, path)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				} else {
					if !filepath.IsAbs(path) {
						t.Errorf("%s: expected absolute path but got: %s", tt.description, path)
					}
				}
			}
		})
	}
}

func TestStore_DirectoryCreation(t *testing.T) {
	// Test the new directory auto-creation feature in GetFullPath
	fs := FileSystem{}

	t.Run("store with nested path creation", func(t *testing.T) {
		tempDir := t.TempDir()
		option := &base.Option{"PATH": tempDir}

		content := "nested file content"
		reader := strings.NewReader(content)
		filename := "images/thumbnails/nested_file.txt"

		err := fs.Store(filename, option, reader)
		if err != nil {
			t.Fatalf("Unexpected error storing nested file: %v", err)
		}

		fullPath := filepath.Join(tempDir, filename)
		if _, err = os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected nested file to exist at %s", fullPath)
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read nested file: %v", err)
		}
		if string(data) != content {
			t.Errorf("Expected content '%s', got '%s'", content, string(data))
		}
	})
}

func TestStore_ErrorHandling(t *testing.T) {
	fs := FileSystem{}

	t.Run("store with invalid path", func(t *testing.T) {
		// Test with path traversal that should be blocked
		reader := strings.NewReader("test content")
		if err := fs.Store("../../../etc/test.txt", nil, reader); err == nil {
			t.Errorf("Expected error when storing file with path traversal")
		}
	})
}
