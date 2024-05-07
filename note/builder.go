package note

import (
	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

type AfterCreateFunc func(db *gorm.DB) error

type Builder struct {
	db              *gorm.DB
	models          []*presets.ModelBuilder
	afterCreateFunc AfterCreateFunc
}

func New(db *gorm.DB) *Builder {
	b := &Builder{
		db: db,
	}
	return b
}

func (b *Builder) Models(vs ...*presets.ModelBuilder) (r *Builder) {
	b.models = append(b.models, vs...)
	return b
}

func (b *Builder) AfterCreate(f AfterCreateFunc) (r *Builder) {
	b.afterCreateFunc = f
	return b
}

func (b *Builder) Install(pb *presets.Builder) {
	configure(b, pb)
}
