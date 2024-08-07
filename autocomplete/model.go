package autocomplete

import (
	"encoding/json"
	"fmt"
	"github.com/qor5/web/v3"
	"math"
	"net/http"
	"reflect"
	"strings"
)

type (
	ModelBuilder struct {
		p            *Builder
		uriName      string
		model        any
		sQLCondition string
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
func (b *ModelBuilder) Paging(v bool) *ModelBuilder {
	b.paging = v
	return b
}

func (b *ModelBuilder) JsonHref() string {
	return fmt.Sprintf("%s/%s", b.p.prefix, b.uriName)
}

func (mb *ModelBuilder) NewModel() (r interface{}) {
	return reflect.New(mb.modelType.Elem()).Interface()
}

func (mb *ModelBuilder) NewModelSlice() (r interface{}) {
	return reflect.New(reflect.SliceOf(mb.modelType)).Interface()
}

func (b *ModelBuilder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if b.paging {
		pageSize := ctx.ParamAsInt(ParamPageSize)
		page := ctx.ParamAsInt(ParamPage)
		if pageSize != 0 {
			if page == 0 {
				page = 1
			}
			g = g.Offset((page - 1) * pageSize).Limit(pageSize)
			response.Current = int64(page * pageSize)
			if response.Current > response.Total {
				response.Current = response.Total
			}
			response.Pages = int(math.Ceil(float64(response.Total) / float64(pageSize)))
		}

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
	return
}
