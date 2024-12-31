package location

import (
	"context"
	"net"
	"strings"

	"github.com/pkg/errors"

	"github.com/oschwald/geoip2-golang"
	"github.com/samber/lo"
	"golang.org/x/text/language"
)

var ErrInvalidIP = errors.New("invalid IP address")

type GEOIP2 struct {
	*geoip2.Reader
}

func New(file string) (*GEOIP2, error) {
	db, err := geoip2.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GEOIP2")
	}
	return &GEOIP2{
		Reader: db,
	}, nil
}

func NewFromBytes(data []byte) (*GEOIP2, error) {
	db, err := geoip2.FromBytes(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GEOIP2 from bytes")
	}
	return &GEOIP2{
		Reader: db,
	}, nil
}

// https://dev.maxmind.com/geoip/docs/databases/enterprise/#locations-files
// locales: “de”, “en”, “es”, “fr”, “ja”, “pt-BR”, “ru”, “zh-CN”
func LocaleFromLanguage(t language.Tag) string {
	s := t.String()
	switch s {
	case "zh-Hant", "zh-Hans":
		return "zh-CN"
	default:
		return s
	}
}

func (g *GEOIP2) GetLocation(_ context.Context, lang language.Tag, addr string) (string, error) {
	addr = strings.Trim(addr, "[]")
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", ErrInvalidIP
	}

	record, err := g.City(ip)
	if err != nil {
		return "", err
	}

	locale := LocaleFromLanguage(lang)

	sep := ", "
	switch lang {
	case language.Japanese:
		sep = "、"
	case language.SimplifiedChinese, language.TraditionalChinese:
		sep = "，"
	}

	return strings.Join(
		lo.Filter([]string{
			strings.TrimSpace(record.City.Names[locale]),
			strings.TrimSpace(record.Country.Names[locale]),
		}, func(item string, _ int) bool {
			return item != ""
		}),
		sep,
	), nil
}
