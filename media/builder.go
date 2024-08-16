package media

import (
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

type (
	UserIDFunc func(ctx *web.EventContext) uint
	SearchFunc func(db *gorm.DB, ctx *web.EventContext) *gorm.DB
	SaverFunc  func(db *gorm.DB, obj interface{}, id string, ctx *web.EventContext) error
	Builder    struct {
		db                  *gorm.DB
		permVerifier        *perm.Verifier
		mediaLibraryPerPage int
		currentUserID       UserIDFunc
		searcher            SearchFunc
		saverFunc           SaverFunc
	}
)

func New(db *gorm.DB) *Builder {
	b := &Builder{}
	b.db = db
	b.mediaLibraryPerPage = 39
	b.saverFunc = base.SaveUploadAndCropImage
	return b
}

func (b *Builder) MediaLibraryPerPage(v int) *Builder {
	b.mediaLibraryPerPage = v
	return b
}

func (b *Builder) CurrentUserID(v UserIDFunc) *Builder {
	b.currentUserID = v
	return b
}

func (b *Builder) Searcher(v SearchFunc) *Builder {
	b.searcher = v
	return b
}

func (b *Builder) AutoMigrate() *Builder {
	err := AutoMigrate(b.db)
	if err != nil {
		panic(err)
	}
	return b
}

func (b *Builder) Install(pb *presets.Builder) error {
	configure(pb, b, b.db)
	return nil
}
func (b *Builder) WrapSaverFunc(w func(in SaverFunc) SaverFunc) (r *Builder) {
	b.saverFunc = w(b.saverFunc)
	return b
}
