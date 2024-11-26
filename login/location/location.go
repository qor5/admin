package location

import (
	_ "embed"
	"log"
	"net"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/samber/lo"
	"golang.org/x/text/language"
)

//go:embed assets/GeoLite2-City.mmdb
var geoIPData []byte

var db *geoip2.Reader

func init() {
	var err error
	db, err = geoip2.FromBytes(geoIPData)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: close db
	// defer db.Close()
}

// https://dev.maxmind.com/geoip/docs/databases/enterprise/#locations-files
// locales: “de”, “en”, “es”, “fr”, “ja”, “pt-BR”, “ru”, “zh-CN”
func LanguageTagToGeoIP2Locale(t language.Tag) string {
	if t.String() == "zh-Hans" {
		return "zh-CN"
	}
	return t.String()
}

func GetLocation(lang language.Tag, addr string) (string, error) {
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(addr)
	record, err := db.City(ip)
	if err != nil {
		return "", err
	}

	return strings.Join(
		lo.Filter([]string{record.City.Names[LanguageTagToGeoIP2Locale(lang)], record.Country.Names[LanguageTagToGeoIP2Locale(lang)]}, func(item string, _ int) bool {
			return item != ""
		}),
		map[language.Tag]string{
			language.English:           ", ",
			language.Japanese:          "、",
			language.SimplifiedChinese: "，",
		}[lang],
	), nil
}
