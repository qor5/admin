package activity

import (
	"fmt"
	"log"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
)

const (
	ParamResourceKeys    = "resource_keys"
	ParamResourceComment = "comment"
	TimelinePortalName   = "activity-timeline-portal"
)

func createNote(ab *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		keys := ctx.R.FormValue(ParamResourceKeys)
		content := ctx.R.FormValue(ParamResourceComment)

		if strings.TrimSpace(content) == "" {
			presets.ShowMessage(&r, "comment cannot be blank", "error")
			return
		}

		mv := mb.NewModel()
		creator := ab.currentUserFunc(ctx.R.Context())
		activity := ActivityLog{
			CreatorID: creator.ID,
			Creator:   *creator, // TODO:
			ModelName: modelName(mv),
			ModelKeys: keys,
			Action:    ActionCreateNote,
			Comment:   content,
		}

		if err = ab.db.Save(&activity).Error; err != nil {
			handleError(err, &r, "Failed to save activity")
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		notesSection := ab.timelineList(ctx.R.Context(), mv, keys)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: TimelinePortalName,
			Body: notesSection,
		})

		return
	}
}

func updateNote(ab *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		keys := ctx.R.FormValue(ParamResourceKeys)
		mn := modelName(mb.NewModel())

		if keys == "" {
			err = fmt.Errorf("missing required parameters")
			log.Println("updateUserNoteAction error:", err)
			return
		}

		creator := ab.currentUserFunc(ctx.R.Context())

		userNote := ActivityLog{CreatorID: creator.ID, ModelName: mn, ModelKeys: keys}
		if err = ab.db.Where(userNote).FirstOrCreate(&userNote).Error; err != nil {
			log.Println("updateUserNoteAction error:", err)
			return
		}

		var total int64
		ab.db.Model(&ActivityLog{}).Where("model_name = ? AND model_keys = ?", mn, keys).Count(&total)
		userNote.Number = total

		if err = ab.db.Save(&userNote).Error; err != nil {
			log.Println("updateUserNoteAction error:", err)
			return
		}

		r.ReloadPortals = append(r.ReloadPortals, presets.NotificationCenterPortalName)
		return
	}
}

func deleteNote(ab *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		noteID := ctx.R.FormValue(presets.ParamID)
		keys := ctx.R.FormValue(ParamResourceKeys)
		// mn := modelName(mb.NewModel())

		creator := ab.currentUserFunc(ctx.R.Context())

		// Find the note by ID and delete it

		if err = ab.db.Model(&ActivityLog{}).Delete("id = ? AND creator_id = ?", noteID, creator.ID).Error; err != nil {
			presets.ShowMessage(&r, "Failed to delete note", "error")
			err = nil
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nNoteKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		notesSection := ab.timelineList(ctx.R.Context(), mb, keys)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: TimelinePortalName,
			Body: notesSection,
		})

		return
	}
}
