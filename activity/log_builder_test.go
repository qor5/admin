package activity

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db          *gorm.DB
	pb          = presets.New()
	pageModel   = pb.Model(&Page{})
	widgetModel = pb.Model(&Widget{})
)

type (
	Page struct {
		ID          uint `gorm:"primary_key"`
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

	TestActivityLog struct {
		ActivityLog
	}

	TestActivityModel struct {
		ID          uint `gorm:"primary_key"`
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

	if err = db.AutoMigrate(&TestActivityModel{}, &ActivityLog{}); err != nil {
		panic(err)
	}

	m.Run()
}

func resetDB() {
	db.Exec("delete from test_activity_logs;")
	db.Exec("delete from test_activity_models;")

	// TODO: rename the table name
	db.Exec("DELETE FROM activity_logs")
}

func TestModelKeys(t *testing.T) {
	resetDB()

	if err := db.AutoMigrate(&ActivityLog{}); err != nil {
		panic(err)
	}

	builder := New(db)
	pb.Use(builder)
	builder.RegisterModel(pageModel).AddKeys("ID", "VersionName")

	var err error

	err = builder.AddCreateRecord("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	record := &ActivityLog{}

	db.Debug().First(&record)

	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.ModelKeys != "1:v1" {
		fmt.Println(record.ModelKeys)
		t.Errorf("want the keys %v, but got %v", "1:v1", record.ModelKeys)
	}

	resetDB()

	builder.RegisterModel(widgetModel).AddKeys("Name")
	err = builder.AddCreateRecord("b", Widget{Name: "Text 01", Title: "123"}, db)
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	record2 := &ActivityLog{}

	db.Debug().First(&record)

	if err := db.First(record2).Error; err != nil {
		t.Fatal(err)
	}
	if record2.ModelKeys != "Text 01" {
		t.Errorf("want the keys %v, but got %v", "Text 01", record2.ModelKeys)
	}
}

func TestModelLink(t *testing.T) {
	builder := New(db)
	builder.Install(pb)
	builder.RegisterModel(pageModel).LinkFunc(func(v any) string {
		page := v.(Page)
		return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
	})

	resetDB()

	builder.AddCreateRecord("a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.ModelLink != "/admin/pages/1?version=v1" {
		t.Errorf("want the link %v, but got %v", "/admin/pages/1?version=v1", record.ModelLink)
	}
}

func TestModelTypeHanders(t *testing.T) {
	builder := New(db)
	builder.Install(pb)
	builder.RegisterModel(pageModel).AddTypeHanders(Widgets{}, func(old, now any, prefixField string) (diffs []Diff) {
		oldWidgets := old.(Widgets)
		nowWidgets := now.(Widgets)

		var (
			oldLen  = len(oldWidgets)
			nowLen  = len(nowWidgets)
			minLen  int
			added   bool
			deleted bool
		)

		if oldLen > nowLen {
			minLen = nowLen
			deleted = true
		}

		if oldLen < nowLen {
			minLen = oldLen
			added = true
		}

		if oldLen == nowLen {
			minLen = oldLen
		}

		for i := 0; i < minLen; i++ {
			if oldWidgets[i].Name != nowWidgets[i].Name {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: oldWidgets[i].Name, Now: nowWidgets[i].Name})
			}
		}

		if added {
			for i := minLen; i < nowLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: "", Now: nowWidgets[i].Name})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				diffs = append(diffs, Diff{Field: newPrefixField, Old: oldWidgets[i].Name, Now: ""})
			}
		}
		return diffs
	})

	resetDB()
	builder.AddEditRecordWithOld("a",
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
		}, db)
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	wants := `[{"Field":"VersionName","Old":"v1","Now":"v2"},{"Field":"Title","Old":"test","Now":"test1"},{"Field":"Widgets.0","Old":"Text 01","Now":"Text 011"},{"Field":"Widgets.1","Old":"HeroBanner 02","Now":"HeroBanner 022"},{"Field":"Widgets.3","Old":"","Now":"Video 03"}]`
	if record.ModelDiffs != wants {
		t.Errorf("want the diffs %v, but got %v", wants, record.ModelDiffs)
	}
}

type user struct{}

func (u user) GetID() uint {
	return 10
}

func (u user) GetName() string {
	return "user a"
}

func TestCreatorInferface(t *testing.T) {
	builder := New(db)
	builder.Install(pb)

	builder.RegisterModel(pageModel)
	resetDB()

	builder.AddCreateRecord(user{}, Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := &ActivityLog{}
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.Creator != "user a" {
		t.Errorf("want the creator %v, but got %v", "user a", record.Creator)
	}
	if record.UserID != 10 {
		t.Errorf("want the creator id %v, but got %v", 10, record.UserID)
	}
}

func TestGetActivityLogs(t *testing.T) {
	builder := New(db)
	builder.Install(pb)

	builder.RegisterModel(Page{}).AddKeys("ID", "VersionName")
	resetDB()

	err := builder.AddCreateRecord("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	if err != nil {
		t.Fatalf("failed to add create record: %v", err)
	}
	err = builder.AddEditRecordWithOld("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, Page{ID: 1, VersionName: "v1", Title: "test1"}, db)
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}
	err = builder.AddEditRecordWithOld("creator a", Page{ID: 1, VersionName: "v1", Title: "test1"}, Page{ID: 1, VersionName: "v1", Title: "test2"}, db)
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}
	err = builder.AddEditRecordWithOld("creator a", Page{ID: 2, VersionName: "v1", Title: "test1"}, Page{ID: 2, VersionName: "v1", Title: "test2"}, db)
	if err != nil {
		t.Fatalf("failed to add edit record: %v", err)
	}

	logs := builder.GetActivityLogs(Page{ID: 1, VersionName: "v1"}, db)
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, but got %d", len(logs))
	}

	expectedActions := []string{"Create", "Edit", "Edit"}
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
		if log.Creator != "creator a" {
			t.Errorf("expected creator 'creator a', but got %s", log.Creator)
		}
	}
}

func TestMutliModelBuilder(t *testing.T) {
	builder := New(db).CreatorContextKey("creator")
	builder.Install(pb)
	pb.DataOperator(gorm2op.DataOperator(db))

	pageModel2 := pb.Model(&TestActivityModel{}).URIName("page-02").Label("Page-02")
	pageModel3 := pb.Model(&TestActivityModel{}).URIName("page-03").Label("Page-03")

	builder.RegisterModel(&TestActivityModel{}).Keys("ID")
	builder.RegisterModel(pageModel2).Keys("ID").SkipDelete().AddIgnoredFields("VersionName")
	builder.RegisterModel(pageModel3).Keys("ID").SkipCreate().AddIgnoredFields("Description")

	data1 := &TestActivityModel{ID: 1, VersionName: "v1", Title: "test1", Description: "Description1"}
	data2 := &TestActivityModel{ID: 2, VersionName: "v2", Title: "test2", Description: "Description3"}
	data3 := &TestActivityModel{ID: 3, VersionName: "v3", Title: "test3", Description: "Description3"}

	resetDB()
	// add create record
	db.Create(data1)
	builder.AddCreateRecord("Test User", data1, db)
	pageModel2.Editing().Saver(data2, "2", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})
	pageModel3.Editing().Saver(data3, "3", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})
	{
		for _, id := range []string{"1", "2"} {
			var log TestActivityLog
			if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Create", "TestActivityModel", id).Find(&log); log.ID == 0 {
				t.Errorf("want the log %v, but got %v", "TestActivityModel:"+id, log)
			}
		}

		var log TestActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Create", "TestActivityModel", 3).Find(&log); log.ID != 0 {
			t.Errorf("want skip the create, but still got the record %v", log)
		}
	}

	// add edit record
	data1.Title = "test1-1"
	data1.Description = "Description1-1"
	builder.AddEditRecord("Test User", data1, db)
	db.Save(data1)

	data2.Title = "test2-1"
	data2.Description = "Description2-1"
	pageModel2.Editing().Saver(data2, "2", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})

	data3.Title = "test3-1"
	data3.Description = "Description3-1"
	pageModel3.Editing().Saver(data3, "3", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})

	{
		var log1 TestActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "1").Find(&log1); log1.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:1", log1)
		}
		if log1.ModelDiffs != `[{"Field":"Title","Old":"test1","Now":"test1-1"},{"Field":"Description","Old":"Description1","Now":"Description1-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test1","Now":"test1-1"},{"Field":"Description","Old":"Description1","Now":"Description1-1"}]`, log1.ModelDiffs)
		}

		var log2 TestActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "2").Find(&log2); log2.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:2", log2)
		}
		if log2.ModelDiffs != `[{"Field":"Title","Old":"test2","Now":"test2-1"},{"Field":"Description","Old":"Description3","Now":"Description2-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test2","Now":"test2-1"},{"Field":"Description","Old":"Description3","Now":"Description2-1"}]`, log1.ModelDiffs)
		}

		if log2.ModelLabel != "page-02" {
			t.Errorf("want the log %v, but got %v", "page-02", log2.ModelLabel)
		}

		var log3 TestActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Edit", "TestActivityModel", "3").Find(&log3); log3.ID == 0 {
			t.Errorf("want the log %v, but got %v", "TestActivityModel:3", log3)
		}
		if log3.ModelDiffs != `[{"Field":"Title","Old":"test3","Now":"test3-1"}]` {
			t.Errorf("want the log %v, but got %v", `[{"Field":"Title","Old":"test3","Now":"test3-1"}]`, log1.ModelDiffs)
		}

		if log3.ModelLabel != "page-03" {
			t.Errorf("want the log %v, but got %v", "page-03", log2.ModelLabel)
		}

	}

	// // add delete record
	builder.AddDeleteRecord("Test User", data1, db)
	db.Delete(data1)

	pageModel2.Editing().Deleter(data2, "2", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-01/2", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})
	pageModel3.Editing().Deleter(data3, "3", &web.EventContext{R: httptest.NewRequest("POST", "/admin/page-02/3", nil).WithContext(context.WithValue(context.Background(), "creator", "Test User"))})
	{
		for _, id := range []string{"1", "3"} {
			var log TestActivityLog
			if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Delete", "TestActivityModel", id).Find(&log); log.ID == 0 {
				t.Errorf("want the log %v, but got %v", "TestActivityModel:"+id, log)
			}
		}

		var log TestActivityLog
		if db.Where("action = ? AND model_name = ? AND model_keys = ?", "Delete", "TestActivityModel", "2").Find(&log); log.ID != 0 {
			t.Errorf("want skip the create, but still got the record %v", log)
		}
	}
}
