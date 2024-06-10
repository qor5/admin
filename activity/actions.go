package activity

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"gorm.io/gorm"
)

const (
	I18nNoteKey i18n.ModuleKey = "I18nNoteKey"
)

func createNoteAction(b *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (web.EventResponse, error) {
		return handleNoteAction(ctx, b, true)
	}
}

func updateUserNoteAction(b *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (web.EventResponse, error) {
		return handleNoteAction(ctx, b, false)
	}
}

func handleNoteAction(ctx *web.EventContext, b *Builder, isCreate bool) (web.EventResponse, error) {
	db := b.db
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

	var r web.EventResponse
	if isCreate {
		if err := db.Save(&note).Error; err != nil {
			_ = presets.ShowMessage(&r, err.Error(), "error")
			return r, err
		}
	} else {
		userNote := UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
		if err := db.Where(userNote).FirstOrCreate(&userNote).Error; err != nil {
			return r, err
		}
	}

	updateUserNotes(db, rt, ri, userID)

	msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
	if isCreate {
		_ = presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")
		return r, nil
	}
	r.ReloadPortals = append(r.ReloadPortals, presets.NotificationCenterPortalName)
	return r, nil
}

func updateUserNotes(db *gorm.DB, rt, ri string, userID uint) {
	var total int64
	db.Model(&QorNote{}).Where("resource_type = ? AND resource_id = ?", rt, ri).Count(&total)
	userNote := UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
	db.Model(&userNote).UpdateColumn("Number", total)
}
