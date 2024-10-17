package examples_admin

import (
	"net/http"
	"strings"

	"github.com/qor5/admin/v3/autosync"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func AutoSyncExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return autoSyncExample(b, db, nil)
}

func autoSyncExample(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder)) http.Handler {
	b.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)

	b.DataOperator(gorm2op.DataOperator(db))

	type WithSlugProduct struct {
		ID          uint
		Title       string
		TitleSlug   string
		Description string
	}

	err := db.AutoMigrate(&WithSlugProduct{})
	if err != nil {
		panic(err)
	}

	mb := b.Model(&WithSlugProduct{})

	lazyWrapperEditCompoSync := autosync.NewLazyWrapComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) *autosync.Config {
		return &autosync.Config{
			SyncFromFromKey: strings.TrimSuffix(field.FormKey, "Slug"),
			InitialChecked:  autosync.InitialCheckedAuto,
			CheckboxLabel:   "Auto Sync",
			SyncCall:        autosync.SyncCallSlug,
		}
	})

	mb.Editing().Field("TitleSlug").LazyWrapComponentFunc(lazyWrapperEditCompoSync)
	dp := mb.Detailing("Detail").Drawer(true)
	detailSection := presets.NewSectionBuilder(mb, "Detail").Editing("Title", "TitleSlug", "Description")
	detailSection.EditingField("TitleSlug").LazyWrapComponentFunc(lazyWrapperEditCompoSync)
	dp.Section(detailSection)

	if customize != nil {
		customize(mb)
	}
	return b
}
