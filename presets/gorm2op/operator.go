package gorm2op

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
	"gorm.io/gorm"
)

var (
	wildcardReg = regexp.MustCompile(`[%_]`)
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
	text := "TEXT"
	if op.db.Dialector.Name() == "sqlite" {
		ilike = "LIKE"
	}
	if op.db.Dialector.Name() == "mysql" {
		text = "CHAR"
	}

	wh := op.db.Model(obj)
	if len(params.KeywordColumns) > 0 && len(params.Keyword) > 0 {
		var segs []string
		var args []interface{}
		for _, c := range params.KeywordColumns {
			segs = append(segs, fmt.Sprintf("CAST(%s AS %s) %s ?", c, text, ilike))
			kw := wildcardReg.ReplaceAllString(params.Keyword, `\$0`)
			args = append(args, fmt.Sprintf("%%%s%%", kw))
		}
		wh = wh.Where(strings.Join(segs, " OR "), args...)
	}

	for _, cond := range params.SQLConditions {
		wh = wh.Where(strings.Replace(cond.Query, " ILIKE ", " "+ilike+" ", -1), cond.Args...)
	}

	var c int64
	err = wh.Count(&c).Error
	if err != nil {
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
		for key, value := range cs {
			wh = wh.Where(fmt.Sprintf("%s = ?", key), value)
		}
	} else {
		wh = wh.Where("id =  ?", id)
	}

	return wh
}

func (op *DataOperatorBuilder) Fetch(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
	err = op.primarySluggerWhere(obj, id).First(obj).Error
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
	err = op.primarySluggerWhere(obj, id).Save(obj).Error
	return
}

func (op *DataOperatorBuilder) Delete(obj interface{}, id string, ctx *web.EventContext) (err error) {
	err = op.primarySluggerWhere(obj, id).Delete(obj).Error
	return
}
