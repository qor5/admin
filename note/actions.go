package note

import (
	"net/url"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	"gorm.io/gorm"
)

func createNoteAction(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")
		content := ctx.R.FormValue("Content")

		userID, creator := getUserData(ctx)
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

		var pageURL *url.URL
		if pageURL, err = url.Parse(ctx.R.Referer()); err == nil {
			mb.Listing().ReloadList(ctx, &r, pageURL)
		}

		return
	}
}

func updateUserNoteAction(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")

		userID, _ := getUserData(ctx)
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

		var pageURL *url.URL
		if pageURL, err = url.Parse(ctx.R.Referer()); err == nil {
			mb.Listing().ReloadList(ctx, &r, pageURL)
		}

		return
	}
}
