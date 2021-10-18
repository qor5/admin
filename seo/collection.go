package seo

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor/qor5/media"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// New initialize a SeoCollection instance
func New(name string) *Collection {
	return &Collection{Name: name}
}

// Collection will hold registered seo configures and global setting definition and other configures
type Collection struct {
	Name          string
	registeredSEO []*SEO
	settingModel  interface{}
	globalSetting interface{}
}

// SEO represents a seo object for a page
type SEO struct {
	Name       string
	Model      interface{}
	Variables  []string
	OpenGraph  *OpenGraphConfig
	Context    func(...interface{}) map[string]string
	collection *Collection
}

// OpenGraphConfig open graph config
type OpenGraphConfig struct {
	Size *media.Size
}

// RegisterGlobalVariables register global setting struct and will represents as 'Site-wide Settings' part in admin
func (collection *Collection) RegisterGlobalVariables(s interface{}) {
	collection.globalSetting = s
}

// RegisterSettingModel register setting model struct
func (collection *Collection) RegisterSettingModel(s interface{}) {
	collection.settingModel = s
}

// RegisterSEO register a seo
func (collection *Collection) RegisterSEO(seo *SEO) {
	seo.collection = collection
	if seo.OpenGraph == nil {
		seo.OpenGraph = &OpenGraphConfig{}
	}
	collection.registeredSEO = append(collection.registeredSEO, seo)
}

// GetSEO get a Seo by name
func (collection *Collection) GetSEOByName(name string) *SEO {
	for _, s := range collection.registeredSEO {
		if s.Name == name {
			return s
		}
	}

	return &SEO{Name: name, collection: collection}
}

func (collection *Collection) GetSEOByModel(model interface{}) *SEO {
	for _, s := range collection.registeredSEO {
		if reflect.TypeOf(model) == reflect.TypeOf(s.Model) {
			return s
		}
	}

	return &SEO{Name: "", collection: collection}
}

func (collection Collection) Render(db *gorm.DB, req *http.Request, name string, objects ...interface{}) h.HTMLComponent {
	seoSetting := collection.GetSEOSetting(db, name, objects...)

	return seoSetting.Component(req)
}

func (collection Collection) GetSEOSetting(db *gorm.DB, name string, objects ...interface{}) Setting {
	var (
		seoSetting Setting
		seo        = collection.GetSEOByName(name)
	)

	// If passed objects has customzied SEO Setting field
	for _, obj := range objects {
		if value := reflect.Indirect(reflect.ValueOf(obj)); value.IsValid() && value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
					seoSetting = value.Field(i).Interface().(Setting)
					break
				}
			}
		}
	}

	if !seoSetting.EnabledCustomize {
		globalSeoSetting := reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface().(QorSEOSettingInterface)
		if db.Where("name = ?", name).First(globalSeoSetting); globalSeoSetting.GetName() != "" {
			seoSetting = globalSeoSetting.GetSEOSetting()
		}
	}

	siteWideSetting := reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface()
	db.Where("is_global_seo = ? AND name = ?", true, collection.Name).First(siteWideSetting)
	tagValues := siteWideSetting.(QorSEOSettingInterface).GetGlobalSetting()

	if tagValues == nil {
		tagValues = map[string]string{}
	}

	if seo.Context != nil {
		for key, value := range seo.Context(objects...) {
			tagValues[key] = value
		}
	}

	return replaceTags(seoSetting, seo.Variables, tagValues)
}

func replaceTags(seoSetting Setting, validTags []string, values map[string]string) Setting {
	replace := func(str string) string {
		re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
		matches := re.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			str = strings.Replace(str, match[0], values[match[1]], 1)
		}
		return str
	}

	seoSetting.Title = replace(seoSetting.Title)
	seoSetting.Description = replace(seoSetting.Description)
	seoSetting.Keywords = replace(seoSetting.Keywords)
	seoSetting.Type = replace(seoSetting.Type)
	seoSetting.OpenGraphURL = replace(seoSetting.OpenGraphURL)
	seoSetting.OpenGraphImageURL = replace(seoSetting.OpenGraphImageURL)
	seoSetting.OpenGraphType = replace(seoSetting.OpenGraphType)
	for idx, metadata := range seoSetting.OpenGraphMetadata {
		seoSetting.OpenGraphMetadata[idx] = OpenGraphMetadata{
			Property: replace(metadata.Property),
			Content:  replace(metadata.Content),
		}
	}
	return seoSetting
}
