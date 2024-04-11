package pagebuilder

import (
	"fmt"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/seo"
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

func settings(db *gorm.DB, pm *presets.ModelBuilder, b *Builder, seoBuilder *seo.Builder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		p := obj.(*Page)
		c := &Category{}
		db.First(c, "id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode)
		var previewDevelopUrl = "http://localhost:9500/page_builder/preview?id=2&version=2024-04-11-v01&locale=International"
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
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		var categories []*Category
		if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
			panic(err)
		}

		infoComponentTab := h.Div(
			VTabs(
				VTab(h.Text("Page")).Size(SizeXSmall).Value("Page"),
				VTab(h.Text("Seo")).Size(SizeXSmall).Value("Seo"),
			).Attr("v-model", "editLocals.infoTab"),
			h.Div(
				VBtn("Save").AppendIcon("mdi-check").Color("black").Class("text-none").Size(SizeSmall).Variant(VariantFlat).
					Attr("@click", web.POST().
						EventFunc(actions.Update).
						Queries(ctx.Queries()).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").
						Go(),
					),
				VBtn("").AppendIcon("mdi-plus").Color("black").Class("text-none").Size(SizeSmall).Variant(VariantFlat),
			),
		).Class("d-flex justify-space-between align-center")

		seoForm := seoBuilder.EditingComponentFunc(obj, nil, ctx)

		infoComponentContent := VWindow(
			VWindowItem(
				VTextField().Label("Title").Variant(VariantOutlined).Attr(web.VField("Title", p.Title)...),
				VTextField().Label("Slug").Variant(VariantOutlined).Attr(web.VField("Slug", p.Slug)...),
				VAutocomplete().Label(msgr.Category).Variant(VariantOutlined).
					Attr(web.VField("CategoryID", p.CategoryID)...).
					Multiple(false).Chips(false).
					Items(categories).ItemTitle("Path").ItemValue("ID"),
			).Value("Page").Class("mt-9"),
			VWindowItem(
				seoForm,
			).Value("Seo"),
		).Attr("v-model", "editLocals.infoTab")

		detailComponentTab :=
			VTabs(
				VTab(h.Text("Activity")).Size(SizeXSmall).Value("Activity"),
				VTab(h.Text("Notes")).Size(SizeXSmall).Value("Notes"),
			).Attr("v-model", "editLocals.detailTab").AlignTabs(Center).FixedTabs(true)
		detailComponentContent := VWindow(
			VWindowItem(
				VBtn("New note").PrependIcon("mdi-plus").Variant(VariantOutlined).Color("blue").Class("w-100").
					Attr("@click", web.POST().
						EventFunc(createNoteDialogEvent).
						Query(presets.ParamOverlay, actions.Dialog).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").Go(),
					),
				notesSetcion,
			).Class("pa-5").Value("Notes"),
			VWindowItem(
				VTimeline(
					VTimelineItem(
						h.Div(h.Text("2 hours ago")).Class("text-caption"),
						h.Div(
							VAvatar().Size(SizeXSmall).Image("https://cdn.vuetifyjs.com/images/lists/1.jpg"),
							h.Strong("Peterson  Lee").Class("ml-1"),
						),
						h.Div(h.Text("Edited the page")).Class("text-caption"),
					).DotColor("success").Size(SizeXSmall),
					VTimelineItem(
						h.Div(h.Text("3 hours ago")).Class("text-caption"),
						h.Div(
							VAvatar().Size(SizeXSmall).Image("https://cdn.vuetifyjs.com/images/lists/2.jpg"),
							h.Strong("Peterson  Lee").Class("ml-1"),
						),
						h.Div(h.Text("Edited the page")).Class("text-caption"),
					).DotColor("success").Size(SizeXSmall),
				).Density(DensityCompact).TruncateLine("start").Side("end").Align(LocationStart).Class("mt-16 mr-4"),
			).Class("mt-5").Value("Activity"),
		).Attr("v-model", "editLocals.detailTab")
		versionSwitch := VChip(
			VChip(h.Text(fmt.Sprintf("%d", versionCount(db, p)))).Label(true).Color("#E0E0E0").Size(SizeSmall).Class("px-1 mx-1 text-black").Attr("style", "height:20px"),
			h.Text(p.GetVersionName()+" | "),
			VChip(h.Text(pv.GetStatusText(p.GetStatus(), pvMsgr))).Label(true).Color(pv.GetStatusColor(p.GetStatus())).Size(SizeSmall).Class("px-1  mx-1 text-black").Attr("style", "height:20px"),
			VIcon("chevron_right"),
		).Label(true).Variant(VariantOutlined).Class("px-1 ml-8 rounded-r-0 text-black").Attr("style", "height:40px;background-color:#FFFFFF!important;").
			Attr("@click", web.Plaid().EventFunc(actions.OpenListingDialog).
				URL(b.prefix+"/version-list-dialog").
				Query("select_id", p.PrimarySlug()).
				Go())

		return VContainer(
			web.Scope(
				VLayout(
					VCard(
						web.Slot(
							VBtn("").Size(SizeSmall).Icon("mdi-arrow-left").Color("neutral").Variant(VariantText).Attr("@click",
								web.GET().URL(mi.PresetsPrefix()+"/pages").PushState(true).Go(),
							),
						).Name("prepend"),
						web.Slot(h.Text("Heading"),
							VChip(h.Text(p.GetStatus())).Color("warning").Class("ml-2"),
						).Name("title"),

						VCardText(
							h.Div(
								h.Iframe().Src(previewDevelopUrl).Style(`height:320px;width:100%;`),
								h.Div(
									versionSwitch.Class("w-75"),
									VBtn("").Icon("mdi-file-document-multiple").Color("white").Class("text-none rounded-sm").Size(SizeSmall).Variant(VariantFlat),
									VBtn("").Icon("mdi-pencil").Color("black").Class("text-none rounded-sm ml-2").Size(SizeSmall).Variant(VariantFlat),
								).Class("w-100 d-inline-flex").Style(`position:absolute;bottom:24px;left:24px`),
							).Style(`position:relative`).Class("w-100"),
							h.Div(
								h.A(h.Text(previewDevelopUrl)).Href(previewDevelopUrl),
								VBtn("").Icon("mdi-file-document-multiple").Color("accent").Variant(VariantText).Size(SizeXSmall).Class("ml-1"),
							).Class("d-inline-flex align-center"),

							infoComponentTab.Class("mt-7"),
							infoComponentContent,
						),
					).Class("w-75"),
					VCard(
						VCardText(
							detailComponentTab,
							detailComponentContent),
					).Class("w-25").Class("ml-5"),
				).Class("d-inline-flex w-100"),
			).VSlot(" { locals : editLocals }").Init(`{ infoTab:"Page",detailTab:"Notes" } `),

			h.If(false, VRow(
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
			)),
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
