package seo

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/qor/qor5/media/media_library"
	h "github.com/theplant/htmlgo"
)

// QorSEOSettingInterface support customize Seo model
type QorSEOSettingInterface interface {
	GetName() string
	SetName(string)
	GetSEOSetting() Setting
	SetSEOSetting(Setting)
	GetGlobalSetting() map[string]string
	SetGlobalSetting(map[string]string)
	GetSEOType() string
	SetSEOType(string)
	GetIsGlobalSEO() bool
	SetIsGlobalSEO(bool)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	SetCollection(*Collection)
	GetOpenGraphTitle() string
	GetOpenGraphDescription() string
	GetOpenGraphURL() string
	GetOpenGraphType() string
	GetOpenGraphImageURL() string
	GetOpenGraphImageFromMediaLibrary() media_library.MediaBox
	GetOpenGraphMetadata() []OpenGraphMetadata
}

// QorSEOSetting default seo model
type QorSEOSetting struct {
	Name        string `gorm:"primary_key"`
	Setting     Setting
	IsGlobalSEO bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	collection *Collection
}

// Setting defined meta's attributes
type Setting struct {
	Title                          string `gorm:"size:4294967295"`
	Description                    string
	Keywords                       string
	Type                           string
	OpenGraphTitle                 string
	OpenGraphDescription           string
	OpenGraphURL                   string
	OpenGraphType                  string
	OpenGraphImageURL              string
	OpenGraphImageFromMediaLibrary media_library.MediaBox
	OpenGraphMetadata              []OpenGraphMetadata
	EnabledCustomize               bool
	GlobalSetting                  map[string]string
}

// OpenGraphMetadata open graph meta data
type OpenGraphMetadata struct {
	Property string
	Content  string
}

func (s *QorSEOSetting) SetSEOSetting(setting Setting) {
	s.Setting = setting
}

// GetSEOSetting get seo setting
func (s QorSEOSetting) GetSEOSetting() Setting {
	return s.Setting
}

// GetName get QorSeoSetting's name
func (s QorSEOSetting) GetName() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSEOSetting) SetName(name string) {
	s.Name = name
}

// GetSEOType get QorSeoSetting's type
func (s QorSEOSetting) GetSEOType() string {
	return s.Setting.Type
}

// SetSEOType set QorSeoSetting's type
func (s *QorSEOSetting) SetSEOType(t string) {
	s.Setting.Type = t
}

// GetIsGlobalSEO get QorSEOSetting's isGlobal
func (s QorSEOSetting) GetIsGlobalSEO() bool {
	return s.IsGlobalSEO
}

// SetIsGlobalSEO set QorSeoSetting's isGlobal
func (s *QorSEOSetting) SetIsGlobalSEO(isGlobal bool) {
	s.IsGlobalSEO = isGlobal
}

// GetGlobalSetting get QorSeoSetting's globalSetting
func (s QorSEOSetting) GetGlobalSetting() map[string]string {
	return s.Setting.GlobalSetting
}

// SetGlobalSetting set QorSeoSetting's globalSetting
func (s *QorSEOSetting) SetGlobalSetting(globalSetting map[string]string) {
	s.Setting.GlobalSetting = globalSetting
}

func (s QorSEOSetting) GetOpenGraphTitle() string {
	return s.Setting.OpenGraphTitle
}
func (s QorSEOSetting) GetOpenGraphDescription() string {
	return s.Setting.OpenGraphDescription
}
func (s QorSEOSetting) GetOpenGraphURL() string {
	return s.Setting.OpenGraphURL
}
func (s QorSEOSetting) GetOpenGraphType() string {
	return s.Setting.OpenGraphType
}
func (s QorSEOSetting) GetOpenGraphImageURL() string {
	return s.Setting.OpenGraphImageURL
}
func (s QorSEOSetting) GetOpenGraphImageFromMediaLibrary() media_library.MediaBox {
	return s.Setting.OpenGraphImageFromMediaLibrary
}
func (s QorSEOSetting) GetOpenGraphMetadata() []OpenGraphMetadata {
	return s.Setting.OpenGraphMetadata
}

// GetTitle get Setting's title
func (s QorSEOSetting) GetTitle() string {
	return s.Setting.Title
}

// GetDescription get Setting's description
func (s QorSEOSetting) GetDescription() string {
	return s.Setting.Description
}

// GetKeywords get Setting's keywords
func (s QorSEOSetting) GetKeywords() string {
	return s.Setting.Keywords
}

// SetCollection set Setting's collection
func (s *QorSEOSetting) SetCollection(collection *Collection) {
	s.collection = collection
}

// GetSEO get Setting's SEO configure
// func (s QorSEOSetting) GetSEO() *SEO {
// 	return s.collection.GetSEO(s.Name)
// }

// Scan scan value from database into struct
func (setting *Setting) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		json.Unmarshal(bytes, setting)
	} else if str, ok := value.(string); ok {
		json.Unmarshal([]byte(str), setting)
	} else if strs, ok := value.([]string); ok {
		for _, str := range strs {
			json.Unmarshal([]byte(str), setting)
		}
	}
	return nil
}

// Value get value from struct, and save into database
func (setting Setting) Value() (driver.Value, error) {
	result, err := json.Marshal(setting)
	return string(result), err
}

func (setting Setting) Component(req *http.Request) h.HTMLComponent {
	toAbsoluteURL := func(str string) string {
		if u, err := url.Parse(str); err == nil {
			if u.IsAbs() {
				return str
			}

			if u.Host == "" && req != nil {
				u.Host = req.Host
			}

			if u.Scheme == "" {
				if req != nil && req.URL.Scheme != "" {
					u.Scheme = req.URL.Scheme
				} else {
					u.Scheme = "http"
				}
			}
			return u.String()
		}
		return str
	}

	openGraphData := map[string]string{}

	if setting.OpenGraphURL != "" {
		openGraphData["og:url"] = toAbsoluteURL(setting.OpenGraphURL)
	}

	if setting.OpenGraphType != "" {
		openGraphData["og:type"] = setting.OpenGraphType
	}

	if setting.OpenGraphImageFromMediaLibrary.FileName != "" {
		openGraphData["og:image"] = toAbsoluteURL(setting.OpenGraphImageFromMediaLibrary.URL())
	} else {
		openGraphData["og:image"] = toAbsoluteURL(setting.OpenGraphImageURL)
	}

	for _, metavalue := range setting.OpenGraphMetadata {
		openGraphData[metavalue.Property] = metavalue.Content
	}

	if _, ok := openGraphData["og:title"]; !ok {
		openGraphData["og:title"] = setting.Title
	}

	if _, ok := openGraphData["og:description"]; !ok {
		openGraphData["og:description"] = setting.Description
	}

	var openGraphDataComponents h.HTMLComponents
	for key, value := range openGraphData {
		openGraphDataComponents = append(openGraphDataComponents, h.Meta().Attr("property", key).Attr("name", key).Attr("content", value))
	}

	return h.HTMLComponents{
		h.Title(setting.Title),
		h.Meta().Attr("name", "description").Attr("content", setting.Description),
		h.Meta().Attr("name", "keywords").Attr("content", setting.Keywords),
		openGraphDataComponents,
	}
}
