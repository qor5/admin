package media

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
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

func storedFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	require.NoError(t, filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, p)
			files = append(files, rel)
		}
		return nil
	}))
	return files
}

func uploadTestImage(t *testing.T, root string) *media_library.MediaLibrary {
	t.Helper()
	m := &media_library.MediaLibrary{SelectedType: media_library.ALLOW_TYPE_IMAGE}
	require.NoError(t, m.File.Scan(testFileHeader(t, "photo.png", testPNG(t))))
	require.NoError(t, base.SaveUploadAndCropImage(testDB, m, "", &web.EventContext{}))
	require.NotEmpty(t, storedFiles(t, root), "upload must write files")
	return m
}

func TestDoDeleteFilesOnDelete(t *testing.T) {
	root := t.TempDir()
	prev := oss.Storage
	oss.Storage = filesystem.New(root)
	t.Cleanup(func() { oss.Storage = prev })
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)

	run := func(b *Builder, m *media_library.MediaLibrary) {
		t.Helper()
		pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
		require.NoError(t, b.Install(pb))
		_, err := doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
			ParamMediaIDS: []string{fmt.Sprint(m.ID)},
		}))
		require.NoError(t, err)
	}

	// Default: files are kept, since MediaBox values elsewhere still point at them.
	kept := uploadTestImage(t, root)
	before := storedFiles(t, root)
	run(New(testDB), kept)
	require.Equal(t, before, storedFiles(t, root))
	require.NoError(t, os.RemoveAll(root))
	require.NoError(t, os.MkdirAll(root, 0o755))

	// Opt-in: every stored variant of the item is removed.
	m := uploadTestImage(t, root)
	reloaded := reload(t, m.ID)
	for _, f := range storedFiles(t, root) {
		require.Contains(t, reloaded.File.StoredPaths(), "/"+filepath.ToSlash(f),
			"StoredPaths must cover every file written on upload")
	}
	run(New(testDB).DeleteFilesOnDelete(true), m)
	require.Empty(t, storedFiles(t, root))
}
