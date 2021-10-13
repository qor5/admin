package pagebuilder

import (
	"context"
	"path"

	"github.com/qor/qor5/publish"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func (p *Page) GetPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
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

	if p.GetStatus() == publish.StatusOnline && p.GetOnlineUrl() != p.getPublishUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      p.GetOnlineUrl(),
			IsDelete: true,
		})
	}

	p.SetOnlineUrl(p.getPublishUrl())
	return
}
func (p *Page) GetUnPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
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
	var comps []h.HTMLComponent
	comps, err = b.renderContainers(nil, p.ID, true)
	if err != nil {
		return
	}
	r = h.MustString(h.Components(comps...), ctx)
	return
}
