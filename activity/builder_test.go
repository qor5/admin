package activity

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/stretchr/testify/require"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

type (
	Page struct {
		ID          uint `gorm:"primaryKey"`
		VersionName string
		Title       string
		Description string
		Widgets     Widgets
	}
	Widgets []Widget
	Widget  struct {
		Name  string
		Title string
	}

	TestActivityModel struct {
		ID          uint `gorm:"primaryKey"`
		VersionName string
		Title       string
		Description string
	}
)

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()

	db = env.DB
	db.Logger = db.Logger.LogMode(logger.Info)

	if err = AutoMigrate(db, ""); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&TestActivityModel{}); err != nil {
		panic(err)
	}

	m.Run()
}

var currentUser = &User{
	ID:     "1",
	Name:   "John",
	Avatar: "https://i.pravatar.cc/300",
}

var anotherUser = &User{
	ID:     "2",
	Name:   "Sam",
	Avatar: "https://i.pravatar.cc/300",
}

type ctxKeyCurrentUser struct{}

func testCurrentUser(ctx context.Context) (*User, error) {
	u, ok := ctx.Value(ctxKeyCurrentUser{}).(*User)
	if ok {
		return u, nil
	}
	return currentUser, nil
}

func resetDB() {
	db.Exec("delete from test_activity_models;")
	db.Exec("DELETE FROM activity_logs")
	db.Exec("DELETE FROM activity_users")
}

func TestModelKeys(t *testing.T) {
	pb := presets.New()
	pageModel := pb.Model(&Page{})
	widgetModel := pb.Model(&Widget{})

	resetDB()

	if err := AutoMigrate(db, ""); err != nil {
		panic(err)
	}

	builder := New(db, testCurrentUser)
	pb.Use(builder)
	builder.RegisterModel(pageModel).AddKeys("ID", "VersionName")

	var err error
	ctx := context.Background()

	_, err = builder.OnCreate(ctx, Page{ID: 1, VersionName: "v1", Title: "test"})
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.ModelKeys != "1:v1" {
		t.Errorf("want the keys %v, but got %v", "1:v1", record.ModelKeys)
	}

	resetDB()

	builder.RegisterModel(widgetModel).AddKeys("Name")
	_, err = builder.OnCreate(ctx, Widget{Name: "Text 01", Title: "123"})
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	record2 := &ActivityLog{}
	if err := db.First(record2).Error; err != nil {
		t.Fatal(err)
	}
	if record2.ModelKeys != "Text 01" {
		t.Errorf("want the keys %v, but got %v", "Text 01", record2.ModelKeys)
	}
}

func TestModelLink(t *testing.T) {
	pb := presets.New()
	pageModel := pb.Model(&Page{})

	builder := New(db, testCurrentUser)
	builder.Install(pb)
	builder.RegisterModel(pageModel).LinkFunc(func(v any) string {
		page := v.(Page)
		return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
	})

	resetDB()

	ctx := context.Background()
	builder.OnCreate(ctx, Page{ID: 1, VersionName: "v1", Title: "test"})
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.ModelLink != "/admin/pages/1?version=v1" {
		t.Errorf("want the link %v, but got %v", "/admin/pages/1?version=v1", record.ModelLink)
	}
}

func TestModelTypeHanders(t *testing.T) {
	pb := presets.New()
	pageModel := pb.Model(&Page{})

	builder := New(db, testCurrentUser)
	builder.Install(pb)
	builder.RegisterModel(pageModel).AddTypeHanders(Widgets{}, func(oldObj, newObj any, prefixField string) (diffs []Diff) {
		oldWidgets := oldObj.(Widgets)
		newWidgets := newObj.(Widgets)

		var (
			oldLen  = len(oldWidgets)
			newLen  = len(newWidgets)
			minLen  int
			added   bool
			deleted bool
		)

		if oldLen > newLen {
			minLen = newLen
			deleted = true
		}

		if oldLen < newLen {
			minLen = oldLen
			added = true
		}

		if oldLen == newLen {
			minLen = oldLen
		}

		for i := 0; i < minLen; i++ {
			if oldWidgets[i].Name != newWidgets[i].Name {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: oldWidgets[i].Name, New: newWidgets[i].Name})
			}
		}

		if added {
			for i := minLen; i < newLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: "", New: newWidgets[i].Name})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: oldWidgets[i].Name, New: ""})
			}
		}
		return diffs
	})

	resetDB()
	ctx := context.Background()
	builder.OnEdit(ctx,
		Page{
			ID: 1, VersionName: "v1", Title: "test",
			Widgets: []Widget{
				{Name: "Text 01", Title: "test1"},
				{Name: "HeroBanner 02", Title: "banner 1"},
				{Name: "Card 03", Title: "cards 1"},
			},
		},
		Page{
			ID: 1, VersionName: "v2", Title: "test1",
			Widgets: []Widget{
				{Name: "Text 011", Title: "test1"},
				{Name: "HeroBanner 022", Title: "banner 1"},
				{Name: "Card 03", Title: "cards 1"},
				{Name: "Video 03", Title: "video 1"},
			},
		})
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	wants := `[{"Field":"VersionName","Old":"v1","New":"v2"},{"Field":"Title","Old":"test","New":"test1"},{"Field":"Widgets.0","Old":"Text 01","New":"Text 011"},{"Field":"Widgets.1","Old":"HeroBanner 02","New":"HeroBanner 022"},{"Field":"Widgets.3","Old":"","New":"Video 03"}]`
	if record.Detail != wants {
		t.Errorf("want the diffs %v, but got %v", wants, record.Detail)
	}
}

func TestUser(t *testing.T) {
	pb := presets.New()
	pageModel := pb.Model(&Page{})

	builder := New(db, testCurrentUser)
	builder.Install(pb)

	builder.RegisterModel(pageModel)
	resetDB()

	ctx := context.Background()
	_, err := builder.OnCreate(ctx, Page{ID: 1, VersionName: "v1", Title: "test"})
	require.NoError(t, err)

	record := &ActivityLog{}
	err = db.First(record).Error
	require.NoError(t, err)
	require.Equal(t, "1", record.UserID)
}

func TestScope(t *testing.T) {
	pb := presets.New()
	pageModel := pb.Model(&Page{})

	builder := New(db, testCurrentUser)
	builder.Install(pb)

	builder.RegisterModel(pageModel)
	resetDB()

	ctx := context.Background()
	{
		_, err := builder.OnCreate(ctx, Page{ID: 1, VersionName: "v1", Title: "test"})
		require.NoError(t, err)

		record := &ActivityLog{}
		err = db.Order("created_at DESC").First(record).Error
		require.NoError(t, err)
		require.Equal(t, record.Scope, ",owner:1,")
	}
	{
		_, err := builder.OnCreate(
			ContextWithScope(ctx, fmt.Sprintf(",role:editor%s", ScopeWithOwner(currentUser.ID))),
			Page{ID: 2, VersionName: "v1", Title: "test"},
		)
		require.NoError(t, err)

		record := &ActivityLog{}
		err = db.Order("created_at DESC").First(record).Error
		require.NoError(t, err)
		require.Equal(t, ",role:editor,owner:1,", record.Scope)
	}
}

func TestGetActivityLogs(t *testing.T) {
	pb := presets.New()

	builder := New(db, testCurrentUser)
	pb.Use(builder)

	amb := builder.RegisterModel(Page{})
	amb.AddKeys("ID", "VersionName")
	resetDB()

	ctx := context.Background()
	_, err := builder.OnCreate(ctx, Page{ID: 1, VersionName: "v1", Title: "test"})
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	_, err = builder.OnEdit(ctx, Page{ID: 1, VersionName: "v1", Title: "test"}, Page{ID: 1, VersionName: "v1", Title: "test1"})
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}
	_, err = builder.OnEdit(
		context.WithValue(ctx, ctxKeyCurrentUser{}, anotherUser),
		Page{ID: 1, VersionName: "v1", Title: "test1"},
		Page{ID: 1, VersionName: "v1", Title: "test2"},
	)
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}
	_, err = builder.OnEdit(ctx, Page{ID: 2, VersionName: "v1", Title: "test1"}, Page{ID: 2, VersionName: "v1", Title: "test2"})
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}

	page := Page{ID: 1, VersionName: "v1"}
	logs, hasMore, err := builder.getActivityLogs(context.Background(), ParseModelName(page), amb.ParseModelKeys(page))
	require.NoError(t, err)
	require.Len(t, logs, 3)
	require.False(t, hasMore)

	expectedActions := []string{"Edit", "Edit", "Create"}
	for i, log := range logs {
		if log.Action != expectedActions[i] {
			t.Errorf("expected action %s, but got %s", expectedActions[i], log.Action)
		}
		if log.ModelName != "Page" {
			t.Errorf("expected model name 'Page', but got %s", log.ModelName)
		}
		if log.ModelKeys != "1:v1" {
			t.Errorf("expected model keys '1:v1', but got %s", log.ModelKeys)
		}
		if i == 0 {
			require.Equal(t, log.UserID, anotherUser.ID)
		} else {
			require.Equal(t, log.UserID, currentUser.ID)
		}
		require.Equal(t, log.Scope, ",owner:1,")
	}
}

type activityTestKey struct{}

func TestMutliModelBuilder(t *testing.T) {
	pb := presets.New()

	builder := New(db, testCurrentUser)
	builder.Install(pb)
	pb.DataOperator(gorm2op.DataOperator(db))

	pageModel2 := pb.Model(&TestActivityModel{}).URIName("page-02").Label("Page-02")
	pageModel3 := pb.Model(&TestActivityModel{}).URIName("page-03").Label("Page-03")

	builder.RegisterModel(&TestActivityModel{}).Keys("ID")
	builder.RegisterModel(pageModel2).Keys("ID").SkipDelete().AddIgnoredFields("VersionName")
	builder.RegisterModel(pageModel3).Keys("ID").SkipCreate().AddIgnoredFields("Description")

	data1 := &TestActivityModel{ID: 1, VersionName: "v1", Title: "test1", Description: "Description1"}
	data2 := &TestActivityModel{ID: 2, VersionName: "v2", Title: "test2", Description: "Description2"}
	data3 := &TestActivityModel{ID: 3, VersionName: "v3", Title: "test3", Description: "Description3"}

	resetDB()
	ctx := context.Background()

	// add create record
	db.Create(data1)
	builder.OnCreate(ctx, data1)
	pageModel2.Editing().Saver(data2, "", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})
	pageModel3.Editing().Saver(data3, "", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})
	{
		for _, id := range []string{"1", "2"} {
			var log ActivityLog
			if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Create", "TestActivityModel", id).Find(&log); log.ID == 0 {
				t.Errorf("want the log %v, but got %v", "TestActivityModel:"+id, log)
			}
		}

		var log ActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Create", "TestActivityModel", "3").Find(&log); log.ID != 0 {
			t.Errorf("want skip the create, but still got the record %v", log)
		}
	}

	// add edit record
	data1.Title = "test1-1"
	data1.Description = "Description1-1"

	old, err := FetchOld(db, data1)
	require.NoError(t, err)
	require.NoError(t, db.Save(data1).Error)
	_, err = builder.OnEdit(ctx, old, data1)
	require.NoError(t, err)

	data2.Title = "test2-1"
	data2.Description = "Description2-1"
	pageModel2.Editing().Saver(data2, "2", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})

	data3.Title = "test3-1"
	data3.Description = "Description3-1"
	pageModel3.Editing().Saver(data3, "3", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})

	{
		var log1 ActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "1").Find(&log1); log1.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:1", log1)
		}
		if log1.Detail != `[{"Field":"Title","Old":"test1","New":"test1-1"},{"Field":"Description","Old":"Description1","New":"Description1-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test1","New":"test1-1"},{"Field":"Description","Old":"Description1","New":"Description1-1"}]`, log1.Detail)
		}

		var log2 ActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "2").Find(&log2); log2.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:2", log2)
		}
		if log2.Detail != `[{"Field":"Title","Old":"test2","New":"test2-1"},{"Field":"Description","Old":"Description2","New":"Description2-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test2","New":"test2-1"},{"Field":"Description","Old":"Description2","New":"Description2-1"}]`, log1.Detail)
		}

		if log2.ModelLabel != "page-02" {
			t.Errorf("want the log %v, but got %v", "page-02", log2.ModelLabel)
		}

		var log3 ActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "3").Find(&log3); log3.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:3", log3)
		}
		if log3.Detail != `[{"Field":"Title","Old":"test3","New":"test3-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test3","New":"test3-1"}]`, log1.Detail)
		}

		if log3.ModelLabel != "page-03" {
			t.Errorf("want the log %v, but got %v", "page-03", log2.ModelLabel)
		}
	}

	// add delete record
	db.Delete(data1)
	builder.OnDelete(ctx, data1)
	pageModel2.Editing().Deleter(data2, "2", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})
	pageModel3.Editing().Deleter(data3, "3", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", http.NoBody).WithContext(context.WithValue(context.Background(), activityTestKey{}, "Test User"))})
	{
		for _, id := range []string{"1", "3"} {
			var log ActivityLog
			if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Delete", "TestActivityModel", id).Find(&log); log.ID == 0 {
				t.Errorf("want the log %v, but got %v", "TestActivityModel:"+id, log)
			}
		}

		var log ActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Delete", "TestActivityModel", "2").Find(&log); log.ID != 0 {
			t.Errorf("want skip the create, but still got the record %v", log)
		}
	}
}

func TestContextWithDB(t *testing.T) {
	{
		resetDB()

		ab := New(db, testCurrentUser).AutoMigrate()
		ab.RegisterModel(presets.New().Model(&Page{}))

		ctx := context.Background()
		{
			ctx := ContextWithDB(ctx, db)
			_, err := ab.Log(ctx, "Review", Page{ID: 1, VersionName: "v1", Title: "test"}, nil)
			require.NoError(t, err)
		}
		{
			// simulate a error to ensure the ContextWithDB is actually used
			cctx, ccancel := context.WithCancel(ctx)
			ccancel()

			ctx := ContextWithDB(ctx, db.WithContext(cctx))
			_, err := ab.Log(ctx, "Review", Page{ID: 1, VersionName: "v1", Title: "test"}, nil)
			require.ErrorIs(t, err, context.Canceled)
		}
	}

	// with table prefix
	{
		ab := New(db, testCurrentUser).TablePrefix("cms_").AutoMigrate()
		ab.RegisterModel(presets.New().Model(&Page{}))

		ctx := context.Background()
		{
			ctx := ContextWithDB(ctx, db)
			_, err := ab.Log(ctx, "Review", Page{ID: 1, VersionName: "v1", Title: "test"}, nil)
			require.NoError(t, err)
		}
		{
			// simulate a error to ensure the ContextWithDB is actually used
			cctx, ccancel := context.WithCancel(ctx)
			ccancel()

			ctx := ContextWithDB(ctx, db.WithContext(cctx))
			_, err := ab.Log(ctx, "Review", Page{ID: 1, VersionName: "v1", Title: "test"}, nil)
			require.ErrorIs(t, err, context.Canceled)
		}
	}
}
