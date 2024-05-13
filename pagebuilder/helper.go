package pagebuilder

import (
	"context"
	"path"
	"regexp"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

var directoryRe = regexp.MustCompile(`^([\/]{1}[a-zA-Z0-9._-]+)+(\/?){1}$|^([\/]{1})$`)

const (
	queryLocaleCodeCategoryPathSlugSQL = `
	SELECT pages.id AS id,
	       pages.version AS version,
	       pages.locale_code AS locale_code,
	       categories.path AS category_path,
	       pages.slug AS slug
FROM page_builder_pages pages
LEFT JOIN page_builder_categories categories ON category_id = categories.id AND pages.locale_code = categories.locale_code
WHERE pages.deleted_at IS NULL AND categories.deleted_at IS NULL
`
	invalidPathMsg  = "Invalid Path"
	invalidSlugMsg  = "Invalid Slug"
	conflictSlugMsg = "Conflicting Slug"
	conflictPathMsg = "Conflicting Path"
	existingPathMsg = "Existing Path"

	unableDeleteCategoryMsg = "this category cannot be deleted because it has used with pages"
)

type pagePathInfo struct {
	ID           uint
	Version      string
	LocaleCode   string
	CategoryPath string
	Slug         string
}

func pageValidator(ctx context.Context, p *Page, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	if p.Slug != "" {
		pagePath := path.Clean(p.Slug)
		if !directoryRe.MatchString(pagePath) {
			err.FieldError("Page.Slug", invalidSlugMsg)
			return
		}
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(p.LocaleCode)
	}

	currentPageCategory, inErr := p.GetCategory(db)
	if inErr != nil {
		panic(err)
	}
	currentPagePublishUrl := p.getPublishUrl(localePath, currentPageCategory.Path)

	var pagePathInfos []pagePathInfo
	if err := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&pagePathInfos).Error; err != nil {
		panic(err)
	}

	for _, info := range pagePathInfos {
		if info.ID == p.ID && info.LocaleCode == p.LocaleCode {
			continue
		}
		var localePath string
		if l10nB != nil {
			localePath = l10nB.GetLocalePath(info.LocaleCode)
		}

		if generatePublishUrl(localePath, info.CategoryPath, info.Slug) == currentPagePublishUrl {
			err.FieldError("Page.Slug", conflictSlugMsg)
			return
		}
	}

	return
}

func categoryValidator(category *Category, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	categoryPath := path.Clean(category.Path)
	if !directoryRe.MatchString(categoryPath) {
		err.FieldError("Category.Category", invalidPathMsg)
		return
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(category.LocaleCode)
	}

	currentCategoryPathPublishUrl := generatePublishUrl(localePath, categoryPath, "")

	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}

	for _, c := range categories {
		if c.ID == category.ID && c.LocaleCode == category.LocaleCode {
			continue
		}
		var localePath string
		if l10nB != nil {
			localePath = l10nB.GetLocalePath(c.LocaleCode)
		}
		if generatePublishUrl(localePath, c.Path, "") == currentCategoryPathPublishUrl {
			err.FieldError("Category.Category", existingPathMsg)
			return
		}
	}

	return
}
