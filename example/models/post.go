package models

import (
	"time"

	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/seo"
	"github.com/qor/qor5/slug"
)

type Post struct {
	ID            uint
	Title         string
	TitleWithSlug slug.Slug
	Seo           seo.Setting
	Body          string
	HeroImage     media_library.MediaBox `sql:"type:text;"`
	BodyImage     media_library.MediaBox `sql:"type:text;"`
	UpdatedAt     time.Time
	CreatedAt     time.Time
}

func (post Post) GetSEO() *seo.SEO {
	return SeoCollection.GetSEO("Post")
}
