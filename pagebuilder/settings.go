package pagebuilder

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/qor5/admin/note"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/publish"
	pv "github.com/qor5/admin/publish/views"
	"github.com/qor5/admin/utils"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
)

func settings(db *gorm.DB, pm *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		p := obj.(*Page)
		c := &Category{}
		db.First(c, "id = ?", p.CategoryID)

		overview := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Title).ZeroLabel("No Title")).Label("Title"),
				vx.DetailField(vx.OptionalText(c.Path).ZeroLabel("No Category")).Label("Category"),
			),
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Slug).ZeroLabel("No Slug")).Label("Slug"),
			),
		)
		var start, end, se string
		if p.GetScheduledStartAt() != nil {
			start = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if p.GetScheduledEndAt() != nil {
			end = p.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		if start != "" || end != "" {
			se = start + " ~ " + end
		}
		pageState := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.GetStatus()).ZeroLabel("No State")).Label("State"),
				vx.DetailField(vx.OptionalText(se).ZeroLabel("No Set")).Label("SchedulePublishTime"),
			),
		)
		var notes []note.QorNote
		db.Where("resource_type = ? and resource_id = ?", pm.Info().Label(), p.PrimarySlug()).
			Order("id DESC").Find(&notes)

		var notesSetcion h.HTMLComponent
		if len(notes) > 0 {
			s := VContainer()
			for _, note := range notes {
				s.AppendChildren(VRow(h.Text(note.Content)).Class("ma-1"))
				s.AppendChildren(VRow(h.Text(fmt.Sprintf("%v - %v", note.Creator, note.CreatedAt.Format("2006-01-02 15:04:05 MST")))).Class("ma-1 mb-2"))
			}
			notesSetcion = s
		}
		var editBtn h.HTMLComponent
		var pageStateBtn h.HTMLComponent
		pvMsgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, utils.Messages_en_US).(*pv.Messages)
		if p.GetStatus() == publish.StatusDraft {
			editBtn = VBtn("Edit").Depressed(true).
				Attr("@click", web.POST().
					EventFunc(actions.Edit).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, p.PrimarySlug()).
					URL(mi.PresetsPrefix()+"/pages").Go(),
				)
		}
		if p.GetStatus() == publish.StatusOnline {
			pageStateBtn = VBtn(pvMsgr.Unpublish).Depressed(true).Class("mr-2").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.UnpublishEvent))
		} else {
			pageStateBtn = VBtn("Schedule Publish").Depressed(true).
				Attr("@click", web.POST().EventFunc(schedulePublishDialogEvent).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, p.PrimarySlug()).
					URL(mi.PresetsPrefix()+"/pages").Go(),
				)
		}
		return VContainer(VRow(VCol(
			vx.Card(overview).HeaderTitle("Overview").
				Actions(
					h.If(editBtn != nil, editBtn),
				).Class("mb-4 rounded-lg").Outlined(true),
			vx.Card(pageState).HeaderTitle("Page State").
				Actions(
					h.If(pageStateBtn != nil, pageStateBtn),
				).Class("mb-4 rounded-lg").Outlined(true),
			vx.Card(notesSetcion).HeaderTitle("Notes").
				Actions(
					VBtn("Create").Depressed(true).
						Attr("@click", web.POST().
							EventFunc(createNoteDialogEvent).
							Query(presets.ParamOverlay, actions.Dialog).
							Query(presets.ParamID, p.PrimarySlug()).
							URL(mi.PresetsPrefix()+"/pages").Go(),
						),
				).Class("mb-4 rounded-lg").Outlined(true),
		).Cols(8)))
	}
}
