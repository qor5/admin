package seo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/qor5/admin/l10n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultGlobalSEOName = "Global SEO"
	defaultLocale        = "en"
)

type (
	ContextVarFunc func(interface{}, *Setting, *http.Request) string
	Option         func(*Builder)
)

func WithInherit(inherited bool) Option {
	return func(b *Builder) {
		b.inherited = inherited
	}
}

func WithGlobalSEOName(name string) Option {
	return func(b *Builder) {
		name = strings.TrimSpace(name)
		if name == "" {
			panic("The global seo name must be not empty")
		}
		b.seoRoot.name = name
		delete(b.registeredSEO, defaultGlobalSEOName)
		b.registeredSEO[name] = b.seoRoot
	}
}

func WithLocales(locales ...string) Option {
	return func(b *Builder) {
		b.locales = locales
	}
}

func NewBuilder(db *gorm.DB, ops ...Option) *Builder {
	globalSEO := &SEO{name: defaultGlobalSEOName}
	globalSEO.RegisterSettingVariables("SiteName")
	b := &Builder{
		registeredSEO: make(map[interface{}]*SEO),
		seoRoot:       globalSEO,
		inherited:     true,
		db:            db,
	}
	b.registeredSEO[defaultGlobalSEOName] = b.seoRoot

	for _, opFunc := range ops {
		opFunc(b)
	}

	if err := db.AutoMigrate(&QorSEOSetting{}); err != nil {
		panic(err)
	}

	if err := insertIfNotExists(db, b.seoRoot.name, b.locales); err != nil {
		panic(err)
	}
	return b
}

// Builder will hold registered SEO configures and global setting definition and other configures
// @snippet_begin(SeoBuilderDefinition)
type Builder struct {
	// key == val.Name
	registeredSEO map[interface{}]*SEO

	locales   []string
	db        *gorm.DB
	seoRoot   *SEO
	inherited bool
	afterSave func(ctx context.Context, settingName string, locale string) error // hook called after saving
}

// @snippet_end

// RegisterSEO registers a SEO through name or model.
// If the SEO to be registered already exists, it will panic.
// The optional second parameter names `model` is an instance of a type
// that has a field of type `Setting`, if the type of model does not have
// such field or len(model) > 1, the program will panic.
// The default parent of the registered SEO is global seo. If you need to set
// its parent, Please call the SetParent method of SEO after invoking RegisterSEO method.
// For Example: b.RegisterSEO("Region", &Region{}).SetParent(parentSEO)
func (b *Builder) RegisterSEO(name string, model ...interface{}) *SEO {
	seoName := strings.TrimSpace(name)
	if seoName == "" {
		panic("the seo name must not be empty")
	}
	if _, isExist := b.registeredSEO[seoName]; isExist {
		panic(fmt.Sprintf("The %v SEO already exists!", seoName))
	}

	seo := &SEO{name: seoName}
	// default parent is seoRoot
	seo.SetParent(b.seoRoot)
	b.registeredSEO[seoName] = seo
	if len(model) > 0 {
		if len(model) > 1 {
			panic("too many arguments")
		}
		modelType := reflect.Indirect(reflect.ValueOf(model[0])).Type()
		isSettingNested := false
		if modelType.Kind() == reflect.Struct {
			for i := 0; i < modelType.NumField(); i++ {
				if modelType.Field(i).Type == reflect.TypeOf(Setting{}) {
					isSettingNested = true
					break
				}
			}
		}
		if !isSettingNested {
			panic("model must be of struct type that embed Setting struct")
		}
		if _, isExist := b.registeredSEO[modelType]; isExist {
			panic(fmt.Sprintf("the seo for %v model has been registered", modelType.Name()))
		}
		b.registeredSEO[modelType] = seo
	}

	if err := insertIfNotExists(b.db, seoName, b.locales); err != nil {
		panic(err)
	}
	return seo
}

// RemoveSEO removes the specified SEO,
// if the SEO has children, the parent of the children will
// be the parent of the SEO
func (b *Builder) RemoveSEO(obj interface{}) *Builder {
	seoToBeRemoved := b.GetSEO(obj)
	if seoToBeRemoved == nil || seoToBeRemoved == b.seoRoot {
		return b
	}
	seoToBeRemoved.removeSelf()
	delete(b.registeredSEO, seoToBeRemoved.name)
	return b
}

// GetSEO retrieves the specified SEO, It accepts two types of parameters.
// One is a string, where the literal value of the parameter is the name of the SEO.
// The other is an instance of a struct embedded with the Setting type, in which case
// the SEO name is obtained from the type name that is retrieved through reflection.
// If no SEO with the specified name is found, it returns nil.
func (b *Builder) GetSEO(obj interface{}) *SEO {
	switch res := obj.(type) {
	case string:
		return b.registeredSEO[res]
	default:
		modelType := reflect.Indirect(reflect.ValueOf(obj)).Type()
		return b.registeredSEO[modelType]
	}
}

func (b *Builder) GetGlobalSEO() *SEO {
	return b.seoRoot
}

// GetSEOPriority gets the priority of the specified SEO,
// with higher number indicating higher priority.
// The priority of Global SEO is 1 (the lowest priority)
func (b *Builder) GetSEOPriority(name string) int {
	node := b.GetSEO(name)
	depth := 0
	for node != nil && node.name != "" {
		node = node.parent
		depth++
	}
	return depth
}

func (b *Builder) SortSEOs(SEOs []*QorSEOSetting) {
	orders := make(map[string]int)
	order := 0
	var dfs func(root *SEO)
	dfs = func(seo *SEO) {
		if seo == nil {
			return
		}
		orders[seo.name] = order
		order++
		for _, child := range seo.children {
			dfs(child)
		}
	}
	dfs(b.seoRoot)
	sort.Slice(SEOs, func(i, j int) bool {
		return orders[SEOs[i].Name] < orders[SEOs[j].Name]
	})
}

// AfterSave sets the hook called after saving
func (b *Builder) AfterSave(v func(ctx context.Context, settingName string, locale string) error) *Builder {
	b.afterSave = v
	return b
}

func (b *Builder) Render(obj interface{}, req *http.Request) h.HTMLComponent {
	seo := b.GetSEO(obj)
	if seo == nil {
		return h.RawHTML("")
	}

	locale := defaultLocale
	if v, ok := obj.(l10n.L10nInterface); ok {
		if v.GetLocale() != "" {
			locale = v.GetLocale()
		}
	}
	localeFinalSeoSetting := seo.getLocaleFinalQorSEOSetting(locale, b.db)
	return b.render(obj, localeFinalSeoSetting, seo, req)
}

// BatchRender rendering multiple SEOs at once.
// objs must be a slice, and each element in objs must be of the same type.
// It is the responsibility of the caller to ensure that every element in objs
// is of the same type, as it is performance-intensive to check whether each element
// in `objs` if of the same type through reflection.
func (b *Builder) BatchRender(objs interface{}, req *http.Request) []h.HTMLComponent {
	v := reflect.ValueOf(objs)
	if v.Kind() != reflect.Slice {
		panic("the objs must be a slice")
	}
	if v.Len() == 0 {
		return nil
	}
	seo := b.GetSEO(v.Index(0).Interface())
	if seo == nil {
		return nil
	}
	finalSeoSettings := seo.getFinalQorSEOSetting(b.db)
	comps := make([]h.HTMLComponent, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		obj := v.Index(i).Interface()
		locale := defaultLocale
		if v, ok := obj.(l10n.L10nInterface); ok {
			if v.GetLocale() != "" {
				locale = v.GetLocale()
			}
		}
		defaultSetting := finalSeoSettings[locale]
		if defaultSetting == nil {
			panic(fmt.Sprintf("There are no available seo configuration for %v locale", locale))
		}
		comp := b.render(obj, finalSeoSettings[locale], seo, req)
		comps = append(comps, comp)
	}
	return comps
}

func (b *Builder) render(obj interface{}, defaultSEOSetting *QorSEOSetting, seo *SEO, req *http.Request) h.HTMLComponent {
	// get setting
	var setting Setting
	{
		setting = defaultSEOSetting.Setting
		if _, ok := obj.(string); !ok {
			if value := reflect.Indirect(reflect.ValueOf(obj)); value.IsValid() && value.Kind() == reflect.Struct {
				for i := 0; i < value.NumField(); i++ {
					if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
						if tSetting := value.Field(i).Interface().(Setting); tSetting.EnabledCustomize {
							// if the obj embeds Setting, then overrides `finalSeoSetting.Setting` with `tSetting`
							if b.inherited {
								mergeSetting(&defaultSEOSetting.Setting, &tSetting)
							}
							setting = tSetting
						}
						break
					}
				}
			}
		}
	}

	// replace placeholders
	{
		variables := defaultSEOSetting.Variables
		finalContextVars := seo.getFinalContextVars()
		// execute function for context var
		for varName, varFunc := range finalContextVars {
			variables[varName] = varFunc(obj, &setting, req)
		}
		setting = replaceVariables(setting, variables)
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
	}

	metaProperties := map[string]string{}
	finalMetaProperties := seo.getFinalMetaProps()
	for propName, propFunc := range finalMetaProperties {
		metaProperties[propName] = propFunc(obj, &setting, req)
	}

	return setting.HTMLComponent(metaProperties)
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

func insertIfNotExists(db *gorm.DB, seoName string, locales []string) error {
	settings := make([]QorSEOSetting, 0, len(locales))
	if len(locales) == 0 {
		settings = append(settings, QorSEOSetting{
			Name:   seoName,
			Locale: l10n.Locale{LocaleCode: defaultLocale},
		})
	} else {
		for _, locale := range locales {
			settings = append(settings, QorSEOSetting{
				Name:   seoName,
				Locale: l10n.Locale{LocaleCode: locale},
			})
		}
	}
	// The aim to use `Clauses(clause.OnConflict{DoNothing: true})` is it will not affect the existing data
	// or cause the create function to fail When the data to be inserted already exists in the database,
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&settings).Error; err != nil {
		return err
	}
	return nil
}
