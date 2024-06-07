package note

import (
	"fmt"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	I18nNoteKey         i18n.ModuleKey = "I18nNoteKey"
	createNoteEvent                    = "note_CreateNoteEvent"
	updateUserNoteEvent                = "note_UpdateUserNoteEvent"
)

func tabsPanel(db *gorm.DB, mb *presets.ModelBuilder) presets.TabComponentFunc {
	return func(obj any, ctx *web.EventContext) (h.HTMLComponent, h.HTMLComponent) {
		id := ctx.Param(presets.ParamID)
		if id == "" {
			return nil, nil
		}

		tn := mb.Info().Label()
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		notesSection := getNotesTab(ctx, db, tn, id)
		count, _ := GetUnreadNotesCount(db, ctx.R.Context().Value(UserIDKey).(uint), tn, id)

		tab := VTab(
			VBadge(h.Text(msgr.Notes)).Attr(":content", "locals.unreadNotesCount").Attr(":value", "locals.unreadNotesCount").Color("red"),
		).Attr(web.VAssign("locals", fmt.Sprintf("{unreadNotesCount:%v}", count))...).Attr("@click", getClickEvent(count, id, tn))
		content := VTabsWindowItem(web.Portal().Name("notes-section").Children(notesSection)).Value("notesTab")

		return tab, content
	}
}

func getClickEvent(count int64, id, tn string) string {
	if count == 0 {
		return ""
	}
	return web.Plaid().
		BeforeScript("locals.unreadNotesCount=0").
		EventFunc(updateUserNoteEvent).
		Query("resource_id", id).
		Query("resource_type", tn).
		Go() + ";" + web.Plaid().
		EventFunc(actions.ReloadList).
		Go()
}

func noteFunc(db *gorm.DB, mb *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		tn := mb.Info().Label()
		id := fmt.Sprint(reflectutils.MustGet(obj, "ID"))
		if ps, ok := obj.(interface{ PrimarySlug() string }); ok {
			id = ps.PrimarySlug()
		}

		latestNote := QorNote{}
		db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", tn, id).Order("created_at DESC").First(&latestNote)

		content := latestNote.Content
		if len(content) > 60 {
			content = content[:60] + "..."
		}
		count, _ := GetUnreadNotesCount(db, ctx.R.Context().Value(UserIDKey).(uint), tn, id)
		if count > 0 {
			return h.Td(VBadge(h.Text(content)).Content(count).Color("red"))
		}
		return h.Td(h.Text(content))
	}
}
