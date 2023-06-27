package note

import (
	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"gorm.io/gorm"
)

func createNoteAction(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")

		content := ctx.R.FormValue("Content")

		userID, creator := GetUserData(ctx)
		note := QorNote{
			UserID:       userID,
			Creator:      creator,
			ResourceID:   ri,
			ResourceType: rt,
			Content:      content,
		}

		if err = db.Save(&note).Error; err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			err = nil
			return
		}

		userNote := UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
		db.Where(userNote).FirstOrCreate(&userNote)

		var total int64
		db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", rt, ri).Count(&total)
		db.Model(&userNote).UpdateColumn("Number", total)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		notesSection := getNotesTab(ctx, db, rt, ri)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "notes-section",
			Body: notesSection,
		})

		if err = AfterCreateFunc(db); err != nil {
			return
		}

		return
	}
}

func updateUserNoteAction(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")

		userID, _ := GetUserData(ctx)
		if userID == 0 {
			return
		}

		userNote := UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
		if err = db.Where(userNote).FirstOrCreate(&userNote).Error; err != nil {
			return
		}

		var total int64
		db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", rt, ri).Count(&total)
		userNote.Number = total

		if err = db.Save(&userNote).Error; err != nil {
			return
		}

		if err = AfterCreateFunc(db); err != nil {
			return
		}

		// notify notification center after note read. if notification center is not enabled, this one would just do nothing
		r.ReloadPortals = append(r.ReloadPortals, presets.NotificationCenterPortalName)

		return
	}
}
