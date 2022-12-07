package gormop

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
)

func DataOperator(db *gorm.DB) (r *DataOperatorBuilder) {
	r = &DataOperatorBuilder{db: db}
	return
}

type DataOperatorBuilder struct {
	db *gorm.DB
}

func (op *DataOperatorBuilder) Search(obj interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
	ilike := "ILIKE"
	if op.db.Dialect().GetName() == "sqlite3" {
		ilike = "LIKE"
	}

	wh := op.db.Model(obj)
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

	err = wh.Count(&totalCount).Error
	if err != nil {
		return
	}

	if params.PerPage > 0 {
		wh = wh.Limit(params.PerPage)
		page := params.Page
		if page == 0 {
			page = 1
		}
		offset := (page - 1) * params.PerPage
		wh = wh.Offset(offset)
	}

	orderBy := params.OrderBy
	if len(orderBy) > 0 {
		wh = wh.Order(orderBy)
	}

	err = wh.Find(obj).Error
	if err != nil {
		return
	}
	r = reflect.ValueOf(obj).Elem().Interface()
	return
}

func (op *DataOperatorBuilder) primarySluggerWhere(obj interface{}, id string) *gorm.DB {
	wh := op.db.Model(obj)

	if id == "" {
		return wh
	}

	if slugger, ok := obj.(presets.SlugDecoder); ok {
		cs := slugger.PrimaryColumnValuesBySlug(id)
		for _, cond := range cs {
			wh = wh.Where(fmt.Sprintf("%s = ?", cond[0]), cond[1])
		}
	} else {
		wh = wh.Where("id =  ?", id)
	}

	return wh
}

func (op *DataOperatorBuilder) Fetch(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
	err = op.primarySluggerWhere(obj, id).Find(obj).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, presets.ErrRecordNotFound
		}
		return
	}
	r = obj
	return
}

func (op *DataOperatorBuilder) Save(obj interface{}, id string, ctx *web.EventContext) (err error) {
	if id == "" {
		err = op.db.Create(obj).Error
		return
	}
	err = op.primarySluggerWhere(obj, id).Update(obj).Error
	return
}

func (op *DataOperatorBuilder) Delete(obj interface{}, id string, ctx *web.EventContext) (err error) {
	err = op.primarySluggerWhere(obj, id).Delete(obj).Error
	return
}
