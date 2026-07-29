package media

import (
	"slices"
	"strconv"

	"github.com/qor5/web/v3"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
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
		mb                  *presets.ModelBuilder
		mediaLibraryPerPage int
		currentUserID       UserIDFunc
		searcher            SearchFunc
		saverFunc           SaverFunc
		allowTypes          []string
		fileAccept          string
		ab                  *activity.Builder
	}
)

func New(db *gorm.DB) *Builder {
	b := &Builder{}
	b.db = db
	b.mediaLibraryPerPage = 39
	b.saverFunc = base.SaveUploadAndCropImage
	return b
}

func (b *Builder) GetPresetsModelBuilder() *presets.ModelBuilder {
	return b.mb
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

// scopedDB returns a media_libraries query with the configured searcher
// applied, so by-ID lookups and folder-tree queries see the same subset of rows
// the searcher gives the listing. Without a searcher the query is unscoped,
// preserving the historical behavior — including for a builder that only sets
// CurrentUserID, whose listing-only user_id filter is deliberately not extended
// to these queries.
func (b *Builder) scopedDB(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
	q := db.Model(&media_library.MediaLibrary{})
	if b.searcher != nil {
		q = b.searcher(q, ctx)
	}
	return q
}

// folderIsVisible reports whether folderID may be used as an upload / move
// target for this request: the root folder always may, and any other folder
// only when the searcher lets the request see it. Without a searcher every id
// is accepted, keeping the historical behavior.
func (b *Builder) folderIsVisible(ctx *web.EventContext, folderID uint) bool {
	if b.searcher == nil || folderID == 0 {
		return true
	}
	var count int64
	if err := b.scopedDB(b.db, ctx).Where("id = ? and folder = true", folderID).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// recordIsVisible reports whether the searcher lets this request see the row
// addressed by the primary-key slug id. It backs the presets CRUD event funcs,
// which address rows by id through the generic DataOperator and would otherwise
// bypass the searcher entirely. Without a searcher every id is accepted,
// keeping the historical behavior.
func (b *Builder) recordIsVisible(ctx *web.EventContext, id string) bool {
	if b.searcher == nil || id == "" {
		return true
	}
	recordID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return false
	}
	var count int64
	if err := b.scopedDB(b.db, ctx).Where("id = ?", recordID).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func (b *Builder) Activity(v *activity.Builder) *Builder {
	b.ab = v
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

func (b *Builder) FileAccept(v string) *Builder {
	b.fileAccept = v
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
				selectItem{Text: msgr.Videos, Value: typeVideo})
		case media_library.ALLOW_TYPE_FILE:
			items = append(items,
				selectItem{Text: msgr.Files, Value: typeFile})
		}
	}
	return
}

func (b *Builder) onEdit(ctx *web.EventContext, old, obj media_library.MediaLibrary) {
	if b.ab == nil {
		return
	}
	_, _ = b.ab.OnEdit(ctx.R.Context(), old, obj)
}

func (b *Builder) onCreate(ctx *web.EventContext, obj media_library.MediaLibrary) {
	if b.ab == nil {
		return
	}
	_, _ = b.ab.OnCreate(ctx.R.Context(), obj)
}

func (b *Builder) onDelete(ctx *web.EventContext, objs []media_library.MediaLibrary) {
	if b.ab == nil {
		return
	}
	for _, obj := range objs {
		_, _ = b.ab.OnDelete(ctx.R.Context(), obj)
	}
}
