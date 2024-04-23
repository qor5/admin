package views

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const SchedulePublishDialogPortalName = "SchedulePublishDialogPortalName"

func sidePanel(db *gorm.DB, mb *presets.ModelBuilder) presets.ComponentFunc {
	return func(ctx *web.EventContext) h.HTMLComponent {
		var (
			msgr                = i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
			activeClass         = "primary white--text"
			selected            = ctx.R.FormValue("selected")
			selectVersionsEvent = web.Plaid().EventFunc(selectVersionsEvent).Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).Query("selected", web.Var("$event")).Go()
			selectItems         = []map[string]string{
				{"text": msgr.AllVersions, "value": "all-versions"},
				{"text": msgr.NamedVersions, "value": "named-versions"},
			}
		)

		table, currentVersion, err := versionListTable(db, mb, msgr, ctx)
		if err != nil || table == nil {
			return nil
		}

		if selected == "" {
			selected = "all-versions"
		}

		var onlineVersionComp h.HTMLComponent
		if currentVersion != nil {
			onlineVersionComp = VTable(h.Tbody(h.Tr(h.Td(h.Text(currentVersion.VersionName)), h.Td(h.Text(currentVersion.Status))).Class(activeClass)))
		}

		return h.Div(
			VCard(
				VCardTitle(h.Text(msgr.OnlineVersion)),
				onlineVersionComp,
			),
			h.Br(),
			VCard(
				VCardTitle(
					h.Text(msgr.VersionsList),
				).Attr("style", "padding-bottom: 0px;"),
				VCardText(
					VSelect().
						Items(selectItems).
						Value(selected).
						On("change", selectVersionsEvent),
				).Attr("style", "padding-bottom: 0px;"),
				web.Portal(
					table,
				).Name("versions-list"),
			),
		)
	}
}

func findVersionItems(db *gorm.DB, mb *presets.ModelBuilder, ctx *web.EventContext, paramId string) (list interface{}, err error) {
	list = mb.NewModelSlice()
	primaryKeys, err := utils.GetPrimaryKeys(mb.NewModel(), db)
	if err != nil {
		return
	}
	err = utils.PrimarySluggerWhere(db.Session(&gorm.Session{NewDB: true}).Select(strings.Join(primaryKeys, ",")), mb.NewModel(), paramId, "version").
		Order("version DESC").
		Find(list).
		Error
	return list, err
}

type versionListTableItem struct {
	ID          string
	Version     string
	VersionName string
	Status      string
	ItemClass   string
	ParamID     string
}

func versionListTable(db *gorm.DB, mb *presets.ModelBuilder, msgr *Messages, ctx *web.EventContext) (table h.HTMLComponent, currentVersion *versionListTableItem, err error) {
	var obj = mb.NewModel()
	slugger := obj.(presets.SlugDecoder)
	paramID := ctx.R.FormValue(presets.ParamID)
	if paramID == "" {
		return nil, nil, nil
	}
	cs := slugger.PrimaryColumnValuesBySlug(paramID)
	id, currentVersionName := cs["id"], cs["version"]
	if id == "" || currentVersionName == "" {
		return nil, nil, fmt.Errorf("invalid version id: %s", paramID)
	}

	var (
		versions      []*versionListTableItem
		namedVersions []*versionListTableItem
		activeClass   = "vx-list-item--active primary--text"
		selected      = ctx.R.FormValue("selected")
		page          = ctx.R.FormValue("page")
		currentPage   = 1
	)

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			currentPage = p
		}
	}

	var results = mb.NewModelSlice()
	primaryKeys, err := utils.GetPrimaryKeys(mb.NewModel(), db)
	if err != nil {
		return
	}
	err = utils.PrimarySluggerWhere(db.Session(&gorm.Session{NewDB: true}).Select(strings.Join(append(primaryKeys, "version_name", "status"), ",")), mb.NewModel(), paramID, "version").
		Order("version DESC").
		Find(results).Error
	if err != nil {
		panic(err)
	}

	vO := reflect.ValueOf(results).Elem()

	for i := 0; i < vO.Len(); i++ {
		v := vO.Index(i).Interface()

		version := &versionListTableItem{}
		ID, _ := reflectutils.Get(v, "ID")
		version.ID = fmt.Sprintf("%v", ID)
		version.Version = v.(publish.VersionInterface).GetVersion()
		version.VersionName = v.(publish.VersionInterface).GetVersionName()
		version.Status = v.(publish.StatusInterface).GetStatus()

		if version.Status == publish.StatusOnline {
			currentVersion = version
		}

		version.Status = GetStatusText(version.Status, msgr)

		if version.Version == currentVersionName {
			version.ItemClass = activeClass
		}

		if version.VersionName == "" {
			version.VersionName = version.Version
		}

		if version.VersionName != version.Version {
			namedVersions = append(namedVersions, version)
		}
		version.ParamID = v.(presets.SlugEncoder).PrimarySlug()
		versions = append(versions, version)
	}

	if selected == "named-versions" {
		versions = namedVersions
	}

	var (
		switchVersionEvent = web.Plaid().EventFunc(switchVersionEvent).Query(presets.ParamID, web.Var(`$event.ParamID`)).Query("selected", selected).Query("page", web.Var("locals.versionPage")).Go()
		deleteVersionEvent = web.Plaid().EventFunc(actions.DeleteConfirmation).Query(presets.ParamID, web.Var(`props.item.ParamID`)).
					Query(presets.ParamAfterDeleteEvent, afterDeleteVersionEvent).
					Query("current_selected_id", ctx.R.FormValue(presets.ParamID)).
					Query("selected", selected).
					Query("page", web.Var("locals.versionPage")).
					Go() + ";event.stopPropagation();"
		renameVersionEvent = web.Plaid().EventFunc(renameVersionEvent).Query(presets.ParamID, web.Var(`props.item.ParamID`)).Query("name", web.Var("props.item.VersionName")).Go()
	)

	table = web.Scope(
		VDataTable(
			web.Slot(
				VDialog(
					VIcon("edit").Size(SizeSmall).Class("mr-2").Attr(":class", "props.item.ItemClass"),
					web.Slot(
						VTextField().Attr("v-model", "props.item.VersionName").Label(msgr.RenameVersion),
					).Name("input"),
				).Bind("return-value.sync", "props.item.VersionName").On("save", renameVersionEvent).Width(600).Transition("slide-x-reverse-transition"),
			).Name("item.Edit").Scope("props"),
			web.Slot(
				VIcon("delete").Size(SizeSmall).Class("mr-2").Attr("@click", deleteVersionEvent).Attr(":class", "props.item.ItemClass"),
			).Name("item.Delete").Scope("props"),
		).
			Items(versions).
			Headers(
				[]map[string]interface{}{
					{"title": "VersionName", "value": "VersionName", "width": "60%", "sortable": false},
					{"title": "Status", "value": "Status", "width": "20%", "sortable": false},
					{"title": "Edit", "value": "Edit", "width": "10%", "sortable": false},
					{"title": "Delete", "value": "Delete", "width": "10%", "sortable": false},
				}).
			// HideDefaultFooter(len(versions) <= 10).
			On("click:row", switchVersionEvent).
			On("pagination", "locals.versionPage = $event.page").
			ReturnObject(true).
			ItemsPerPageOptions([]int{5, 10, 20}).
			PageText("").ItemsPerPageText("").
			Page(currentPage),
	).Init(fmt.Sprintf(`{versionPage: %d}`, currentPage)).
		VSlot("{ locals }")

	return table, currentVersion, nil
}

func switchVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramId := ctx.R.FormValue(presets.ParamID)

		eb := mb.Editing()

		obj := mb.NewModel()
		obj, err = eb.Fetcher(obj, paramId, ctx)

		eb.UpdateOverlayContent(ctx, &r, obj, "", err)

		if ctx.Queries().Get("no_msg") == "true" {
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SwitchedToNewVersion, "")

		return
	}
}

func saveNewVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var toObj = mb.NewModel()
		slugger := toObj.(presets.SlugDecoder)
		currentVersionName := slugger.PrimaryColumnValuesBySlug(ctx.R.FormValue(presets.ParamID))["version"]
		paramID := ctx.R.FormValue(presets.ParamID)

		me := mb.Editing()
		vErr := me.RunSetterFunc(ctx, false, toObj)

		if vErr.HaveErrors() {
			me.UpdateOverlayContent(ctx, &r, toObj, "", &vErr)
			return
		}

		var fromObj = mb.NewModel()
		utils.PrimarySluggerWhere(db, mb.NewModel(), paramID).First(fromObj)
		if err = utils.SetPrimaryKeys(fromObj, toObj, db, paramID); err != nil {
			return
		}

		if err = reflectutils.Set(toObj, "Version.ParentVersion", currentVersionName); err != nil {
			return
		}

		if me.Validator != nil {
			if vErr := me.Validator(toObj, ctx); vErr.HaveErrors() {
				me.UpdateOverlayContent(ctx, &r, toObj, "", &vErr)
				return
			}
		}

		if err = me.Saver(toObj, paramID, ctx); err != nil {
			me.UpdateOverlayContent(ctx, &r, toObj, "", err)
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		if ctx.R.URL.Query().Get(presets.ParamInDialog) == "true" {
			web.AppendRunScripts(&r,
				"vars.presetsDialog = false",
				web.Plaid().
					URL(ctx.R.RequestURI).
					EventFunc(actions.UpdateListingDialog).
					StringQuery(ctx.R.URL.Query().Get(presets.ParamListingQueries)).
					Go(),
			)
		} else {
			r.Reload = true
		}

		return
	}
}

func duplicateVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var toObj = mb.NewModel()
		slugger := toObj.(presets.SlugDecoder)
		currentVersionName := slugger.PrimaryColumnValuesBySlug(ctx.R.FormValue(presets.ParamID))["version"]
		paramID := ctx.R.FormValue(presets.ParamID)
		me := mb.Editing()

		var fromObj = mb.NewModel()
		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), paramID).First(fromObj).Error; err != nil {
			return
		}
		if err = utils.SetPrimaryKeys(fromObj, toObj, db, paramID); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		if vErr := me.SetObjectFields(fromObj, toObj, &presets.FieldContext{
			ModelInfo: mb.Info(),
		}, false, presets.ContextModifiedIndexesBuilder(ctx).FromHidden(ctx.R), ctx); vErr.HaveErrors() {
			presets.ShowMessage(&r, vErr.Error(), "error")
			return
		}

		if err = reflectutils.Set(toObj, "Version.ParentVersion", currentVersionName); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		if me.Validator != nil {
			if vErr := me.Validator(toObj, ctx); vErr.HaveErrors() {
				presets.ShowMessage(&r, vErr.Error(), "error")
				return
			}
		}

		if err = me.Saver(toObj, paramID, ctx); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")
		se := toObj.(presets.SlugEncoder)
		newQueries := ctx.Queries()
		newQueries.Del(presets.ParamID)
		r.PushState = web.Location(newQueries).URL(mb.Info().DetailingHref(se.PrimarySlug()))
		return
	}
}

func searcher(db *gorm.DB, mb *presets.ModelBuilder) presets.SearchFunc {
	return func(obj interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		ilike := "ILIKE"
		if db.Dialector.Name() == "sqlite" {
			ilike = "LIKE"
		}

		wh := db.Model(obj)

		if len(params.KeywordColumns) > 0 && len(params.Keyword) > 0 {
			var segs []string
			var args []interface{}
			for _, c := range params.KeywordColumns {
				segs = append(segs, fmt.Sprintf("%s %s ?", c, ilike))
				args = append(args, fmt.Sprintf("%%%s%%", params.Keyword))
			}
			wh = wh.Where(strings.Join(segs, " OR "), args...)
		}

		for _, cond := range params.SQLConditions {
			wh = wh.Where(strings.Replace(cond.Query, " ILIKE ", " "+ilike+" ", -1), cond.Args...)
		}

		stmt := &gorm.Statement{DB: db}
		stmt.Parse(mb.NewModel())
		tn := stmt.Schema.Table

		var pks []string
		condition := ""
		var c int64
		for _, f := range stmt.Schema.Fields {
			if f.Name == "DeletedAt" {
				condition = "WHERE deleted_at IS NULL"
			}
		}
		for _, f := range stmt.Schema.PrimaryFields {
			if f.Name != "Version" {
				pks = append(pks, f.DBName)
			}
		}
		pkc := strings.Join(pks, ",")
		sql := fmt.Sprintf("(%v,version) IN (SELECT %v, MAX(version) FROM %v %v GROUP BY %v)", pkc, pkc, tn, condition, pkc)
		if err = wh.Where(sql).Count(&c).Error; err != nil {
			return
		}
		totalCount = int(c)

		if params.PerPage > 0 {
			wh = wh.Limit(int(params.PerPage))
			page := params.Page
			if page == 0 {
				page = 1
			}
			offset := (page - 1) * params.PerPage
			wh = wh.Offset(int(offset))
		}

		orderBy := params.OrderBy
		if len(orderBy) > 0 {
			wh = wh.Order(orderBy)
		}

		if err = wh.Find(obj).Error; err != nil {
			return
		}
		r = reflect.ValueOf(obj).Elem().Interface()
		return
	}
}

func versionActionsFunc(m *presets.ModelBuilder) presets.ObjectComponentFunc {
	return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		gmsgr := presets.MustGetMessages(ctx.R)
		var buttonLabel = gmsgr.Create
		m.RightDrawerWidth("800")
		var disableUpdateBtn bool
		var isCreateBtn = true
		if ctx.R.FormValue(presets.ParamID) != "" {
			isCreateBtn = false
			buttonLabel = gmsgr.Update
			m.RightDrawerWidth("1200")
			disableUpdateBtn = m.Info().Verifier().Do(presets.PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		updateBtn := VBtn(buttonLabel).
			Color("primary").
			Attr("@click", web.Plaid().
				EventFunc(actions.Update).
				Queries(ctx.Queries()).
				Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
				URL(m.Info().ListingHref()).
				Go(),
			)
		if disableUpdateBtn {
			updateBtn = updateBtn.Disabled(disableUpdateBtn)
		} else {
			updateBtn = updateBtn.Attr(":disabled", "isFetching").Attr(":loading", "isFetching")
		}
		if isCreateBtn {
			return h.Components(
				VSpacer(),
				updateBtn,
			)
		}

		saveNewVersionBtn := VBtn(msgr.SaveAsNewVersion).
			Color("secondary").
			Attr("@click", web.Plaid().
				EventFunc(SaveNewVersionEvent).
				Queries(ctx.Queries()).
				Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
				URL(m.Info().ListingHref()).
				Go(),
			)
		if disableUpdateBtn {
			saveNewVersionBtn = saveNewVersionBtn.Disabled(disableUpdateBtn)
		} else {
			saveNewVersionBtn = saveNewVersionBtn.Attr(":disabled", "isFetching").Attr(":loading", "isFetching")
		}

		return h.Components(
			VSpacer(),
			saveNewVersionBtn,
			updateBtn,
		)
	}
}

func DefaultVersionComponentFunc(b *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			version        publish.VersionInterface
			status         publish.StatusInterface
			primarySlugger presets.SlugEncoder
			ok             bool
			versionSwitch  *VChipBuilder
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
		).Class("w-100 d-inline-flex pa-6 pb-6")

		if version, ok = obj.(publish.VersionInterface); ok {
			versionSwitch = VChip(
				h.Text(version.GetVersionName()),
			).Label(true).Variant(VariantOutlined).Class("rounded-r-0 text-black").
				Attr("style", "height:40px;background-color:#FFFFFF!important;").
				Attr("@click", web.Plaid().EventFunc(actions.OpenListingDialog).
					URL(b.Info().PresetsPrefix()+"/"+field.ModelInfo.URIName()+"-version-list-dialog").
					Query("select_id", primarySlugger.PrimarySlug()).
					Go()).
				Class(W100)
			if status, ok = obj.(publish.StatusInterface); ok {
				versionSwitch.AppendChildren(VChip(h.Text(GetStatusText(status.GetStatus(), msgr))).Label(true).Color(GetStatusColor(status.GetStatus())).Size(SizeSmall).Class("px-1  mx-1 text-black ml-2"))
			}
			versionSwitch.AppendIcon("mdi-chevron-down")

			div.AppendChildren(versionSwitch)
			div.AppendChildren(VBtn("").Icon("mdi-file-document-multiple").
				Height(40).Color("white").Class("rounded-sm").Variant(VariantFlat).
				Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, SaveNewVersionEvent)))
		}

		if status, ok = obj.(publish.StatusInterface); ok {
			switch status.GetStatus() {
			case publish.StatusDraft, publish.StatusOffline:
				publishBtn = h.Div(
					VBtn(msgr.Publish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, PublishEvent)).
						Class("rounded-sm ml-2").Variant(VariantFlat).Color("primary").Height(40),
				)
			case publish.StatusOnline:
				publishBtn = h.Div(
					VBtn(msgr.Unpublish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, UnpublishEvent)).
						Class("rounded-sm ml-2").Variant(VariantFlat).Color(presets.ColorPrimary).Height(40),
					VBtn(msgr.Republish).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, RepublishEvent)).
						Class("rounded-sm ml-2").Variant(VariantFlat).Color(presets.ColorPrimary).Height(40),
				).Class("d-inline-flex")
			}
			div.AppendChildren(publishBtn)
		}

		if _, ok = obj.(publish.ScheduleInterface); ok {
			scheduleBtn := VBtn("").Icon("mdi-alarm").Class("rounded-sm ml-1").
				Variant(VariantFlat).Color("primary").Height(40).Attr("@click", web.POST().
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
			VRadio().ModelValue(p.PrimarySlug()).TrueValue(id).Attr("@change", web.Plaid().EventFunc(actions.UpdateListingDialog).
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
				VBadge().Content(count).Color("red"),
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

		return h.Td(VBtn("Delete").Disabled(disable).PrependIcon("mdi-delete").Size(SizeXSmall).Color("primary").Variant(VariantText).Attr("@click", web.Plaid().
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
		return VBtn("Cancel").Variant(VariantElevated).Attr("@click", "vars.presetsListingDialog=false")
	})
	lb.FooterAction("Save").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}
		return VBtn("Save").Variant(VariantElevated).Color("secondary").Attr("@click", web.Plaid().
			Query("select_id", id).
			URL(pm.Info().PresetsPrefix()+"/"+mb.Info().URIName()).
			EventFunc(selectVersionEventV2).
			Go())
	})
	lb.RowMenu().Empty()
	mb.RegisterEventFunc(selectVersionEventV2, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("select_id")
		refer, _ := url.Parse(ctx.R.Referer())
		newQueries := refer.Query()
		r.PushState = web.Location(newQueries).URL(pm.Info().DetailingHref(id))
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
