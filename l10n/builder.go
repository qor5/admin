package l10n

import (
	"context"
	"net/http"
	"path"
	"time"

	"github.com/biter777/countries"
)

type Builder struct {
	supportLocales                   []countries.CountryCode
	localesCodes                     map[countries.CountryCode]string
	localesPaths                     map[countries.CountryCode]string
	localesLabels                    map[countries.CountryCode]string
	getSupportLocalesFromRequestFunc func(R *http.Request) []countries.CountryCode
	cookieName                       string
	queryName                        string
}

func New() *Builder {
	b := &Builder{
		supportLocales: []countries.CountryCode{},
		localesCodes:   make(map[countries.CountryCode]string),
		localesPaths:   make(map[countries.CountryCode]string),
		localesLabels:  make(map[countries.CountryCode]string),
		cookieName:     "locale",
		queryName:      "locale",
	}
	return b
}

func (b *Builder) IsTurnedOn() bool {
	return len(b.GetSupportLocales()) > 0
}

func (b *Builder) GetCookieName() string {
	return b.cookieName
}

func (b *Builder) GetQueryName() string {
	return b.queryName
}

func (b *Builder) RegisterLocales(locale countries.CountryCode, localeCode, localePath, localeLabel string) (r *Builder) {
	b.supportLocales = append(b.supportLocales, locale)
	b.localesCodes[locale] = localeCode
	b.localesPaths[locale] = path.Join("/", localePath)
	b.localesLabels[locale] = localeLabel
	return b
}

func (b *Builder) GetLocaleCode(locale countries.CountryCode) string {
	code, exist := b.localesCodes[locale]
	if exist {
		return code
	}
	return "Unkonw"
}

func (b *Builder) GetLocalePath(locale countries.CountryCode) string {
	p, exist := b.localesPaths[locale]
	if exist {
		return p
	}
	return ""
}

func (b *Builder) GetLocaleLabel(locale countries.CountryCode) string {
	label, exist := b.localesLabels[locale]
	if exist {
		return label
	}
	return "Unkonw"
}

func (b *Builder) GetSupportLocales() []countries.CountryCode {
	return b.supportLocales
}

func (b *Builder) GetSupportLocalesFromRequest(R *http.Request) []countries.CountryCode {
	if b.getSupportLocalesFromRequestFunc != nil {
		return b.getSupportLocalesFromRequestFunc(R)
	}
	return b.GetSupportLocales()
}

func (b *Builder) GetSupportLocalesFromRequestFunc(v func(R *http.Request) []countries.CountryCode) (r *Builder) {
	b.getSupportLocalesFromRequestFunc = v
	return b
}

func (b *Builder) GetCurrentLocaleFromCookie(r *http.Request) (locale string) {
	localeCookie, _ := r.Cookie(b.cookieName)
	if localeCookie != nil {
		locale = localeCookie.Value
	}
	return
}

func (b *Builder) GetCorrectLocale(r *http.Request) countries.CountryCode {
	locale := r.FormValue(b.queryName)
	if locale == "" {
		locale = b.GetCurrentLocaleFromCookie(r)
	}

	supportLocales := b.GetSupportLocalesFromRequest(r)
	for _, v := range supportLocales {
		if locale == b.GetLocaleCode(v) {
			return v
		}
	}

	return supportLocales[0]
}

type l10nContextKey int

const LocaleCode l10nContextKey = iota

func (b *Builder) EnsureLocale(in http.Handler) (out http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(b.GetSupportLocalesFromRequest(r)) == 0 {
			in.ServeHTTP(w, r)
			return
		}

		var locale = b.GetCorrectLocale(r)

		if !locale.IsValid() {
			in.ServeHTTP(w, r)
			return
		}

		maxAge := 365 * 24 * 60 * 60
		http.SetCookie(w, &http.Cookie{
			Name:    b.cookieName,
			Value:   locale.String(),
			Path:    "/",
			MaxAge:  maxAge,
			Expires: time.Now().Add(time.Duration(maxAge) * time.Second),
		})
		ctx := context.WithValue(r.Context(), LocaleCode, b.GetLocaleCode(locale))

		in.ServeHTTP(w, r.WithContext(ctx))
	})
}
