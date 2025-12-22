package media_library

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/qor5/admin/v3/media/base"
)

const (
	ALLOW_TYPE_FILE  = "file"
	ALLOW_TYPE_IMAGE = "image"
	ALLOW_TYPE_VIDEO = "video"
)

type MediaBox struct {
	ID          json.Number
	Url         string
	VideoLink   string
	FileName    string
	Description string
	FileSizes   map[string]int `json:",omitempty"`
	// for default image
	Width       int                         `json:",omitempty"`
	Height      int                         `json:",omitempty"`
	CropOptions map[string]*base.CropOption `json:",omitempty"`
	Sizes       map[string]*base.Size       `json:",omitempty"`
	CropID      map[string]string           `json:",omitempty"`
}

// MediaBoxConfig configure MediaBox metas
type MediaBoxConfig struct {
	Sizes      map[string]*base.Size
	Max        uint
	AllowType  string
	FileAccept string
	// the background color of MediaBox
	BackgroundColor string
	// disable crop
	DisableCrop bool
	// allow to accept media_box only with URL
	SimpleIMGURL bool
}

func (mediaBox *MediaBox) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		if len(values) > 0 {
			return json.Unmarshal(values, mediaBox)
		}
	case string:
		return mediaBox.Scan([]byte(values))
	}
	return nil
}

func (mediaBox MediaBox) Value() (driver.Value, error) {
	if mediaBox.ID.String() == "0" || mediaBox.ID.String() == "" {
		return nil, nil
	}
	results, err := json.Marshal(mediaBox)
	return string(results), err
}

// IsImage return if it is an image
func (mediaBox *MediaBox) IsImage() bool {
	return base.IsImageFormat(mediaBox.Url)
}

func (mediaBox *MediaBox) IsVideo() bool {
	return base.IsVideoFormat(mediaBox.Url)
}

func (mediaBox *MediaBox) IsSVG() bool {
	return base.IsSVGFormat(mediaBox.Url)
}

func (mediaBox *MediaBox) URL(styles ...string) (s string) {
	var cropID string
	if len(styles) == 0 {
		cropID = mediaBox.CropID[base.DefaultSizeKey]
	} else {
		cropID = mediaBox.CropID[styles[0]]
	}
	ext := path.Ext(mediaBox.Url)

	defer func() {
		if cropID != "" {
			s = fmt.Sprintf("%v_%v%v", s, cropID, ext)
			return
		}
		s = fmt.Sprintf("%v%v", s, ext)
	}()
	if mediaBox.Url != "" && len(styles) > 0 {
		return fmt.Sprintf("%v.%v", strings.TrimSuffix(mediaBox.Url, ext), styles[0])
	}
	return strings.TrimSuffix(mediaBox.Url, ext)
}

func (mediaBox *MediaBox) URLNoCached(styles ...string) string {
	i := mediaBox.URL(styles...)
	if i != "" && !strings.Contains(i, "?") {
		return i + "?" + fmt.Sprint(time.Now().Nanosecond())
	}
	return i
}

func (mediaBox MediaBox) WebpURL(styles ...string) string {
	url := mediaBox.URL(styles...)
	ext := path.Ext(url)
	extArr := strings.Split(ext, "?")
	i := strings.LastIndex(url, ext)
	return url[:i] + strings.Replace(url[i:], extArr[0], ".webp", 1)
}
