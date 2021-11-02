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
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nPublishKey i18n.ModuleKey = "I18nPublishKey"

func Configure(b *presets.Builder, db *gorm.DB, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		m.RightDrawerWidth("1200")
		m.Editing().SidePanelFunc(sidePanel(db, m)).ActionsFunc(func(ctx *web.EventContext) h.HTMLComponent {
			gmsgr := presets.MustGetMessages(ctx.R)
			var buttonLabel = gmsgr.Create
			if ctx.Event.Params[0] != "" {
				buttonLabel = gmsgr.Update
			}

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

			return h.Components(
				VSpacer(),
				VBtn(msgr.SaveAsNewVersion).
					Color("secondary").
					Attr("@click", web.Plaid().
						EventFunc(saveNewVersionEvent, ctx.Event.Params...).
						URL(m.Info().ListingHref()).
						Go(),
					).Disabled(ctx.Event.Params[0] == ""),
				VBtn(buttonLabel).
					Color("primary").
					Attr("@click", web.Plaid().
						EventFunc(actions.Update, ctx.Event.Params...).
						URL(m.Info().ListingHref()).
						Go(),
					),
			)
		})
		m.Listing().Searcher(searcher(db, m))

		m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			if ctx.Event.Params[0] == "" {
				version := db.NowFunc().Format("2006-01-02")
				if err := reflectutils.Set(obj, "Version.Version", fmt.Sprintf("%s-v01", version)); err != nil {
					return
				}
			}

			return
		})

		registerEventFuncs(db, m, publisher)
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
