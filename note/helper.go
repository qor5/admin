package note

import (
	"fmt"

	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/ui/vuetify"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type contextUserIDKey int

const (
	UserIDKey contextUserIDKey = iota
	UserKey
)

func GetUserData(ctx *web.EventContext) (userID uint, creator string) {
	if ctx.R.Context().Value(UserIDKey) != nil {
		userID = ctx.R.Context().Value(UserIDKey).(uint)
	}
	if ctx.R.Context().Value(UserKey) != nil {
		creator = ctx.R.Context().Value(UserKey).(string)
	}

	return
}

func getNotesTab(ctx *web.EventContext, db *gorm.DB, resourceType string, resourceId string) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)

	c := h.Div(
		web.Scope(
			VCardText(
				h.Text(msgr.NewNote),
				VRow(VCol(VTextField().Attr(web.VFieldName("Content")...).Clearable(true))),
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
		).VSlot("{plaidForm}"),
	)

	var notes []QorNote
	db.Where("resource_type = ? and resource_id = ?", resourceType, resourceId).
		Order("id DESC").Find(&notes)

	var panels []h.HTMLComponent
	for _, note := range notes {
		panels = append(panels, vuetify.VExpansionPanel(
			vuetify.VExpansionPanelHeader(h.Span(fmt.Sprintf("%v - %v", note.Creator, note.CreatedAt.Format("2006-01-02 15:04:05 MST")))),
			vuetify.VExpansionPanelContent(h.Text(note.Content)),
		))
	}
	c.AppendChildren(vuetify.VExpansionPanels(panels...).Attr("style", "padding:10px;"))
	return c
}

var AfterCreateFunc = func(db *gorm.DB) (err error) {
	return
}
