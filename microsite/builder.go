package microsite

import (
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/oss/filesystem"
	"gorm.io/gorm"
)

var (
	MaximumNumberOfFilesUploadedAtTheSameTime = 10
	putSemaphore                              = make(chan struct{}, MaximumNumberOfFilesUploadedAtTheSameTime)
)

var (
	MaximumNumberOfFilesCopiedAtTheSameTime = 10
	copySemaphore                           = make(chan struct{}, MaximumNumberOfFilesCopiedAtTheSameTime)
)

type Builder struct {
	db                       *gorm.DB
	packageAndPreviewPrepath string
	publisher                *publish.Builder
	storage                  oss.StorageInterface
}

func New(db *gorm.DB) *Builder {
	b := &Builder{}
	b.db = db
	b.packageAndPreviewPrepath = "microsite"
	b.storage = filesystem.New("public/microsite")
	return b
}

func (b *Builder) PackageAndPreviewPrepath(v string) *Builder {
	b.packageAndPreviewPrepath = v
	return b
}

func (b *Builder) Publisher(v *publish.Builder) (r *Builder) {
	b.publisher = v
	return b
}

func (b *Builder) Storage(v oss.StorageInterface) (r *Builder) {
	b.storage = v
	return b
}
