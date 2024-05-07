package pagebuilder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/x/v3/i18n"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
)

func settings(db *gorm.DB, b *Builder, activityB *activity.Builder) presets.FieldComponentFunc {
	// TODO: refactor versionDialog to use publish/views
	pm := b.mb
	seoBuilder := b.seoBuilder
	// publish.ConfigureVersionListDialog(db, b.ps, pm)
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// TODO: init default VersionComponent
		var (
			versionComponent  = publish.DefaultVersionComponentFunc(pm)(obj, field, ctx)
			mi                = field.ModelInfo
			p                 = obj.(*Page)
			c                 = &Category{}
			previewDevelopUrl = b.previewHref(strconv.Itoa(int(p.GetID())), p.GetVersion(), p.GetLocale())
		)
		var (
			start, end, se string
			notes          []note.QorNote
			categories     []*Category
			notesItems     []h.HTMLComponent
			timelineItems  []h.HTMLComponent
			onlineHint     h.HTMLComponent
		)
		db.First(c, "id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode)
		if p.GetScheduledStartAt() != nil {
			start = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if p.GetScheduledEndAt() != nil {
			end = p.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		if start != "" || end != "" {
			se = "Scheduled at: " + start + " ~ " + end
		}
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
		if len(notes) > 0 {
			for _, n := range notes {
				notesItems = append(notesItems, VTimelineItem(
					h.Div(h.Text(n.CreatedAt.Format("2006-01-02 15:04:05 MST"))).Class("text-caption"),
					h.Div(
						VAvatar().Text(strings.ToUpper(string(n.Creator[0]))).Color(ColorSecondary).Class("text-h6 rounded-lg").Size(SizeXSmall),
						h.Strong(n.Creator).Class("ml-1"),
					),
					h.Div(h.Text(n.Content)).Class("text-caption"),
				).DotColor(ColorSuccess).Size(SizeXSmall),
				)
			}
		}

		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		noteMsgr := i18n.MustGetModuleMessages(ctx.R, note.I18nNoteKey, note.Messages_en_US).(*note.Messages)
		seoMsgr := i18n.MustGetModuleMessages(ctx.R, seo.I18nSeoKey, seo.Messages_en_US).(*seo.Messages)
		if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
			panic(err)
		}

		infoComponentTab := h.Div(
			VTabs(
				VTab(h.Text("Page")).Size(SizeXSmall).Value("Page"),
				VTab(h.Text(seoMsgr.Seo)).Size(SizeXSmall).Value("Seo"),
			).Attr("v-model", "editLocals.infoTab"),
			h.Div(
				VBtn("Save").AppendIcon("mdi-check").Color(ColorSecondary).Size(SizeSmall).Variant(VariantFlat).
					Attr("@click", fmt.Sprintf(`editLocals.infoTab=="Page"?%s:%s`, web.POST().
						EventFunc(actions.Update).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").
						Go(), web.Plaid().
						EventFunc(updateSEOEvent).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").
						Go()),
					),
			),
		).Class("d-flex justify-space-between align-center")

		seoForm := seoBuilder.EditingComponentFunc(obj, nil, ctx)

		infoComponentContent := VTabsWindow(
			VTabsWindowItem(
				b.mb.Editing().ToComponent(pm.Info(), obj, ctx),
			).Value("Page").Class("pt-8"),
			VTabsWindowItem(
				seoForm,
			).Value("Seo").Class("pt-8"),
		).Attr("v-model", "editLocals.infoTab")

		detailComponentTab :=
			VTabs(
				VTab(h.Text("Activity")).Size(SizeXSmall).Value("Activity"),
				VTab(h.Text(noteMsgr.Notes)).Size(SizeXSmall).Value("Notes"),
			).Attr("v-model", "editLocals.detailTab").AlignTabs(Center).FixedTabs(true)
		if activityB != nil {
			for _, i := range activityB.GetActivityLogs(p, db.Order("created_at desc")) {
				timelineItems = append(timelineItems,
					VTimelineItem(
						h.Div(h.Text(i.GetCreatedAt().Format("2006-01-02 15:04:05 MST"))).Class("text-caption"),
						h.Div(
							VAvatar().Text(strings.ToUpper(string(i.GetCreator()[0]))).Color(ColorSecondary).Class("text-h6 rounded-lg").Size(SizeXSmall),
							h.Strong(i.GetCreator()).Class("ml-1"),
						),
						h.Div(h.Text(i.GetAction())).Class("text-caption"),
					).DotColor(ColorSuccess).Size(SizeXSmall),
				)
			}

		}
		detailComponentContent := VTabsWindow(
			VTabsWindowItem(
				VBtn(noteMsgr.NewNote).PrependIcon("mdi-plus").Variant(VariantTonal).Class("w-100").
					Attr("@click", web.POST().
						EventFunc(createNoteDialogEvent).
						Query(presets.ParamOverlay, actions.Dialog).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").Go(),
					),
				VTimeline(
					notesItems...,
				).Density(DensityCompact).TruncateLine("start").Side("end").Align(LocationStart).Class("mt-5"),
			).Value("Notes").Class("pa-5"),
			VTabsWindowItem(
				VTimeline(
					timelineItems...,
				).Density(DensityCompact).TruncateLine("start").Side("end").Align(LocationStart),
			).Value("Activity").Class("pa-5"),
		).Attr("v-model", "editLocals.detailTab")
		versionBadge := VChip(h.Text(fmt.Sprintf("%d versions", versionCount(db, p)))).Color(ColorPrimary).Size(SizeSmall).Class("px-1 mx-1").Attr("style", "height:20px")
		if p.GetStatus() == publish.StatusOnline {
			onlineHint = VAlert(h.Text("The version cannot be edited directly after it is released. Please copy the version and edit it.")).Density(DensityCompact).Type(TypeInfo).Variant(VariantTonal).Closable(true).Class("mb-2")
		}
		return VContainer(
			web.Scope(
				VLayout(
					VCard(
						web.Slot(
							VBtn("").Size(SizeSmall).Icon("mdi-arrow-left").Variant(VariantText).Attr("@click",
								web.GET().URL(mi.PresetsPrefix()+"/pages").PushState(true).Go(),
							),
						).Name(VSlotPrepend),
						web.Slot(h.Text(p.Title),
							versionBadge,
						).Name(VSlotTitle),
						VCardText(
							onlineHint,
							versionComponent,
							h.Div(
								h.Iframe().Src(previewDevelopUrl).Style(`height:320px;width:100%;pointer-events: none;`),
								h.Div(
									h.Div(
										h.Text(se),
									).Class(fmt.Sprintf("bg-%s", ColorSecondaryLighten2)),
									VBtn("Page Builder").PrependIcon("mdi-pencil").Color(ColorSecondary).
										Class("rounded-sm").Height(40).Variant(VariantFlat),
								).Class("pa-6 w-100 d-flex justify-space-between align-center").Style(`position:absolute;top:0;left:0`),
							).Style(`position:relative`).Class("w-100 mt-4").
								Attr("@click", web.Plaid().Query("tab", "content").PushState(true).Go()),
							h.Div(
								h.A(h.Text(previewDevelopUrl)).Href(previewDevelopUrl),
								VBtn("").Icon("mdi-file-document-multiple").Variant(VariantText).Size(SizeXSmall).Class("ml-1").
									Attr("@click", fmt.Sprintf(`$event.view.window.navigator.clipboard.writeText($event.view.window.location.origin+"%s");vars.presetsMessage = { show: true, message: "success", color: "%s"}`, previewDevelopUrl, ColorSuccess)),
							).Class("d-inline-flex align-center"),

							web.Scope(
								infoComponentTab.Class("mt-7"),
								infoComponentContent,
							).VSlot("{form}"),
						),
					).Class("w-75"),
					VCard(
						VCardText(
							detailComponentTab,
							detailComponentContent),
					).Class("w-25").Class("ml-5"),
				).Class("d-inline-flex w-100"),
			).VSlot(" { locals : editLocals }").Init(`{ infoTab:"Page",detailTab:"Activity" } `),
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
