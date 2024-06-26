package media

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

type (
	GetUserIDFunc func(ctx *web.EventContext) uint
	Builder       struct {
		db                  *gorm.DB
		permVerifier        *perm.Verifier
		mediaLibraryPerPage int
		getCurrentUserID    GetUserIDFunc
	}
)

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

func (b *Builder) GetCurrentUserID(v GetUserIDFunc) *Builder {
	b.getCurrentUserID = v
	return b
}

func (b *Builder) Install(pb *presets.Builder) error {
	configure(pb, b, b.db)
	return nil
}
