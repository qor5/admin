package views

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nPublishKey i18n.ModuleKey = "I18nPublishKey"

func Configure(b *presets.Builder, db *gorm.DB, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		if _, ok := m.NewModel().(publish.VersionInterface); ok {
			m.Editing().SidePanelFunc(sidePanel(db, m)).ActionsFunc(versionActionsFunc(m))
			m.Listing().Searcher(searcher(db, m))

			m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
				if ctx.R.FormValue("id") == "" {
					version := db.NowFunc().Format("2006-01-02")
					if err := reflectutils.Set(obj, "Version.Version", fmt.Sprintf("%s-v01", version)); err != nil {
						return
					}
				}
			})

			m.Listing().Field("Draft Count").ComponentFunc(draftCountFunc(db))
			m.Listing().Field("Online").ComponentFunc(onlineFunc(db))
		}

		registerEventFuncs(db, m, publisher)
	}

	b.FieldDefaults(presets.LIST).
		FieldType(publish.Status{}).
		ComponentFunc(StatusListFunc())
	b.FieldDefaults(presets.WRITE).
		FieldType(publish.Status{}).
		ComponentFunc(StatusEditFunc()).
		SetterFunc(StatusEditSetterFunc)

	b.FieldDefaults(presets.WRITE).
		FieldType(publish.Schedule{}).
		ComponentFunc(ScheduleEditFunc()).
		SetterFunc(ScheduleEditSetterFunc)

	b.I18n().
		RegisterForModule(language.English, I18nPublishKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPublishKey, Messages_zh_CN)
}
