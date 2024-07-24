package publish

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const filterKeySelected = "f_select_id"

func MustFilterQuery(compo *presets.ListingCompo) url.Values {
	if compo == nil {
		panic("compo is nil")
	}
	qs, err := url.ParseQuery(compo.FilterQuery)
	if err != nil {
		panic(err)
	}
	return qs
}

const versionListDialogURISuffix = "-version-list-dialog"

type VersionComponentConfig struct {
	// If you want to use custom publish dialog, you can update the portal named PublishCustomDialogPortalName
	PublishEvent     func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	UnPublishEvent   func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	RePublishEvent   func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	Top              bool
	DisableListeners bool
}

func DefaultVersionComponentFunc(mb *presets.ModelBuilder, cfg ...VersionComponentConfig) presets.FieldComponentFunc {
	var config VersionComponentConfig
	if len(cfg) > 0 {
		config = cfg[0]
	}
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			version        VersionInterface
			status         StatusInterface
			primarySlugger presets.SlugEncoder
			ok             bool
			versionSwitch  *v.VChipBuilder
			publishBtn     h.HTMLComponent
			verifier       = mb.Info().Verifier()
		)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		primarySlugger, ok = obj.(presets.SlugEncoder)
		if !ok {
			panic("obj should be SlugEncoder")
		}

		div := h.Div().Class("w-100 d-inline-flex")
		div.AppendChildren(
			utils.ConfirmDialog(msgr.Areyousure, web.Plaid().EventFunc(web.Var("locals.action")).
				Query(presets.ParamID, primarySlugger.PrimarySlug()).Go(),
				utilsMsgr),
		)

		if !config.Top {
			div.Class("pb-4")
		}

		urlSuffix := field.ModelInfo.URIName() + versionListDialogURISuffix
		if version, ok = obj.(VersionInterface); ok {
			versionSwitch = v.VChip(
				h.Text(version.EmbedVersion().VersionName),
			).Label(true).Variant(v.VariantOutlined).
				Attr("style", "height:40px;").
				On("click", web.Plaid().EventFunc(actions.OpenListingDialog).
					URL(mb.Info().PresetsPrefix()+"/"+urlSuffix).
					Query(filterKeySelected, primarySlugger.PrimarySlug()).
					Go()).
				Class(v.W100)
			if status, ok = obj.(StatusInterface); ok {
				versionSwitch.AppendChildren(statusChip(status.EmbedStatus().Status, msgr).Class("mx-2"))
			}
			versionSwitch.AppendChildren(v.VSpacer())
			versionSwitch.AppendIcon("mdi-chevron-down")

			div.AppendChildren(versionSwitch)

			if !DeniedDo(verifier, obj, ctx.R, presets.PermUpdate, PermDuplicate) {
				div.AppendChildren(v.VBtn(msgr.Duplicate).PrependIcon("mdi-file-document-multiple").
					Height(40).Class("ml-2").Variant(v.VariantOutlined).
					Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, EventDuplicateVersion)))
			}
		}

		deniedPublish := DeniedDo(verifier, obj, ctx.R, PermPublish)
		deniedUnpublish := DeniedDo(verifier, obj, ctx.R, PermUnpublish)
		if status, ok = obj.(StatusInterface); ok {
			switch status.EmbedStatus().Status {
			case StatusDraft, StatusOffline:
				if !deniedPublish {
					publishEvent := fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, EventPublish)
					if config.PublishEvent != nil {
						publishEvent = config.PublishEvent(obj, field, ctx)
					}
					publishBtn = h.Div(
						v.VBtn(msgr.Publish).Attr("@click", publishEvent).Rounded("0").
							Class("rounded-s ml-2").Variant(v.VariantFlat).Color(v.ColorPrimary).Height(40),
					)
				}
			case StatusOnline:
				var unPublishEvent, rePublishEvent string
				if !deniedUnpublish {
					unPublishEvent = fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, EventUnpublish)
					if config.UnPublishEvent != nil {
						unPublishEvent = config.UnPublishEvent(obj, field, ctx)
					}
				}
				if !deniedPublish {
					rePublishEvent = fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, EventRepublish)
					if config.RePublishEvent != nil {
						rePublishEvent = config.RePublishEvent(obj, field, ctx)
					}
				}
				if unPublishEvent != "" || rePublishEvent != "" {
					publishBtn = h.Div(
						h.Iff(unPublishEvent != "", func() h.HTMLComponent {
							return v.VBtn(msgr.Unpublish).Attr("@click", unPublishEvent).
								Class("ml-2").Variant(v.VariantFlat).Color(v.ColorError).Height(40)
						}),
						h.Iff(rePublishEvent != "", func() h.HTMLComponent {
							return v.VBtn(msgr.Republish).Attr("@click", rePublishEvent).
								Class("ml-2").Variant(v.VariantFlat).Color(v.ColorPrimary).Height(40)
						}),
					).Class("d-inline-flex")
				}
			}
			if publishBtn != nil {
				div.AppendChildren(publishBtn)
				// Publish/Unpublish/Republish CustomDialog
				if config.UnPublishEvent != nil || config.RePublishEvent != nil || config.PublishEvent != nil {
					div.AppendChildren(web.Portal().Name(PortalPublishCustomDialog))
				}
			}
		}

		if _, ok = obj.(ScheduleInterface); ok {
			deniedSchedule := deniedPublish || deniedUnpublish || DeniedDo(verifier, obj, ctx.R, PermSchedule)
			if !deniedSchedule {
				var scheduleBtn h.HTMLComponent
				clickEvent := web.Plaid().
					EventFunc(eventSchedulePublishDialog).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, primarySlugger.PrimarySlug()).
					URL(mb.Info().ListingHref()).Go()
				if config.Top {
					scheduleBtn = v.VAutocomplete().PrependInnerIcon("mdi-alarm").Density(v.DensityCompact).
						Variant(v.FieldVariantSoloFilled).ModelValue("Schedule Publish Time").
						BgColor(v.ColorPrimaryLighten2).Readonly(true).
						Width(600).HideDetails(true).Attr("@click", clickEvent).Class("ml-2 text-caption")
				} else {
					scheduleBtn = v.VBtn("").Children(v.VIcon("mdi-alarm").Size(v.SizeXLarge)).Rounded("0").Class("ml-1 rounded-e").
						Variant(v.VariantFlat).Color(v.ColorPrimary).Height(40).Attr("@click", clickEvent)
				}
				div.AppendChildren(scheduleBtn)
				// SchedulePublishDialog
				div.AppendChildren(web.Portal().Name(PortalSchedulePublishDialog))
			}
		}

		children := []h.HTMLComponent{div}
		if !config.DisableListeners {
			slug := primarySlugger.PrimarySlug()
			children = append(children,
				NewListenerVersionSelected(mb, slug),
				NewListenerModelsDeleted(mb, slug),
			)
		}
		return web.Scope(children...).VSlot(" { locals } ").Init(`{action: "", commonConfirmDialog: false }`)
	}
}

func DefaultVersionBar(db *gorm.DB) presets.ObjectComponentFunc {
	return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		res := h.Div().Class("d-inline-flex align-center")

		slugEncoderIf := obj.(presets.SlugEncoder)
		slugDncoderIf := obj.(presets.SlugDecoder)
		mp := slugDncoderIf.PrimaryColumnValuesBySlug(slugEncoderIf.PrimarySlug())

		currentObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
		err := db.Where("id = ?", mp["id"]).Where("status = ?", StatusOnline).First(&currentObj).Error
		if err != nil {
			return res
		}
		versionIf := currentObj.(VersionInterface)
		currentVersionStr := fmt.Sprintf("%s: %s", msgr.OnlineVersion, versionIf.EmbedVersion().VersionName)
		res.AppendChildren(v.VChip(h.Span(currentVersionStr)).Density(v.DensityCompact).Color(v.ColorSuccess))

		if _, ok := currentObj.(ScheduleInterface); !ok {
			return res
		}

		nextObj := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
		flagTime := db.NowFunc()
		count := int64(0)
		err = db.Model(nextObj).Where("id = ?", mp["id"]).Where("scheduled_start_at >= ?", flagTime).Count(&count).Error
		if err != nil {
			return res
		}

		if count == 0 {
			return res
		}

		err = db.Where("id = ?", mp["id"]).Where("scheduled_start_at >= ?", flagTime).Order("scheduled_start_at ASC").First(&nextObj).Error
		if err != nil {
			return res
		}
		res.AppendChildren(
			h.Div(
				h.Div().Class(fmt.Sprintf(`w-100 bg-%s`, v.ColorSuccessLighten2)).Style("height:4px"),
				v.VIcon("mdi-circle").Size(v.SizeXSmall).Color(v.ColorSuccess).Attr("style", "position:absolute;left:0;right:0;margin-left:auto;margin-right:auto"),
			).Class("h-100 d-flex align-center").Style("position:relative;width:40px"),
		)
		versionIf = nextObj.(VersionInterface)
		// TODO use nextVersion GetI18n
		nextText := fmt.Sprintf("%s: %s", msgr.OnlineVersion, versionIf.EmbedVersion().VersionName)
		res.AppendChildren(v.VChip(h.Span(nextText)).Density(v.DensityCompact).Color(v.ColorSecondary))
		if count >= 2 {
			res.AppendChildren(
				h.Div(
					h.Div().Class(fmt.Sprintf(`w-100 bg-%s`, v.ColorSecondaryLighten1)).Style("height:4px"),
				).Class("h-100 d-flex align-center").Style("width:40px"),
				h.Div(
					h.Text(fmt.Sprintf(`+%v`, count)),
				).Class(fmt.Sprintf(`text-caption bg-%s`, v.ColorSecondaryLighten1)),
			)
		}
		return res
	}
}

func configureVersionListDialog(db *gorm.DB, pb *Builder, b *presets.Builder, pm *presets.ModelBuilder) {
	// actually, VersionListDialog is a listing
	// use this URL : URLName-version-list-dialog
	mb := b.Model(pm.NewModel()).
		URIName(pm.Info().URIName() + versionListDialogURISuffix).
		InMenu(false)

	fmt.Println(fmt.Sprintf("%s", pm.Info().URIName()))
	b.GetPermission().CreatePolicies(
		perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(fmt.Sprintf("*:presets:%s_version_list_dialog:*", pm.Info().URIName())),
	)

	listingHref := mb.Info().ListingHref()
	registerEventFuncsForVersion(mb, db)
	listingFields := []string{"Version", "Status", "StartAt", "EndAt", "Option"}
	if pb.ab != nil {
		defer func() { pb.ab.RegisterModel(mb) }()
		listingFields = []string{"Version", "Status", "StartAt", "EndAt", activity.ListFieldNotes, "Option"}
	}

	lb := mb.Listing(listingFields...).
		DialogWidth("900px").
		TitleFunc(func(evCtx *web.EventContext) (string, error) {
			msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPublishKey, Messages_en_US).(*Messages)
			return msgr.VersionsList, nil
		}).
		SearchColumns("version", "version_name").
		PerPage(10).
		WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
			return func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
				compo := presets.ListingCompoFromEventContext(ctx)
				selected := MustFilterQuery(compo).Get(filterKeySelected)

				if selected != "" {
					cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(selected)
					con := presets.SQLCondition{
						Query: "id = ?",
						Args:  []interface{}{cs["id"]},
					}
					params.SQLConditions = append(params.SQLConditions, &con)
				}
				params.OrderBy = "created_at DESC"

				return in(model, params, ctx)
			}
		})
	lb.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		return cell
	})
	lb.Field("Version").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		compo := presets.ListingCompoFromEventContext(ctx)
		filter := MustFilterQuery(compo)
		selected := filter.Get(filterKeySelected)
		slug := obj.(presets.SlugEncoder).PrimarySlug()
		filter.Set(filterKeySelected, slug)
		return h.Td().Children(
			h.Div().Class("d-inline-flex align-center").Children(
				v.VRadio().ModelValue(slug).TrueValue(selected).Attr("@change",
					stateful.ReloadAction(ctx.R.Context(), compo, func(target *presets.ListingCompo) {
						target.FilterQuery = filter.Encode()
					}).Go(),
				),
				h.Text(
					obj.(VersionInterface).EmbedVersion().VersionName,
				),
			),
		)
	})
	lb.Field("Status").ComponentFunc(StatusListFunc())
	lb.Field("StartAt").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(ScheduleInterface)

		return h.Td(
			h.Text(ScheduleTimeString(p.EmbedSchedule().ScheduledStartAt)),
		)
	}).Label("Start at")
	lb.Field("EndAt").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(ScheduleInterface)
		return h.Td(
			h.Text(ScheduleTimeString(p.EmbedSchedule().ScheduledEndAt)),
		)
	}).Label("End at")

	lb.Field("Option").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		pmsgr := presets.MustGetMessages(ctx.R)

		id := obj.(presets.SlugEncoder).PrimarySlug()
		versionName := obj.(VersionInterface).EmbedVersion().VersionName
		status := obj.(StatusInterface).EmbedStatus().Status
		disable := status == StatusOnline || status == StatusOffline
		verifier := mb.Info().Verifier()
		deniedUpdate := DeniedDo(verifier, obj, ctx.R, presets.PermUpdate)
		deniedDelete := DeniedDo(verifier, obj, ctx.R, presets.PermDelete)
		return h.Td().Children(
			v.VBtn(msgr.Rename).Disabled(disable || deniedUpdate).PrependIcon("mdi-rename-box").Size(v.SizeXSmall).Color(v.ColorPrimary).Variant(v.VariantText).
				On("click.stop", web.Plaid().
					URL(listingHref).
					EventFunc(eventRenameVersionDialog).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, id).
					Query(paramVersionName, versionName).
					Go(),
				),
			v.VBtn(pmsgr.Delete).Disabled(disable || deniedDelete).PrependIcon("mdi-delete").Size(v.SizeXSmall).Color(v.ColorPrimary).Variant(v.VariantText).
				On("click.stop", web.Plaid().
					URL(listingHref).
					EventFunc(eventDeleteVersionDialog).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, id).
					Query(paramVersionName, versionName).
					Go(),
				),
		)
	})
	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.DisableModelListeners(true)
	lb.FooterAction("Cancel").ButtonCompFunc(func(evCtx *web.EventContext) h.HTMLComponent {
		utilsMsgr := i18n.MustGetModuleMessages(evCtx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		ctx := evCtx.R.Context()
		c := presets.ListingCompoFromContext(ctx)
		filter := MustFilterQuery(c)
		selected := filter.Get(filterKeySelected)
		filter.Del(filterKeySelected)
		return h.Components(
			web.Listen(
				mb.NotifModelsCreated(), stateful.ReloadAction(ctx, c, nil).Go(),
				mb.NotifModelsUpdated(), stateful.ReloadAction(ctx, c, nil).Go(),
				mb.NotifModelsDeleted(), fmt.Sprintf(`(payload, addon) => { %s%s }`, presets.ListingCompo_JsPreFixWhenNotifModelsDeleted,
					stateful.ReloadAction(ctx, c, nil, stateful.WithAppendFix(fmt.Sprintf(`
						if (payload.ids.includes(%q) && addon && addon.next_version_slug) {
							v.compo.filter_query = [%q, %q + "=" + addon.next_version_slug].join("&");
						}
					`, selected, filter.Encode(), filterKeySelected))).Go(),
				),
			),
			v.VBtn(utilsMsgr.Cancel).Variant(v.VariantElevated).Attr("@click", "vars.presetsListingDialog=false"),
		)
	})
	lb.FooterAction("OK").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		compo := presets.ListingCompoFromEventContext(ctx)
		selected := MustFilterQuery(compo).Get(filterKeySelected)
		return v.VBtn(utilsMsgr.OK).Disabled(selected == "").Variant(v.VariantElevated).Color(v.ColorSecondary).Attr("@click",
			fmt.Sprintf(`%s;%s;`,
				presets.CloseListingDialogVarScript,
				web.Emit(NotifVersionSelected(mb), PayloadVersionSelected{Slug: selected}),
			),
		)
	})
	lb.RowMenu().Empty()

	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:          "all",
				Invisible:    true,
				SQLCondition: ``,
			},
			{
				Key:          "online_versions",
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
		compo := presets.ListingCompoFromEventContext(ctx)
		selected := MustFilterQuery(compo).Get(filterKeySelected)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabAllVersions,
				ID:    "all",
				Query: url.Values{"all": []string{"1"}, "select_id": []string{selected}},
			},
			{
				Label: msgr.FilterTabOnlineVersion,
				ID:    "online_versions",
				Query: url.Values{"online_versions": []string{"1"}, "select_id": []string{selected}},
			},
			{
				Label: msgr.FilterTabNamedVersions,
				ID:    "named_versions",
				Query: url.Values{"named_versions": []string{"1"}, "select_id": []string{selected}},
			},
		}
	})
}
