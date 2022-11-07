package gorm2op

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/l10n"
	"github.com/qor/qor5/publish"
	"github.com/qor/qor5/utils"
	"gorm.io/gorm"
)

func Searcher(db *gorm.DB, mb *presets.ModelBuilder) presets.SearchFunc {
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

		var c int64
		if publish.IsVersion(obj) {
			stmt := &gorm.Statement{DB: db}
			stmt.Parse(mb.NewModel())
			tn := stmt.Schema.Table

			var pks []string
			condition := ""
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
		} else {
			err = wh.Count(&c).Error
			if err != nil {
				return
			}
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

func PrimarySluggerWhere(db *gorm.DB, obj interface{}, id string, ctx *web.EventContext, withoutKeys ...string) *gorm.DB {
	wh := db.Model(obj)

	if id == "" {
		return wh
	}

	if slugger, ok := obj.(presets.SlugDecoder); ok {
		cs := slugger.PrimaryColumnValuesBySlug(id)
		for _, cond := range cs {
			if !utils.Contains(withoutKeys, cond[0]) {
				wh = wh.Where(fmt.Sprintf("%s = ?", cond[0]), cond[1])
			}
		}
	} else {
		wh = wh.Where("id =  ?", id)
	}

	if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
		if l10n.IsLocalizable(obj) {
			wh = wh.Where("locale_code = ?", localeCode)
		}
	}

	return wh
}

func Fetcher(db *gorm.DB, mb *presets.ModelBuilder) presets.FetchFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		err = PrimarySluggerWhere(db, obj, id, ctx).First(obj).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, presets.ErrRecordNotFound
			}
			return
		}
		r = obj
		return
	}
}

func Saver(db *gorm.DB, mb *presets.ModelBuilder) presets.SaveFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
			if l10n.IsLocalizable(obj) {
				obj := obj.(l10n.L10nInterface)
				obj.SetLocale(localeCode.(string))
			}
		}
		if id == "" {
			err = db.Create(obj).Error
			return
		}
		err = PrimarySluggerWhere(db, obj, id, ctx).Save(obj).Error
		return
	}
}

func Deleter(db *gorm.DB, mb *presets.ModelBuilder) presets.DeleteFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		err = PrimarySluggerWhere(db, obj, id, ctx).Delete(obj).Error
		return
	}
}
