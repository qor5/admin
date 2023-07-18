package pagebuilder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/publish"
	"github.com/qor5/admin/seo"
	"gorm.io/gorm"
)

type Page struct {
	gorm.Model
	Title      string
	Slug       string
	CategoryID uint

	SEO seo.Setting
	publish.Status
	publish.Schedule
	publish.Version
	l10n.Locale
}

func (*Page) TableName() string {
	return "page_builder_pages"
}

var l10nON bool

func (p *Page) L10nON() {
	l10nON = true
	return
}

func (p *Page) PrimarySlug() string {
	if !l10nON {
		return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
	}
	return fmt.Sprintf("%v_%v_%v", p.ID, p.Version.Version, p.LocaleCode)
}

func (p *Page) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if !l10nON {
		if len(segs) != 2 {
			panic("wrong slug")
		}

		return map[string]string{
			"id":      segs[0],
			"version": segs[1],
		}
	}
	if len(segs) != 3 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"version":     segs[1],
		"locale_code": segs[2],
	}
}

func (p *Page) PermissionRN() []string {
	rn := []string{"pages", strconv.Itoa(int(p.ID)), p.Version.Version}
	if l10nON {
		rn = append(rn, p.LocaleCode)
	}
	return rn
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

	l10n.Locale
}

func (this *Container) PrimarySlug() string {
	if !l10nON {
		return fmt.Sprintf("%v", this.ID)
	}
	return fmt.Sprintf("%v_%v", this.ID, this.LocaleCode)
}

func (this *Container) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if !l10nON {
		if len(segs) != 1 {
			panic("wrong slug")
		}

		return map[string]string{
			"id": segs[0],
		}
	}
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"locale_code": segs[1],
	}
}

func (*Container) TableName() string {
	return "page_builder_containers"
}

type DemoContainer struct {
	gorm.Model
	ModelName string
	ModelID   uint

	l10n.Locale
}

func (this *DemoContainer) PrimarySlug() string {
	if !l10nON {
		return fmt.Sprintf("%v", this.ID)
	}
	return fmt.Sprintf("%v_%v", this.ID, this.LocaleCode)
}

func (this *DemoContainer) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if !l10nON {
		if len(segs) != 1 {
			panic("wrong slug")
		}

		return map[string]string{
			"id": segs[0],
		}
	}
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"locale_code": segs[1],
	}
}

func (*DemoContainer) TableName() string {
	return "page_builder_demo_containers"
}

type Template struct {
	gorm.Model
	Name        string
	Description string

	l10n.Locale
}

func (this *Template) PrimarySlug() string {
	if !l10nON {
		return fmt.Sprintf("%v", this.ID)
	}
	return fmt.Sprintf("%v_%v", this.ID, this.LocaleCode)
}

func (this *Template) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if !l10nON {
		if len(segs) != 1 {
			panic("wrong slug")
		}

		return map[string]string{
			"id": segs[0],
		}
	}
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"locale_code": segs[1],
	}
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
		Locale: m.Locale,
	}
}
