package pagebuilder

import (
	"fmt"
	"strings"
	"time"

	"github.com/qor/qor5/publish"
)

type Page struct {
	ID    uint `gorm:"primarykey"`
	Title string
	Slug  string
	Ti    *time.Time

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
	ID           uint
	PageID       uint
	Name         string
	ModelID      uint
	DisplayOrder float64
}

func (*Container) TableName() string {
	return "page_builder_containers"
}
