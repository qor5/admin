package pagebuilder

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path"

	"github.com/qor/oss"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/publish"
	"gorm.io/gorm"
)

func (p *Page) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	var b *Builder
	var ok bool
	if b, ok = ctx.Value("pagebuilder").(*Builder); !ok || b == nil {
		return
	}
	content, err := p.getPublishContent(b, ctx)
	if err != nil {
		return
	}

	var localePath string
	var l10nBuilder *l10n.Builder
	if l10nBuilder, ok = ctx.Value("l10nbuilder").(*l10n.Builder); ok && l10nBuilder != nil {
		if locale, isLocalizable := l10n.IsLocalizableFromCtx(ctx); isLocalizable && l10nON {
			localePath = l10nBuilder.GetLocalePath(locale)
		}
	}

	objs = append(objs, &publish.PublishAction{
		Url:      p.getPublishUrl(localePath),
		Content:  content,
		IsDelete: false,
	})
	p.SetOnlineUrl(p.getPublishUrl(localePath))

	var liveRecord Page
	{
		lrdb := db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline)
		if l10nON {
			lrdb = lrdb.Where("locale_code = ?", p.LocaleCode)
		}
		lrdb.First(&liveRecord)
	}
	if liveRecord.ID == 0 {
		return
	}

	if liveRecord.GetOnlineUrl() != p.GetOnlineUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      liveRecord.getPublishUrl(localePath),
			IsDelete: true,
		})
	}

	return
}
func (p *Page) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.GetOnlineUrl(),
		IsDelete: true,
	})
	return
}

func (p Page) getPublishUrl(localePath string) string {
	return path.Join(localePath, p.Slug, "/index.html")
}

func (p Page) getPublishContent(b *Builder, ctx context.Context) (r string, err error) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%d&version=%s&locale=%s", p.ID, p.GetVersion(), p.GetLocale()), nil)
	b.preview.ServeHTTP(w, req)

	r = w.Body.String()
	return
}
