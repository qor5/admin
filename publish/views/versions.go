package views

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func sidePanel(db *gorm.DB, mb *presets.ModelBuilder) presets.ComponentFunc {
	return func(ctx *web.EventContext) h.HTMLComponent {
		var (
			msgr                = i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
			activeClass         = "deep-purple white--text"
			selected            = ctx.R.FormValue("selected")
			selectVersionsEvent = web.Plaid().EventFunc(selectVersionsEvent).Query("id", ctx.R.FormValue("id")).Query("selected", web.Var("$event")).Go()
			selectItems         = []map[string]string{
				{"text": msgr.AllVersions, "value": "all-versions"},
				{"text": msgr.NamedVersions, "value": "named-versions"},
			}
		)

		table, currentVersion, err := versionListTable(db, mb, msgr, ctx.R)
		if err != nil {
			return nil
		}

		if selected == "" {
			selected = "all-versions"
		}

		return h.Div(
			VCard(
				VCardTitle(h.Text(msgr.OnlineVersion)),
				VSimpleTable(h.Tbody(h.Tr(h.Td(h.Text(currentVersion.VersionName)), h.Td(h.Text(currentVersion.Status))).Class(activeClass))),
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

type versionListTableItem struct {
	ID          string
	Version     string
	VersionName string
	Status      string
	ItemClass   string
}

func versionListTable(db *gorm.DB, mb *presets.ModelBuilder, msgr *Messages, req *http.Request) (table h.HTMLComponent, currentVersion versionListTableItem, err error) {
	segs := strings.Split(req.FormValue("id"), "_")
	if len(segs) != 2 {
		return nil, currentVersion, fmt.Errorf("invalid version id: %s", req.FormValue("id"))
	}

	id, currentVersionName := segs[0], segs[1]
	if id == "" || currentVersionName == "" {
		return nil, currentVersion, fmt.Errorf("invalid version id: %s", req.FormValue("id"))
	}

	var (
		versions    []versionListTableItem
		activeClass = "deep-purple white--text"
		selected    = req.FormValue("selected")
		page        = req.FormValue("page")
		currentPage = 1
	)

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			currentPage = p
		}
	}

	if selected == "named-versions" {
		db.Session(&gorm.Session{NewDB: true}).Model(mb.NewModel()).Select("id,version,version_name,status").Where("id = ? and (version_name != version and version_name != '')", id).Order("created_at DESC").Find(&versions)
	} else {
		db.Session(&gorm.Session{NewDB: true}).Model(mb.NewModel()).Select("id,version,version_name,status").Where("id = ?", id).Order("created_at DESC").Find(&versions)
	}

	for index := range versions {
		if versions[index].Version == currentVersionName {
			versions[index].ItemClass = activeClass
		}

		if versions[index].VersionName == "" {
			versions[index].VersionName = versions[index].Version
		}

		if versions[index].Status == publish.StatusOnline {
			currentVersion = versions[index]
		}
	}

	var (
		swithVersionEvent  = web.Plaid().EventFunc(switchVersionEvent).Query("id", web.Var(`$event.ID+"_"+$event.Version`)).Query("selected", selected).Query("page", web.Var("locals.versionPage")).Go()
		deleteVersionEvent = web.Plaid().EventFunc(actions.DeleteConfirmation).Query("id", web.Var(`item.ID+"_"+item.Version`)).Go() + ";event.stopPropagation();"
		renameVersionEvent = web.Plaid().EventFunc(renameVersionEvent).Query("id", web.Var(`props.item.ID+"_"+props.item.Version`)).Query("name", web.Var("props.item.VersionName")).Go()
	)

	table = web.Scope(
		VDataTable(
			web.Slot(
				VIcon("delete").Small(true).Class("mr-2").Attr("@click", deleteVersionEvent),
			).Name("item.actions").Scope("{ item }"),
			web.Slot(
				VEditDialog(
					h.Text(" {{ props.item.VersionName }}"),
					web.Slot(
						VTextField().Attr("v-model", "props.item.VersionName").Label(msgr.RenameVersion),
					).Name("input"),
				).Bind("return-value.sync", "props.item.VersionName").On("save", renameVersionEvent),
			).Name("item.VersionName").Scope("props"),
		).
			Items(versions).
			Headers(
				[]map[string]interface{}{
					{"text": "VersionName", "value": "VersionName"},
					{"text": "Status", "value": "Status"},
					{"text": "Actions", "value": "actions"},
				}).
			HideDefaultHeader(true).
			On("click:row", swithVersionEvent).
			On("pagination", "locals.versionPage = $event.page").
			ItemClass("ItemClass").
			FooterProps(
				map[string]interface{}{
					"items-per-page-text":    "",
					"items-per-page-options": []int{5, 10, 20, -1},
				},
			).
			Page(currentPage),
	).Init(fmt.Sprintf(`{versionPage: %d}`, currentPage)).
		VSlot("{ locals }")

	return table, currentVersion, nil
}

func switchVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		eb := mb.Editing()

		obj := mb.NewModel()
		obj, err = eb.Fetcher(obj, id, ctx)

		eb.UpdateOverlayContent(ctx, &r, obj, "", err)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SwitchedToNewVersion, "")

		return
	}
}

func saveNewVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		segs := strings.Split(ctx.R.FormValue("id"), "_")
		id := segs[0]

		var obj = mb.NewModel()

		me := mb.Editing()
		vErr := me.RunSetterFunc(ctx, false, obj)

		if vErr.HaveErrors() {
			me.UpdateOverlayContent(ctx, &r, obj, "", &vErr)
			return
		}

		if err = reflectutils.Set(obj, "ID", id); err != nil {
			return
		}

		version := db.NowFunc().Format("2006-01-02")
		var count int64
		newObj := mb.NewModel()
		db.Model(newObj).Unscoped().Where("id = ? AND version like ?", id, version+"%").Count(&count)

		versionName := fmt.Sprintf("%s-v%02v", version, count+1)
		if err = reflectutils.Set(obj, "Version.Version", versionName); err != nil {
			return
		}
		if err = reflectutils.Set(obj, "Version.VersionName", versionName); err != nil {
			return
		}
		if err = reflectutils.Set(obj, "Version.ParentVersion", segs[1]); err != nil {
			return
		}

		if me.Validator != nil {
			if vErr := me.Validator(obj, ctx); vErr.HaveErrors() {
				me.UpdateOverlayContent(ctx, &r, obj, "", &vErr)
				return
			}
		}

		if err = me.Saver(obj, ctx.R.FormValue("id"), ctx); err != nil {
			me.UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")
		r.Reload = true

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
				pks = append(pks, stmt.Quote(strings.ToLower(f.Name)))
			}
		}
		pkc := strings.Join(pks, ",")
		sql := fmt.Sprintf("(%v.id, %v.version) IN (SELECT %v, MAX(version) FROM %v %v GROUP BY %v)", tn, tn, pkc, tn, condition, pkc)
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
		if ctx.R.FormValue("id") != "" {
			buttonLabel = gmsgr.Update
			m.RightDrawerWidth("1200")
			disableUpdateBtn = m.Info().Verifier().Do(presets.PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		updateBtn := VBtn(buttonLabel).
			Color("primary").
			Attr("@click", web.Plaid().
				EventFunc(actions.Update).Query("id", ctx.R.FormValue("id")).
				URL(m.Info().ListingHref()).
				Go(),
			)
		saveNewVersionBtn := VBtn(msgr.SaveAsNewVersion).
			Color("secondary").
			Attr("@click", web.Plaid().
				EventFunc(SaveNewVersionEvent).Query("id", ctx.R.FormValue("id")).
				URL(m.Info().ListingHref()).
				Go(),
			)
		if disableUpdateBtn {
			updateBtn = updateBtn.Disabled(disableUpdateBtn)
			saveNewVersionBtn = saveNewVersionBtn.Disabled(disableUpdateBtn)
		} else {
			updateBtn = updateBtn.Attr(":disabled", "isFetching").Attr(":loading", "isFetching")
			saveNewVersionBtn = saveNewVersionBtn.Attr(":disabled", "isFetching").Attr(":loading", "isFetching")
		}
		return h.Components(
			VSpacer(),
			saveNewVersionBtn,
			updateBtn,
		)
	}
}
