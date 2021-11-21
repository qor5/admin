package activity

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	table = Activity().logModel
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
)

func init() {
	var err error
	if db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{}); err != nil {
		panic(err)
	}

	db.AutoMigrate(table)
}

func TestModelKeys(t *testing.T) {
	builder := Activity()
	builder.RegisterModel(Page{}).AddKeys("ID", "VersionName")

	db.Where("1 = 1").Delete(table)
	builder.AddCreateRecord("creator a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModel().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetModelKeys() != "1:v1" {
		t.Errorf("want the keys %v, but got %v", "1:v1", record.GetModelKeys())
	}

	db.Where("1 = 1").Delete(table)
	builder.RegisterModel(Widget{}).AddKeys("Name")
	builder.AddCreateRecord("b", Widget{Name: "Text 01", Title: "123"}, db)
	record2 := builder.NewLogModel().(ActivityLogInterface)
	if err := db.First(record2).Error; err != nil {
		t.Fatal(err)
	}
	if record2.GetModelKeys() != "Text 01" {
		t.Errorf("want the keys %v, but got %v", "Text 01", record2.GetModelKeys())
	}
}

func TestModelLink(t *testing.T) {
	builder := Activity()
	builder.RegisterModel(Page{}).SetLink(func(v interface{}) string {
		page := v.(Page)
		return fmt.Sprintf("/admin/pages/%d?version=%s", page.ID, page.VersionName)
	})

	db.Where("1 = 1").Delete(table)
	builder.AddCreateRecord("a", Page{ID: 1, VersionName: "v1", Title: "test"}, db)
	record := builder.NewLogModel().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	if record.GetModelLink() != "/admin/pages/1?version=v1" {
		t.Errorf("want the link %v, but got %v", "/admin/pages/1?version=v1", record.GetModelLink())
	}
}

func TestModelTypeHanders(t *testing.T) {
	builder := Activity()
	builder.RegisterModel(Page{}).AddTypeHanders(Widgets{}, func(old, now interface{}, prefixField string) (diffs []Diff) {
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

	db.Where("1 = 1").Delete(table)
	builder.AddEditRecord("a",
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
	record := builder.NewLogModel().(ActivityLogInterface)
	if err := db.First(record).Error; err != nil {
		t.Fatal(err)
	}
	wants := `[{"Field":".VersionName","Old":"v1","Now":"v2"},{"Field":".Title","Old":"test","Now":"test1"},{"Field":".Widgets.0","Old":"Text 01","Now":"Text 011"},{"Field":".Widgets.1","Old":"HeroBanner 02","Now":"HeroBanner 022"},{"Field":".Widgets.3","Old":"","Now":"Video 03"}]`
	if record.GetModelDiffs() != wants {
		t.Errorf("want the diffs %v, but got %v", wants, record.GetModelDiffs())
	}
}
