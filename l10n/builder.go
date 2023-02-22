package l10n

import (
	"context"
	"net/http"
	"path"
	"time"
)

type Builder struct {
	supportLocaleCodes                   []string
	localesPaths                         map[string]string
	localesLabels                        map[string]string
	getSupportLocaleCodesFromRequestFunc func(R *http.Request) []string
	cookieName                           string
	queryName                            string
}

func New() *Builder {
	b := &Builder{
		supportLocaleCodes: []string{},
		localesPaths:       make(map[string]string),
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

func (b *Builder) RegisterLocales(localeCode, localePath, localeLabel string) (r *Builder) {
	b.supportLocaleCodes = append(b.supportLocaleCodes, localeCode)
	b.localesPaths[localeCode] = path.Join("/", localePath)
	b.localesLabels[localeCode] = localeLabel
	return b
}

func (b *Builder) GetLocalePath(localeCode string) string {
	p, exist := b.localesPaths[localeCode]
	if exist {
		return p
	}
	return ""
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

		var localeCode = b.GetCorrectLocaleCode(r)

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
