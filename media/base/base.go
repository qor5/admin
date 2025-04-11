package base

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gosimple/slug"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// CropOption includes crop options
type CropOption struct {
	X, Y, Width, Height int
}

// FileHeader is an interface, for matched values, when call its `Open` method will return `multipart.File`
type FileHeader interface {
	Open() (multipart.File, error)
}

type fileWrapper struct {
	*os.File
}

func (fileWrapper *fileWrapper) Open() (multipart.File, error) {
	return fileWrapper.File, nil
}

// Base defined a base struct for storages
type Base struct {
	FileName    string
	Url         string
	CropOptions map[string]*CropOption `json:"-"`
	Delete      bool                   `json:"-"`
	Crop        bool                   `json:"-"`
	FileHeader  FileHeader             `json:"-"`
	Reader      io.Reader              `json:"-"`
	Options     map[string]string      `json:",omitempty"`
	cropped     bool
	Width       int            `json:",omitempty"`
	Height      int            `json:",omitempty"`
	FileSizes   map[string]int `json:",omitempty"`
}

// Scan scan files, crop options, db values into struct
func (b *Base) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case *os.File:
		b.FileHeader = &fileWrapper{values}
		b.FileName = filepath.Base(values.Name())
	case *multipart.FileHeader:
		b.FileHeader, b.FileName = values, values.Filename
	case []*multipart.FileHeader:
		if len(values) > 0 {
			if file := values[0]; file.Size > 0 {
				b.FileHeader, b.FileName = file, file.Filename
				b.Delete = false
			}
		}
	case []byte:
		if len(values) != 0 {
			if err = json.Unmarshal(values, b); err == nil {
				var options struct {
					Crop   bool
					Delete bool
				}
				if err = json.Unmarshal(values, &options); err == nil {
					if options.Crop {
						b.Crop = true
					}
					if options.Delete {
						b.Delete = true
					}
				}
			}
		}
	case string:
		return b.Scan([]byte(values))
	case []string:
		for _, str := range values {
			if err := b.Scan(str); err != nil {
				return err
			}
		}
	case *MemoryFile:
		b.FileHeader = values
		b.FileName = filepath.Base(values.name)
		return
	default:
		err = errors.New("unsupported driver -> Scan pair for MediaLibrary")
	}

	// If image is deleted, then clean up all values, for serialized fields
	if b.Delete {
		b.Url = ""
		b.FileName = ""
		b.CropOptions = nil
	}
	return
}

// Value return struct's Value
func (b Base) Value() (driver.Value, error) {
	if b.Delete {
		return nil, nil
	}

	results, err := json.Marshal(b)
	return string(results), err
}

func (b *Base) Ext() string {
	return strings.ToLower(path.Ext(b.Url))
}

// URL return file's url with given style
func (b *Base) URL(styles ...string) string {
	if b.Url != "" && len(styles) > 0 {
		ext := path.Ext(b.Url)
		return fmt.Sprintf("%v.%v%v", strings.TrimSuffix(b.Url, ext), styles[0], ext)
	}
	return b.Url
}

func (b *Base) URLNoCached(styles ...string) string {
	i := b.URL(styles...)
	if i != "" {
		return i + "?" + fmt.Sprint(time.Now().Nanosecond())
	}
	return i
}

// String return file's url
func (b *Base) String() string {
	return b.URL()
}

// GetFileName get file's name
func (b *Base) GetFileName() string {
	if b.FileName != "" {
		return b.FileName
	}
	if b.Url != "" {
		return filepath.Base(b.Url)
	}
	return ""
}

// GetFileHeader get file's header, this value only exists when saving files
func (b *Base) GetFileHeader() FileHeader {
	return b.FileHeader
}

// GetURLTemplate get url template
func (*Base) GetURLTemplate(option *Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "/system/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}
	return
}

var urlReplacer = regexp.MustCompile(`(\s|\+)+`)

func getFuncMap(db *gorm.DB, field *schema.Field, filename string) template.FuncMap {
	hash := func() string { return strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "") }
	shortHash := func() string { return time.Now().Format("20060102150405") }

	return template.FuncMap{
		"class": func() string { return inflection.Plural(strcase.ToSnake(field.Schema.ModelType.Name())) },
		"primary_key": func() string {
			ppf := db.Statement.Schema.PrioritizedPrimaryField
			if ppf != nil {
				return fmt.Sprintf("%v", ppf.ReflectValueOf(db.Statement.Context, db.Statement.ReflectValue))
			}

			return "0"
		},
		"column":     func() string { return strings.ToLower(field.Name) },
		"filename":   func() string { return filename },
		"basename":   func() string { return strings.TrimSuffix(path.Base(filename), path.Ext(filename)) },
		"hash":       hash,
		"short_hash": shortHash,
		"filename_with_hash": func() string {
			return urlReplacer.ReplaceAllString(fmt.Sprintf("%s.%v%v", slug.Make(strings.TrimSuffix(path.Base(filename), path.Ext(filename))), hash(), path.Ext(filename)), "-")
		},
		"filename_with_short_hash": func() string {
			return urlReplacer.ReplaceAllString(fmt.Sprintf("%s.%v%v", slug.Make(strings.TrimSuffix(path.Base(filename), path.Ext(filename))), shortHash(), path.Ext(filename)), "-")
		},
		"extension": func() string { return strings.TrimPrefix(path.Ext(filename), ".") },
	}
}

// GetURL get default URL for a model based on its options
func (b *Base) GetURL(option *Option, db *gorm.DB, field *schema.Field, templater URLTemplater) string {
	if path := templater.GetURLTemplate(option); path != "" {
		tmpl := template.New("").Funcs(getFuncMap(db, field, b.GetFileName()))
		if tmpl, err := tmpl.Parse(path); err == nil {
			result := bytes.NewBufferString("")
			if err := tmpl.Execute(result, db.Statement.Dest); err == nil {
				return result.String()
			}
		}
	}
	return ""
}

// Cropped mark the image to be cropped
func (b *Base) Cropped(values ...bool) (result bool) {
	result = b.cropped
	for _, value := range values {
		b.cropped = value
	}
	return result
}

// NeedCrop return the file needs to be cropped or not
func (b *Base) NeedCrop() bool {
	return b.Crop
}

// GetCropOption get crop options
func (b *Base) GetCropOption(name string) *image.Rectangle {
	if cropOption := b.CropOptions[strings.Split(name, "@")[0]]; cropOption != nil {
		return &image.Rectangle{
			Min: image.Point{X: cropOption.X, Y: cropOption.Y},
			Max: image.Point{X: cropOption.X + cropOption.Width, Y: cropOption.Y + cropOption.Height},
		}
	}
	return nil
}

// GetFileSizes get file sizes
func (b *Base) GetFileSizes() map[string]int {
	if b.FileSizes != nil {
		return b.FileSizes
	}
	return make(map[string]int)
}

// Retrieve retrieve file content with url
func (*Base) Retrieve(url string) (*os.File, error) {
	return nil, errors.New("not implemented")
}

// GetSizes get configured sizes, it will be used to crop images accordingly
func (*Base) GetSizes() map[string]*Size {
	return map[string]*Size{}
}

// IsImage return if it is an image
func (b *Base) IsImage() bool {
	return IsImageFormat(b.URL())
}

func (b *Base) IsVideo() bool {
	return IsVideoFormat(b.URL())
}

func (b *Base) IsSVG() bool {
	return IsSVGFormat(b.URL())
}

type MemoryFile struct {
	reader *bytes.Reader
	name   string
}

func (*MemoryFile) Close() error {
	return nil
}

func (m *MemoryFile) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func (m *MemoryFile) Seek(offset int64, whence int) (int64, error) {
	return m.reader.Seek(offset, whence)
}

func (m *MemoryFile) ReadAt(p []byte, off int64) (int, error) {
	return m.reader.ReadAt(p, off)
}

func (m *MemoryFile) Open() (multipart.File, error) {
	return m, nil
}

func NewMemoryFile(filename string, data []byte) *MemoryFile {
	return &MemoryFile{
		name:   filename,
		reader: bytes.NewReader(data),
	}
}
