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

type Category struct {
	gorm.Model
	Name      string
	Path      string
	Desc      string
	ModelType string
	// BreadcrumbsPosition       int
	// ShowOnSubNav              bool
	// HideCategoryOnBreadcrumbs bool
}

func (*Category) TableName() string {
	return "page_builder_categories"
}

type Container struct {
	gorm.Model
	PageID       uint
	PageVersion  string
	Name         string
	ModelID      uint
	DisplayOrder float64
	Shared       bool
	DisplayName  string
}

func (*Container) TableName() string {
	return "page_builder_containers"
}
