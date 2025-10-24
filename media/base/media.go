package base

import (
	"database/sql/driver"
	"image"
	"io"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Media is an interface including methods that needs for a media library storage
type Media interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)

	GetURLTemplate(*Option) string
	GetURL(option *Option, db *gorm.DB, field *schema.Field, templater URLTemplater) string

	GetFileHeader() FileHeader
	GetFileName() string

	GetSizes() map[string]*Size
	NeedCrop() bool
	Cropped(values ...bool) bool
	GetCropOption(name string) *image.Rectangle
	GetFileSizes() map[string]int

	Store(url string, option *Option, reader io.Reader) error
	Retrieve(url string) (FileInterface, error)

	IsImage() bool

	URL(style ...string) string
	Ext() string
	String() string
}

// FileInterface media file interface
type FileInterface interface {
	io.ReadSeeker
	io.Closer
}

// Size is a struct, used for `GetSizes` method, it will return a slice of Size, media library will crop images automatically based on it
type Size struct {
	Width   int
	Height  int
	Padding bool
	// v-col sm
	Sm int
	// v-col col
	Cols int
}

// URLTemplater is a interface to return url template
type URLTemplater interface {
	GetURLTemplate(*Option) string
}

// Option media library option
type Option map[string]string

// get option with name
func (option Option) Get(key string) string {
	return option[strings.ToUpper(key)]
}

// set option
func (option Option) Set(key string, val string) {
	option[strings.ToUpper(key)] = val
}
