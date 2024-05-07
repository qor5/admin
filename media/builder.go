package media

import (
	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

type Builder struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Builder {
	b := &Builder{db}
	return b
}

func (b *Builder) Install(pb *presets.Builder) {
	configure(pb, b.db)
}
