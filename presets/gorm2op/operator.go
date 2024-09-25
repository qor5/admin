package gorm2op

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/samber/lo"
	relay "github.com/theplant/gorelay"
	"github.com/theplant/gorelay/cursor"
	"github.com/theplant/gorelay/gormrelay"
	"gorm.io/gorm"
)

var wildcardReg = regexp.MustCompile(`[%_]`)

func DataOperator(db *gorm.DB) (r *DataOperatorBuilder) {
	r = &DataOperatorBuilder{db: db}
	return
}

type ctxKeyDB struct{}

type DataOperatorBuilder struct {
	db *gorm.DB
}

func (op *DataOperatorBuilder) Search(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
	ilike := "ILIKE"
	if op.db.Dialector.Name() == "sqlite" {
		ilike = "LIKE"
	}

	wh := op.db.Model(params.Model)
	if len(params.KeywordColumns) > 0 && len(params.Keyword) > 0 {
		var segs []string
		var args []interface{}
		for _, c := range params.KeywordColumns {
			segs = append(segs, fmt.Sprintf("%s %s ?", c, ilike))
			kw := wildcardReg.ReplaceAllString(params.Keyword, `\$0`)
			args = append(args, fmt.Sprintf("%%%s%%", kw))
		}
		wh = wh.Where(strings.Join(segs, " OR "), args...)
	}

	for _, cond := range params.SQLConditions {
		wh = wh.Where(strings.Replace(cond.Query, " ILIKE ", " "+ilike+" ", -1), cond.Args...)
	}

	var p relay.Pagination[any]
	var req *relay.PaginateRequest[any]
	if params.RelayPagination != nil {
		ctx.WithContextValue(ctxKeyDB{}, wh)
		p, err = params.RelayPagination(ctx)
		if err != nil {
			return nil, err
		}
		req = params.RelayPaginateRequest
		if req == nil {
			return nil, errors.New("RelayPaginateRequest is required")
		}
	} else {
		if params.RelayPaginateRequest != nil {
			return nil, errors.New("RelayPagination is required")
		}

		p = relay.New(
			true, // nodesOnly
			presets.PerPageMax, presets.PerPageDefault,
			gormrelay.NewOffsetAdapter[any](wh),
		)
		req = &relay.PaginateRequest[any]{
			OrderBys: params.OrderBys,
		}
		if params.PerPage > 0 {
			req.First = lo.ToPtr(int(params.PerPage))
			page := params.Page
			if page == 0 {
				page = 1
			}
			offset := int((page - 1) * params.PerPage)
			if offset > 0 {
				req.After = lo.ToPtr(cursor.EncodeOffsetCursor(offset - 1))
			}
		}
	}

	resp, err := p.Paginate(ctx.R.Context(), req)
	if err != nil {
		return
	}

	// []any => []modelType
	nodes := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(params.Model)), len(resp.Nodes), len(resp.Nodes))
	for i := 0; i < len(resp.Nodes); i++ {
		nodes.Index(i).Set(reflect.ValueOf(resp.Nodes[i]))
	}
	return &presets.SearchResult{
		PageInfo: resp.PageInfo,
		Nodes:    nodes.Interface(),
	}, nil
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
	err = op.saveOrUpdate(obj, id)
	return
}

func (op *DataOperatorBuilder) saveOrUpdate(obj interface{}, id string) (err error) {
	var count int64
	if op.primarySluggerWhere(obj, id).Count(&count).Error != nil {
		return
	}
	if count > 0 {
		return op.primarySluggerWhere(obj, id).Select("*").Updates(obj).Error
	}
	return op.primarySluggerWhere(obj, id).Save(obj).Error
}

func (op *DataOperatorBuilder) Delete(obj interface{}, id string, ctx *web.EventContext) (err error) {
	err = op.primarySluggerWhere(obj, id).Delete(obj).Error
	return
}
