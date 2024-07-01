package activity

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	I18nNoteKey i18n.ModuleKey = "I18nNoteKey"

	createNoteEvent     = "note_CreateNoteEvent"
	updateUserNoteEvent = "note_UpdateUserNoteEvent"
	deleteNoteEvent     = "note_DeleteNoteEvent"
)

func getNotesTab(ctx *web.EventContext, db *gorm.DB, resourceType string, resourceId string) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)

	c := h.Div(
		web.Scope(
			VCardText(
				h.Text(msgr.NewNote),
				VRow(VCol(VTextField().Attr(web.VField("Content", "")...).Clearable(true))),
			),
			VCardActions(h.Components(
				VSpacer(),
				VBtn(presets.MustGetMessages(ctx.R).Create).
					Color("primary").
					Attr("@click", web.Plaid().
						EventFunc(createNoteEvent).
						Query("resource_id", resourceId).
						Query("resource_type", resourceType).
						Go()+";"+web.Plaid().
						EventFunc(actions.ReloadList).
						Go(),
					),
			)),
		).VSlot("{form}"),
	)

	var notes []ActivityLog
	db.Where("resource_type = ? and resource_id = ?", resourceType, resourceId).
		Order("id DESC").Find(&notes)

	var panels []h.HTMLComponent
	for _, note := range notes {
		panels = append(panels, VCard(
			VCardText(
				h.Div(h.Text(fmt.Sprintf("%v - %v", note.Creator, note.CreatedAt.Format("2006-01-02 15:04:05 MST")))).
					Class("text-h6"),
				h.Text(note.Content)),
		))
	}
	c.AppendChildren(panels...).Class("p-2")
	return c
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

		latestNote := ActivityLog{}
		db.Model(&ActivityLog{}).Where("resource_type = ? AND resource_id = ?", tn, id).Order("created_at DESC").First(&latestNote)

		content := []rune(latestNote.Content)
		result := string(content[:])
		if len(content) > 60 {
			result = string(content[0:60]) + "..."
		}
		userID, _ := GetUserData(ctx)

		count, err := GetUnreadNotesCount(db, userID, tn, id)
		if err != nil {
			panic(err)
		}

		return h.Td(
			h.If(count > 0,
				VBadge(h.Text(result)).Content(count).Color("red"),
			).Else(
				h.Text(fmt.Sprint(result)),
			),
		)
	}
}
