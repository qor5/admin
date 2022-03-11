package pagebuilder

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path"

	"github.com/qor/oss"
	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

func (p *Page) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	var b *Builder
	var ok bool
	if b, ok = ctx.Value("pagebuilder").(*Builder); !ok || b == nil {
		return
	}
	content, err := p.getPublishContent(b, ctx)
	if err != nil {
		return
	}
	objs = append(objs, &publish.PublishAction{
		Url:      p.getPublishUrl(),
		Content:  content,
		IsDelete: false,
	})
	p.SetOnlineUrl(p.getPublishUrl())

	var liveRecord Page
	db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline).First(&liveRecord)
	if liveRecord.ID == 0 {
		return
	}

	if liveRecord.GetOnlineUrl() != p.GetOnlineUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      liveRecord.getPublishUrl(),
			IsDelete: true,
		})
	}

	return
}
func (p *Page) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.GetOnlineUrl(),
		IsDelete: true,
	})
	return
}

func (p Page) getPublishUrl() string {
	return path.Join(p.Slug, "/index.html")
}

func (p Page) getPublishContent(b *Builder, ctx context.Context) (r string, err error) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%d", p.ID), nil)
	b.preview.ServeHTTP(w, req)

	r = w.Body.String()
	return
}
