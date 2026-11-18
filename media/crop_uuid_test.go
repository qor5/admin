package media

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/stretchr/testify/require"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

// useTestStorage points oss.Storage at a temp dir and, if urlTemplate is set,
// base.MediaLibraryURL at it, restoring both afterwards.
func useTestStorage(t *testing.T, urlTemplate string) string {
	t.Helper()
	root := t.TempDir()
	prevStorage, prevURL := oss.Storage, base.MediaLibraryURL
	oss.Storage = filesystem.New(root)
	base.MediaLibraryURL = urlTemplate
	t.Cleanup(func() { oss.Storage, base.MediaLibraryURL = prevStorage, prevURL })
	return root
}

// uploadImageWithSizes stores a PNG media library item with the given sizes.
func uploadImageWithSizes(t *testing.T, sizes ...string) *media_library.MediaLibrary {
	t.Helper()
	m := &media_library.MediaLibrary{SelectedType: media_library.ALLOW_TYPE_IMAGE}
	require.NoError(t, m.File.Scan(testFileHeader(t, "photo.png", testPNG(t))))
	m.File.Sizes = map[string]*base.Size{}
	for _, s := range sizes {
		m.File.Sizes[s] = &base.Size{Width: 1, Height: 1}
	}
	require.NoError(t, base.SaveUploadAndCropImage(testDB, m, "", &web.EventContext{}))
	return reload(t, m.ID)
}

func cropThumb(t *testing.T, b *Builder, pb *presets.Builder, id uint, thumb string) {
	t.Helper()
	_, err := cropImage(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS:  []string{fmt.Sprint(id)},
		"field":        []string{"media"},
		"media.Values": []string{`{"FileName":"photo.png"}`},
		"thumb":        []string{thumb},
		"CropOption":   []string{`{"x":0,"y":0,"width":1,"height":1}`},
	}))
	require.NoError(t, err)
}

func filesWithPrefix(t *testing.T, root, urlDir, prefix string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(urlDir)))
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			names = append(names, e.Name())
		}
	}
	return names
}

// Re-cropping must keep a {{uuid}} directory stable, or crop copies made
// earlier (referenced by MediaBox CropIDs) end up in an abandoned directory
// while the URL points at a new one.
func TestCropKeepsUUIDDirectory(t *testing.T) {
	root := useTestStorage(t, "/system/{{class}}/{{uuid}}/{{column}}.{{extension}}")
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)
	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB)
	require.NoError(t, b.Install(pb))

	m := uploadImageWithSizes(t, "A", "B")
	dir := path.Dir(m.File.Url)
	require.NotEmpty(t, base.URLDirUUID(m.File.Url), "url %q must contain the uuid", m.File.Url)

	cropThumb(t, b, pb, m.ID, "A")
	require.Equal(t, dir, path.Dir(reload(t, m.ID).File.Url), "crop must keep the uuid directory")

	cropThumb(t, b, pb, m.ID, "B")
	require.Equal(t, dir, path.Dir(reload(t, m.ID).File.Url), "crop must keep the uuid directory")

	require.Len(t, filesWithPrefix(t, root, dir, "file.A_"), 1, "crop copy of A must sit in the current directory")
	require.Len(t, filesWithPrefix(t, root, dir, "file.B_"), 1, "crop copy of B must sit in the current directory")
}
