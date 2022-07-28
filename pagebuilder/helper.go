package pagebuilder

import (
	"path"
	"regexp"

	"github.com/goplaid/web"
)

var (
	pathRe = regexp.MustCompile(`^/[0-9a-zA-Z\/^/s]*$`)
)

func pageValidator(p *Page) (err web.ValidationErrors) {
	if p.CategoryID == 0 {
		err.FieldError("Page.Category", "Category is required")
	}

	return
}

func categoryValidator(c *Category) (err web.ValidationErrors) {
	p := path.Clean(c.Path)
	if p != "" && !pathRe.MatchString(p) {
		err.FieldError("Category.Category", "Invalid Path format")
	}

	return
}
