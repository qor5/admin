package examples_admin

// @snippet_begin(L10nFullExample)
import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"gorm.io/gorm"
)

// @snippet_begin(L10nModelExample)
type L10nModel struct {
	gorm.Model
	Title string

	l10n.Locale
}

func (lm *L10nModel) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", lm.ID, lm.LocaleCode)
}

func (lm *L10nModel) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"id":                segs[0],
		l10n.SlugLocaleCode: segs[1],
	}
}

// @snippet_end

func LocalizationExample(b *presets.Builder, db *gorm.DB) http.Handler {
	if err := db.AutoMigrate(&L10nModel{}); err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))

	// @snippet_begin(L10nBuilderExample)
	l10nBuilder := l10n.New(db)
	l10nBuilder.
		// RegisterLocales("International", "international", "International", l10n.InternationalSvg).
		RegisterLocales("Japan", "jp", "Japan", l10n.JapanSvg).
		RegisterLocales("China", "cn", "China", l10n.ChinaSvg).
		SupportLocalesFunc(func(R *http.Request) []string {
			return l10nBuilder.GetSupportLocaleCodes()[:]
		})
	// @snippet_end

	// @snippet_begin(L10nConfigureExample)
	mb := b.Model(&L10nModel{}).URIName("l10n-models")
	b.Use(l10nBuilder)
	mb.Use(l10nBuilder)
	mb.Listing("ID", "Title", "Locale")
	// @snippet_end
	// @snippet_end
	return l10nBuilder.EnsureLocale(b)
}
