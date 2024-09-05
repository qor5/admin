package autocomplete

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"

	"github.com/qor5/web/v3"
)

type (
	ModelBuilder struct {
		p            *Builder
		uriName      string
		model        any
		sQLCondition string
		orderBy      string
		modelType    reflect.Type
		paging       bool
		columns      []string
	}

	Response struct {
		Data    []map[string]interface{} `json:"data"`
		Total   int64                    `json:"total"`
		Pages   int                      `json:"pages"`
		Current int64                    `json:"current"`
	}
)

const (
	ParamPage       = "page"
	ParamPageSize   = "pageSize"
	ParamSearch     = "search"
	ResponseItems   = "data"
	ResponseTotal   = "total"
	ResponsePages   = "pages"
	ResponseCurrent = "current"
)

func (b *ModelBuilder) Columns(v ...string) *ModelBuilder {
	b.columns = v
	return b
}

func (b *ModelBuilder) SQLCondition(v string) *ModelBuilder {
	b.sQLCondition = v
	return b
}

func (b *ModelBuilder) UriName(v string) *ModelBuilder {
	b.uriName = v
	return b
}

func (b *ModelBuilder) OrderBy(v string) *ModelBuilder {
	b.orderBy = v
	return b
}

func (b *ModelBuilder) Paging(v bool) *ModelBuilder {
	b.paging = v
	return b
}

func (b *ModelBuilder) JsonHref() string {
	return fmt.Sprintf("%s/%s", b.p.prefix, b.uriName)
}

func (b *ModelBuilder) NewModel() (r interface{}) {
	return reflect.New(b.modelType.Elem()).Interface()
}

func (b *ModelBuilder) NewModelSlice() (r interface{}) {
	return reflect.New(reflect.SliceOf(b.modelType)).Interface()
}

func (b *ModelBuilder) crossOrigin(w http.ResponseWriter) {
	if !b.p.allowCrossOrigin {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (b *ModelBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.crossOrigin(w)
	var (
		db       = b.p.db
		response = Response{
			Data: []map[string]interface{}{},
		}
		g   = db.Model(b.NewModel())
		ctx = &web.EventContext{
			R: r,
			W: w,
		}
	)
	if b.sQLCondition != "" || ctx.Param(ParamSearch) != "" {
		g = g.Where(b.sQLCondition, fmt.Sprintf("%%%s%%", ctx.Param(ParamSearch)))
	}

	if err := g.Count(&response.Total).Error; err != nil {
		return
	}
	page := ctx.ParamAsInt(ParamPage)
	if page == 0 {
		page = 1
	}
	pageSize := ctx.ParamAsInt(ParamPageSize)
	if pageSize == 0 {
		pageSize = 20
	}
	g = g.Offset((page - 1) * pageSize).Limit(pageSize)
	if b.paging {
		response.Pages = int(math.Ceil(float64(response.Total) / float64(pageSize)))
	}
	response.Current = int64(page * pageSize)
	if response.Current > response.Total {
		response.Current = response.Total
	}
	if b.orderBy != "" {
		g = g.Order(b.orderBy)
	}
	if err := g.Select(strings.Join(b.columns, ",")).Find(&response.Data).Error; err != nil {
		return
	}
	// 将结构体编码为 JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
