package pagebuilder

import (
	"fmt"
	"net/url"
	"os"

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
	"gorm.io/gorm"
)

func settings(db *gorm.DB, pm *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		p := obj.(*Page)
		c := &Category{}
		db.First(c, "id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode)

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
		var publishURL string
		if p.GetStatus() == publish.StatusOnline {
			var err error
			publishURL, err = url.JoinPath(os.Getenv("PUBLISH_URL"), p.getAccessUrl(p.GetOnlineUrl()))
			if err != nil {
				panic(err)
			}
		}
		pageState := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.GetStatus()).ZeroLabel("No State")).Label("State"),
				vx.DetailField(h.A(h.Text(publishURL)).Href(publishURL).Target("_blank").Class("text-truncate")).Label("URL"),
				vx.DetailField(vx.OptionalText(se).ZeroLabel("No Set")).Label("SchedulePublishTime"),
			),
		)
		var notes []note.QorNote
		ri := p.PrimarySlug()
		rt := pm.Info().Label()
		db.Where("resource_type = ? and resource_id = ?", rt, ri).
			Order("id DESC").Find(&notes)

		if len(notes) > 0 {
			userID, _ := note.GetUserData(ctx)
			userNote := note.UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
			db.Where(userNote).FirstOrCreate(&userNote)
			if userNote.Number != int64(len(notes)) {
				userNote.Number = int64(len(notes))
				db.Save(&userNote)
			}
		}
		var notesSetcion h.HTMLComponent
		if len(notes) > 0 {
			s := VContainer()
			for _, n := range notes {
				s.AppendChildren(VRow(VCardText(h.Text(n.Content)).Class("pb-0")))
				s.AppendChildren(VRow(VCardText(h.Text(fmt.Sprintf("%v - %v", n.Creator, n.CreatedAt.Format("2006-01-02 15:04:05 MST")))).Class("pt-0")))
			}
			notesSetcion = s
		}
		var editBtn h.HTMLComponent
		var pageStateBtn h.HTMLComponent
		var seoBtn h.HTMLComponent
		pvMsgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, utils.Messages_en_US).(*pv.Messages)
		if p.GetStatus() == publish.StatusDraft {
			editBtn = VBtn("Edit").Variant(VariantFlat).
				Attr("@click", web.POST().
					EventFunc(actions.Edit).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, p.PrimarySlug()).
					URL(mi.PresetsPrefix()+"/pages").Go(),
				)
			seoBtn = VBtn("Edit").Variant(VariantFlat).
				Attr("@click", web.POST().
					EventFunc(editSEODialogEvent).
					Query(presets.ParamOverlay, actions.Drawer).
					Query(presets.ParamID, p.PrimarySlug()).
					URL(mi.PresetsPrefix()+"/pages").Go(),
				)
		}
		if p.GetStatus() == publish.StatusOnline {
			pageStateBtn = VBtn(pvMsgr.Unpublish).Variant(VariantFlat).Class("mr-2").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.UnpublishEvent))
		} else {
			pageStateBtn = VBtn("Schedule Publish").Variant(VariantFlat).
				Attr("@click", web.POST().
					EventFunc(schedulePublishDialogEvent).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, p.PrimarySlug()).
					URL(mi.PresetsPrefix()+"/pages").Go(),
				)
		}

		seoState := "Default"
		if p.SEO.EnabledCustomize {
			seoState = "Customized"
		}
		seo := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(seoState)).Label("State"),
			),
		)
		return VContainer(
			VRow(
				VCol(
					vx.Card(overview).HeaderTitle("Overview").
						Actions(
							h.If(editBtn != nil, editBtn),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
					vx.Card(pageState).HeaderTitle("Page State").
						Actions(
							h.If(pageStateBtn != nil, pageStateBtn),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
					vx.Card(seo).HeaderTitle("SEO").
						Actions(
							h.If(seoBtn != nil, seoBtn),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
				).Cols(8),
				VCol(
					vx.Card(notesSetcion).HeaderTitle("Notes").
						Actions(
							VBtn("Create").Variant(VariantFlat).
								Attr("@click", web.POST().
									EventFunc(createNoteDialogEvent).
									Query(presets.ParamOverlay, actions.Dialog).
									Query(presets.ParamID, p.PrimarySlug()).
									URL(mi.PresetsPrefix()+"/pages").Go(),
								),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
				).Cols(4),
			),
		)
	}
}

func templateSettings(db *gorm.DB, pm *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Template)

		overview := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Name)).Label("Title"),
				vx.DetailField(vx.OptionalText(p.Description)).Label("Description"),
			),
		)

		editBtn := VBtn("Edit").Variant(VariantFlat).
			Attr("@click", web.POST().
				EventFunc(actions.Edit).
				Query(presets.ParamOverlay, actions.Dialog).
				Query(presets.ParamID, p.PrimarySlug()).
				URL(pm.Info().ListingHref()).Go(),
			)

		return VContainer(
			VRow(
				VCol(
					vx.Card(overview).HeaderTitle("Overview").
						Actions(
							h.If(editBtn != nil, editBtn),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
				).Cols(8),
			),
		)
	}
}
