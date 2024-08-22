package pagebuilder

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/sunfmin/reflectutils"
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

type PageTitleInterface interface {
	GetTitle() string
}

func (p *Page) GetTitle() string {
	return p.Title
}

func (p *Page) GetID() uint {
	return p.ID
}

func (*Page) TableName() string {
	return "page_builder_pages"
}

func primarySlug(v interface{}) string {
	locale := v.(l10n.LocaleInterface).EmbedLocale().LocaleCode
	version := v.(publish.VersionInterface).EmbedVersion().Version
	id := reflectutils.MustGet(v, "ID")
	if locale == "" {
		return fmt.Sprintf("%v_%v", id, version)
	}

	return fmt.Sprintf("%v_%v_%v", id, version, locale)
}

func primarySlugWithoutVersion(v interface{}) string {
	locale := v.(l10n.LocaleInterface).EmbedLocale().LocaleCode
	id := reflectutils.MustGet(v, "ID")
	if locale == "" {
		return fmt.Sprintf("%v", id)
	}

	return fmt.Sprintf("%v_%v", id, locale)
}

func primaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 && len(segs) != 3 {
		panic("wrong slug")
	}
	if len(segs) == 2 {
		return map[string]string{
			"id":                segs[0],
			publish.SlugVersion: segs[1],
		}
	}
	return map[string]string{
		"id":                segs[0],
		publish.SlugVersion: segs[1],
		l10n.SlugLocaleCode: segs[2],
	}
}

func primaryColumnValuesBySlugWithoutVersion(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) > 2 {
		panic("wrong slug")
	}
	if len(segs) == 1 {
		return map[string]string{
			"id": segs[0],
		}
	}
	return map[string]string{
		"id":                segs[0],
		l10n.SlugLocaleCode: segs[1],
	}
}

func (p *Page) PrimarySlug() string {
	return primarySlug(p)
}

func (p *Page) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlug(slug)
}

func (p *Page) PermissionRN() []string {
	rn := []string{"pages", strconv.Itoa(int(p.ID)), p.Version.Version}
	if len(p.LocaleCode) > 0 {
		rn = append(rn, p.LocaleCode)
	}
	return rn
}

func (p *Page) GetCategory(db *gorm.DB) (category Category, err error) {
	err = db.Where("id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type Category struct {
	gorm.Model
	Name        string
	Path        string
	Description string

	IndentLevel int `gorm:"-"`

	l10n.Locale
}

func (c *Category) PrimarySlug() string {
	return primarySlugWithoutVersion(c)
}

func (c *Category) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlugWithoutVersion(slug)
}

func (*Category) TableName() string {
	return "page_builder_categories"
}

type Container struct {
	gorm.Model
	PageID        uint
	PageVersion   string
	PageModelName string
	ModelName     string
	ModelID       uint
	DisplayOrder  float64
	Shared        bool
	Hidden        bool
	DisplayName   string

	l10n.Locale
	LocalizeFromModelID uint
}

func (c *Container) PrimarySlug() string {
	return primarySlugWithoutVersion(c)
}

func (c *Container) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlugWithoutVersion(slug)
}

func (*Container) TableName() string {
	return "page_builder_containers"
}

type DemoContainer struct {
	gorm.Model
	ModelName string
	ModelID   uint
	Filled    bool

	l10n.Locale
}

func (c *DemoContainer) PrimarySlug() string {
	return primarySlugWithoutVersion(c)
}

func (c *DemoContainer) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlugWithoutVersion(slug)
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

func (t *Template) GetID() uint {
	return t.ID
}

func (t *Template) PrimarySlug() string {
	return primarySlugWithoutVersion(t)
}

func (t *Template) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlugWithoutVersion(slug)
}

func (*Template) TableName() string {
	return "page_builder_templates"
}

const templateVersion = "tpl"

func (t *Template) Page() *Page {
	return &Page{
		Model: t.Model,
		Title: t.Name,
		Slug:  "",
		Status: publish.Status{
			Status:    publish.StatusDraft,
			OnlineUrl: "",
		},
		Schedule: publish.Schedule{},
		Version: publish.Version{
			Version: templateVersion,
		},
		Locale: t.Locale,
	}
}

type (
	PrimarySlugInterface interface {
		PrimarySlug() string
		PrimaryColumnValuesBySlug(slug string) map[string]string
	}
)
