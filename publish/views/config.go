package views

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nPublishKey i18n.ModuleKey = "I18nPublishKey"

func Configure(b *presets.Builder, db *gorm.DB, ab *activity.ActivityBuilder, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		obj := m.NewModel()
		_ = obj.(presets.SlugEncoder)
		_ = obj.(presets.SlugDecoder)
		if model, ok := obj.(publish.VersionInterface); ok {
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

			// listing-delete deletes all versions
			{
				// rewrite Delete row menu item to show correct id in prompt message
				m.Listing().RowMenu().RowMenuItem("Delete").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
					msgr := presets.MustGetMessages(ctx.R)
					if m.Info().Verifier().Do(presets.PermDelete).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
						return nil
					}

					promptID := id
					if slugger, ok := obj.(presets.SlugDecoder); ok {
						fvs := []string{}
						for f, v := range slugger.PrimaryColumnValuesBySlug(id) {
							if f == "id" {
								fvs = append([]string{v}, fvs...)
							} else {
								if f != "version" {
									fvs = append(fvs, v)
								}
							}
						}
						promptID = strings.Join(fvs, "_")
					}

					onclick := web.Plaid().
						EventFunc(actions.DeleteConfirmation).
						Query(presets.ParamID, id).
						Query("all_versions", true).
						Query("prompt_id", promptID)
					if presets.IsInDialog(ctx.R.Context()) {
						onclick.URL(ctx.R.RequestURI).
							Query(presets.ParamOverlay, actions.Dialog).
							Query(presets.ParamInDialog, true).
							Query(presets.ParamListingQueries, ctx.Queries().Encode())
					}
					return VListItem(
						web.Slot(
							VIcon("mdi-delete"),
						).Name("prepend"),
						VListItemTitle(h.Text(msgr.Delete)),
					).Attr("@click", onclick.Go())
				})
				// rewrite Deleter to ignore version condition
				m.Editing().DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
					allVersions := ctx.R.URL.Query().Get("all_versions") == "true"

					wh := db.Model(obj)

					if id != "" {
						if slugger, ok := obj.(presets.SlugDecoder); ok {
							cs := slugger.PrimaryColumnValuesBySlug(id)
							for key, value := range cs {
								if allVersions && key == "version" {
									continue
								}
								wh = wh.Where(fmt.Sprintf("%s = ?", key), value)
							}
						} else {
							wh = wh.Where("id =  ?", id)
						}
					}

					return wh.Delete(obj).Error
				})
			}

			setter := m.Editing().Setter
			m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
				if ctx.R.FormValue(presets.ParamID) == "" {
					version := fmt.Sprintf("%s-v01", db.NowFunc().Format("2006-01-02"))
					if err := reflectutils.Set(obj, "Version.Version", version); err != nil {
						return
					}
					if err := reflectutils.Set(obj, "Version.VersionName", version); err != nil {
						return
					}
				}
				if setter != nil {
					setter(obj, ctx)
				}
			})

			m.Listing().Field("Draft Count").ComponentFunc(draftCountFunc(db))
			m.Listing().Field("Online").ComponentFunc(onlineFunc(db))
			if m.Editing().GetField("StatusBar") != nil {
				m.Editing().Field("StatusBar").ComponentFunc(StatusEditFunc())
			}
			// Version V2
			{
				m.Detailing().Field("defaultVersion").ComponentFunc(DefaultVersionComponentFunc(m))
				m.Editing().Field("defaultVersion").ComponentFunc(DefaultVersionComponentFunc(m))
				ConfigureVersionListDialog(db, b, m)
			}
		} else {
			if schedulePublishModel, ok := obj.(publish.ScheduleInterface); ok {
				publish.NonVersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
			}
		}

		if _, ok := obj.(publish.ScheduleInterface); ok {
			if m.Editing().GetField("ScheduleBar") != nil {
				m.Editing().Field("ScheduleBar").ComponentFunc(ScheduleEditFunc()).SetterFunc(ScheduleEditSetterFunc)
			}
		}

		if model, ok := obj.(publish.ListInterface); ok {
			if schedulePublishModel, ok := model.(publish.ScheduleInterface); ok {
				publish.ListPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
			}
		}

		registerEventFuncs(db, m, publisher, ab)
	}

	b.FieldDefaults(presets.LIST).
		FieldType(publish.Status{}).
		ComponentFunc(StatusListFunc())

	b.I18n().
		RegisterForModule(language.English, I18nPublishKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPublishKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPublishKey, Messages_ja_JP)
}
