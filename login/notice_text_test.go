package login

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/i18n"
)

type noticeTextMessages struct {
	LoginNoticePasswordTooShort string
	LoginNoticeEmpty            string
}

func TestNoticeText(t *testing.T) {
	i18nB := i18n.New().
		SupportLanguages(language.English, language.Japanese).
		RegisterForModule(language.English, presets.ModelsI18nModuleKey, &noticeTextMessages{
			LoginNoticePasswordTooShort: "Password is too short",
		}).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, &noticeTextMessages{
			LoginNoticePasswordTooShort: "パスワードが短すぎます",
		})

	tests := []struct {
		name           string
		acceptLanguage string
		message        string
		want           string
	}{
		{name: "key in English", acceptLanguage: "en", message: "PasswordTooShort", want: "Password is too short"},
		{name: "key in Japanese", acceptLanguage: "ja", message: "PasswordTooShort", want: "パスワードが短すぎます"},
		{name: "text with no key registered is shown as it is", acceptLanguage: "ja", message: "Password cannot be less than 12 characters. Try again.", want: "Password cannot be less than 12 characters. Try again."},
		{name: "key registered with no text is shown as it is", acceptLanguage: "ja", message: "Empty", want: "Empty"},
		{name: "empty message", acceptLanguage: "ja", message: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			h := i18nB.EnsureLanguage(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = NoticeText(r, tt.message)
			}))
			r := httptest.NewRequest(http.MethodGet, "/auth/reset-password", nil)
			r.Header.Set("Accept-Language", tt.acceptLanguage)
			h.ServeHTTP(httptest.NewRecorder(), r)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("no language resolved", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/auth/reset-password", nil)
		assert.Equal(t, "PasswordTooShort", NoticeText(r, "PasswordTooShort"))
	})
}
