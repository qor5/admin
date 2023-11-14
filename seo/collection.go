package seo

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor5/admin/l10n"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	GlobalSEO    = "Global SEO"
	GlobalDB     *gorm.DB
	DBContextKey contextKey = "DB"
)

type (
	contextKey           string
	contextVariablesFunc func(interface{}, *Setting, *http.Request) string
)

// NewCollection creates a new SeoCollection instance
func NewCollection() *Collection {
	collection := &Collection{
		settingModel: &QorSEOSetting{},
		dbContextKey: DBContextKey,
		globalName:   GlobalSEO,
		inherited:    true,
	}

	collection.RegisterSEO(GlobalSEO).RegisterSettingVaribles(struct{ SiteName string }{}).
		RegisterContextVariables(
			"og:url", func(_ interface{}, _ *Setting, req *http.Request) string {
				return req.URL.String()
			},
		)

	return collection
}

// Collection will hold registered seo configures and global setting definition and other configures
// @snippet_begin(SeoCollectionDefinition)
type Collection struct {
	registeredSEO []*SEO
	globalName    string                                                             // default name is GlobalSEO
	inherited     bool                                                               // default is true. the order is model seo setting, system seo setting, global seo setting
	dbContextKey  interface{}                                                        // get db from context
	settingModel  interface{}                                                        // db model
	afterSave     func(ctx context.Context, settingName string, locale string) error // hook called after saving
}

// @snippet_end

// SEO represents a seo object for a page
// @snippet_begin(SeoDefinition)
type SEO struct {
	name             string
	modelTyp         reflect.Type
	contextVariables map[string]contextVariablesFunc // fetch context variables from request
	settingVariables interface{}                     // fetch setting variables from db
}

// @snippet_end

func (seo *SEO) SetModel(model interface{}) *SEO {
	seo.modelTyp = reflect.Indirect(reflect.ValueOf(model)).Type()
	return seo
}

// SetName set seo name
func (seo *SEO) SetName(name string) *SEO {
	seo.name = name
	return seo
}

// RegisterContextVariables register context variables. the registered variables will be rendered to the page
func (seo *SEO) RegisterContextVariables(key string, f contextVariablesFunc) *SEO {
	if seo.contextVariables == nil {
		seo.contextVariables = map[string]contextVariablesFunc{}
	}
	seo.contextVariables[key] = f
	return seo
}

// RegisterSettingVaribles register a setting variable
func (seo *SEO) RegisterSettingVaribles(setting interface{}) *SEO {
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
	sliceType := reflect.SliceOf(reflect.PtrTo(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()))
	slice := reflect.New(sliceType)
	slice.Elem().Set(reflect.MakeSlice(sliceType, 0, 0))
	return slice.Interface()
}

func (collection *Collection) SetInherited(b bool) *Collection {
	collection.inherited = b
	return collection
}

func (collection *Collection) SetSettingModel(s interface{}) *Collection {
	collection.settingModel = s
	return collection
}

// SetDBContextKey sets the key to get db instance from context
func (collection *Collection) SetDBContextKey(key interface{}) *Collection {
	collection.dbContextKey = key
	return collection
}

// RegisterSEOByNames registers multiple SEOs at once through name
func (collection *Collection) RegisterSEOByNames(names ...string) *Collection {
	for index := range names {
		collection.registeredSEO = append(collection.registeredSEO, &SEO{name: names[index]})
	}
	return collection
}

// RegisterSEO registers a seo through name or model
func (collection *Collection) RegisterSEO(obj interface{}) (seo *SEO) {
	if name, ok := obj.(string); ok {
		seo = &SEO{name: name}
	} else {
		typ := reflect.Indirect(reflect.ValueOf(obj)).Type()
		seo = &SEO{name: typ.Name(), modelTyp: typ}
	}

	collection.registeredSEO = append(collection.registeredSEO, seo)
	return
}

// RemoveSEO removes the specified seo
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

// GetSEO gets the specified SEO by name or model.
// It calls methods GetSEOByName and GetSEOByModel to realize its functionality.
func (collection *Collection) GetSEO(obj interface{}) *SEO {
	if name, ok := obj.(string); ok {
		return collection.GetSEOByName(name)
	} else {
		return collection.GetSEOByModel(obj)
	}
}

// GetSEOByName gets the specified SEO by name
func (collection *Collection) GetSEOByName(name string) *SEO {
	for _, s := range collection.registeredSEO {
		if s.name == name {
			return s
		}
	}

	return nil
}

// GetSEOByModel gets a seo by model
func (collection *Collection) GetSEOByModel(model interface{}) *SEO {
	for _, s := range collection.registeredSEO {
		if reflect.Indirect(reflect.ValueOf(model)).Type() == s.modelTyp {
			return s
		}
	}

	return nil
}

// AfterSave sets the hook called after saving
func (collection *Collection) AfterSave(v func(ctx context.Context, settingName string, locale string) error) *Collection {
	collection.afterSave = v
	return collection
}

// RenderGlobal renders global SEO
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
		locale           string
	)

	// sort all SEOs
	globalSeo := collection.GetSEO(collection.globalName)
	if globalSeo == nil {
		return h.RawHTML("")
	}

	sortedSEOs = append(sortedSEOs, globalSeo)

	if name, ok := obj.(string); !ok || name != collection.globalName {
		if seo := collection.GetSEO(obj); seo != nil {
			sortedSeoNames = append(sortedSeoNames, seo.name)
			sortedSEOs = append(sortedSEOs, seo)
		}
	}
	sortedSeoNames = append(sortedSeoNames, globalSeo.name)

	if v, ok := obj.(l10n.L10nInterface); ok {
		locale = v.GetLocale()
	}

	// sort all QorSEOSettingInterface
	var settingModelSlice = collection.NewSettingModelSlice()
	if db.Find(settingModelSlice, "name in (?) AND locale_code = ?", sortedSeoNames, locale).Error != nil {
		return h.RawHTML("")
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(settingModelSlice))

	for _, name := range sortedSeoNames {
		for i := 0; i < reflectValue.Len(); i++ {
			if modelSetting, ok := reflectValue.Index(i).Interface().(QorSEOSettingInterface); ok && modelSetting.GetName() == name {
				sortedDBSettings = append(sortedDBSettings, modelSetting)
			}
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
	for i, s := range sortedSettings {
		if !collection.inherited && i >= 1 {
			break
		}

		if s.Title != "" && setting.Title == "" {
			setting.Title = s.Title
		}
		if s.Description != "" && setting.Description == "" {
			setting.Description = s.Description
		}
		if s.Keywords != "" && setting.Keywords == "" {
			setting.Keywords = s.Keywords
		}
		if s.OpenGraphTitle != "" && setting.OpenGraphTitle == "" {
			setting.OpenGraphTitle = s.OpenGraphTitle
		}
		if s.OpenGraphDescription != "" && setting.OpenGraphDescription == "" {
			setting.OpenGraphDescription = s.OpenGraphDescription
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

	for i, seo := range sortedSEOs {
		for key, f := range seo.contextVariables {
			value := f(obj, &setting, req)
			if strings.Contains(key, ":") && collection.inherited {
				tags[key] = value
			} else if strings.Contains(key, ":") && !collection.inherited && i == 0 {
				tags[key] = value
			} else {
				variables[key] = f(obj, &setting, req)
			}
		}
	}
	setting = replaceVariables(setting, variables)
	return setting.HTMLComponent(tags)
}

// getDBFromContext get the db from the ctx
func (collection Collection) getDBFromContext(ctx context.Context) *gorm.DB {
	if ctxDB := ctx.Value(collection.dbContextKey); ctxDB != nil {
		return ctxDB.(*gorm.DB)
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
	setting.OpenGraphTitle = replace(setting.OpenGraphTitle)
	setting.OpenGraphDescription = replace(setting.OpenGraphDescription)
	setting.OpenGraphURL = replace(setting.OpenGraphURL)
	setting.OpenGraphType = replace(setting.OpenGraphType)
	setting.OpenGraphImageURL = replace(setting.OpenGraphImageURL)
	var metadata []OpenGraphMetadata
	for _, m := range setting.OpenGraphMetadata {
		metadata = append(metadata, OpenGraphMetadata{
			Property: m.Property,
			Content:  replace(m.Content),
		})
	}
	setting.OpenGraphMetadata = metadata
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
