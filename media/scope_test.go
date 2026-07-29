package media

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/gormx"
	"github.com/stretchr/testify/require"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	testSuite := gormx.MustStartTestSuite(ctx)
	defer func() {
		if err := testSuite.Stop(context.Background()); err != nil {
			fmt.Printf("Error during teardown: %v\n", err)
		}
	}()

	testDB = testSuite.DB()
	if err := AutoMigrate(testDB); err != nil {
		panic(err)
	}

	m.Run()
}

const scopeTestUserHeader = "X-Test-User"

// scopeTestBuilder scopes every query to the user id carried by the request
// header, mimicking a multi-tenant Searcher.
func scopeTestBuilder(t *testing.T) (*Builder, *presets.Builder) {
	t.Helper()
	pb := presets.New()
	b := New(testDB).Searcher(func(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
		userID, err := strconv.ParseUint(ctx.R.Header.Get(scopeTestUserHeader), 10, 64)
		require.NoError(t, err, "every scoped request must carry a test user id")
		return db.Where("user_id = ?", userID)
	})
	require.NoError(t, b.Install(pb))
	return b, pb
}

func scopeEventContext(t *testing.T, pb *presets.Builder, userID uint, params url.Values) *web.EventContext {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/media-libraries?"+params.Encode(), http.NoBody)
	req.Header.Set(scopeTestUserHeader, fmt.Sprint(userID))
	var out *http.Request
	pb.GetI18n().EnsureLanguage(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		out = r
	})).ServeHTTP(httptest.NewRecorder(), req)
	return &web.EventContext{R: out, W: httptest.NewRecorder()}
}

func mkRow(t *testing.T, userID uint, folder bool, parentID uint, name string) *media_library.MediaLibrary {
	t.Helper()
	m := &media_library.MediaLibrary{Folder: folder, ParentId: parentID, UserID: userID}
	m.File.FileName = name
	require.NoError(t, testDB.Create(m).Error)
	return m
}

func seedScopeRows(t *testing.T) (ownFolder, ownFile, foreignFolder, foreignFile *media_library.MediaLibrary) {
	t.Helper()
	require.NoError(t, testDB.Exec("DELETE FROM media_libraries").Error)
	ownFolder = mkRow(t, 1, true, 0, "own-folder")
	ownFile = mkRow(t, 1, false, 0, "own.png")
	foreignFolder = mkRow(t, 2, true, 0, "foreign-folder")
	foreignFile = mkRow(t, 2, false, 0, "foreign.png")
	return
}

func reload(t *testing.T, id uint) *media_library.MediaLibrary {
	t.Helper()
	var m media_library.MediaLibrary
	require.NoError(t, testDB.First(&m, id).Error)
	return &m
}

func TestScopedDBRestrictsByIDLookups(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	var m media_library.MediaLibrary
	require.NoError(t, b.scopedDB(testDB, ctx).Find(&m, foreignFile.ID).Error)
	require.Zero(t, m.ID, "foreign row must not be visible through scopedDB")

	m = media_library.MediaLibrary{}
	require.NoError(t, b.scopedDB(testDB, ctx).Find(&m, ownFile.ID).Error)
	require.Equal(t, ownFile.ID, m.ID)

	unscoped := New(testDB)
	m = media_library.MediaLibrary{}
	require.NoError(t, unscoped.scopedDB(testDB, ctx).Find(&m, foreignFile.ID).Error)
	require.Equal(t, foreignFile.ID, m.ID, "without a searcher the historical unscoped behavior is preserved")
}

func TestFolderTreeScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, _, foreignFolder, _ := seedScopeRows(t)
	// Children make the expandable-node counts meaningful: without them the
	// scoped and unscoped count queries cannot differ.
	mkRow(t, 1, false, ownFolder.ID, "own-child.png")
	mkRow(t, 2, false, foreignFolder.ID, "foreign-child.png")
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	items := folderGroupsComponents(b, ctx, 0)
	require.Len(t, items, 1, "move-to tree must list only the current tenant's folders")
}

func TestFolderChildCountScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, _, _, _ := seedScopeRows(t)
	mkRow(t, 1, false, ownFolder.ID, "own-child.png")
	mkRow(t, 2, false, ownFolder.ID, "foreign-child-in-own-folder.png")
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	_, content := folderComponent(b, ctx, ownFolder)
	require.Contains(t, h.MustString(content, ctx.R.Context()), "1 items",
		"folder tiles must count only the current tenant's children")
}

func TestParentFoldersScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, _, foreignFolder, _ := seedScopeRows(t)
	ctx := scopeEventContext(t, pb, 1, url.Values{})
	cfg := &media_library.MediaBoxConfig{}

	require.NotEmpty(t, parentFolders("f", ctx, cfg, b, ownFolder.ID, ownFolder.ID, nil, true))
	require.Empty(t, parentFolders("f", ctx, cfg, b, foreignFolder.ID, foreignFolder.ID, nil, true),
		"foreign folder names must not leak into breadcrumbs")
}

func TestWrapFirstScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)

	var r web.EventResponse
	_, ok := wrapFirst(b, scopeEventContext(t, pb, 1, url.Values{ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)}}), &r)
	require.False(t, ok)

	obj, ok := wrapFirst(b, scopeEventContext(t, pb, 1, url.Values{ParamMediaIDS: []string{fmt.Sprint(ownFile.ID)}}), &r)
	require.True(t, ok)
	require.Equal(t, ownFile.ID, obj.ID)
}

func TestRenameScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, _, _, foreignFile := seedScopeRows(t)
	var before int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Count(&before).Error)

	_, err := rename(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
		ParamName:     []string{"hijacked"},
	}))
	require.NoError(t, err)

	require.Equal(t, "foreign.png", reload(t, foreignFile.ID).File.FileName)
	var after int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Count(&after).Error)
	require.Equal(t, before, after, "a scoped-out rename must not insert a new row")
}

func TestDoDeleteScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)

	_, err := doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
	}))
	require.NoError(t, err)
	require.NotZero(t, reload(t, foreignFile.ID).ID, "foreign row must survive a scoped-out delete")

	_, err = doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(ownFile.ID)},
	}))
	require.NoError(t, err)
	var count int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Where("id = ?", ownFile.ID).Count(&count).Error)
	require.Zero(t, count)
}

// A mixed batch pins that only the scoped-visible ids are deleted: deleting the
// raw request ids would take the foreign row down with the own one.
func TestDoDeleteScopedMixedBatch(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)

	_, err := doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprintf("%d,%d", ownFile.ID, foreignFile.ID)},
	}))
	require.NoError(t, err)

	var ownCount int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Where("id = ?", ownFile.ID).Count(&ownCount).Error)
	require.Zero(t, ownCount, "own row in the batch must be deleted")
	require.NotZero(t, reload(t, foreignFile.ID).ID, "foreign row in the same batch must survive")
}

// Deleting a folder reparents its children to root; that UPDATE must not touch
// another tenant's rows that happen to sit under the deleted folder.
func TestDoDeleteReparentScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, _, _, _ := seedScopeRows(t)
	ownChild := mkRow(t, 1, false, ownFolder.ID, "own-child.png")
	foreignChild := mkRow(t, 2, false, ownFolder.ID, "foreign-child.png")

	_, err := doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(ownFolder.ID)},
	}))
	require.NoError(t, err)

	require.Zero(t, reload(t, ownChild.ID).ParentId, "own child must be reparented to root")
	require.Equal(t, ownFolder.ID, reload(t, foreignChild.ID).ParentId,
		"foreign child must not be reparented by another tenant's delete")
}

func TestMoveToFolderScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, ownFile, foreignFolder, foreignFile := seedScopeRows(t)

	move := func(userID uint, fileID, targetID uint) {
		t.Helper()
		_, err := moveToFolder(b)(scopeEventContext(t, pb, userID, url.Values{
			ParamSelectIDS:      []string{fmt.Sprint(fileID)},
			ParamSelectFolderID: []string{fmt.Sprint(targetID)},
		}))
		require.NoError(t, err)
	}

	move(1, ownFile.ID, foreignFolder.ID)
	require.Zero(t, reload(t, ownFile.ID).ParentId, "own file must not move into a foreign folder")

	move(1, foreignFile.ID, ownFolder.ID)
	require.Zero(t, reload(t, foreignFile.ID).ParentId, "foreign file must not be movable")

	move(1, ownFile.ID, ownFolder.ID)
	require.Equal(t, ownFolder.ID, reload(t, ownFile.ID).ParentId)

	// Unfiling back to the root folder stays possible — the target check
	// special-cases id 0.
	move(1, ownFile.ID, 0)
	require.Zero(t, reload(t, ownFile.ID).ParentId, "moving to root must keep working")
}

func TestUploadAndCreateFolderRejectForeignParent(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, _, foreignFolder, _ := seedScopeRows(t)
	var before int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Count(&before).Error)

	_, err := createFolder(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamName:     []string{"sneaky"},
		ParamParentID: []string{fmt.Sprint(foreignFolder.ID)},
	}))
	require.NoError(t, err)

	var after int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Count(&after).Error)
	require.Equal(t, before, after, "no row may be created under a foreign parent")
}

// Without a searcher every guard added by the scoping change must be inert, so
// existing single-tenant apps keep working exactly as before.
func TestNoSearcherKeepsEveryIDReachable(t *testing.T) {
	pb := presets.New()
	b := New(testDB)
	require.NoError(t, b.Install(pb))
	_, _, foreignFolder, foreignFile := seedScopeRows(t)
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	require.True(t, b.folderIsVisible(ctx, foreignFolder.ID))
	require.True(t, b.folderIsVisible(ctx, 4242), "a stale folder id stays acceptable")
	require.True(t, b.recordIsVisible(ctx, fmt.Sprint(foreignFile.ID)))
	require.True(t, b.recordIsVisible(ctx, "4242"))

	obj, ok := wrapFirst(b, scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
	}), &web.EventResponse{})
	require.True(t, ok)
	require.Equal(t, foreignFile.ID, obj.ID)

	_, err := moveToFolder(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamSelectIDS:      []string{fmt.Sprint(foreignFile.ID)},
		ParamSelectFolderID: []string{fmt.Sprint(foreignFolder.ID)},
	}))
	require.NoError(t, err)
	require.Equal(t, foreignFolder.ID, reload(t, foreignFile.ID).ParentId)
}

func TestChooseFileScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)

	r, err := chooseFile(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamField:    []string{"f"},
		ParamCfg:      []string{"{}"},
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
	}))
	require.NoError(t, err)
	require.Empty(t, r.UpdatePortals, "a foreign file must not be attachable through the chooser")

	r, err = chooseFile(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamField:    []string{"f"},
		ParamCfg:      []string{"{}"},
		ParamMediaIDS: []string{fmt.Sprint(ownFile.ID)},
	}))
	require.NoError(t, err)
	require.NotEmpty(t, r.UpdatePortals, "an own file must still be attachable")
}

func TestCropImageScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, _, _, foreignFile := seedScopeRows(t)
	params := url.Values{
		ParamField:    []string{"f"},
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
		"thumb":       []string{"original"},
		"cfg":         []string{"{}"},
		"CropOption":  []string{`{"x":0,"y":0,"width":10,"height":10}`},
		"f.Values":    []string{"{}"},
	}

	_, err := cropImage(b)(scopeEventContext(t, pb, 1, params))
	require.NoError(t, err)
	require.Empty(t, reload(t, foreignFile.ID).File.CropID,
		"a foreign image must not be re-cropped")
}
