package views

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/jsontyperegistry"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nPublishKey i18n.ModuleKey = "I18nPublishKey"

func Configure(b *presets.Builder, db *gorm.DB, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		m.Editing().SidePanelFunc(sidePanel(db, m)).ActionsFunc(func(ctx *web.EventContext) h.HTMLComponent {
			gmsgr := presets.MustGetMessages(ctx.R)
			var buttonLabel = gmsgr.Create
			m.RightDrawerWidth("800")
			if ctx.R.FormValue("id") != "" {
				buttonLabel = gmsgr.Update
				m.RightDrawerWidth("1200")
			}

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

			return h.Components(
				VSpacer(),
				VBtn(msgr.SaveAsNewVersion).
					Color("secondary").
					Attr("@click", web.Plaid().
						EventFunc(saveNewVersionEvent).Query("id", ctx.R.FormValue("id")).
						URL(m.Info().ListingHref()).
						Go(),
					).Disabled(ctx.R.FormValue("id") == ""),
				VBtn(buttonLabel).
					Color("primary").
					Attr("@click", web.Plaid().
						EventFunc(actions.Update).Query("id", ctx.R.FormValue("id")).
						URL(m.Info().ListingHref()).
						Go(),
					),
			)
		})
		m.Listing().Searcher(searcher(db, m))

		m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			if ctx.R.FormValue("id") == "" {
				version := db.NowFunc().Format("2006-01-02")
				if err := reflectutils.Set(obj, "Version.Version", fmt.Sprintf("%s-v01", version)); err != nil {
					return
				}
			}
		})

		registerEventFuncs(db, m, publisher)

		jsontyperegistry.MustRegisterType(m.NewModel())

		m.Listing("Draft Count", "Online")
		m.Listing().Field("Draft Count").ComponentFunc(draftCountFunc(db))
		m.Listing().Field("Online").ComponentFunc(onlineFunc(db))
	}

	b.FieldDefaults(presets.LIST).
		FieldType(publish.Status{}).
		ComponentFunc(StatusListFunc())
	b.FieldDefaults(presets.WRITE).
		FieldType(publish.Status{}).
		ComponentFunc(StatusEditFunc()).
		SetterFunc(ScheduleEditSetterFunc)

	b.FieldDefaults(presets.WRITE).
		FieldType(publish.Schedule{}).
		ComponentFunc(ScheduleEditFunc()).
		SetterFunc(ScheduleEditSetterFunc)

	b.I18n().
		RegisterForModule(language.English, I18nPublishKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPublishKey, Messages_zh_CN)
}
