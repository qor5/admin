package media

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func TestRowStorageDir(t *testing.T) {
	const id = "0b9f6f9e-6a51-4d2a-9d6c-0b1e3f1c2a10"
	for _, c := range []struct {
		url  string
		id   uint
		want string
	}{
		{"/system/media_libraries/7/file.png", 7, "/system/media_libraries/7"},
		{"/system/media_libraries/7/file/photo.2024.png", 7, "/system/media_libraries/7"},
		{"//cdn.example.com/system/media_libraries/7/file.png", 7, "/system/media_libraries/7"},
		{"https://cdn.example.com/system/media_libraries/7/file.png", 7, "/system/media_libraries/7"},
		{"/system/media_libraries/" + id + "/file.png", 7, "/system/media_libraries/" + id},
		{"/system/" + id + "/x/file.png", 7, "/system/" + id},
		// Not per-row: no id under the class segment, no uuid.
		{"/system/media_libraries/file_7.png", 7, ""},
		{"/uploads/7/file.png", 7, ""},
		{"/system/media_libraries/8/file.png", 7, ""},
		{"/system/media_libraries/7/file.png", 0, ""},
		{"file.png", 7, ""},
		{"", 7, ""},
	} {
		require.Equal(t, c.want, rowStorageDir(c.url, c.id), "url %q id %d", c.url, c.id)
	}
}

func writeStray(t *testing.T, root, p string) string {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(p))
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte("x"), 0o644))
	return full
}

func deleteMedia(t *testing.T, b *Builder, pb *presets.Builder, id uint) {
	t.Helper()
	_, err := doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(id)},
	}))
	require.NoError(t, err)
}

// With a per-row directory, delete also removes crop copies and leftovers
// but nothing outside that directory.
func TestDoDeleteFilesRemovesRowDirectory(t *testing.T) {
	root := useTestStorage(t, "")
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)
	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB).DeleteFilesOnDelete(true)
	require.NoError(t, b.Install(pb))

	m := uploadImageWithSizes(t, "A")
	cropThumb(t, b, pb, m.ID, "A")
	dir := rowStorageDir(reload(t, m.ID).File.Url, m.ID)
	require.NotEmpty(t, dir)
	require.Len(t, filesWithPrefix(t, root, dir, "file.A_"), 1)
	leftover := writeStray(t, root, dir+"/old/file.png")
	neighbour := writeStray(t, root, fmt.Sprintf("/system/media_libraries/%d0/file.png", m.ID))

	deleteMedia(t, b, pb, m.ID)
	var left []string
	require.NoError(t, filepath.WalkDir(filepath.Join(root, filepath.FromSlash(dir)), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			left = append(left, p)
		}
		return err
	}))
	require.Empty(t, left, "row directory must be emptied")
	require.NoFileExists(t, leftover)
	require.FileExists(t, neighbour, "another row's directory must be untouched")
}

// Without a per-row directory only the row's own paths go; other files in the
// shared directory stay.
func TestDoDeleteFilesSharedDirectory(t *testing.T) {
	root := useTestStorage(t, "/system/shared/{{column}}_{{primary_key}}.{{extension}}")
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)
	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB).DeleteFilesOnDelete(true)
	require.NoError(t, b.Install(pb))

	m := uploadImageWithSizes(t)
	stored := reload(t, m.ID).File.Url
	require.Empty(t, rowStorageDir(stored, m.ID))
	require.FileExists(t, filepath.Join(root, filepath.FromSlash(stored)))
	other := writeStray(t, root, "/system/shared/other.png")

	deleteMedia(t, b, pb, m.ID)
	require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(stored)))
	require.FileExists(t, other, "files of other rows in a shared directory must survive")
}
