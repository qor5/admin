package pagebuilder

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/spf13/cast"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
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
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	m := map[string]string{
		"id":                segs[0],
		publish.SlugVersion: segs[1],
	}
	if len(segs) > 2 {
		m[l10n.SlugLocaleCode] = segs[2]
	}
	return m
}

func primaryColumnValuesBySlugWithoutVersion(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 1 && len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	m := map[string]string{
		"id": segs[0],
	}
	if len(segs) == 2 {
		m[l10n.SlugLocaleCode] = segs[1]
	}
	return m
}

func (p *Page) PrimarySlug() string {
	return primarySlug(p)
}

func (p *Page) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return primaryColumnValuesBySlug(slug)
}

var PagePermissionRN = func(p *Page) []string {
	rn := []string{"pages", strconv.Itoa(int(p.ID)), p.Version.Version}
	if p.LocaleCode != "" {
		rn = append(rn, p.LocaleCode)
	}
	return rn
}

func (p *Page) PermissionRN() []string {
	return PagePermissionRN(p)
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

	ModelUpdatedAt time.Time
	ModelUpdatedBy string
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

type (
	Template struct {
		gorm.Model
		Name        string
		Description string

		l10n.Locale
	}
)

func (t *Template) GetID() uint {
	return t.ID
}

func (t *Template) GetName(_ *web.EventContext) string {
	return t.Name
}

func (t *Template) GetDescription(_ *web.EventContext) string {
	return t.Description
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
