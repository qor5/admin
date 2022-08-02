package pagebuilder

import (
	"path"
	"regexp"

	"github.com/goplaid/web"
	"gorm.io/gorm"
)

var (
	pathRe             = regexp.MustCompile(`^/[0-9a-zA-Z\/]*$`)
	slugWithCategoryRe = regexp.MustCompile(`^/[0-9a-zA-Z]*$`)
)

const queryPathWithSlugSQL = `
SELECT pages.id, categories.path || pages.slug AS path_with_slug
FROM page_builder_pages pages
         LEFT JOIN page_builder_categories categories ON category_id = categories.id
WHERE pages.deleted_at IS NULL
`

func pageValidator(p *Page, db *gorm.DB) (err web.ValidationErrors) {
	if p.CategoryID == 0 && p.Slug == "" {
		err.FieldError("Page.Category", "Category or Slug is required")
		err.FieldError("Page.Slug", "Category or Slug is required")
		return
	}

	if p.CategoryID == 0 {
		s := path.Clean(p.Slug)
		if s != "" && !pathRe.MatchString(s) {
			err.FieldError("Page.Slug", "Invalid Slug format")
			return
		}
	} else {
		if p.Slug != "" {
			s := path.Clean(p.Slug)
			if !slugWithCategoryRe.MatchString(s) {
				err.FieldError("Page.Slug", "Invalid Slug format")
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
		if e.ID == p.CategoryID {
			c = *e
			break
		}
	}

	urlPath := c.Path + p.Slug

	type result struct {
		ID           uint
		PathWithSlug string
	}

	var results []result
	if err := db.Raw(queryPathWithSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		if r.PathWithSlug == urlPath {
			err.FieldError("Page.Slug", "Conflicting Slug")
			return
		}
	}

	for _, e := range categories {
		if e.Path == urlPath {
			err.FieldError("Page.Slug", "Conflicting Slug")
			return
		}
	}

	return
}

func categoryValidator(category *Category, db *gorm.DB) (err web.ValidationErrors) {
	p := path.Clean(category.Path)
	if p != "" && !pathRe.MatchString(p) {
		err.FieldError("Category.Category", "Invalid Path format")
		return
	}

	// Verify category does not conflict the category with slug.
	type result struct {
		ID           uint
		PathWithSlug string
	}

	var results []result
	if err := db.Raw(queryPathWithSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		if r.PathWithSlug == p {
			err.FieldError("Category.Category", "Conflicting Path")
			return
		}
	}

	// Verify category not duplicate.
	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}

	for _, c := range categories {
		if c.Path == p {
			err.FieldError("Category.Category", "Existing Path")
			return
		}
	}

	return
}
