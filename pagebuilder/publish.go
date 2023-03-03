package pagebuilder

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path"

	"github.com/qor/oss"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/publish"
	"github.com/qor5/web"
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
	if eventCtx, ok := ctx.Value(publish.PublishContextKeyEventContext).(*web.EventContext); ok && eventCtx != nil {
		if l10nBuilder, ok := ctx.Value(publish.PublishContextKeyL10nBuilder).(*l10n.Builder); ok && l10nBuilder != nil {
			if locale, isLocalizable := l10n.IsLocalizableFromCtx(eventCtx.R.Context()); isLocalizable && l10nON {
				localePath = l10nBuilder.GetLocalePath(locale)
			}
		}
	}

	var categoryPath string
	var category Category
	if err = db.Where("id = ?", p.CategoryID).First(&category).Error; err != nil && err != gorm.ErrRecordNotFound {
		return
	}
	err = nil
	categoryPath = category.Path
	objs = append(objs, &publish.PublishAction{
		Url:      p.getPublishUrl(localePath, categoryPath),
		Content:  content,
		IsDelete: false,
	})
	p.SetOnlineUrl(p.getPublishUrl(localePath, categoryPath))

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
			Url:      liveRecord.getPublishUrl(localePath, categoryPath),
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

func (p *Page) getPublishUrl(localePath, categoryPath string) string {
	return path.Join(localePath, categoryPath, p.Slug, "/index.html")
}

func (p *Page) getAccessUrl(publishUrl string) string {
	dir, _ := path.Split(publishUrl)
	return dir
}

func (p *Page) getPublishContent(b *Builder, ctx context.Context) (r string, err error) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%d&version=%s&locale=%s", p.ID, p.GetVersion(), p.GetLocale()), nil)
	b.preview.ServeHTTP(w, req)

	r = w.Body.String()
	return
}
