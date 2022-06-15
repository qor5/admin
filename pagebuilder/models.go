package pagebuilder

import (
	"context"
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

func (p *Page) AfterSaveNewVersion(db *gorm.DB, ctx context.Context) (err error) {
	var b *Builder
	var ok bool
	if b, ok = ctx.Value("pagebuilder").(*Builder); !ok || b == nil {
		return
	}
	if err = b.CopyContainers(int(p.ID), p.ParentVersion, p.GetVersion()); err != nil {
		return err
	}

	return
}

type Container struct {
	gorm.Model
	PageID       uint
	PageVersion  string
	Name         string
	ModelID      uint
	DisplayOrder float64
}

func (*Container) TableName() string {
	return "page_builder_containers"
}
