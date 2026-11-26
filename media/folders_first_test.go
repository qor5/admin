package media

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/qor5/admin/v3/media/media_library"
)

// A page of the folders tab mixes folders and files, so with FoldersFirst the
// query has to put folders ahead of files itself: sorting by upload time alone
// lets newer files push older folders onto later pages.
func TestFoldersTabListsFoldersFirst(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(folder bool, name string, age int) uint {
		m := mkRow(t, 1, folder, 0, name)
		require.NoError(t, testDB.Model(m).UpdateColumn("created_at", base.Add(time.Duration(age)*time.Hour)).Error)
		return m.ID
	}
	oldFolder := mk(true, "old-folder", 1)
	oldFile := mk(false, "old.png", 2)
	newFolder := mk(true, "new-folder", 3)
	newFile := mk(false, "new.png", 4)

	ctx := scopeEventContext(t, pb, 1, url.Values{})
	list := func(tab, orderBy string) []uint {
		var rows []*media_library.MediaLibrary
		require.NoError(t, b.mediaLibraryFilter(tab, "", "", orderBy, 0, ctx, &media_library.MediaBoxConfig{}).Find(&rows).Error)
		ids := make([]uint, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		return ids
	}

	require.Equal(t, []uint{newFile, newFolder, oldFile, oldFolder}, list(tabFolders, orderByCreatedAtDESC),
		"off by default: upload time only")

	b.FoldersFirst(true)
	require.Equal(t, []uint{newFolder, oldFolder, newFile, oldFile}, list(tabFolders, orderByCreatedAtDESC))
	require.Equal(t, []uint{oldFolder, newFolder, oldFile, newFile}, list(tabFolders, orderByCreatedAt))
	require.Equal(t, []uint{newFile, oldFile}, list(tabFiles, orderByCreatedAtDESC))
}
