package media

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/gormx"
	"github.com/qor5/x/v3/i18n"
	"github.com/stretchr/testify/require"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
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
	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB).Searcher(func(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
		userID, err := strconv.ParseUint(ctx.R.Header.Get(scopeTestUserHeader), 10, 64)
		require.NoError(t, err, "every scoped request must carry a test user id")
		return db.Where(qualified("user_id")+" = ?", userID)
	})
	require.NoError(t, b.Install(pb))
	return b, pb
}

// scopeTestJoinBuilder scopes through a Joins clause, which is what makes
// unqualified column conditions ambiguous and silently drops from UPDATEs.
func scopeTestJoinBuilder(t *testing.T) (*Builder, *presets.Builder) {
	t.Helper()
	require.NoError(t, testDB.Exec(`
CREATE TABLE IF NOT EXISTS scope_test_owners (user_id bigint primary key)`).Error)
	require.NoError(t, testDB.Exec(`
INSERT INTO scope_test_owners (user_id) VALUES (1) ON CONFLICT DO NOTHING`).Error)
	t.Cleanup(func() {
		if err := testDB.Exec(`DROP TABLE IF EXISTS scope_test_owners`).Error; err != nil {
			t.Errorf("drop owners table: %v", err)
		}
	})

	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB).Searcher(func(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
		userID, err := strconv.ParseUint(ctx.R.Header.Get(scopeTestUserHeader), 10, 64)
		require.NoError(t, err, "every scoped request must carry a test user id")
		return db.
			Joins("join scope_test_owners on scope_test_owners.user_id = "+qualified("user_id")).
			Where("scope_test_owners.user_id = ?", userID)
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

func TestCreateFolderRejectsForeignParent(t *testing.T) {
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

// uploadFile checks the parent before it touches the multipart body, so an
// empty request is enough to pin the guard.
func TestUploadFileRejectsForeignParent(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	ownFolder, _, foreignFolder, _ := seedScopeRows(t)
	msgr := i18n.MustGetModuleMessages(
		scopeEventContext(t, pb, 1, url.Values{}).R, presets.CoreI18nModuleKey, presets.Messages_en_US,
	).(*presets.Messages)

	r, err := uploadFile(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamField:    []string{"f"},
		ParamCfg:      []string{"{}"},
		ParamParentID: []string{fmt.Sprint(foreignFolder.ID)},
	}))
	require.NoError(t, err)
	require.Contains(t, r.RunScript, msgr.RecordNotFound, "upload into a foreign folder must be refused")

	r, err = uploadFile(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamField:    []string{"f"},
		ParamCfg:      []string{"{}"},
		ParamParentID: []string{fmt.Sprint(ownFolder.ID)},
	}))
	require.NoError(t, err)
	require.NotContains(t, r.RunScript, msgr.RecordNotFound, "upload into an own folder must pass the guard")
}

// loadImageCropper returns early on a scoped-out record instead of rendering a
// cropper for a zero-value row.
func TestLoadImageCropperScoped(t *testing.T) {
	b, pb := scopeTestBuilder(t)
	_, ownFile, _, foreignFile := seedScopeRows(t)
	params := func(id uint) url.Values {
		return url.Values{
			ParamField:    []string{"f"},
			ParamMediaIDS: []string{fmt.Sprint(id)},
			"thumb":       []string{base.DefaultSizeKey},
			"cfg":         []string{"{}"},
			"f.Values":    []string{"{}"},
		}
	}

	r, err := loadImageCropper(b)(scopeEventContext(t, pb, 1, params(foreignFile.ID)))
	require.NoError(t, err)
	require.Empty(t, r.UpdatePortals, "no cropper may be rendered for a foreign image")

	r, err = loadImageCropper(b)(scopeEventContext(t, pb, 1, params(ownFile.ID)))
	require.NoError(t, err)
	require.NotEmpty(t, r.UpdatePortals, "an own image must still open the cropper")
}

// Without a searcher every guard added by the scoping change must be inert, so
// existing single-tenant apps keep working exactly as before.
func TestNoSearcherKeepsEveryIDReachable(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(testDB))
	b := New(testDB)
	require.NoError(t, b.Install(pb))
	_, _, foreignFolder, foreignFile := seedScopeRows(t)
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	requireVisible := func(msg string, visible bool, err error) {
		t.Helper()
		require.NoError(t, err)
		require.True(t, visible, msg)
	}
	visible, err := b.folderIsVisible(ctx, foreignFolder.ID)
	requireVisible("a foreign folder stays an acceptable target", visible, err)
	visible, err = b.folderIsVisible(ctx, 4242)
	requireVisible("a stale folder id stays acceptable", visible, err)
	visible, err = b.recordIsVisible(ctx, fmt.Sprint(foreignFile.ID))
	requireVisible("a foreign row stays reachable by id", visible, err)
	visible, err = b.recordIsVisible(ctx, "4242")
	requireVisible("an unknown id stays acceptable", visible, err)

	obj, ok := wrapFirst(b, scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(foreignFile.ID)},
	}), &web.EventResponse{})
	require.True(t, ok)
	require.Equal(t, foreignFile.ID, obj.ID)

	_, err = moveToFolder(b)(scopeEventContext(t, pb, 1, url.Values{
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

// serveEvent drives a presets event func through the presets handler, so the
// generic CRUD guards are exercised the way a real request reaches them.
func serveEvent(t *testing.T, pb *presets.Builder, userID uint, params url.Values) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/media-libraries?"+params.Encode(), http.NoBody)
	req.Header.Set(scopeTestUserHeader, fmt.Sprint(userID))
	rec := httptest.NewRecorder()
	pb.ServeHTTP(rec, req)
	return rec.Body.String()
}

// The model's generic presets event funcs address rows by primary key through
// the DataOperator, bypassing the media event funcs entirely.
func TestPresetsCRUDEventsScoped(t *testing.T) {
	_, pb := scopeTestBuilder(t)
	pb.Build()
	_, ownFile, _, foreignFile := seedScopeRows(t)

	body := serveEvent(t, pb, 1, url.Values{
		"__execute_event__": []string{"presets_DoDelete"},
		"id":                []string{fmt.Sprint(foreignFile.ID)},
	})
	require.Contains(t, body, "record not found")
	require.NotZero(t, reload(t, foreignFile.ID).ID, "a foreign row must survive presets_DoDelete")

	body = serveEvent(t, pb, 1, url.Values{
		"__execute_event__": []string{"presets_DetailingDrawer"},
		"id":                []string{fmt.Sprint(foreignFile.ID)},
	})
	require.NotContains(t, body, foreignFile.File.FileName, "presets_DetailingDrawer must not read a foreign row back")

	body = serveEvent(t, pb, 1, url.Values{
		"__execute_event__": []string{"presets_DoDelete"},
		"id":                []string{fmt.Sprint(ownFile.ID)},
	})
	require.NotContains(t, body, "record not found")
	var count int64
	require.NoError(t, testDB.Model(&media_library.MediaLibrary{}).Where("id = ?", ownFile.ID).Count(&count).Error)
	require.Zero(t, count, "an own row must still be deletable through presets_DoDelete")
}

// ListingCompo actions stay dispatchable even though the media library replaces
// the listing page, so the listing's generic search needs scoping too.
func TestListingCompoReloadScoped(t *testing.T) {
	_, pb := scopeTestBuilder(t)
	pb.Build()
	_, ownFile, _, foreignFile := seedScopeRows(t)

	action := fmt.Sprintf(`{"compo_type":"*presets.ListingCompo","compo":{"id":"media_libraries_page","per_page":100},"injector":"media_libraries","method":"OnReload","request":"{}"}`)
	req := httptest.NewRequest(http.MethodPost, "/media-libraries?__execute_event__=__dispatch_stateful_action__",
		strings.NewReader(url.Values{"__action__": []string{action}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(scopeTestUserHeader, "1")
	rec := httptest.NewRecorder()
	pb.ServeHTTP(rec, req)
	body := rec.Body.String()

	rowLink := func(id uint) string {
		// Row click renders as presets_Edit with the id as a query, JSON-escaped.
		return fmt.Sprintf(`query(\"id\", \"%d\")`, id)
	}
	require.NotContains(t, body, rowLink(foreignFile.ID),
		"the listing must not enumerate a foreign row")
	require.Contains(t, body, rowLink(ownFile.ID),
		"the listing must still enumerate own rows")
}

// Installing media must not enable detailing for the model: that would mount a
// detail route and change how listing rows behave for every app.
func TestInstallDoesNotEnableDetailing(t *testing.T) {
	b, _ := scopeTestBuilder(t)
	require.False(t, b.GetPresetsModelBuilder().HasDetailing(),
		"configScopedCRUD must guard detailing without enabling it")
}

// A Joins-based searcher is the case that breaks unqualified conditions and
// silently drops from UPDATEs.
func TestJoinSearcherScoped(t *testing.T) {
	b, pb := scopeTestJoinBuilder(t)
	ownFolder, ownFile, _, foreignFile := seedScopeRows(t)
	ownChild := mkRow(t, 1, false, ownFolder.ID, "own-child.png")
	foreignChild := mkRow(t, 2, false, ownFolder.ID, "foreign-child.png")
	ctx := scopeEventContext(t, pb, 1, url.Values{})

	visible, err := b.folderIsVisible(ctx, ownFolder.ID)
	require.NoError(t, err, "a join searcher must not make the condition ambiguous")
	require.True(t, visible, "an own folder must stay usable as a target")

	visible, err = b.recordIsVisible(ctx, fmt.Sprint(ownFile.ID))
	require.NoError(t, err)
	require.True(t, visible)

	visible, err = b.recordIsVisible(ctx, fmt.Sprint(foreignFile.ID))
	require.NoError(t, err)
	require.False(t, visible)

	// Deleting the folder reparents children; the scope must survive the write.
	_, err = doDelete(b)(scopeEventContext(t, pb, 1, url.Values{
		ParamMediaIDS: []string{fmt.Sprint(ownFolder.ID)},
	}))
	require.NoError(t, err)
	require.Zero(t, reload(t, ownChild.ID).ParentId, "own child must be reparented to root")
	require.Equal(t, ownFolder.ID, reload(t, foreignChild.ID).ParentId,
		"foreign child must not be reparented by another tenant's delete")
}
