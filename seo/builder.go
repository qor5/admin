package seo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultGlobalSEOName = "Global SEO"
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

func WithLocales(locales ...string) Option {
	return func(b *Builder) {
		b.locales = locales
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

func New(db *gorm.DB, ops ...Option) *Builder {
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
	mb        *presets.ModelBuilder
}

// @snippet_end

// RegisterSEO registers a SEO through name or model.
// There are two types of SEOs, one is SEO with model, the other is SEO without model aka 'non-model seo'.
// if you want to register a non-model SEO, you can call RegisterSEO method like this:
// seoBuilder.RegisterSEO("About Us")
// if you want to register a SEO with model, you can call RegisterSEO method like this:
// seoBuilder.RegisterSEO("Product", &Product{})
// the first parameter of RegisterSEO method is the name of the SEO, it will be displayed in the admin user interface.
//
// If the SEO to be registered already exists, it will panic.
// The optional second parameter names `model` is an instance of a type
// that has a field of type `Setting`, if the type of model does not have
// such field or len(model) > 1, the program will panic.
//
// The default parent of the registered SEO is global seo. If you need to set
// its parent, Please call the SetParent method of SEO after invoking RegisterSEO method.
// For Example: b.RegisterSEO("Region", &Region{}).SetParent(parentSEO)
// Or you can call appendChild method to add the SEO to the specified parent.
// For Example: b.GetGlobalSEO().appendChild(b.RegisterSEO("Region", &Region{}))
func (b *Builder) RegisterSEO(name string, model ...interface{}) *SEO {
	seoName := strings.TrimSpace(name)
	if seoName == "" {
		panic("the seo name must not be empty")
	}
	if _, isExist := b.registeredSEO[seoName]; isExist {
		panic(fmt.Sprintf("the %v SEO already exists!", seoName))
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
// The other is an instance of a struct that has a field of type `Setting`, in which case
// the SEO corresponding to the type of the struct will be returned.
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

// SortSEOs sorts the SEOs in the order of their priority.
// The global SEO is always the first element in the sorted slice.
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

// NonModelSEO is used to store the name and locale of non-model SEO.
// If you register a SEO without model like the following:
// b.RegisterSEO("About Us")
// When you want to render the SEO, you must pass a NonModelSEO instance to Render method.
// For Example:
// b.Render(NewNonModelSEO("About Us", "en"))
// or
// b.render(NewNonModelSEO("About Us")) when the locales passed to New method is only one.
type NonModelSEO interface {
	GetName() string
	l10n.LocaleInterface
}

type nonModelSEO struct {
	Name string
	l10n.Locale
}

func (n *nonModelSEO) GetName() string {
	return n.Name
}

// NewNonModelSEO creates a nonModelSEO instance that implements NonModelSEO interface.
// This function is only used to create a NonModelSEO instance passed to Render method.
// the name parameter must be the same as the name you passed to RegisterSEO method.
// For Example:
// If you register a SEO like this: b.RegisterSEO("About Us"),
// when you want render "About US" SEO, you must pass "About Us" to NewNonModelSEO method like this:
// b := New(db, []string{"en", "zh"})
// b.Render(NewNonModelSEO("About Us", "en")).
//
// For convenience, you can call NewNonModelSEO function without passing the locale parameter when
// the locales passed to New method is only one.
// For Example:
// b := newBuilder(db, []string{"en"}) // only one locale(en) is passed to New method.
// b.render(NewNonModelSEO("About Us")) // the only locale(en) passed to New method will be used.
func NewNonModelSEO(name string, locale ...string) NonModelSEO {
	var loc string
	if len(locale) > 0 {
		loc = locale[0]
	}
	return &nonModelSEO{
		Name:   name,
		Locale: l10n.Locale{LocaleCode: loc},
	}
}

// NewNonModelSEOSlice creates a slice of NonModelSEO.
// The function is only used to create a slice of NonModelSEO passed to BatchRender method.
// For Example: following code will render the non-model SEO named "About Us" with locale "en" and "zh".
// b.BatchRender(NewNonModelSEOSlice("About Us", "en", "zh"))
func NewNonModelSEOSlice(name string, locales ...string) []NonModelSEO {
	ret := make([]NonModelSEO, 0, len(locales))
	for _, locale := range locales {
		ret = append(ret, &nonModelSEO{
			Name:   name,
			Locale: l10n.Locale{LocaleCode: locale},
		})
	}
	return ret
}

// Render renders the SEO according to the specified object.
// obj must be of type NonModelSEO or a pointer to a struct that has a field of type `Setting`.
//
// If the obj is an instance of NonModelSEO, it will render the non-model SEO.
// For Example: following code will render the non-model SEO named "About Us" with locale "en".
// b.render(NewNonModelSEO("About Us", "en"))
//
// When the locales passed to New method is only one,
// you can call NewNonModelSEO without passing the locale parameter.
// in this case, the only locale passed to New method will be used.
// For Example: following code will render the non-model SEO named "About Us" with locale "en".
// b := New(db, []string{"en"})
// b.Render(NewNonModelSEO("About Us"))
func (b *Builder) Render(obj interface{}, req *http.Request) h.HTMLComponent {
	var seo *SEO
	var locale string
	objV := reflect.ValueOf(obj)
	if objV.Kind() != reflect.Ptr {
		panic("the obj must be a pointer to a struct")
	}
	if nModelSEO, ok := obj.(NonModelSEO); ok {
		seo = b.registeredSEO[nModelSEO.GetName()]
	} else {
		objV = reflect.Indirect(objV)
		seo = b.registeredSEO[objV.Type()]
	}
	// if the seo is not registered, return empty html component.
	if seo == nil {
		return h.RawHTML("")
	}
	if v, ok := obj.(l10n.LocaleInterface); ok {
		locale = v.EmbedLocale().LocaleCode
	}
	if locale == "" && len(b.locales) == 1 {
		locale = b.locales[0]
	}
	localeFinalSeoSetting := seo.getLocaleFinalQorSEOSetting(locale, b.db)
	return b.render(obj, localeFinalSeoSetting, seo, req)
}

// BatchRender rendering multiple SEOs at once.
// objs must be a slice, and each element in objs must be of the same type.
// It is the responsibility of the caller to ensure that every element in objs
// is of the same type, as it is performance-intensive to check whether each element
// in `objs` if of the same type through reflection.
//
// If you want to render non-model SEO, you must pass a slice of NonModelSEO to objs.
// For convenience, you can call NewNonModelSEOSlice function to create a slice of NonModelSEO.
// For Example: Following code will render the non-model SEO named "About Us" with locale "en" and "zh".
// b.BatchRender(NewNonModelSEOSlice("About Us", "en", "zh"))
func (b *Builder) BatchRender(objs interface{}, req *http.Request) []h.HTMLComponent {
	v := reflect.ValueOf(objs)
	if v.Kind() != reflect.Slice {
		panic("the objs must be a slice")
	}
	if v.Len() == 0 {
		return nil
	}

	// The seo retrieved from the first element in objs will be used to render all elements in objs.
	// If the objs consists of different types of elements, the rendering result will be incorrect.
	var seo *SEO
	isNonModelSEO := false
	vAtIndex0 := reflect.Indirect(v.Index(0))
	if v, ok := vAtIndex0.Interface().(NonModelSEO); ok {
		seo = b.registeredSEO[v.GetName()]
		isNonModelSEO = true
	} else {
		seo = b.registeredSEO[vAtIndex0.Type()]
	}
	if seo == nil {
		return nil
	}

	finalSeoSettings := seo.getFinalQorSEOSetting(b.db)
	comps := make([]h.HTMLComponent, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		objV := reflect.Indirect(v.Index(i))
		var obj interface{}
		var locale string
		if isNonModelSEO {
			obj = objV.Interface()
			locale = obj.(NonModelSEO).EmbedLocale().LocaleCode
		} else {
			// the purpose of objV.addr().Interface() is to get the pointer of the element in objs.
			// if the element in objs is a pointer, it will return the pointer itself.
			// if the element in objs is not a pointer, it will return a pointer to the element.
			// this will prevent unnecessary value copy and caller can use a slice of value type
			// to call BatchRender method.
			// For Example:
			// b.BatchRender([]Product{...}) // []Product is a slice of value type.
			obj = objV.Addr().Interface()
			if v, ok := obj.(l10n.LocaleInterface); ok {
				locale = v.EmbedLocale().LocaleCode
			}
		}
		if locale == "" && len(b.locales) == 1 {
			locale = b.locales[0]
		}

		defaultSetting := finalSeoSettings[locale]
		if defaultSetting == nil {
			panic(fmt.Sprintf("There are no available seo configuration for %v locale", locale))
		}
		comps = append(comps, b.render(obj, defaultSetting, seo, req))
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

// insertIfNotExists inserts the specified seo with the specified locales into the database.
// if the seo already exists, it will not be inserted into the database.
func insertIfNotExists(db *gorm.DB, seoName string, locales []string) error {
	settings := make([]QorSEOSetting, 0, len(locales))
	for _, locale := range locales {
		settings = append(settings, QorSEOSetting{
			Name:   seoName,
			Locale: l10n.Locale{LocaleCode: locale},
		})
	}
	if len(locales) == 0 {
		settings = append(settings, QorSEOSetting{
			Name: seoName,
		})
	}
	// The aim to use `Clauses(clause.OnConflict{DoNothing: true})` is it will not affect the existing data
	// or cause the create function to fail When the data to be inserted already exists in the database,
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&settings).Error; err != nil {
		return err
	}
	return nil
}

func (b *Builder) AutoMigrate() (r *Builder) {
	if err := AutoMigrate(b, b.db); err != nil {
		return nil
	}
	return b
}

// Deprecated: use Migrate instead.
func AutoMigrate(b *Builder, db *gorm.DB) (err error) {
	if err = db.AutoMigrate(&QorSEOSetting{}); err != nil {
		panic(err)
	}
	// NOTE: do not replace b.seoRoot.name with defaultGlobalSEOName.
	// because the name of global seo may be changed by user through WithGlobalSEOName option.
	if err = insertIfNotExists(db, b.seoRoot.name, b.locales); err != nil {
		return
	}
	return
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&QorSEOSetting{})
}
