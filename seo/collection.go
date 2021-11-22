package seo

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	GlobalSEO    = "Global SEO"
	GlobalDB     *gorm.DB
	DBContextKey contextKey = "DB"
)

type (
	contextKey            string
	contextVariablesFunc  func(interface{}, *Setting, *http.Request) string
	globalSettingVaribles struct {
		SiteName string
	}
)

// Create a SeoCollection instance
func NewCollection() *Collection {
	collection := &Collection{
		settingModel: &QorSEOSetting{},
		dbContextKey: DBContextKey,
		globalName:   GlobalSEO,
	}

	collection.RegisterSEO(GlobalSEO).RegisterVariblesSetting(&globalSettingVaribles{}).
		RegisterContextVariables(
			"og:url", func(_ interface{}, _ *Setting, req *http.Request) string {
				return req.URL.String()
			},
		)

	return collection
}

// Collection will hold registered seo configures and global setting definition and other configures
type Collection struct {
	registeredSEO []*SEO
	globalName    string      //default is GlobalSEO
	dbContextKey  interface{} // context key to get db from context
	settingModel  interface{} // db model
}

// SEO represents a seo object for a page
type SEO struct {
	name             string
	model            interface{}
	contextVariables map[string]contextVariablesFunc // fetch context variables from request
	settingVariables interface{}                     // fetch setting variables from db
}

// RegisterModel register a model to seo
func (seo *SEO) SetModel(model interface{}) *SEO {
	seo.model = model
	return seo
}

// SetName set seo name
func (seo *SEO) SetName(name string) *SEO {
	seo.name = name
	return seo
}

// RegisterContextVariables register context variables
func (seo *SEO) RegisterContextVariables(key string, f contextVariablesFunc) *SEO {
	if seo.contextVariables == nil {
		seo.contextVariables = map[string]contextVariablesFunc{}
	}
	seo.contextVariables[key] = f
	return seo
}

// RegisterSettingModel register a setting
func (seo *SEO) RegisterVariblesSetting(setting interface{}) *SEO {
	seo.settingVariables = setting
	return seo
}

func (collection *Collection) SetGlobalName(name string) *Collection {
	collection.globalName = name
	if globalSeo := collection.GetSEOByName(GlobalSEO); globalSeo != nil {
		globalSeo.SetName(name)
	}
	return collection
}

func (collection *Collection) NewSettingModelInstance() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface()
}

func (collection *Collection) NewSettingModelSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type())
	slice := reflect.New(sliceType)
	slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))
	return slice.Interface()
}

// RegisterVariblesSetting register variables setting
func (collection *Collection) RegisterSettingModel(s interface{}) *Collection {
	collection.settingModel = s
	return collection
}

// RegisterDBContextKey register a key to get db from context
func (collection *Collection) RegisterDBContextKey(key interface{}) *Collection {
	collection.dbContextKey = key
	return collection
}

// RegisterSEO register a seo
func (collection *Collection) RegisterSEOByNames(names ...string) *Collection {
	for index := range names {
		collection.registeredSEO = append(collection.registeredSEO, &SEO{name: names[index]})
	}
	return collection
}

// RegisterSEO register a seo
func (collection *Collection) RegisterSEO(obj interface{}) (seo *SEO) {
	if name, ok := obj.(string); ok {
		seo = &SEO{name: name}
	} else {
		name := reflect.Indirect(reflect.ValueOf(obj)).Type().Name()
		seo = &SEO{name: name, model: obj}
	}

	collection.registeredSEO = append(collection.registeredSEO, seo)
	return
}

// RegisterSEO remove a seo
func (collection *Collection) RemoveSEO(obj interface{}) *Collection {
	var name string
	if n, ok := obj.(string); ok {
		name = n
	} else {
		name = reflect.Indirect(reflect.ValueOf(obj)).Type().Name()
	}

	for index, s := range collection.registeredSEO {
		if s.name == name {
			collection.registeredSEO = append(collection.registeredSEO[:index], collection.registeredSEO[index+1:]...)
			break
		}
	}

	return collection
}

// GetSEO get a Seo
func (collection *Collection) GetSEO(obj interface{}) *SEO {
	if name, ok := obj.(string); ok {
		return collection.GetSEOByName(name)
	} else {
		return collection.GetSEOByModel(obj)
	}
}

// GetSEO get a Seo by name
func (collection *Collection) GetSEOByName(name string) *SEO {
	for _, s := range collection.registeredSEO {
		if s.name == name {
			return s
		}
	}

	return nil
}

// GetSEOByModel get a seo by model
func (collection *Collection) GetSEOByModel(model interface{}) *SEO {
	for _, s := range collection.registeredSEO {
		if reflect.TypeOf(model) == reflect.TypeOf(s.model) {
			return s
		}
	}

	return nil
}

func (collection Collection) RenderGlobal(req *http.Request) h.HTMLComponent {
	return collection.Render(collection.globalName, req)
}

// Render render seo tags
func (collection Collection) Render(obj interface{}, req *http.Request) h.HTMLComponent {
	var (
		db               = collection.getDBFromContext(req.Context())
		sortedSEOs       []*SEO
		sortedSeoNames   []string
		sortedDBSettings []QorSEOSettingInterface
		sortedSettings   []Setting
		setting          Setting
	)

	// sort all SEOs
	globalSeo := collection.GetSEO(collection.globalName)
	if globalSeo != nil {
		return nil
	}
	sortedSeoNames = append(sortedSeoNames, globalSeo.name)
	sortedSEOs = append(sortedSEOs, globalSeo)

	if name, ok := obj.(string); !ok || name != collection.globalName {
		if seo := collection.GetSEO(obj); seo != nil {
			sortedSeoNames = append(sortedSeoNames, seo.name)
			sortedSEOs = append(sortedSEOs, seo)
		}
	}

	// sort all QorSEOSettingInterface
	var settingModelSlice = collection.NewSettingModelSlice()
	if db.Find(settingModelSlice, "name in (?)", sortedSeoNames).Error != nil {
		return nil
	}

	reflectVlaue := reflect.Indirect(reflect.ValueOf(settingModelSlice))
	for i := 0; i < reflectVlaue.Len(); i++ {
		if modelSetting, ok := reflectVlaue.Index(i).Interface().(QorSEOSettingInterface); ok {
			sortedDBSettings = append(sortedDBSettings, modelSetting)
		}
	}

	// sort all settings
	if _, ok := obj.(string); !ok {
		if value := reflect.Indirect(reflect.ValueOf(obj)); value.IsValid() && value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
					if setting := value.Field(i).Interface().(Setting); setting.EnabledCustomize {
						sortedSettings = append(sortedSettings, setting)
					}
					break
				}
			}
		}
	}

	for _, s := range sortedDBSettings {
		sortedSettings = append(sortedSettings, s.GetSEOSetting())
	}

	// get the final setting from sortedSettings
	for _, s := range sortedSettings {
		if s.Title != "" && setting.Title == "" {
			setting.Title = s.Title
		}
		if s.Description != "" && setting.Description == "" {
			setting.Description = s.Description
		}
		if s.Keywords != "" && setting.Keywords == "" {
			setting.Keywords = s.Keywords
		}
		if s.OpenGraphURL != "" && setting.OpenGraphURL == "" {
			setting.OpenGraphURL = s.OpenGraphURL
		}
		if s.OpenGraphType != "" && setting.OpenGraphType == "" {
			setting.OpenGraphType = s.OpenGraphType
		}
		if s.OpenGraphImageURL != "" && setting.OpenGraphImageURL == "" {
			setting.OpenGraphImageURL = s.OpenGraphImageURL
		}
		if s.OpenGraphImageFromMediaLibrary.URL("og") != "" && setting.OpenGraphImageURL == "" {
			setting.OpenGraphImageURL = s.OpenGraphImageFromMediaLibrary.URL("og")
		}
		if len(s.OpenGraphMetadata) > 0 && len(setting.OpenGraphMetadata) == 0 {
			setting.OpenGraphMetadata = s.OpenGraphMetadata
		}
	}

	if setting.OpenGraphURL != "" && !isAbsoluteURL(setting.OpenGraphURL) {
		var u url.URL
		u.Host = req.Host
		if req.URL.Scheme != "" {
			u.Scheme = req.URL.Scheme
		} else {
			u.Scheme = "http"
		}
		setting.OpenGraphURL = path.Join(u.String(), setting.OpenGraphURL)
	}

	// fetch all variables and tags from context
	var (
		variables = map[string]string{}
		tags      = map[string]string{}
	)

	for _, s := range sortedDBSettings {
		for key, val := range s.GetVariables() {
			variables[key] = val
		}
	}

	for _, seo := range sortedSEOs {
		for key, f := range seo.contextVariables {
			value := f(obj, &setting, req)
			if strings.Contains(key, ":") {
				tags[key] = value
			} else {
				variables[key] = f(obj, &setting, req)
			}
		}
	}

	setting = replaceVariables(setting, variables)
	return setting.HTMLComponent(tags)
}

// GetDB get db from context
func (collection Collection) getDBFromContext(ctx context.Context) *gorm.DB {
	if contextdb := ctx.Value(collection.dbContextKey); contextdb != nil {
		return contextdb.(*gorm.DB)
	}
	return GlobalDB
}

var regex = regexp.MustCompile("{{([a-zA-Z0-9]*)}}")

func replaceVariables(setting Setting, values map[string]string) Setting {
	replace := func(str string) string {
		matches := regex.FindAllStringSubmatch(str, -1)
		for _, match := range matches {
			str = strings.Replace(str, match[0], values[match[1]], 1)
		}
		return str
	}

	setting.Title = replace(setting.Title)
	setting.Description = replace(setting.Description)
	setting.Keywords = replace(setting.Keywords)
	return setting
}

func isAbsoluteURL(str string) bool {
	if u, err := url.Parse(str); err == nil && u.IsAbs() {
		return true
	}
	return false
}

func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBContextKey, db)
}
