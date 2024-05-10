package publish

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	SchedulePublishDialogPortalName = "publish_SchedulePublishDialogPortalName"
	PublishCustomDialogPortalName   = "publish_PublishCustomDialogPortalName"
)

func sidePanel(db *gorm.DB, mb *presets.ModelBuilder) presets.ObjectComponentFunc {
	return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
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
		version.Version = v.(VersionInterface).GetVersion()
		version.VersionName = v.(VersionInterface).GetVersionName()
		version.Status = v.(StatusInterface).GetStatus()

		if version.Status == StatusOnline {
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

func switchVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder) web.EventFunc {
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

func saveNewVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder) web.EventFunc {
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

func duplicateVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder) web.EventFunc {
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
