package media_library

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/qor5/admin/v3/media/base"
)

func TestStoredPathsEmpty(t *testing.T) {
	require.Nil(t, (&MediaLibraryStorage{}).StoredPaths())
}

func TestStoredPathsNonImage(t *testing.T) {
	s := &MediaLibraryStorage{}
	s.Url = "/system/media_libraries/7/file.pdf"
	require.Equal(t, []string{"/system/media_libraries/7/file.pdf"}, s.StoredPaths())
}

func TestStoredPathsImage(t *testing.T) {
	s := &MediaLibraryStorage{
		Sizes:  map[string]*base.Size{"thumb": {Width: 10, Height: 10}},
		CropID: map[string]string{"thumb": "transient"},
	}
	s.Url = "/system/media_libraries/7/file.png"
	s.FileSizes = map[string]int{"original": 1, "default": 1, "thumb": 1, "@qor_preview": 1, "legacy": 1}

	got := s.StoredPaths()
	want := []string{
		"/system/media_libraries/7/file.png",
		"/system/media_libraries/7/file.@qor_preview.png",
		"/system/media_libraries/7/file.legacy.png",
		"/system/media_libraries/7/file.original.png",
		"/system/media_libraries/7/file.thumb.png",
		"/system/media_libraries/7/file.webp",
		"/system/media_libraries/7/file.@qor_preview.webp",
		"/system/media_libraries/7/file.legacy.webp",
		"/system/media_libraries/7/file.original.webp",
		"/system/media_libraries/7/file.thumb.webp",
	}
	require.Equal(t, want, got)
	require.Equal(t, "transient", s.CropID["thumb"], "StoredPaths must not mutate the receiver")
}
