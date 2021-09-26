package seo

import (
	"github.com/qor/qor5/media"
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
	Varibles   []string
	OpenGraph  *OpenGraphConfig
	Context    func(...interface{}) map[string]string
	collection *Collection
}

// OpenGraphConfig open graph config
type OpenGraphConfig struct {
	Size *media.Size
}

// RegisterGlobalVaribles register global setting struct and will represents as 'Site-wide Settings' part in admin
func (collection *Collection) RegisterGlobalVaribles(s interface{}) {
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
func (collection *Collection) GetSEO(name string) *SEO {
	for _, s := range collection.registeredSEO {
		if s.Name == name {
			return s
		}
	}

	return &SEO{Name: name, collection: collection}
}
