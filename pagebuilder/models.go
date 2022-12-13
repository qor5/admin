package pagebuilder

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/publish"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Title      string
	Slug       string
	CategoryID uint

	publish.Status
	publish.Schedule
	publish.Version
	l10n.Locale
}

func (*Page) TableName() string {
	return "page_builder_pages"
}

func (p *Page) PrimarySlug() string {
	return fmt.Sprintf("%v_%v_%v", p.ID, p.Version.Version, p.LocaleCode)
}

func (p *Page) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 3 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"version":     segs[1],
		"locale_code": segs[2],
	}
}

type Category struct {
	gorm.Model
	Name        string
	Path        string
	Description string

	IndentLevel int `gorm:"-"`
}

func (*Category) TableName() string {
	return "page_builder_categories"
}

type Container struct {
	gorm.Model
	PageID       uint
	PageVersion  string
	ModelName    string
	ModelID      uint
	DisplayOrder float64
	Shared       bool
	Hidden       bool
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
	Name        string
	Description string
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
