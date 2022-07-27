package pagebuilder

import (
	"fmt"
	"strings"

	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Title string
	Slug  string

	publish.Status
	publish.Schedule
	publish.Version
}

func (*Page) TableName() string {
	return "page_builder_pages"
}

func (p *Page) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Page) PrimaryColumnValuesBySlug(slug string) [][]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return [][]string{
		{"id", segs[0]},
		{"version", segs[1]},
	}
}

type Container struct {
	gorm.Model
	PageID       uint
	PageVersion  string
	ModelName    string
	ModelID      uint
	DisplayOrder float64
	Shared       bool
	DisplayName  string
}

func (*Container) TableName() string {
	return "page_builder_containers"
}

type DemoContainer struct {
	gorm.Model
	ModelName string
	ModelID   uint
}

func (*DemoContainer) TableName() string {
	return "page_builder_demo_containers"
}

type Template struct {
	gorm.Model
	Name string
	Desc string
}

func (*Template) TableName() string {
	return "page_builder_templates"
}

const templateVersion = "tpl"

func (m *Template) Page() *Page {
	return &Page{
		Model: m.Model,
		Title: m.Name,
		Slug:  "",
		Status: publish.Status{
			Status:    publish.StatusDraft,
			OnlineUrl: "",
		},
		Schedule: publish.Schedule{},
		Version: publish.Version{
			Version: templateVersion,
		},
	}
}
