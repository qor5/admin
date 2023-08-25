package seo

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/media/media_library"
	h "github.com/theplant/htmlgo"
)

// QorSEOSettingInterface support customize Seo model
// @snippet_begin(QorSEOSettingInterface)
type QorSEOSettingInterface interface {
	GetName() string
	SetName(string)
	GetSEOSetting() Setting
	SetSEOSetting(Setting)
	GetVariables() Variables
	SetVariables(Variables)
	GetLocale() string
	SetLocale(string)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	GetOpenGraphURL() string
	GetOpenGraphType() string
	GetOpenGraphImageURL() string
	GetOpenGraphImageFromMediaLibrary() media_library.MediaBox
	GetOpenGraphMetadata() []OpenGraphMetadata
}

// @snippet_end

// QorSEOSetting default seo model
type QorSEOSetting struct {
	Name      string `gorm:"primary_key"`
	Setting   Setting
	Variables Variables `sql:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
	l10n.Locale
}

// Setting defined meta's attributes
type Setting struct {
	Title                          string `gorm:"size:4294967295"`
	Description                    string
	Keywords                       string
	OpenGraphURL                   string
	OpenGraphType                  string
	OpenGraphImageURL              string
	OpenGraphImageFromMediaLibrary media_library.MediaBox
	OpenGraphMetadata              []OpenGraphMetadata
	EnabledCustomize               bool
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

// SetVariablesSetting set variables setting
func (s *QorSEOSetting) SetVariables(setting Variables) {
	s.Variables = setting
}

// GetVariablesSetting get variables setting
func (s QorSEOSetting) GetVariables() Variables {
	return s.Variables
}

// GetName get QorSeoSetting's name
func (s QorSEOSetting) GetLocale() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSEOSetting) SetLocale(locale string) {
	s.LocaleCode = locale
}

// GetName get QorSeoSetting's name
func (s QorSEOSetting) GetName() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSEOSetting) SetName(name string) {
	s.Name = name
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

func (setting Setting) IsEmpty() bool {
	return setting.Title == "" && setting.Description == "" && setting.Keywords == "" && setting.OpenGraphURL == "" && setting.OpenGraphType == "" && setting.OpenGraphImageURL == "" && setting.OpenGraphImageFromMediaLibrary.Url == "" && len(setting.OpenGraphMetadata) == 0
}

type Variables map[string]string

// Scan scan value from database into struct
func (setting *Variables) Scan(value interface{}) error {
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
func (setting Variables) Value() (driver.Value, error) {
	result, err := json.Marshal(setting)
	return string(result), err
}

func (setting Setting) HTMLComponent(tags map[string]string) h.HTMLComponent {
	openGraphData := map[string]string{
		"og:title":       setting.Title,
		"og:description": setting.Description,
		"og:url":         setting.OpenGraphURL,
		"og:type":        setting.OpenGraphType,
		"og:image":       setting.OpenGraphImageURL,
	}

	for _, metavalue := range setting.OpenGraphMetadata {
		openGraphData[metavalue.Property] = metavalue.Content
	}

	for _, key := range []string{"og:url", "og:type", "og:image", "og:title", "og:description"} {
		if v := openGraphData[key]; v == "" {
			if v, ok := tags[key]; ok {
				openGraphData[key] = v
			}
		}
	}

	if openGraphData["og:type"] == "" {
		openGraphData["og:type"] = "website"
	}

	for key, value := range tags {
		if _, ok := openGraphData[key]; !ok {
			openGraphData[key] = value
		}
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
