package activity

import (
	"fmt"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"log"
)

func createNoteAction(b *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return b.wrapper.Wrap(func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := b.db
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")
		content := ctx.R.FormValue("Content")

		userID, creator := GetUserData(ctx)
		activity := ActivityLog{
			UserID:    userID,
			Creator:   creator,
			ModelName: rt,
			ModelKeys: ri,
			Action:    "create_note",
			Content:   content,
		}

		if err = db.Save(&activity).Error; err != nil {
			handleError(err, &r, "Failed to save activity")
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		notesSection := getNotesTab(ctx, db, rt, ri)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "notes-section",
			Body: notesSection,
		})

		return
	})
}

func updateUserNoteAction(b *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return b.wrapper.Wrap(func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := b.db
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")

		if ri == "" || rt == "" {
			err = fmt.Errorf("missing required parameters")
			log.Println("updateUserNoteAction error:", err)
			return
		}

		userID, _ := GetUserData(ctx)
		if userID == 0 {
			err = fmt.Errorf("user not authenticated")
			log.Println("updateUserNoteAction error:", err)
			return
		}

		userNote := ActivityLog{UserID: userID, ModelName: rt, ModelKeys: ri}
		if err = db.Where(userNote).FirstOrCreate(&userNote).Error; err != nil {
			log.Println("updateUserNoteAction error:", err)
			return
		}

		var total int64
		db.Model(&ActivityLog{}).Where("model_name = ? AND model_keys = ?", rt, ri).Count(&total)
		userNote.Number = total
		userNote.Action = fmt.Sprintf("update_note: %d", total)

		if err = db.Save(&userNote).Error; err != nil {
			log.Println("updateUserNoteAction error:", err)
			return
		}

		r.ReloadPortals = append(r.ReloadPortals, presets.NotificationCenterPortalName)
		return
	})
}

func deleteNoteAction(b *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return b.wrapper.Wrap(func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := b.db
		noteID := ctx.R.FormValue("note_id")
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")

		userID, _ := GetUserData(ctx)
		if userID == 0 {
			return
		}

		// Find the note by ID and delete it
		var note ActivityLog
		if err = db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
			presets.ShowMessage(&r, "Note not found or access denied", "error")
			err = nil
			return
		}

		if err = db.Delete(&note).Error; err != nil {
			presets.ShowMessage(&r, "Failed to delete note", "error")
			err = nil
			return
		}

		// Update user note count
		userNote := ActivityLog{UserID: userID, ModelName: rt, ModelKeys: ri}
		if err = db.Where(userNote).First(&userNote).Error; err != nil {
			return
		}

		var total int64
		db.Model(&ActivityLog{}).Where("model_name = ? AND model_keys = ?", rt, ri).Count(&total)
		userNote.Number = total
		userNote.Action = fmt.Sprintf("delete_note: %d", total)

		if err = db.Save(&userNote).Error; err != nil {
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		notesSection := getNotesTab(ctx, db, rt, ri)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "notes-section",
			Body: notesSection,
		})

		return
	})
}
