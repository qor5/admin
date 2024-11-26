package location

import (
	"testing"

	"golang.org/x/text/language"
)

func TestLanguageTagToGeoIP2Locale(t *testing.T) {
	type args struct {
		t language.Tag
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"de", args{t: language.German}, "de"},
		{"en", args{t: language.English}, "en"},
		{"es", args{t: language.Spanish}, "es"},
		{"fr", args{t: language.French}, "fr"},
		{"ja", args{t: language.Japanese}, "ja"},
		{"pt-BR", args{t: language.BrazilianPortuguese}, "pt-BR"},
		{"ru", args{t: language.Russian}, "ru"},
		{"zh-CN", args{t: language.SimplifiedChinese}, "zh-CN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LanguageTagToGeoIP2Locale(tt.args.t); got != tt.want {
				t.Errorf("LanguageTagToGeoIP2Locale() = %v, want %v", got, tt.want)
			}
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
		wantErr bool
	}{
		{"en", args{lang: language.English, addr: "81.2.69.160"}, "Bickley, United Kingdom", false},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "81.2.69.160"}, "英国", false},
		{"ja", args{lang: language.Japanese, addr: "81.2.69.160"}, "英国", false},

		{"en", args{lang: language.English, addr: "35.77.239.1"}, "Tokyo, Japan", false},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "35.77.239.1"}, "东京，日本", false},
		{"ja", args{lang: language.Japanese, addr: "35.77.239.1"}, "東京、日本", false},

		{"en", args{lang: language.English, addr: "60.176.242.26"}, "Hangzhou, China", false},
		{"zh-CN", args{lang: language.SimplifiedChinese, addr: "60.176.242.26"}, "杭州，中国", false},
		{"ja", args{lang: language.Japanese, addr: "60.176.242.26"}, "杭州市、中国", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocation(tt.args.lang, tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLocation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
