package examples_presets

import (
	"net/http"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var sectionDemoData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO section_demos (id, name, email, age, description, created_at, updated_at) VALUES 
(1, 'Test User', 'test@example.com', 25, 'Test description', '2024-01-01 00:00:00', '2024-01-01 00:00:00');
`, []string{"section_demos"}))

var sectionDemoDataWithItems = gofixtures.Data(gofixtures.Sql(`
INSERT INTO section_demos (id, name, email, age, description, created_at, updated_at) VALUES 
(1, 'Test User', 'test@example.com', 25, 'Test description', '2024-01-01 00:00:00', '2024-01-01 00:00:00');
`, []string{"section_demos", "section_demo_items"}))

// TestPresetsSectionSingleton tests singleton editing with sections
func TestPresetsSectionSingleton(t *testing.T) {
	pb := presets.New()
	PresetsSectionSingleton(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "singleton page renders",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"BasicInfo", "AdditionalInfo"},
		},
		{
			Name:  "singleton save with valid data",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update+"&id=1").
					AddField("Name", "Updated Name").
					AddField("Email", "updated@example.com").
					AddField("Age", "30").
					AddField("Description", "Updated description").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var demo SectionDemo
				TestDB.First(&demo, 1)
				if demo.Name != "Updated Name" {
					t.Errorf("expected Name to be 'Updated Name', got '%s'", demo.Name)
				}
				if demo.Email != "updated@example.com" {
					t.Errorf("expected Email to be 'updated@example.com', got '%s'", demo.Email)
				}
			},
		},
		{
			Name:  "singleton validation error - empty name",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update+"&id=1").
					AddField("Name", "").
					AddField("Email", "test@example.com").
					AddField("Age", "30").
					AddField("Description", "Description").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name is required"},
		},
		{
			Name:  "singleton cross-field validation",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update+"&id=1").
					AddField("Name", "Test").
					AddField("Email", "test@example.com").
					AddField("Age", "105").
					AddField("Description", "Young person").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"senior"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

// TestPresetsSectionDetailingNormal tests normal editing + detailing with section
func TestPresetsSectionDetailingNormal(t *testing.T) {
	pb := presets.New()
	PresetsSectionDetailingNormal(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "create new record via editing",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				TestDB.Exec("DELETE FROM section_demos")
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update).
					AddField("Name", "John").
					AddField("Email", "john@example.com").
					AddField("Age", "25").
					AddField("Description", "A developer").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var count int64
				TestDB.Model(&SectionDemo{}).Count(&count)
				if count != 1 {
					t.Errorf("expected 1 record, got %d", count)
				}
			},
		},
		{
			Name:  "editing validation - empty fields",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update).
					AddField("Name", "").
					AddField("Email", "").
					AddField("Age", "25").
					AddField("Description", "Test").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name is required"},
		},
		{
			Name:  "detailing page renders section",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__=presets_DetailingDrawer&id=1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Details"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

// TestPresetsSectionEditingClone tests editing with cloned sections
func TestPresetsSectionEditingClone(t *testing.T) {
	pb := presets.New()
	PresetsSectionEditingClone(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "create new record with section clone",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				TestDB.Exec("DELETE FROM section_demos")
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update).
					AddField("Name", "Alice").
					AddField("Email", "alice@example.com").
					AddField("Age", "28").
					AddField("Description", "Engineer").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var demo SectionDemo
				TestDB.First(&demo)
				if demo.Name != "Alice" {
					t.Errorf("expected Name to be 'Alice', got '%s'", demo.Name)
				}
				if demo.Email != "alice@example.com" {
					t.Errorf("expected Email to be 'alice@example.com', got '%s'", demo.Email)
				}
			},
		},
		{
			Name:  "edit existing record with section clone",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update+"&id=1").
					AddField("Name", "Alice Updated").
					AddField("Email", "alice.updated@example.com").
					AddField("Age", "29").
					AddField("Description", "Senior Engineer").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var demo SectionDemo
				TestDB.First(&demo, 1)
				if demo.Name != "Alice Updated" {
					t.Errorf("expected Name to be 'Alice Updated', got '%s'", demo.Name)
				}
			},
		},
		{
			Name:  "section validation - name too long",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update).
					AddField("Name", strings.Repeat("a", 60)).
					AddField("Email", "test@example.com").
					AddField("Age", "25").
					AddField("Description", "Test").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"less than 50"},
		},
		{
			Name:  "cross-field validation - age without description",
			Debug: true,
			ReqFunc: func() *http.Request {
				sectionDemoData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/section-demos?__execute_event__="+actions.Update).
					AddField("Name", "Test").
					AddField("Email", "test@example.com").
					AddField("Age", "25").
					AddField("Description", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Description is required when age is specified"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
