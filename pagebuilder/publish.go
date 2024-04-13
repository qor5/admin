package pagebuilder

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path"
	"path/filepath"

	"github.com/qor/oss"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

func (p *Page) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	var b *Builder
	var ok bool
	if b, ok = ctx.Value(publish.PublishContextKeyPageBuilder).(*Builder); !ok || b == nil {
		return
	}
	content, err := p.getPublishContent(b, ctx)
	if err != nil {
		return
	}
	var localePath string
	if l10nBuilder, ok := ctx.Value(publish.PublishContextKeyL10nBuilder).(*l10n.Builder); ok && l10nBuilder != nil && l10nON {
		if eventCtx, ok := ctx.Value(publish.PublishContextKeyEventContext).(*web.EventContext); ok && eventCtx != nil {
			if locale, ok := l10n.IsLocalizableFromCtx(eventCtx.R.Context()); ok {
				localePath = l10nBuilder.GetLocalePath(locale)
			}
		}
		if localeCode, err := reflectutils.Get(p, "LocaleCode"); err == nil {
			localePath = l10nBuilder.GetLocalePath(localeCode.(string))
		}
	}

	var category Category
	category, err = p.GetCategory(db)
	if err != nil {
		return
	}
	objs = append(objs, &publish.PublishAction{
		Url:      p.getPublishUrl(localePath, category.Path),
		Content:  content,
		IsDelete: false,
	})
	p.SetOnlineUrl(p.getPublishUrl(localePath, category.Path))

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
			Url:      liveRecord.GetOnlineUrl(),
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

func generatePublishUrl(localePath, categoryPath, slug string) string {
	return path.Join("/", localePath, categoryPath, slug, "/index.html")
}

func (p *Page) getPublishUrl(localePath, categoryPath string) string {
	return generatePublishUrl(localePath, categoryPath, p.Slug)
}

func (p *Page) getAccessUrl(publishUrl string) string {
	return filepath.Dir(publishUrl)
}

func (p *Page) getPublishContent(b *Builder, ctx context.Context) (r string, err error) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%d&version=%s&locale=%s", p.ID, p.GetVersion(), p.GetLocale()), nil)
	b.preview.ServeHTTP(w, req)

	r = w.Body.String()
	return
}
