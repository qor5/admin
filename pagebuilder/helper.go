package pagebuilder

import (
	"path"
	"regexp"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/web"
	"gorm.io/gorm"
)

var (
	pathRe             = regexp.MustCompile(`^/[0-9a-zA-Z-_().\/]*$`)
	slugWithCategoryRe = regexp.MustCompile(`^/[0-9a-zA-Z-_().]*$`)
)

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
	missingCategoryOrSlugMsg   = "Category or Slug is required"
	invalidSlugMsg             = "The slug must start with a '/' followed by one or more characters"
	invalidSlugWithCategoryMsg = "The slug must start with a '/' followed by one or more characters excluding '/'"
	invalidPathMsg             = "The path must start with a '/' followed by one or more characters"
	conflictSlugMsg            = "Conflicting Slug"
	conflictPathMsg            = "Conflicting Path"
	existingPathMsg            = "Existing Path"

	unableDeleteCategoryMsg = "this category cannot be deleted because it has used with pages"
)

func pageValidator(p *Page, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	if p.CategoryID == 0 && p.Slug == "" {
		err.FieldError("Page.Category", missingCategoryOrSlugMsg)
		err.FieldError("Page.Slug", missingCategoryOrSlugMsg)
		return
	}

	if p.CategoryID == 0 {
		s := path.Clean(p.Slug)
		if s != "" && !pathRe.MatchString(s) {
			err.FieldError("Page.Slug", invalidSlugMsg)
			return
		}
	} else {
		if p.Slug != "" {
			s := path.Clean(p.Slug)
			if !slugWithCategoryRe.MatchString(s) {
				err.FieldError("Page.Slug", invalidSlugWithCategoryMsg)
				return
			}
		}
	}

	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}
	var c Category
	for _, e := range categories {
		if e.ID == p.CategoryID && e.LocaleCode == p.LocaleCode {
			c = *e
			break
		}
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(p.LocaleCode)
	}
	urlPath := p.getPublishUrl(localePath, c.Path)
	type result struct {
		ID           uint
		Version      string
		LocaleCode   string
		CategoryPath string
		Slug         string
	}

	var results []result
	if err := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		if r.ID == p.ID && r.LocaleCode == p.LocaleCode {
			continue
		}
		var localePath string
		if l10nB != nil {
			localePath = l10nB.GetLocalePath(r.LocaleCode)
		}

		if generatePublishUrl(localePath, r.CategoryPath, r.Slug) == urlPath {
			err.FieldError("Page.Slug", conflictSlugMsg)
			return
		}
	}

	if p.Slug != "" && p.Slug != "/" {
		var allLocalePaths []string
		if l10nB != nil {
			allLocalePaths = l10nB.GetAllLocalePaths()
		} else {
			allLocalePaths = []string{""}
		}
		for _, e := range categories {
			for _, localePath := range allLocalePaths {

				if generatePublishUrl(localePath, e.Path, "") == urlPath {
					err.FieldError("Page.Slug", conflictSlugMsg)
					return
				}
			}
		}
	}

	return
}

func categoryValidator(category *Category, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	p := path.Clean(category.Path)
	if p != "" && !pathRe.MatchString(p) {
		err.FieldError("Category.Category", invalidPathMsg)
		return
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(category.LocaleCode)
	}

	var publishUrl = generatePublishUrl(localePath, p, "")
	// Verify category does not conflict the pages' PublishUrl.
	type result struct {
		ID           uint
		Version      string
		LocaleCode   string
		CategoryPath string
		Slug         string
	}

	var results []result
	if err := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		var resultLocalePath string
		if l10nB != nil {
			resultLocalePath = l10nB.GetLocalePath(r.LocaleCode)
		}
		if generatePublishUrl(resultLocalePath, r.CategoryPath, r.Slug) == publishUrl {
			err.FieldError("Category.Category", conflictPathMsg)
			return
		}
	}

	// Verify category not duplicate.
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

		if generatePublishUrl(localePath, c.Path, "") == publishUrl {
			err.FieldError("Category.Category", existingPathMsg)
			return
		}
	}

	return
}
