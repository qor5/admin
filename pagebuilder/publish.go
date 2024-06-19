package pagebuilder

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path"
	"path/filepath"

	"github.com/qor5/admin/v3/utils"

	"github.com/qor5/admin/v3/l10n"

	"github.com/qor/oss"
	"github.com/qor5/admin/v3/publish"
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
	return p.getPublishUrl(localePath, category.Path)
}

func (p *Page) LiveUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	var liveRecord Page

	builder := ctx.Value(utils.GetObjectName(p))
	mb, ok := builder.(ModelBuilder)
	if !ok {
		return
	}
	{
		lrdb := db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline)
		if mb.builder.l10n != nil {
			lrdb = lrdb.Where("locale_code = ?", p.LocaleCode)
		}
		lrdb.First(&liveRecord)
	}
	if liveRecord.ID == 0 {
		return
	}
	return
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

func (p *Page) getPublishContent(b *ModelBuilder, _ context.Context) (r string, err error) {
	w := httptest.NewRecorder()

	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%s", p.PrimarySlug()), nil)
	b.preview.ServeHTTP(w, req)
	r = w.Body.String()
	return
}
