package media

import (
	"slices"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
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
		allowTypes          []string
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

func (b *Builder) AllowTypes(v ...string) *Builder {
	b.allowTypes = append(b.allowTypes, v...)
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

func (b *Builder) checkAllowType(v string) bool {
	if len(b.allowTypes) == 0 {
		return true
	}
	return slices.Contains(b.allowTypes, v)
}

func (b *Builder) allowTypeSelectOptions(msgr *Messages) (items []selectItem) {
	items = []selectItem{
		{Text: msgr.All, Value: typeAll},
	}
	allTypes := b.allowTypes
	if len(allTypes) == 0 {
		allTypes = []string{media_library.ALLOW_TYPE_IMAGE, media_library.ALLOW_TYPE_VIDEO, media_library.ALLOW_TYPE_FILE}
	}
	for _, t := range allTypes {
		switch t {
		case media_library.ALLOW_TYPE_IMAGE:
			items = append(items,
				selectItem{Text: msgr.Images, Value: typeImage})
		case media_library.ALLOW_TYPE_VIDEO:
			items = append(items,
				selectItem{Text: msgr.Images, Value: typeVideo})
		case media_library.ALLOW_TYPE_FILE:
			items = append(items,
				selectItem{Text: msgr.Images, Value: typeFile})

		}
	}
	return
}
