package note

import (
	"fmt"

	. "github.com/goplaid/ui/vuetify"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/qor/qor5/presets"
	"github.com/qor/qor5/presets/actions"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const (
	I18nNoteKey i18n.ModuleKey = "I18nNoteKey"

	createNoteEvent     = "note_CreateNoteEvent"
	updateUserNoteEvent = "note_UpdateUserNoteEvent"
)

func Configure(db *gorm.DB, pb *presets.Builder, models ...*presets.ModelBuilder) {
	if err := db.AutoMigrate(QorNote{}, UserNote{}); err != nil {
		panic(err)
	}

	for _, m := range models {
		if m.Info().HasDetailing() {
			m.Detailing().AppendTabsPanelFunc(tabsPanel(db, m))
		}
		m.Editing().AppendTabsPanelFunc(tabsPanel(db, m))
		m.RegisterEventFunc(createNoteEvent, createNoteAction(db, m))
		m.RegisterEventFunc(updateUserNoteEvent, updateUserNoteAction(db, m))
		m.Listing().Field("Notes").ComponentFunc(noteFunc(db, m))
	}

	pb.I18n().
		RegisterForModule(language.English, I18nNoteKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nNoteKey, Messages_zh_CN)
}

func tabsPanel(db *gorm.DB, mb *presets.ModelBuilder) presets.ObjectComponentFunc {
	return func(obj interface{}, ctx *web.EventContext) (c h.HTMLComponent) {
		id := ctx.R.FormValue(presets.ParamID)
		if len(id) == 0 {
			return
		}

		tn := mb.Info().Label()

		notesSection := getNotesTab(ctx, db, tn, id)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)

		userID, _ := getUserData(ctx)
		count := GetUnreadNotesCount(db, userID, tn, id)

		notesTab := VBadge(h.Text(msgr.Notes)).
			Attr(":content", "locals.unreadNotesCount").
			Attr(":value", "locals.unreadNotesCount").
			Color("red")

		clickEvent := web.Plaid().
			BeforeScript("locals.unreadNotesCount=0").
			EventFunc(updateUserNoteEvent).
			Query("resource_id", id).
			Query("resource_type", tn).
			Go() + ";" + web.Plaid().
			EventFunc(actions.ReloadList).
			Go()
		if count == 0 {
			clickEvent = ""
		}
		c = h.Components(
			VTab(notesTab).
				Attr(web.InitContextLocals, fmt.Sprintf("{unreadNotesCount:%v}", count)).
				Attr("@click", clickEvent),
			VTabItem(web.Portal().Name("notes-section").Children(notesSection)),
		)

		return
	}
}

func noteFunc(db *gorm.DB, mb *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (c h.HTMLComponent) {
		tn := mb.Info().Label()

		id := fmt.Sprint(reflectutils.MustGet(obj, "ID"))
		if ps, ok := obj.(interface {
			PrimarySlug() string
		}); ok {
			id = ps.PrimarySlug()
		}

		latestNote := QorNote{}
		db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", tn, id).Order("created_at DESC").First(&latestNote)

		content := []rune(latestNote.Content)
		result := string(content[:])
		if len(content) > 60 {
			result = string(content[0:60]) + "..."
		}
		userID, _ := getUserData(ctx)
		count := GetUnreadNotesCount(db, userID, tn, id)
		return h.Td(
			h.If(count > 0,
				VBadge(h.Text(result)).Content(count).Color("red"),
			).Else(
				h.Text(fmt.Sprint(result)),
			),
		)
	}
}
