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
	Width  int `json:",omitempty"`
	Height int `json:",omitempty"`
}

// MediaBoxConfig configure MediaBox metas
type MediaBoxConfig struct {
	Sizes     map[string]*base.Size
	Max       uint
	AllowType string
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

func (mediaBox *MediaBox) URL(styles ...string) string {
	if mediaBox.Url != "" && len(styles) > 0 {
		ext := path.Ext(mediaBox.Url)
		return fmt.Sprintf("%v.%v%v", strings.TrimSuffix(mediaBox.Url, ext), styles[0], ext)
	}
	return mediaBox.Url
}
func (mediaBox *MediaBox) URLNoCached(styles ...string) string {
	url := mediaBox.URL(styles...)
	if url != "" {
		url = fmt.Sprintf("%s?%d", url, time.Now().Nanosecond())
	}
	return url
}

func (mediaBox MediaBox) WebpURL(styles ...string) string {
	url := mediaBox.URL(styles...)
	ext := path.Ext(url)
	extArr := strings.Split(ext, "?")
	i := strings.LastIndex(url, ext)
	return url[:i] + strings.Replace(url[i:], extArr[0], ".webp", 1)
}
