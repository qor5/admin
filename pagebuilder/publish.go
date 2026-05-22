package pagebuilder

import (
	"context"
	"path"
	"path/filepath"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/x/v3/oss"
	"gorm.io/gorm"
)

func (p *Page) PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	var (
		err        error
		localePath string
	)
	builder := ctx.Value(utils.GetObjectName(p))

	b, ok := builder.(*ModelBuilder)
	if !ok {
		return
	}
	if b.builder.l10n != nil {
		localePath = l10n.LocalePathFromContext(p, ctx)
	}

	var category Category
	if category, err = p.GetCategory(db); err != nil {
		return
	}
	p.OnlineUrl = p.getPublishUrl(localePath, category.Path)
	return p.OnlineUrl
}

func generatePublishUrl(localePath, categoryPath, slug string) string {
	return path.Join("/", localePath, categoryPath, slug, "/index.html")
}

func (p *Page) getPublishUrl(localePath, categoryPath string) string {
	return generatePublishUrl(localePath, categoryPath, p.Slug)
}

func (p *Page) getAccessUrl(publishUrl string) string {
	return filepath.Dir(publishUrl)
}
