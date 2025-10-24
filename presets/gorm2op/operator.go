package gorm2op

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/hook"
	"github.com/samber/lo"
	"github.com/theplant/relay"
	"github.com/theplant/relay/cursor"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

var wildcardReg = regexp.MustCompile(`[%_]`)

func DataOperator(db *gorm.DB) (r *DataOperatorBuilder) {
	r = &DataOperatorBuilder{db: db}
	return
}

type (
	ctxKeyDBForRelay struct{}
	CtxKeyDB         struct{}
	ctxKeyHook       struct{}
)

func WithHook(ctx context.Context, hooks ...hook.Hook[*gorm.DB]) context.Context {
	previousHook, _ := ctx.Value(ctxKeyHook{}).(hook.Hook[*gorm.DB])
	hook := hook.Prepend(previousHook, hooks...)
	return context.WithValue(ctx, ctxKeyHook{}, hook)
}

func EventContextWithHook(ctx *web.EventContext, hooks ...hook.Hook[*gorm.DB]) *web.EventContext {
	previousHook, _ := ctx.ContextValue(ctxKeyHook{}).(hook.Hook[*gorm.DB])
	hook := hook.Prepend(previousHook, hooks...)
	return ctx.WithContextValue(ctxKeyHook{}, hook)
}

type DataOperatorBuilder struct {
	db *gorm.DB
}

func (op *DataOperatorBuilder) Search(evCtx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
	ilike := "ILIKE"
	db := op.getDB(evCtx)
	if db.Dialector.Name() == "sqlite" {
		ilike = "LIKE"
	}

	wh := db.Model(params.Model)
	if len(params.KeywordColumns) > 0 && params.Keyword != "" {
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
		wh = wh.Where(strings.ReplaceAll(cond.Query, " ILIKE ", " "+ilike+" "), cond.Args...)
	}

	dbHook, _ := evCtx.ContextValue(ctxKeyHook{}).(hook.Hook[*gorm.DB])
	if dbHook != nil {
		wh = dbHook(wh.Session(&gorm.Session{}))
	}

	var p relay.Paginator[any]
	var req *relay.PaginateRequest[any]
	ctx := evCtx.R.Context()
	if params.RelayPagination != nil {
		ctx = context.WithValue(ctx, ctxKeyDBForRelay{}, wh)
		p, err = params.RelayPagination(evCtx)
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

		opts, _ := ctx.Value(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
		opts = appendWithComputedIfHasHook(ctx, opts)
		p = relay.New(
			gormrelay.NewOffsetAdapter(wh, opts...),
			relay.EnsureLimits[any](presets.PerPageDefault, presets.PerPageMax),
		)
		req = &relay.PaginateRequest[any]{
			OrderBy: params.OrderBy,
		}
		if params.PerPage > 0 {
			req.First = lo.ToPtr(int(params.PerPage))
			page := params.Page
			if page <= 0 {
				page = 1
			}
			offset := int((page - 1) * params.PerPage)
			if offset > 0 {
				req.After = lo.ToPtr(cursor.EncodeOffsetCursor(offset - 1))
			}
		}
		ctx = relay.WithSkip(ctx, relay.Skip{Edges: true})
	}

	paginationHook, _ := ctx.Value(ctxKeyRelayPaginationHook{}).(hook.Hook[relay.Paginator[any]])
	if paginationHook != nil {
		p = paginationHook(p)
	}
	resp, err := p.Paginate(ctx, req)
	if err != nil {
		return
	}

	// []any => []modelType
	nodes := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(params.Model)), len(resp.Nodes), len(resp.Nodes))
	for i := 0; i < len(resp.Nodes); i++ {
		nodes.Index(i).Set(reflect.ValueOf(resp.Nodes[i]))
	}

	pageInfo := resp.PageInfo
	if pageInfo == nil {
		// In fact, the current scenario does not use SkipPageInfo, so it will not be triggered here.
		pageInfo = &relay.PageInfo{}
	}
	return &presets.SearchResult{
		PageInfo:   *pageInfo,
		TotalCount: resp.TotalCount,
		Nodes:      nodes.Interface(),
	}, nil
}

func (*DataOperatorBuilder) primarySluggerWhere(db *gorm.DB, obj interface{}, id string) *gorm.DB {
	wh := db.Model(obj)

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
	db := op.getDB(ctx)
	err = op.primarySluggerWhere(db, obj, id).First(obj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, presets.ErrRecordNotFound
		}
		return
	}
	r = obj
	return
}

func (op *DataOperatorBuilder) getDB(ctx *web.EventContext) *gorm.DB {
	if ctx.R != nil {
		db, ok := ctx.ContextValue(CtxKeyDB{}).(*gorm.DB)
		if ok {
			return db
		}
	}
	return op.db
}

func (op *DataOperatorBuilder) Save(obj interface{}, id string, ctx *web.EventContext) (err error) {
	db := op.getDB(ctx)
	if id == "" {
		err = db.Create(obj).Error
		return
	}
	err = op.saveOrUpdate(db, obj, id)
	return
}

func (op *DataOperatorBuilder) saveOrUpdate(db *gorm.DB, obj interface{}, id string) (err error) {
	var count int64
	if op.primarySluggerWhere(db, obj, id).Count(&count).Error != nil {
		return
	}
	if count > 0 {
		return op.primarySluggerWhere(db, obj, id).Select("*").Updates(obj).Error
	}
	return op.primarySluggerWhere(db, obj, id).Save(obj).Error
}

func (op *DataOperatorBuilder) Delete(obj interface{}, id string, ctx *web.EventContext) (err error) {
	db := op.getDB(ctx)

	err = op.primarySluggerWhere(db, obj, id).Delete(obj).Error
	return
}
