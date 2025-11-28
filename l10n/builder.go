package l10n

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"time"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
)

var IncorrectLocaleErr = errors.New("incorrect locale")

type Builder struct {
	db *gorm.DB
	ab *activity.Builder
	// models                               []*presets.ModelBuilder
	locales                              []*loc
	getSupportLocaleCodesFromRequestFunc func(R *http.Request) []string
	cookieName                           string
	queryName                            string
}

type loc struct {
	code  string
	path  string
	label string
	img   string
}

func New(db *gorm.DB) *Builder {
	b := &Builder{
		db:         db,
		cookieName: "locale",
		queryName:  "locale",
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

func (b *Builder) RegisterLocales(localeCode, localePath, localeLabel, img string) (r *Builder) {
	if slices.ContainsFunc(b.locales, func(l *loc) bool {
		return l.code == localeCode
	}) {
		return b
	}

	b.locales = append(b.locales, &loc{
		code:  localeCode,
		path:  path.Join("/", localePath),
		label: localeLabel,
		img:   img,
	})
	return b
}

func (b *Builder) GetLocalePath(localeCode string) string {
	if b == nil {
		return ""
	}
	for _, l := range b.locales {
		if l.code == localeCode {
			return l.path
		}
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

	if locale, ok := IsLocalizableFromContext(ctx); ok {
		localePath = l10nBuilder.GetLocalePath(locale)
	}

	if localeCode, err := reflectutils.Get(m, "LocaleCode"); err == nil {
		localePath = l10nBuilder.GetLocalePath(localeCode.(string))
	}

	return
}

func (b *Builder) GetAllLocalePaths() (r []string) {
	for _, l := range b.locales {
		r = append(r, l.path)
	}
	return
}

func (b *Builder) GetLocaleLabel(localeCode string) string {
	for _, l := range b.locales {
		if l.code == localeCode {
			return l.label
		}
	}
	return "Unknown"
}

func (b *Builder) GetLocaleImg(localeCode string) string {
	for _, l := range b.locales {
		if l.code == localeCode {
			return l.img
		}
	}
	return ""
}

func (b *Builder) GetSupportLocaleCodes() (r []string) {
	for _, l := range b.locales {
		r = append(r, l.code)
	}
	return
}

func (b *Builder) GetSupportLocaleCodesFromRequest(R *http.Request) []string {
	if b.getSupportLocaleCodesFromRequestFunc != nil {
		return b.getSupportLocaleCodesFromRequestFunc(R)
	}
	return b.GetSupportLocaleCodes()
}

func (b *Builder) SupportLocalesFunc(v func(R *http.Request) []string) (r *Builder) {
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
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			value := b.localeValue(obj, field, ctx)
			return Input("").Type("hidden").Attr(web.VField("LocaleCode", value)...)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			value := EmbedLocale(obj).LocaleCode
			if !slices.Contains(b.GetSupportLocaleCodesFromRequest(ctx.R), value) {
				return IncorrectLocaleErr
			}

			return nil
		})

	pb.AddWrapHandler(WrapHandlerKey, b.EnsureLocale)
	pb.AddMenuTopItemFunc(MenuTopItemFunc, runSwitchLocaleFunc(b))
	pb.GetI18n().
		RegisterForModule(language.English, I18nLocalizeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLocalizeKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nLocalizeKey, Messages_ja_JP)
	pb.SwitchLocaleFunc(b.runSwitchLocaleFunc)
	return nil
}

func (b *Builder) localeValue(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string {
	var value string
	id, err := reflectutils.Get(obj, "ID")
	if err == nil && fmt.Sprint(id) != "" && fmt.Sprint(id) != "0" {
		value = EmbedLocale(obj).LocaleCode
	} else {
		value = b.GetCorrectLocaleCode(ctx.R)
	}
	return value
}

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	ab := b.ab
	db := b.db
	obj := m.NewModel()
	_ = obj.(presets.SlugEncoder)
	_ = obj.(presets.SlugDecoder)
	_ = obj.(LocaleInterface)

	m.Listing().Field("Locale")
	m.Editing().Field("Locale")

	m.Listing().WrapSearchFunc(func(searcher presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			if localeCode := ctx.R.Context().Value(LocaleCode); localeCode != nil {
				con := presets.SQLCondition{
					Query: "locale_code = ?",
					Args:  []interface{}{localeCode},
				}
				params.SQLConditions = append(params.SQLConditions, &con)
			}

			return searcher(ctx, params)
		}
	})

	m.Editing().WrapSetterFunc(func(setter presets.SetterFunc) presets.SetterFunc {
		return func(obj interface{}, ctx *web.EventContext) {
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
		}
	})

	m.Editing().WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = in(obj, id, ctx); err != nil {
				return
			}
			locale := obj.(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)["locale_code"]
			locale = fmt.Sprintf("%s(del:%d)", locale, time.Now().UnixMilli())

			var withoutKeys []string
			if ctx.R.URL.Query().Get("all_versions") == "true" {
				withoutKeys = append(withoutKeys, "version")
			}

			if err = utils.PrimarySluggerWhere(db.Unscoped(), obj, id, withoutKeys...).Update("locale_code", locale).Error; err != nil {
				return
			}
			return
		}
	})

	registerEventFuncs(db, m, b, ab)

	pb.FieldDefaults(presets.LIST).
		FieldType(Locale{}).
		ComponentFunc(localeListFunc(db, b))
	pb.FieldDefaults(presets.WRITE).
		FieldType(Locale{}).
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			value := b.localeValue(obj, field, ctx)
			return Input("").Type("hidden").Attr(web.VField("LocaleCode", value)...)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			value := EmbedLocale(obj).LocaleCode
			if !slices.Contains(b.GetSupportLocaleCodesFromRequest(ctx.R), value) {
				return IncorrectLocaleErr
			}

			return nil
		})
	rmb := m.Listing().RowMenu()
	rmb.RowMenuItem("Localize").ComponentFunc(localizeRowMenuItemFunc(m.Info(), "", url.Values{}))

	pb.AddWrapHandler(WrapHandlerKey, b.EnsureLocale)
	pb.AddMenuTopItemFunc(MenuTopItemFunc, runSwitchLocaleFunc(b))
	pb.GetI18n().
		RegisterForModule(language.English, I18nLocalizeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLocalizeKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nLocalizeKey, Messages_ja_JP)
	return nil
}
