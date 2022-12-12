package views

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/publish"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
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
			searcher := m.Listing().Searcher
			mb := m
			m.Listing().SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
				stmt := &gorm.Statement{DB: db}
				stmt.Parse(mb.NewModel())
				tn := stmt.Schema.Table

				var pks []string
				condition := ""
				for _, f := range stmt.Schema.Fields {
					if f.Name == "DeletedAt" {
						condition = "WHERE deleted_at IS NULL"
					}
				}
				for _, f := range stmt.Schema.PrimaryFields {
					if f.Name != "Version" {
						pks = append(pks, f.DBName)
					}
				}
				pkc := strings.Join(pks, ",")
				sql := fmt.Sprintf("(%v,version) IN (SELECT %v, MAX(version) FROM %v %v GROUP BY %v)", pkc, pkc, tn, condition, pkc)

				con := presets.SQLCondition{
					Query: sql,
				}
				params.SQLConditions = append(params.SQLConditions, &con)

				return searcher(model, params, ctx)
			})

			setter := m.Editing().Setter
			m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
				if ctx.R.FormValue("id") == "" {
					version := db.NowFunc().Format("2006-01-02")
					if err := reflectutils.Set(obj, "Version.Version", fmt.Sprintf("%s-v01", version)); err != nil {
						return
					}
				}
				if setter != nil {
					setter(obj, ctx)
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
