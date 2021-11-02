package views

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
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

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		c := h.Div()

		ov := VCard(
			VCardTitle(h.Text(msgr.OnlineVersion)),
		)
		c.AppendChildren(ov)

		lv := map[string]interface{}{}
		db.Session(&gorm.Session{NewDB: true}).Model(mb.NewModel()).
			Where("id = ? AND status = ?", id, publish.StatusOnline).
			First(&lv)
		if len(lv) > 0 {
			tr := h.Tr(
				h.Td(h.Button(fmt.Sprint(lv["version"]))),
				h.Td(h.Button(fmt.Sprint(lv["status"]))),
			).Attr("@click", web.Plaid().EventFunc(switchVersionEvent).Query("id", fmt.Sprintf("%v_%v", lv["id"], lv["version"])).Go())
			if lv["version"] == segs[1] {
				tr.Class("deep-purple white--text")
			}

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
			tr := h.Tr(
				h.Td(h.Button(fmt.Sprint(r["version"]))),
				h.Td(h.Button(fmt.Sprint(r["status"]))),
			).Attr("@click", web.Plaid().EventFunc(switchVersionEvent).Query("id", fmt.Sprintf("%v_%v", r["id"], r["version"])).Go())
			if r["version"] == segs[1] {
				tr.Class("deep-purple white--text")
			}
			tbody.AppendChildren(tr)
		}

		versionsList.AppendChildren(VSimpleTable(tbody))

		return c
	}
}

func switchVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		eb := mb.Editing()

		obj := mb.NewModel()
		obj, err = eb.Fetcher(obj, id, ctx)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		eb.UpdateOverlayContent(ctx, &r, obj, msgr.SwitchedToNewVersion, err)
		return
	}
}

func saveNewVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		segs := strings.Split(ctx.R.FormValue("id"), "_")
		id := segs[0]

		var newObj = mb.NewModel()
		// don't panic for fields that set in SetterFunc
		_ = ctx.UnmarshalForm(newObj)

		var obj = mb.NewModel()

		me := mb.Editing()
		if me.Setter != nil {
			me.Setter(obj, ctx)
		}

		vErr := me.RunSetterFunc(ctx, &r, obj, newObj)

		if vErr.HaveErrors() {
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			return
		}
		if err = reflectutils.Set(obj, "ID", uint(intID)); err != nil {
			return
		}

		version := db.NowFunc().Format("2006-01-02")
		var count int64
		db.Model(newObj).Where("id = ? AND version like ?", id, version+"%").Count(&count)

		if err = reflectutils.Set(obj, "Version.Version", fmt.Sprintf("%s-v%02v", version, count+1)); err != nil {
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

		var c int64
		sql := fmt.Sprintf("(%v.id, %v.version) IN (SELECT %v.id, MAX(%v.version) FROM %v GROUP BY %v.id)", tn, tn, tn, tn, tn, tn)
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
