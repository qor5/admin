package location

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

// Copyright Notice: We respect MaxMind's intellectual property rights and use these files for testing purposes only.
const (
	TestDataURL = "https://raw.githubusercontent.com/maxmind/MaxMind-DB/main/test-data/GeoIP2-City-Test.mmdb"
	TestDataDir = "test-data"
)

// ensureTestData ensures the GeoIP2 test database file exists and returns its path.
// If the file doesn't exist, it downloads it from the official MaxMind test data repository.
// It panics if any error occurs during the process.
func ensureTestData() string {
	filePath := filepath.Join(TestDataDir, "GeoIP2-City-Test.mmdb")

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return filePath
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(TestDataDir, 0o755); err != nil {
		panic(fmt.Errorf("failed to create test data directory: %w", err))
	}

	// Download the file
	resp, err := http.Get(TestDataURL)
	if err != nil {
		panic(fmt.Errorf("failed to download test data: %w", err))
	}
	defer resp.Body.Close()

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		panic(fmt.Errorf("failed to create file: %w", err))
	}
	defer file.Close()

	// Copy the content
	if _, err := io.Copy(file, resp.Body); err != nil {
		panic(fmt.Errorf("failed to write test data: %w", err))
	}

	return filePath
}

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
		{"en(invalid)", args{lang: language.English, addr: "81x.2.69.160x"}, "", "invalid IP address"},

		{"en", args{lang: language.English, addr: "81.2.69.160"}, "London, United Kingdom", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "81.2.69.160"}, "英国，London", ""},
		{"ja", args{lang: language.Japanese, addr: "81.2.69.160"}, "イギリス、ロンドン", ""},
		{"pt-BR", args{lang: language.BrazilianPortuguese, addr: "81.2.69.160"}, "Reino Unido, Londres", ""},

		{"en", args{lang: language.English, addr: "2001:218:abcd:1234:5678:90ab:cdef:1"}, "Japan", ""},
		{"en", args{lang: language.English, addr: "[2001:218:abcd:1234:5678:90ab:cdef:1]"}, "Japan", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "2001:218:abcd:1234:5678:90ab:cdef:1"}, "日本", ""},

		{"en", args{lang: language.English, addr: "175.16.199.26"}, "Changchun, China", ""},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "175.16.199.26"}, "中国，长春", ""},
		{"ja", args{lang: language.Japanese, addr: "175.16.199.26"}, "中国、長春市", ""},
		{"ru", args{lang: language.Russian, addr: "175.16.199.26"}, "Китай, Чанчунь", ""},

		{"ko", args{lang: language.Korean, addr: "81.2.69.160"}, "London, United Kingdom", ""},
	}

	db, err := New(ensureTestData())
	require.Nil(t, err)
	defer db.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			city, err := db.GetCity(context.Background(), tt.args.addr)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			got := GeneralLocalizedCountryCity(city, tt.args.lang, language.English)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
