package location

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func TestLocaleFromLanguage(t *testing.T) {
	tests := []struct {
		name string
		t    language.Tag
		want string
	}{
		{"de", language.German, "de"},
		{"en", language.English, "en"},
		{"es", language.Spanish, "es"},
		{"fr", language.French, "fr"},
		{"ja", language.Japanese, "ja"},
		{"pt-BR", language.BrazilianPortuguese, "pt-BR"},
		{"ru", language.Russian, "ru"},
		{"zh-CN", language.SimplifiedChinese, "zh-CN"},
		{"zh-Hant", language.TraditionalChinese, "zh-CN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LocaleFromLanguage(tt.t)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetLocation(t *testing.T) {
	type args struct {
		lang language.Tag
		addr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr string
	}{
		{"en", args{lang: language.English, addr: "81x.2.69.160x"}, "", "invalid IP address"},

		{"en", args{lang: language.English, addr: "81.2.69.160"}, "London, United Kingdom", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "81.2.69.160"}, "英国", ""},
		{"ja", args{lang: language.Japanese, addr: "81.2.69.160"}, "ロンドン、イギリス", ""},

		{"en", args{lang: language.English, addr: "2001:218:abcd:1234:5678:90ab:cdef:1"}, "Japan", ""},
		{"en", args{lang: language.English, addr: "[2001:218:abcd:1234:5678:90ab:cdef:1]"}, "Japan", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "2001:218:abcd:1234:5678:90ab:cdef:1"}, "日本", ""},

		{"en", args{lang: language.English, addr: "175.16.199.26"}, "Changchun, China", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "175.16.199.26"}, "长春，中国", ""},
		{"ja", args{lang: language.Japanese, addr: "175.16.199.26"}, "長春市、中国", ""},
	}

	db, err := New("test-data/test-data/GeoIP2-City-Test.mmdb")
	require.Nil(t, err)
	defer db.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetLocation(context.Background(), tt.args.lang, tt.args.addr)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}