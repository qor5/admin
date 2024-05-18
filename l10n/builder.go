package l10n

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	"github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type Builder struct {
	db *gorm.DB
	ab *activity.Builder
	// models                               []*presets.ModelBuilder
	supportLocaleCodes                   []string
	localesPaths                         map[string]string
	paths                                []string
	localesLabels                        map[string]string
	getSupportLocaleCodesFromRequestFunc func(R *http.Request) []string
	cookieName                           string
	queryName                            string
}

func New(db *gorm.DB) *Builder {
	b := &Builder{
		db:                 db,
		supportLocaleCodes: []string{},
		localesPaths:       make(map[string]string),
		paths:              []string{},
		localesLabels:      make(map[string]string),
		cookieName:         "locale",
		queryName:          "locale",
	}
	return b
}

func (b *Builder) IsTurnedOn() bool {
	return len(b.GetSupportLocaleCodes()) > 0
}

func (b *Builder) GetCookieName() string {
	return b.cookieName
}

func (b *Builder) GetQueryName() string {
	return b.queryName
}

func (b *Builder) Activity(v *activity.Builder) (r *Builder) {
	b.ab = v
	return b
}

// func (b *Builder) Models(vs ...*presets.ModelBuilder) (r *Builder) {
// 	b.models = append(b.models, vs...)
// 	return b
// }

func (b *Builder) RegisterLocales(localeCode, localePath, localeLabel string) (r *Builder) {
	b.supportLocaleCodes = append(b.supportLocaleCodes, localeCode)
	b.localesPaths[localeCode] = path.Join("/", localePath)
	if !utils.Contains(b.paths, localePath) {
		b.paths = append(b.paths, localePath)
	}
	b.localesLabels[localeCode] = localeLabel
	return b
}

func (b *Builder) GetLocalePath(localeCode string) string {
	if b == nil {
		return ""
	}
	p, exist := b.localesPaths[localeCode]
	if exist {
		return p
	}
	return ""
}

type contextKeyType int

const contextKey contextKeyType = iota

func (b *Builder) ContextValueProvider(in context.Context) context.Context {
	return context.WithValue(in, contextKey, b)
}

func builderFromContext(c context.Context) (b *Builder, ok bool) {
	b, ok = c.Value(contextKey).(*Builder)
	return
}

func LocalePathFromContext(m interface{}, ctx context.Context) (localePath string) {
	l10nBuilder, ok := builderFromContext(ctx)
	if !ok {
		return
	}

	if locale, ok := IsLocalizableFromCtx(ctx); ok {
		localePath = l10nBuilder.GetLocalePath(locale)
	}

	if localeCode, err := reflectutils.Get(m, "LocaleCode"); err == nil {
		localePath = l10nBuilder.GetLocalePath(localeCode.(string))
	}

	return
}

func (b *Builder) GetAllLocalePaths() []string {
	return b.paths
}

func (b *Builder) GetLocaleLabel(localeCode string) string {
	label, exist := b.localesLabels[localeCode]
	if exist {
		return label
	}
	return "Unkonw"
}

func (b *Builder) GetSupportLocaleCodes() []string {
	return b.supportLocaleCodes
}

func (b *Builder) GetSupportLocaleCodesFromRequest(R *http.Request) []string {
	if b.getSupportLocaleCodesFromRequestFunc != nil {
		return b.getSupportLocaleCodesFromRequestFunc(R)
	}
	return b.GetSupportLocaleCodes()
}

func (b *Builder) GetSupportLocaleCodesFromRequestFunc(v func(R *http.Request) []string) (r *Builder) {
	b.getSupportLocaleCodesFromRequestFunc = v
	return b
}

func (b *Builder) GetCurrentLocaleCodeFromCookie(r *http.Request) (localeCode string) {
	localeCookie, _ := r.Cookie(b.cookieName)
	if localeCookie != nil {
		localeCode = localeCookie.Value
	}
	return
}

func (b *Builder) GetCorrectLocaleCode(r *http.Request) string {
	localeCode := r.FormValue(b.queryName)
	if localeCode == "" {
		localeCode = b.GetCurrentLocaleCodeFromCookie(r)
	}

	supportLocaleCodes := b.GetSupportLocaleCodesFromRequest(r)
	for _, v := range supportLocaleCodes {
		if localeCode == v {
			return v
		}
	}

	return supportLocaleCodes[0]
}

type l10nContextKey int

const LocaleCode l10nContextKey = iota

func (b *Builder) EnsureLocale(in http.Handler) (out http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(b.GetSupportLocaleCodesFromRequest(r)) == 0 {
			in.ServeHTTP(w, r)
			return
		}

		localeCode := b.GetCorrectLocaleCode(r)

		maxAge := 365 * 24 * 60 * 60
		http.SetCookie(w, &http.Cookie{
			Name:    b.cookieName,
			Value:   localeCode,
			Path:    "/",
			MaxAge:  maxAge,
			Expires: time.Now().Add(time.Duration(maxAge) * time.Second),
		})
		ctx := context.WithValue(r.Context(), LocaleCode, localeCode)

		in.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (b *Builder) Install(pb *presets.Builder) error {
	db := b.db

	pb.FieldDefaults(presets.LIST).
		FieldType(Locale{}).
		ComponentFunc(localeListFunc(db, b))
	pb.FieldDefaults(presets.WRITE).
		FieldType(Locale{}).
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			var value string
			id, err := reflectutils.Get(obj, "ID")
			if err == nil && len(fmt.Sprint(id)) > 0 && fmt.Sprint(id) != "0" {
				value = field.Value(obj).(Locale).GetLocale()
			} else {
				value = b.GetCorrectLocaleCode(ctx.R)
			}

			return htmlgo.Input("").Type("hidden").Attr(web.VField("LocaleCode", value)...)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			value := field.Value(obj).(Locale).GetLocale()
			if !utils.Contains(b.GetSupportLocaleCodesFromRequest(ctx.R), value) {
				return errors.New("Incorrect locale.")
			}

			return nil
		})

	pb.AddWrapHandler(WrapHandlerKey, b.EnsureLocale)
	pb.AddMenuTopItemFunc(MenuTopItemFunc, runSwitchLocaleFunc(b))
	pb.I18n().
		RegisterForModule(language.English, I18nLocalizeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLocalizeKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nLocalizeKey, Messages_ja_JP)
	return nil
}

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	ab := b.ab
	db := b.db
	obj := m.NewModel()
	_ = obj.(presets.SlugEncoder)
	_ = obj.(presets.SlugDecoder)
	_ = obj.(L10nInterface)
	if l10nONModel, exist := obj.(L10nONInterface); exist {
		l10nONModel.L10nON()
	}
	m.Listing().Field("Locale")
	m.Editing().Field("Locale")

	searcher := m.Listing().Searcher
	m.Listing().SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		if localeCode := ctx.R.Context().Value(LocaleCode); localeCode != nil {
			con := presets.SQLCondition{
				Query: "locale_code = ?",
				Args:  []interface{}{localeCode},
			}
			params.SQLConditions = append(params.SQLConditions, &con)
		}

		return searcher(model, params, ctx)
	})

	setter := m.Editing().Setter
	m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
		if ctx.Param(presets.ParamID) == "" {
			if localeCode := ctx.R.Context().Value(LocaleCode); localeCode != nil {
				if err := reflectutils.Set(obj, "LocaleCode", localeCode); err != nil {
					return
				}
			}
		}
		if setter != nil {
			setter(obj, ctx)
		}
	})

	deleter := m.Editing().Deleter
	m.Editing().DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		if err = deleter(obj, id, ctx); err != nil {
			return
		}
		locale := obj.(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)["locale_code"]
		locale = fmt.Sprintf("%s(del:%d)", locale, time.Now().UnixMilli())

		withoutKeys := []string{}
		if ctx.R.URL.Query().Get("all_versions") == "true" {
			withoutKeys = append(withoutKeys, "version")
		}

		if err = utils.PrimarySluggerWhere(db.Unscoped(), obj, id, withoutKeys...).Update("locale_code", locale).Error; err != nil {
			return
		}
		return
	})

	rmb := m.Listing().RowMenu()
	rmb.RowMenuItem("Localize").ComponentFunc(localizeRowMenuItemFunc(m.Info(), "", url.Values{}))

	registerEventFuncs(db, m, b, ab)

	pb.FieldDefaults(presets.LIST).
		FieldType(Locale{}).
		ComponentFunc(localeListFunc(db, b))
	pb.FieldDefaults(presets.WRITE).
		FieldType(Locale{}).
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) htmlgo.HTMLComponent {
			var value string
			id, err := reflectutils.Get(obj, "ID")
			if err == nil && len(fmt.Sprint(id)) > 0 && fmt.Sprint(id) != "0" {
				value = field.Value(obj).(Locale).GetLocale()
			} else {
				value = b.GetCorrectLocaleCode(ctx.R)
			}

			return htmlgo.Input("").Type("hidden").Attr(web.VField("LocaleCode", value)...)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			value := field.Value(obj).(Locale).GetLocale()
			if !utils.Contains(b.GetSupportLocaleCodesFromRequest(ctx.R), value) {
				return errors.New("Incorrect locale.")
			}

			return nil
		})

	pb.AddWrapHandler(WrapHandlerKey, b.EnsureLocale)
	pb.AddMenuTopItemFunc(MenuTopItemFunc, runSwitchLocaleFunc(b))
	pb.I18n().
		RegisterForModule(language.English, I18nLocalizeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLocalizeKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nLocalizeKey, Messages_ja_JP)
	return nil
}
