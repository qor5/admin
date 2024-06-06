package media

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

type Builder struct {
	db                  *gorm.DB
	permVerifier        *perm.Verifier
	mediaLibraryPerPage int
}

func New(db *gorm.DB) *Builder {
	b := &Builder{}
	b.db = db
	b.mediaLibraryPerPage = 39
	return b
}

func (b *Builder) MediaLibraryPerPage(v int) *Builder {
	b.mediaLibraryPerPage = v
	return b
}

func (b *Builder) Install(pb *presets.Builder) error {
	configure(pb, b, b.db)
	return nil
}
