package publish

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"slices"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/utils"
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

const VersionListDialogURISuffix = "-version-list-dialog"

type VersionComponentConfig struct {
	// If you want to use custom publish dialog, you can update the portal named PublishCustomDialogPortalName
	PublishEvent              func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	UnPublishEvent            func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	RePublishEvent            func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) string
	Top                       bool
	DisableListeners          bool
	DisableDataChangeTracking bool
	WrapActionButtons         func(ctx *web.EventContext, obj interface{}, actionButtons []h.HTMLComponent, phraseHasPresetsDataChanged string) []h.HTMLComponent
}

func DefaultVersionComponentFunc(mb *presets.ModelBuilder, cfg ...VersionComponentConfig) presets.FieldComponentFunc {
	var config VersionComponentConfig
	if len(cfg) > 0 {
		config = cfg[0]
	}
	phraseHasPresetsDataChanged := presets.PhraseHasPresetsDataChanged
	if config.DisableDataChangeTracking {
		phraseHasPresetsDataChanged = "false"
	}
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		primarySlugger, ok := obj.(presets.SlugEncoder)
		if !ok {
			panic("obj should be SlugEncoder")
		}
		slug := primarySlugger.PrimarySlug()
		actionButtons := []h.HTMLComponent{}

		div := h.Div().Class("tagList-bar-warp")

		div.AppendChildren(
			vx.VXDialog().
				Title(utilsMsgr.ModalTitleConfirm).
				Attr(":text", "locals.message").
				HideClose(true).
				OkText(utilsMsgr.OK).
				CancelText(utilsMsgr.Cancel).
				Attr("@click:ok", web.Plaid().URL(mb.Info().ListingHref()).EventFunc(web.Var("locals.action")).Query(presets.ParamID, slug).Go()).
				Attr("v-model", "locals.commonConfirmDialog"))

		if !config.Top {
			div.Class("pb-4")
		}

		urlSuffix := field.ModelInfo.URIName() + VersionListDialogURISuffix
		if _, ok := obj.(VersionInterface); ok {
			div.AppendChildren(buildVersionSwitch(obj, mb, slug, urlSuffix, msgr, phraseHasPresetsDataChanged))

			if !DeniedDo(mb.Info().Verifier(), obj, ctx.R, presets.PermUpdate, PermDuplicate) {
				div.AppendChildren(v.VBtn(msgr.Duplicate).
					Height(36).Class("ml-2").Variant(v.VariantOutlined).
					Attr(":disabled", phraseHasPresetsDataChanged).
					Attr("@click", fmt.Sprintf(`locals.action=%q;locals.commonConfirmDialog = true;locals.message = %q`, EventDuplicateVersion, msgr.ConfirmDuplicate)))
			}
		}
		verifier := mb.Info().Verifier()
		deniedPublish := DeniedDo(verifier, obj, ctx.R, PermPublish)
		deniedUnpublish := DeniedDo(verifier, obj, ctx.R, PermUnpublish)

		if _, ok := obj.(StatusInterface); ok {
			actionButtons = append(actionButtons, buildPublishButton(obj, field, ctx, config, msgr, phraseHasPresetsDataChanged, deniedPublish, deniedUnpublish))
		}

		if _, ok := obj.(ScheduleInterface); ok {
			actionButtons = append(actionButtons, buildScheduleButton(obj, ctx, mb, slug, config, msgr, phraseHasPresetsDataChanged, deniedPublish, deniedUnpublish))
		}
		if config.WrapActionButtons != nil {
			actionButtons = config.WrapActionButtons(ctx, obj, actionButtons, phraseHasPresetsDataChanged)
		}
		for _, actionButton := range actionButtons {
			div.AppendChildren(actionButton)
		}
		children := []h.HTMLComponent{div}
		if !config.DisableListeners {
			children = append(children,
				NewListenerVersionSelected(ctx, mb, slug),
				NewListenerModelsDeleted(mb, slug),
			)
		}
		return web.Scope(children...).VSlot(" { locals } ").Init(`{action: "", commonConfirmDialog: false ,message: ""}`)
	}
}

func buildVersionSwitch(obj interface{}, mb *presets.ModelBuilder, slug, urlSuffix string, msgr *Messages, phraseHasPresetsDataChanged string) h.HTMLComponent {
	version, ok := obj.(VersionInterface)
	if !ok {
		return nil
	}
	versionSwitch := v.VChip(
		h.Span(`{{ xlocals.versionName }}`).Class("ellipsis"),
	).Label(true).Variant(v.VariantOutlined).
		Attr("style", "height:36px;").
		Color(v.ColorAbsGreyDarken3).
		Attr(":disabled", phraseHasPresetsDataChanged).
		On("click", web.Plaid().EventFunc(actions.OpenListingDialog).
			URL(mb.Info().PresetsPrefix()+"/"+urlSuffix).
			Query(filterKeySelected, slug).
			Go()).
		Class(v.W100, "version-select-wrap")
	if status, ok := obj.(StatusInterface); ok {
		versionSwitch.AppendChildren(statusChip(status.EmbedStatus().Status, msgr).Class("mx-2 flex-shrink-0"))
	}
	versionSwitch.AppendIcon("mdi-menu-down").Size(16).Color(v.ColorAbsGreyDarken3)

	return web.Scope().VSlot(" { locals: xlocals } ").Init(fmt.Sprintf("{versionName: %q}", version.EmbedVersion().VersionName)).Children(
		versionSwitch,
		web.Listen(mb.NotifModelsUpdated(), fmt.Sprintf(`xlocals.versionName = payload.models[%q]?.VersionName ?? xlocals.versionName;`, slug)),
	)
}

func buildPublishButton(obj interface{}, field *presets.FieldContext, ctx *web.EventContext, config VersionComponentConfig, msgr *Messages, phraseHasPresetsDataChanged string, deniedPublish, deniedUnpublish bool) h.HTMLComponent {
	status, ok := obj.(StatusInterface)
	if !ok {
		return nil
	}

	var publishBtn h.HTMLComponent
	switch status.EmbedStatus().Status {
	case StatusDraft, StatusOffline:
		if !deniedPublish {
			publishEvent := fmt.Sprintf(`locals.action=%q;locals.commonConfirmDialog = true;locals.message = %q`, EventPublish, msgr.ConfirmPublish)
			if config.PublishEvent != nil {
				publishEvent = config.PublishEvent(obj, field, ctx)
			}
			publishBtn = h.Div(
				v.VBtn(msgr.Publish).
					Attr(":disabled", phraseHasPresetsDataChanged).
					Attr("@click", publishEvent).Class("ml-2").
					ClassIf("rounded", config.Top).ClassIf("rounded-0 rounded-s", !config.Top).
					Variant(v.VariantElevated).Color(v.ColorPrimary).Height(36),
			)
		}
	case StatusOnline:
		var unPublishEvent, rePublishEvent string
		if !deniedUnpublish {
			unPublishEvent = fmt.Sprintf(`locals.action=%q;locals.commonConfirmDialog = true;locals.message = %q`, EventUnpublish, msgr.ConfirmUnpublish)
			if config.UnPublishEvent != nil {
				unPublishEvent = config.UnPublishEvent(obj, field, ctx)
			}
		}
		if !deniedPublish {
			rePublishEvent = fmt.Sprintf(`locals.action=%q;locals.commonConfirmDialog = true;locals.message = %q`, EventRepublish, msgr.ConfirmRepublish)
			if config.RePublishEvent != nil {
				rePublishEvent = config.RePublishEvent(obj, field, ctx)
			}
		}
		if unPublishEvent != "" || rePublishEvent != "" {
			publishBtn = h.Div(
				h.Iff(unPublishEvent != "", func() h.HTMLComponent {
					return v.VBtn(msgr.Unpublish).
						Attr(":disabled", phraseHasPresetsDataChanged).
						Attr("@click", unPublishEvent).
						Class("ml-2").Variant(v.VariantElevated).Color(v.ColorError).Height(36)
				}),
				h.Iff(rePublishEvent != "", func() h.HTMLComponent {
					return v.VBtn(msgr.Republish).
						Attr(":disabled", phraseHasPresetsDataChanged).
						Attr("@click", rePublishEvent).Class("ml-2").
						ClassIf("rounded", config.Top).ClassIf("rounded-0 rounded-s", !config.Top).
						Variant(v.VariantElevated).Color(v.ColorPrimary).Height(36)
				}),
			).Class("d-inline-flex")
		}
	}
	if publishBtn != nil {
		var compos h.HTMLComponents = []h.HTMLComponent{publishBtn}
		// Publish/Unpublish/Republish CustomDialog
		if config.UnPublishEvent != nil || config.RePublishEvent != nil || config.PublishEvent != nil {
			compos = append(compos, web.Portal().Name(PortalPublishCustomDialog))
		}
		return compos
	}
	return nil
}

func buildScheduleButton(obj interface{}, ctx *web.EventContext, mb *presets.ModelBuilder, slug string, config VersionComponentConfig, msgr *Messages, phraseHasPresetsDataChanged string, deniedPublish, deniedUnpublish bool) h.HTMLComponent {
	_, ok := obj.(ScheduleInterface)
	if !ok {
		return nil
	}

	deniedSchedule := deniedPublish || deniedUnpublish || DeniedDo(mb.Info().Verifier(), obj, ctx.R, PermSchedule)
	if !deniedSchedule {
		var scheduleBtn h.HTMLComponent
		clickEvent := web.Plaid().
			EventFunc(eventSchedulePublishDialog).
			Query(presets.ParamOverlay, actions.Dialog).
			Query(presets.ParamID, slug).
			URL(mb.Info().ListingHref()).Go()
		if config.Top {
			scheduleBtn = v.VAutocomplete().PrependInnerIcon("mdi-alarm").Density(v.DensityCompact).
				Variant(v.FieldVariantSoloFilled).ModelValue(msgr.SchedulePublishTime).
				BgColor(v.ColorPrimaryLighten2).Readonly(true).
				Width(500).HideDetails(true).
				Attr(":disabled", phraseHasPresetsDataChanged).
				Attr("@click", clickEvent).Class("ml-2 text-caption page-builder-autoCmp")
		} else {
			scheduleBtn = v.VBtn("").Size(v.SizeSmall).Children(v.VIcon("mdi-alarm").Size(v.SizeXLarge)).Rounded("0").Class("rounded-e ml-abs-1").
				Variant(v.VariantElevated).Color(v.ColorPrimary).Width(36).Height(36).
				Attr(":disabled", phraseHasPresetsDataChanged).
				Attr("@click", clickEvent)
		}
		var compos h.HTMLComponents = []h.HTMLComponent{scheduleBtn}
		// SchedulePublishDialog
		compos = append(compos, web.Portal().Name(PortalSchedulePublishDialog))
		return compos
	}
	return nil
}

func DefaultVersionBar(db *gorm.DB) presets.ObjectComponentFunc {
	return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		res := h.Div().Class("d-inline-flex align-center")
		version := Version{}
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
		res.AppendChildren(v.VChip(h.Span(currentVersionStr)).Density(v.DensityComfortable).Color(v.ColorSuccess).Size(v.SizeSmall))

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
		nextText := fmt.Sprintf("%s: %s", msgr.NextVersion, version.GetNextVersion(nextObj.(ScheduleInterface).EmbedSchedule().ScheduledStartAt))
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

var VersionListDialogStatusSortOrderComputedField = "PublishStatusSortOrder"

func configureVersionListDialog(db *gorm.DB, pb *Builder, b *presets.Builder, pm *presets.ModelBuilder) {
	// actually, VersionListDialog is a listing
	// use this URL : URLName-version-list-dialog
	mb := b.Model(pm.NewModel()).
		URIName(pm.Info().URIName() + VersionListDialogURISuffix).
		InMenu(false)

	mb.WrapVerifier(func(in func() *perm.Verifier) func() *perm.Verifier {
		return func() *perm.Verifier {
			v := mb.GetPresetsBuilder().GetVerifier().Spawn()
			return v.SnakeOn(pm.Info().URIName())
		}
	})

	listingHref := mb.Info().ListingHref()
	registerEventFuncsForVersion(mb, db)
	listingFields := []string{"Version", "Status", "StartAt", "EndAt", "Option"}
	if pb.ab != nil {
		defer func() {
			pb.ab.RegisterModel(mb).LabelFunc(func() string {
				return pm.Info().URIName()
			})
		}()
		listingFields = []string{"Version", "Status", "StartAt", "EndAt", activity.ListFieldNotes, "Option"}
	}

	lb := mb.Listing(listingFields...).
		DialogWidth("900px").
		Title(func(evCtx *web.EventContext, _ presets.ListingStyle, _ string) (string, h.HTMLComponent, error) {
			msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPublishKey, Messages_en_US).(*Messages)
			return msgr.VersionsList, nil, nil
		}).
		WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
			msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPublishKey, Messages_en_US).(*Messages)
			return map[string]string{
				"Version": msgr.HeaderVersion,
				"Status":  msgr.HeaderStatus,
				"StartAt": msgr.HeaderStartAt,
				"EndAt":   msgr.HeaderEndAt,
				"Option":  msgr.HeaderOption,
			}, nil
		})).
		SearchColumns("version", "version_name").
		PerPage(10).
		DefaultOrderBy(relay.Order{Field: "CreatedAt", Direction: relay.OrderDirectionDesc}).
		WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
			return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
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
				if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
					con := presets.SQLCondition{
						Query: "locale_code = ?",
						Args:  []interface{}{localeCode},
					}
					params.SQLConditions = append(params.SQLConditions, &con)
				}

				oldR := ctx.R
				defer func() {
					ctx.R = oldR // restore the original request context
				}()
				ctx = gorm2op.EventContextWithRelayComputedHook(ctx, func(computed *gormrelay.Computed[any]) *gormrelay.Computed[any] {
					computed.Columns[VersionListDialogStatusSortOrderComputedField] = clause.Column{
						Name: fmt.Sprintf("(CASE WHEN status = '%s' THEN 0 WHEN status = '%s' THEN 2 ELSE 1 END)", StatusOnline, StatusOffline),
						Raw:  true,
					}
					return computed
				})
				ctx = gorm2op.EventContextWithRelayPaginationHooks(ctx, func(next relay.Paginator[any]) relay.Paginator[any] {
					return relay.PaginatorFunc[any](func(ctx context.Context, req *relay.PaginateRequest[any]) (*relay.Connection[any], error) {
						req.OrderBy = slices.Concat([]relay.Order{
							{Field: VersionListDialogStatusSortOrderComputedField, Direction: relay.OrderDirectionAsc},
						}, req.OrderBy)
						return next.Paginate(ctx, req)
					})
				})
				return in(ctx, params)
			}
		})
	lb.WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
		return func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
			compo := presets.ListingCompoFromEventContext(evCtx)
			filter := MustFilterQuery(compo)
			slug := obj.(presets.SlugEncoder).PrimarySlug()
			filter.Set(filterKeySelected, slug)
			cell.SetAttr("@click", stateful.ReloadAction(evCtx.R.Context(), compo, func(target *presets.ListingCompo) {
				target.FilterQuery = filter.Encode()
			}).Go())
			return in(evCtx, cell, id, obj)
		}
	})
	lb.Field("Version").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		compo := presets.ListingCompoFromEventContext(ctx)
		filter := MustFilterQuery(compo)
		selected := filter.Get(filterKeySelected)
		slug := obj.(presets.SlugEncoder).PrimarySlug()
		filter.Set(filterKeySelected, slug)
		return h.Td().Children(
			h.Div().Class("d-inline-flex align-center").Children(
				v.VRadio().ModelValue(slug).TrueValue(selected).Attr("@click.native.stop", true).Attr("@change",
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
		disablement := pb.disablementCheckFunc(ctx, obj)
		verifier := mb.Info().Verifier()
		deniedUpdate := DeniedDo(verifier, obj, ctx.R, presets.PermUpdate)
		deniedDelete := DeniedDo(verifier, obj, ctx.R, presets.PermDelete)
		return h.Td().Children(
			v.VBtn(msgr.Rename).Disabled(disablement.DisabledRename || deniedUpdate).PrependIcon("mdi-rename-box").Size(v.SizeXSmall).Color(v.ColorPrimary).Variant(v.VariantText).
				On("click.stop", web.Plaid().
					URL(listingHref).
					EventFunc(eventRenameVersionDialog).
					Query(presets.ParamOverlay, actions.Dialog).
					Query(presets.ParamID, id).
					Query(paramVersionName, versionName).
					Go(),
				),
			v.VBtn(pmsgr.Delete).Disabled(disablement.DisabledDelete || deniedDelete).PrependIcon("mdi-delete").Size(v.SizeXSmall).Color(v.ColorPrimary).Variant(v.VariantText).
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
			v.VBtn(utilsMsgr.Cancel).Variant(v.VariantOutlined).Size(v.SizeSmall).
				Class("text-none text-caption font-weight-regular").
				Attr("@click", "vars.presetsListingDialog=false"),
		)
	})
	lb.FooterAction("OK").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)

		compo := presets.ListingCompoFromEventContext(ctx)
		selected := MustFilterQuery(compo).Get(filterKeySelected)

		return v.VBtn(utilsMsgr.OK).Disabled(selected == "").Variant(v.VariantFlat).Size(v.SizeSmall).Color(v.ColorPrimary).
			Class("text-none text-caption font-weight-regular").Attr("@click",
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
				Label: msgr.FilterTabNamedVersions,
				ID:    "named_versions",
				Query: url.Values{"named_versions": []string{"1"}, "select_id": []string{selected}},
			},
		}
	})
}
