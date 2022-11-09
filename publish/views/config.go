package views

import (
	"fmt"
	"reflect"

	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/gorm2op"
	"github.com/qor5/admin/publish"
	"github.com/sunfmin/reflectutils"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nPublishKey i18n.ModuleKey = "I18nPublishKey"

func Configure(b *presets.Builder, db *gorm.DB, ab *activity.ActivityBuilder, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		if model, ok := m.NewModel().(publish.VersionInterface); ok {
			if schedulePublishModel, ok := model.(publish.ScheduleInterface); ok {
				publish.VersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
			}

			m.Editing().SidePanelFunc(sidePanel(db, m)).ActionsFunc(versionActionsFunc(m))
			m.Listing().SearchFunc(gorm2op.Searcher(db, m))
			m.Editing().FetchFunc(gorm2op.Fetcher(db, m))
			m.Editing().SaveFunc(gorm2op.Saver(db, m))
			m.Editing().DeleteFunc(gorm2op.Deleter(db, m))

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
		} else {
			if schedulePublishModel, ok := m.NewModel().(publish.ScheduleInterface); ok {
				publish.NonVersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
			}
		}

		if model, ok := m.NewModel().(publish.ListInterface); ok {
			if schedulePublishModel, ok := model.(publish.ScheduleInterface); ok {
				publish.ListPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
			}
		}

		registerEventFuncs(db, m, publisher, ab)
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
