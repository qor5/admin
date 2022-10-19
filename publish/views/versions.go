package views

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/l10n"
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
		segs := strings.Split(ctx.R.FormValue("id"), "_")
		id := segs[0]

		if id == "" {
			return nil
		}

		c := h.Div()

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		ov := VCard(
			VCardTitle(h.Text(msgr.OnlineVersion)),
		)
		c.AppendChildren(ov)

		lv := map[string]interface{}{}
		db.Session(&gorm.Session{NewDB: true}).Model(mb.NewModel()).
			Where("id = ? AND status = ?", id, publish.StatusOnline).
			First(&lv)
		if len(lv) > 0 {
			tr := trBuilder(ctx, lv, segs[1])
			ov.AppendChildren(VSimpleTable(h.Tbody(tr)))
		}

		c.AppendChildren(h.Br())

		versionsList := VCard(
			VCardTitle(h.Text(msgr.VersionsList)),
		)
		c.AppendChildren(versionsList)

		var results []map[string]interface{}
		db.Session(&gorm.Session{NewDB: true}).Model(mb.NewModel()).
			Where("id = ?", id).Order("version DESC").
			Find(&results)

		tbody := h.Tbody()

		for _, r := range results {
			tr := trBuilder(ctx, r, segs[1])
			tbody.AppendChildren(tr)
		}

		versionsList.AppendChildren(VSimpleTable(tbody))

		return c
	}
}

func trBuilder(ctx *web.EventContext, r map[string]interface{}, versionName string) *h.HTMLTagBuilder {
	msgr := presets.MustGetMessages(ctx.R)

	attr := web.Plaid().EventFunc(switchVersionEvent).Query("id", fmt.Sprintf("%v_%v", r["id"], r["version"])).Go()
	tr := h.Tr(
		h.Td(h.Button(fmt.Sprint(r["version"]))).Attr("@click", attr),
		h.Td(h.Button(fmt.Sprint(r["status"]))).Attr("@click", attr),
		h.Td(VMenu(
			web.Slot(
				VBtn("").Children(
					VIcon("more_vert"),
				).Attr("v-on", "on").Text(true).Fab(true).Small(true),
			).Name("activator").Scope("{ on }"),

			VList(
				VListItem(
					VListItemIcon(VIcon("delete")),
					VListItemTitle(h.Text(msgr.Delete)),
				).Attr("@click", web.Plaid().
					EventFunc(actions.DeleteConfirmation).
					Query("id", fmt.Sprintf("%v_%v", r["id"], r["version"])).Go(),
				),
			).Dense(true),
		)),
	)
	if r["version"] == versionName {
		tr.Class("deep-purple white--text")
	}
	return tr
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
		if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
			if l10n.IsLocalizable(obj) {
				wh = wh.Where("locale_code = ?", localeCode)
			}
		}

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
