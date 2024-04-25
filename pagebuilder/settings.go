package pagebuilder

import (
	"fmt"
	"github.com/qor5/admin/v3/activity"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	pv "github.com/qor5/admin/v3/publish/views"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func settings(db *gorm.DB, b *Builder, activityB *activity.ActivityBuilder) presets.FieldComponentFunc {
	// TODO: refactor versionDialog to use publish/views
	pm := b.mb
	seoBuilder := b.seoBuilder
	pv.ConfigureVersionListDialog(db, b.ps, pm)
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// TODO: init default VersionComponent
		versionComponent := pv.DefaultVersionComponentFunc(pm)(obj, field, ctx)

		mi := field.ModelInfo
		p := obj.(*Page)
		c := &Category{}
		db.First(c, "id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode)
		var previewDevelopUrl = b.previewHref(strconv.Itoa(int(p.GetID())), p.GetVersion(), p.GetLocale())
		var start, end, se string
		if p.GetScheduledStartAt() != nil {
			start = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if p.GetScheduledEndAt() != nil {
			end = p.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		if start != "" || end != "" {
			se = "Scheduled at: " + start + " ~ " + end
		}
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
				VBtn("Save").AppendIcon("mdi-check").Color("black").Size(SizeSmall).Variant(VariantFlat).
					Attr("@click", web.POST().
						EventFunc(actions.Update).
						Queries(ctx.Queries()).
						Query(presets.ParamID, p.PrimarySlug()).
						URL(mi.PresetsPrefix()+"/pages").
						Go(),
					),
				VBtn("").AppendIcon("mdi-plus").Color("black").Size(SizeSmall).Variant(VariantFlat),
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
		var timelineItems []h.HTMLComponent
		if activityB != nil {
			for _, i := range activityB.GetActivityLogs(p, db) {
				timelineItems = append(timelineItems,
					VTimelineItem(
						h.Div(h.Text(i.GetCreatedAt().Format("2006-01-02 15:04:05"))).Class("text-caption"),
						h.Div(
							VAvatar().Text(strings.ToUpper(string(i.GetCreator()[0]))).Color("secondary").Class("text-h6 rounded-lg").Size(SizeXSmall),
							h.Strong(i.GetCreator()).Class("ml-1"),
						),
						h.Div(h.Text(i.GetAction())).Class("text-caption"),
					).DotColor("success").Size(SizeXSmall),
				)
			}

		}
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
					timelineItems...,
				).Density(DensityCompact).TruncateLine("start").Side("end").Align(LocationStart),
			).Class("pa-5").Value("Activity"),
		).Attr("v-model", "editLocals.detailTab")
		versionBadge := VChip(h.Text(fmt.Sprintf("%d versions", versionCount(db, p)))).Label(true).Color("primary").Size(SizeSmall).Class("px-1 mx-1 text-black").Attr("style", "height:20px")
		return VContainer(
			web.Scope(
				VLayout(
					VCard(
						web.Slot(
							VBtn("").Size(SizeSmall).Icon("mdi-arrow-left").Color("neutral").Variant(VariantText).Attr("@click",
								web.GET().URL(mi.PresetsPrefix()+"/pages").PushState(true).Go(),
							),
						).Name("prepend"),
						web.Slot(h.Text(p.Title),
							versionBadge,
						).Name("title"),

						VCardText(
							h.Div(
								h.Iframe().Src(previewDevelopUrl).Style(`height:320px;width:100%;`),
								h.Div(versionComponent).Class("w-100 pa-6").Style(`position:absolute;top:0;left:0`),
								h.Div(
									h.Div(
										h.Text(se),
									).Class("bg-secondary"),
									VBtn("Page Builder").PrependIcon("mdi-pencil").Color("secondary").
										Class("rounded-sm").Height(40).Variant(VariantFlat).
										Attr("@click", web.Plaid().Query("tab", "content").PushState(true).Go()),
								).Class("pa-6 w-100 d-flex justify-space-between align-center").Style(`position:absolute;bottom:0;left:0`),
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
