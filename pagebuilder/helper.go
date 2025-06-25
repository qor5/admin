package pagebuilder

import (
	"path"
	"regexp"

	"github.com/qor5/x/v3/i18n"

	"github.com/qor5/web/v3"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
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
)

type pagePathInfo struct {
	ID           uint
	Version      string
	LocaleCode   string
	CategoryPath string
	Slug         string
}

func pageValidator(ctx *web.EventContext, p *Page, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	if p.Title == "" {
		err.FieldError("Title", msgr.InvalidTitleMsg)
		return
	}

	if p.Slug != "" {
		pagePath := path.Clean(p.Slug)
		if !directoryRe.MatchString(pagePath) {
			err.FieldError("Slug", msgr.InvalidSlugMsg)
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
	if dbErr := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&pagePathInfos).Error; dbErr != nil {
		panic(dbErr)
	}

	for _, info := range pagePathInfos {
		if info.ID == p.ID && info.LocaleCode == p.LocaleCode {
			continue
		}
		var innerLocalePath string
		if l10nB != nil {
			innerLocalePath = l10nB.GetLocalePath(info.LocaleCode)
		}

		if generatePublishUrl(innerLocalePath, info.CategoryPath, info.Slug) == currentPagePublishUrl {
			err.FieldError("Slug", msgr.ConflictSlugMsg)
			return
		}
	}

	return
}

func categoryValidator(ctx *web.EventContext, category *Category, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	// Validate category name
	if category.Name == "" {
		err.FieldError("Name", msgr.InvalidNameMsg)
	}

	// Validate category path format
	categoryPath := path.Clean(category.Path)
	if !directoryRe.MatchString(categoryPath) {
		err.FieldError("Path", msgr.InvalidPathMsg)
	}

	// Get locale path for the category
	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(category.LocaleCode)
	}

	// Generate the publish URL for this category
	currentCategoryPathPublishUrl := generatePublishUrl(localePath, categoryPath, "")

	// Check for category path conflicts
	var categories []*Category
	if dbErr := db.Model(&Category{}).Find(&categories).Error; dbErr != nil {
		return
	}

	// Check path uniqueness against other categories
	for _, c := range categories {
		if c.ID == category.ID && c.LocaleCode == category.LocaleCode {
			continue
		}

		var innerLocalePath string
		if l10nB != nil {
			innerLocalePath = l10nB.GetLocalePath(c.LocaleCode)
		}

		if generatePublishUrl(innerLocalePath, c.Path, "") == currentCategoryPathPublishUrl {
			err.FieldError("Path", msgr.ExistingPathMsg)
			return
		}
	}

	// If this is an existing category being updated, check for page URL conflicts
	if category.ID != 0 {
		var originalCategory Category
		if dbErr := db.Where("id = ? AND locale_code = ?", category.ID, category.LocaleCode).First(&originalCategory).Error; dbErr != nil {
			return
		} else if originalCategory.Path == categoryPath {
			// If path hasn't changed, no need to check for conflicts
			return
		}

		// Find all pages associated with this category
		var relatedPages []*Page
		if dbErr := db.Model(&Page{}).Where("category_id = ? AND locale_code = ? ",
			category.ID, category.LocaleCode).Find(&relatedPages).Error; dbErr != nil {
			return
		}

		// If no pages are associated, no conflicts can occur
		if len(relatedPages) == 0 {
			return
		}

		// Get all pages for conflict checking
		var pagePathInfos []pagePathInfo
		if dbErr := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&pagePathInfos).Error; dbErr != nil {
			return
		}

		// Check each related page for potential URL conflicts after category update
		for _, page := range relatedPages {
			// Calculate what the new publish URL would be after category update
			pagePublishUrl := generatePublishUrl(localePath, categoryPath, page.Slug)

			// Check against all other pages
			for _, info := range pagePathInfos {
				// Skip self-comparison
				if info.ID == page.ID && info.LocaleCode == page.LocaleCode {
					continue
				}

				var otherLocalePath string
				if l10nB != nil {
					otherLocalePath = l10nB.GetLocalePath(info.LocaleCode)
				}

				// Check if the new URL would conflict with an existing page
				otherPageUrl := generatePublishUrl(otherLocalePath, info.CategoryPath, info.Slug)
				if otherPageUrl == pagePublishUrl {
					err.FieldError("Path", msgr.WouldCausePageConflictMsg)
					return
				}
			}
		}
	}

	return
}
