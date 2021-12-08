package activity

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/goplaid/x/presets"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	pb          = presets.New()
	pageModel   = pb.Model(&Page{})
	widgetModel = pb.Model(&Widget{})
)

type (
	Page struct {
		ID          uint
		VersionName string
		Title       string
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
)

func init() {
	var err error
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{}); err != nil {
		panic(err)
	}
}

func resetDB() {
	db.Exec("truncate test_activity_logs;")
}

func TestModelKeys(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel).AddKeys("ID", "VersionName")
	resetDB()
	builder.AddCreateRecord("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetModelKeys() != "1:v1" {
		t.Errorf("want the keys %v, but got %v", "1:v1", record.GetModelKeys())
	}

	resetDB()
	builder.RegisterModel(widgetModel).AddKeys("Name")
	builder.AddCreateRecord("b", Widget{Name: "Text 01", Title: "123"}, db)
	record2 := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record2).Error; err != nil {
		t.Fatal(err)
	}
	if record2.GetModelKeys() != "Text 01" {
		t.Errorf("want the keys %v, but got %v", "Text 01", record2.GetModelKeys())
	}
}

func TestModelLink(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel).SetLink(func(v interface{}) string {
		page := v.(Page)
		return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
	})

	resetDB()
	builder.AddCreateRecord("a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetModelLink() != "/admin/pages/1?version=v1" {
		t.Errorf("want the link %v, but got %v", "/admin/pages/1?version=v1", record.GetModelLink())
	}
}

func TestModelTypeHanders(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel).AddTypeHanders(Widgets{}, func(old, now interface{}, prefixField string) (diffs []Diff) {
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
		Page{ID: 1, VersionName: "v1", Title: "test",
			Widgets: []Widget{
				{Name: "Text 01", Title: "test1"},
				{Name: "HeroBanner 02", Title: "banner 1"},
				{Name: "Card 03", Title: "cards 1"},
			}},
		Page{ID: 1, VersionName: "v2", Title: "test1",
			Widgets: []Widget{
				{Name: "Text 011", Title: "test1"},
				{Name: "HeroBanner 022", Title: "banner 1"},
				{Name: "Card 03", Title: "cards 1"},
				{Name: "Video 03", Title: "video 1"},
			},
		}, db)
	record := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	wants := `[{"Field":".VersionName","Old":"v1","Now":"v2"},{"Field":".Title","Old":"test","Now":"test1"},{"Field":".Widgets.0","Old":"Text 01","Now":"Text 011"},{"Field":".Widgets.1","Old":"HeroBanner 02","Now":"HeroBanner 022"},{"Field":".Widgets.3","Old":"","Now":"Video 03"}]`
	if record.GetModelDiffs() != wants {
		t.Errorf("want the diffs %v, but got %v", wants, record.GetModelDiffs())
	}
}

func TestCreator(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel)
	resetDB()
	builder.AddCreateRecord("user a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetCreator() != "user a" {
		t.Errorf("want the creator %v, but got %v", "a", record.GetCreator())
	}
}

type user struct {
}

func (u user) GetID() uint {
	return 10
}
func (u user) GetName() string {
	return "user a"
}
func TestCreatorInferface(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel)
	resetDB()

	builder.AddCreateRecord(user{}, Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModelData().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetCreator() != "user a" {
		t.Errorf("want the creator %v, but got %v", "a", record.GetCreator())
	}
	if record.GetUserID() != 10 {
		t.Errorf("want the creator id %v, but got %v", 10, record.GetUserID())
	}
}

func TestGetActivityLogs(t *testing.T) {
	builder := New(pb, db, &TestActivityLog{})
	builder.RegisterModel(pageModel).AddKeys("ID", "VersionName")
	resetDB()

	builder.AddCreateRecord("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	builder.AddEditRecordWithOld("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, Page{ID: 1, VersionName: "v1", Title: "test1"}, db)
	builder.AddEditRecordWithOld("creator a", Page{ID: 1, VersionName: "v1", Title: "test1"}, Page{ID: 1, VersionName: "v1", Title: "test2"}, db)
	builder.AddEditRecordWithOld("creator a", Page{ID: 2, VersionName: "v1", Title: "test1"}, Page{ID: 2, VersionName: "v1", Title: "test2"}, db)

	logs := builder.GetCustomizeActivityLogs(Page{ID: 1, VersionName: "v1"}, db)
	testlogs, ok := logs.(*[]*TestActivityLog)
	if !ok {
		t.Errorf("want the logs type %v, but got %v", "*[]*TestActivityLog", reflect.TypeOf(logs))
	}

	if len(*testlogs) != 3 {
		t.Errorf("want the logs length %v, but got %v", 3, len(*testlogs))
	}

	if (*testlogs)[0].Action != "Create" || (*testlogs)[0].ModelName != "Page" || (*testlogs)[0].ModelKeys != "1:v1" || (*testlogs)[0].Creator != "creator a" {
		t.Errorf("want the logs %v, but got %+v", "Create:Page:1:v1:creator a", (*testlogs)[0])
	}

	if (*testlogs)[1].Action != "Edit" || (*testlogs)[1].ModelName != "Page" || (*testlogs)[1].ModelKeys != "1:v1" || (*testlogs)[1].Creator != "creator a" {
		t.Errorf("want the logs %v, but got %v", "Edit:Page:1:v1:creator a", (*testlogs)[1])
	}

	if (*testlogs)[2].Action != "Edit" || (*testlogs)[2].ModelName != "Page" || (*testlogs)[2].ModelKeys != "1:v1" || (*testlogs)[2].Creator != "creator a" {
		t.Errorf("want the logs %v, but got %v", "Edit:Page:1:v1:creator a", (*testlogs)[2])
	}

}
