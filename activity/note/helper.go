package note

import (
	"errors"
	"fmt"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/ui/v3/vuetify"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type contextUserIDKey int

const (
	UserIDKey contextUserIDKey = iota
	UserKey
)

func GetUserData(ctx *web.EventContext) (userID uint, creator string) {
	if v := ctx.R.Context().Value(UserIDKey); v != nil {
		userID = v.(uint)
	}
	if v := ctx.R.Context().Value(UserKey); v != nil {
		creator = v.(string)
	}
	return
}

func getNotesTab(ctx *web.EventContext, db *gorm.DB, resourceType, resourceId string) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)

	// building form components
	contentField := VTextField().
		Label(msgr.NewNote).
		Clearable(true).
		Attr("v-model", "content")
	form := VCardText(
		VRow(
			VCol(contentField),
		),
	)

	// build actions button
	createAction := web.Plaid().
		EventFunc(createNoteEvent).
		Query("resource_id", resourceId).
		Query("resource_type", resourceType).
		Go()

	reloadAction := web.Plaid().
		EventFunc(actions.ReloadList).
		Go()

	vActions := VCardActions(
		VSpacer(),
		VBtn(presets.MustGetMessages(ctx.R).Create).
			Color("primary").
			Attr("@click", createAction+";"+reloadAction),
	)

	// assemble top container
	container := h.Div(
		web.Scope(form, vActions).VSlot("{form}"),
	).Class("p-2")

	// from database query all note
	var notes []QorNote
	db.Where("resource_type = ? AND resource_id = ?", resourceType, resourceId).
		Order("id DESC").
		Find(&notes)

	// any note support
	for _, note := range notes {
		noteCard := vuetify.VCard(
			vuetify.VCardText(
				h.Div(h.Text(fmt.Sprintf("%v - %v", note.Creator, note.CreatedAt.Format("2006-01-02 15:04:05 MST")))).Class("text-h6"),
				h.Text(note.Content),
			),
		)
		container.AppendChildren(noteCard)
	}

	return container
}

func GetUnreadNotesCount(db *gorm.DB, userID uint, resourceType, resourceID string) (int64, error) {
	var total int64
	if err := db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Count(&total).Error; err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, nil
	}

	var userNote UserNote
	if err := db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", userID, resourceType, resourceID).First(&userNote).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	return total - userNote.Number, nil
}
