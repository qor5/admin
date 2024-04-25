package views

import (
	"fmt"
	"net/url"

	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
	v "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type VersionComponentDesign struct {
	PublishDiglogContent h.HTMLComponent
}

func DefaultVersionComponentFunc(b *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			version        publish.VersionInterface
			status         publish.StatusInterface
			primarySlugger presets.SlugEncoder
			ok             bool
			versionSwitch  *v.VChipBuilder
			publishBtn     h.HTMLComponent
		)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, utils.Messages_en_US).(*Messages)
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		primarySlugger, ok = obj.(presets.SlugEncoder)
		if !ok {
			panic("obj should be SlugEncoder")
		}

		div := h.Div(
			// SchedulePublishDialog
			web.Portal().Name(SchedulePublishDialogPortalName),
			// Publish/Unpublish/Republish ConfirmDialog
			utils.ConfirmDialog(msgr.Areyousure, web.Plaid().EventFunc(web.Var("locals.action")).
				Query(presets.ParamID, primarySlugger.PrimarySlug()).Go(),
				utilsMsgr),
		).Class("w-100 d-inline-flex")

		if version, ok = obj.(publish.VersionInterface); ok {
			versionSwitch = v.VChip(
				h.Text(version.GetVersionName()),
			).Label(true).Variant(v.VariantOutlined).Class("rounded-r-0 text-black").
				Attr("style", "height:40px;background-color:#FFFFFF!important;").
				Attr("@click", web.Plaid().EventFunc(actions.OpenListingDialog).
					URL(b.Info().PresetsPrefix()+"/"+field.ModelInfo.URIName()+"-version-list-dialog").
					Query("select_id", primarySlugger.PrimarySlug()).
					Go()).
				Class(v.W100)
			if status, ok = obj.(publish.StatusInterface); ok {
				versionSwitch.AppendChildren(v.VChip(h.Text(GetStatusText(status.GetStatus(), msgr))).Label(true).Color(GetStatusColor(status.GetStatus())).Size(v.SizeSmall).Class("px-1  mx-1 text-black ml-2"))
			}
			versionSwitch.AppendIcon("mdi-chevron-down")

			div.AppendChildren(versionSwitch)
			div.AppendChildren(v.VBtn("").Icon("mdi-file-document-multiple").
				Height(40).Color("white").Class("rounded-sm").Variant(v.VariantFlat).
				Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, SaveNewVersionEvent)))
		}

		if status, ok = obj.(publish.StatusInterface); ok {
			switch status.GetStatus() {
			case publish.StatusDraft, publish.StatusOffline:
				publishBtn = h.Div(
					v.VBtn(msgr.Publish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, PublishEvent)).
						Class("rounded-sm ml-2").Variant(v.VariantFlat).Color("primary").Height(40),
				)
			case publish.StatusOnline:
				publishBtn = h.Div(
					v.VBtn(msgr.Unpublish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, UnpublishEvent)).
						Class("rounded-sm ml-2").Variant(v.VariantFlat).Color(presets.ColorPrimary).Height(40),
					v.VBtn(msgr.Republish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, RepublishEvent)).
						Class("rounded-sm ml-2").Variant(v.VariantFlat).Color(presets.ColorPrimary).Height(40),
				).Class("d-inline-flex")
			}
			div.AppendChildren(publishBtn)
		}

		if _, ok = obj.(publish.ScheduleInterface); ok {
			scheduleBtn := v.VBtn("").Icon("mdi-alarm").Class("rounded-sm ml-1").
				Variant(v.VariantFlat).Color("primary").Height(40).Attr("@click", web.POST().
				EventFunc(schedulePublishDialogEventV2).
				Query(presets.ParamOverlay, actions.Dialog).
				Query(presets.ParamID, primarySlugger.PrimarySlug()).
				URL(fmt.Sprintf("%s/%s-version-list-dialog", b.Info().PresetsPrefix(), b.Info().URIName())).Go(),
			)
			div.AppendChildren(scheduleBtn)
		}

		return web.Scope(div).VSlot(" { locals } ").Init(fmt.Sprintf(`{action: "", commonConfirmDialog: false }`))
	}
}

func ConfigureVersionListDialog(db *gorm.DB, b *presets.Builder, pm *presets.ModelBuilder) {
	// actually, VersionListDialog is a listing
	// use this URL : URLName-version-list-dialog
	mb := b.Model(pm.NewModel()).
		URIName(pm.Info().URIName() + "-version-list-dialog").
		InMenu(false)

	mb.RegisterEventFunc(schedulePublishDialogEventV2, schedulePublishDialogV2(db, mb))
	mb.RegisterEventFunc(schedulePublishEventV2, schedulePublishV2(db, mb))
	mb.RegisterEventFunc(renameVersionDialogEventV2, renameVersionDialogV2(mb))
	mb.RegisterEventFunc(renameVersionEventV2, renameVersionV2(mb))
	mb.RegisterEventFunc(deleteVersionDialogEventV2, deleteVersionDialogV2(mb))

	searcher := mb.Listing().Searcher
	lb := mb.Listing("Version", "State", "StartAt", "EndAt", "Notes", "Option").
		SearchColumns("version", "version_name").
		PerPage(10).
		SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
			id := ctx.R.FormValue("select_id")
			if id == "" {
				id = ctx.R.FormValue("f_select_id")
			}
			if id != "" {
				cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)
				con := presets.SQLCondition{
					Query: "id = ?",
					Args:  []interface{}{cs["id"]},
				}
				params.SQLConditions = append(params.SQLConditions, &con)
			}
			params.OrderBy = "created_at DESC"

			return searcher(model, params, ctx)
		})
	lb.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		return cell
	})
	lb.Field("Version").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		versionName := obj.(publish.VersionInterface).GetVersionName()
		p := obj.(presets.SlugEncoder)
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}
		return h.Td(
			v.VRadio().ModelValue(p.PrimarySlug()).TrueValue(id).Attr("@change", web.Plaid().EventFunc(actions.UpdateListingDialog).
				URL(b.GetURIPrefix()+"/"+mb.Info().URIName()).
				Query("select_id", p.PrimarySlug()).
				Go()),
			h.Text(versionName),
		).Class("d-inline-flex align-center")
	})
	lb.Field("State").ComponentFunc(StatusListFunc())
	lb.Field("StartAt").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(publish.ScheduleInterface)
		var showTime string
		if p.GetScheduledStartAt() != nil {
			showTime = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}

		return h.Td(
			h.Text(showTime),
		)
	}).Label("Start at")
	lb.Field("EndAt").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(publish.ScheduleInterface)
		var showTime string
		if p.GetScheduledEndAt() != nil {
			showTime = p.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		return h.Td(
			h.Text(showTime),
		)
	}).Label("End at")

	lb.Field("Notes").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(presets.SlugEncoder)
		rt := pm.Info().Label()
		ri := p.PrimarySlug()
		userID, _ := note.GetUserData(ctx)
		count := note.GetUnreadNotesCount(db, userID, rt, ri)

		return h.Td(
			h.If(count > 0,
				v.VBadge().Content(count).Color("red"),
			).Else(
				h.Text(""),
			),
		)
	}).Label("Unread Notes")

	lb.Field("Option").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		versionIf := obj.(publish.VersionInterface)
		statusIf := obj.(publish.StatusInterface)
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}
		versionName := versionIf.GetVersionName()
		var disable bool
		if statusIf.GetStatus() == publish.StatusOnline || statusIf.GetStatus() == publish.StatusOffline {
			disable = true
		}

		return h.Td(v.VBtn("Delete").Disabled(disable).PrependIcon("mdi-delete").Size(v.SizeXSmall).Color("primary").Variant(v.VariantText).Attr("@click", web.Plaid().
			URL(pm.Info().PresetsPrefix()+"/"+mb.Info().URIName()).
			EventFunc(deleteVersionDialogEventV2).
			Queries(ctx.Queries()).
			Query(presets.ParamOverlay, actions.Dialog).
			Query("delete_id", obj.(presets.SlugEncoder).PrimarySlug()).
			Query("version_name", versionName).
			Go()))
	})
	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.FooterAction("Cancel").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return v.VBtn("Cancel").Variant(v.VariantElevated).Attr("@click", "vars.presetsListingDialog=false")
	})
	lb.FooterAction("Save").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}

		return v.VBtn("Save").Variant(v.VariantElevated).Color("secondary").Attr("@click", web.Plaid().
			Query("select_id", id).
			URL(pm.Info().PresetsPrefix()+"/"+mb.Info().URIName()).
			EventFunc(selectVersionEventV2).
			Go())
	})
	lb.RowMenu().Empty()

	mb.RegisterEventFunc(selectVersionEventV2, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		refer, _ := url.Parse(ctx.R.Referer())
		newQueries := refer.Query()
		id := ctx.R.FormValue("select_id")

		if !pm.GetHasDetailing() {
			// close dialog and open editing
			newQueries.Add(presets.ParamID, id)
			r.RunScript = presets.CloseListingDialogVarScript + ";" +
				web.Plaid().EventFunc(actions.Edit).Queries(newQueries).Go()
			return
		}
		if !pm.GetDetailing().GetDrawer() {
			// open detailing without drawer
			// jump URL to support referer
			r.PushState = web.Location(newQueries).URL(pm.Info().DetailingHref(id))
			return
		}
		newQueries.Add(presets.ParamID, id)
		// close dialog and open detailingDrawer
		r.RunScript = presets.CloseListingDialogVarScript + ";" +
			web.Plaid().EventFunc(actions.DetailingDrawer).Queries(newQueries).Go()
		return
	})

	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:          "all",
				Invisible:    true,
				SQLCondition: ``,
			},
			{

				Key:          "online_version",
				Invisible:    true,
				SQLCondition: `status = 'online'`,
			},
			{
				Key:          "named_versions",
				Invisible:    true,
				SQLCondition: `version <> version_name`,
			},
		}
	})

	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabAllVersions,
				ID:    "all",
				Query: url.Values{"all": []string{"1"}, "select_id": []string{id}},
			},
			{
				Label: msgr.FilterTabOnlineVersion,
				ID:    "online_version",
				Query: url.Values{"online_version": []string{"1"}, "select_id": []string{id}},
			},
			{
				Label: msgr.FilterTabNamedVersions,
				ID:    "named_versions",
				Query: url.Values{"named_versions": []string{"1"}, "select_id": []string{id}},
			},
		}
	})
}
