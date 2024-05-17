package presets

import (
	"net/http"
	"net/url"

	"github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type (
	ComponentFunc             func(ctx *web.EventContext) h.HTMLComponent
	ObjectComponentFunc       func(obj interface{}, ctx *web.EventContext) h.HTMLComponent
	TabComponentFunc          func(obj interface{}, ctx *web.EventContext) (tab h.HTMLComponent, content h.HTMLComponent)
	EditingTitleComponentFunc func(obj interface{}, defaultTitle string, ctx *web.EventContext) h.HTMLComponent
)

type FieldComponentFunc func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent

type (
	ActionComponentFunc func(id string, ctx *web.EventContext) h.HTMLComponent
	ActionUpdateFunc    func(id string, ctx *web.EventContext) (err error)
)

type (
	BulkActionComponentFunc                  func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent
	BulkActionUpdateFunc                     func(selectedIds []string, ctx *web.EventContext) (err error)
	BulkActionSelectedIdsProcessorFunc       func(selectedIds []string, ctx *web.EventContext) (processedSelectedIds []string, err error)
	BulkActionSelectedIdsProcessorNoticeFunc func(selectedIds []string, processedSelectedIds []string, unactionableIds []string) string
)

type MessagesFunc func(r *http.Request) *Messages

// Data Layer
type DataOperator interface {
	Search(obj interface{}, params *SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error)
	// return ErrRecordNotFound if record not found
	Fetch(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error)
	Save(obj interface{}, id string, ctx *web.EventContext) (err error)
	Delete(obj interface{}, id string, ctx *web.EventContext) (err error)
}

type (
	SetterFunc      func(obj interface{}, ctx *web.EventContext)
	FieldSetterFunc func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error)
	ValidateFunc    func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors)
)

type (
	SearchFunc func(model interface{}, params *SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error)
	FetchFunc  func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error)
	SaveFunc   func(obj interface{}, id string, ctx *web.EventContext) (err error)
	DeleteFunc func(obj interface{}, id string, ctx *web.EventContext) (err error)
)

type SQLCondition struct {
	Query string
	Args  []interface{}
}

type SearchParams struct {
	KeywordColumns []string
	Keyword        string
	SQLConditions  []*SQLCondition
	PerPage        int64
	Page           int64
	OrderBy        string
	PageURL        *url.URL
}

type SlugDecoder interface {
	PrimaryColumnValuesBySlug(slug string) map[string]string
}

type SlugEncoder interface {
	PrimarySlug() string
}

type FilterDataFunc func(ctx *web.EventContext) vuetifyx.FilterData

type FilterTab struct {
	ID    string
	Label string
	// render AdvancedLabel if it is not nil
	AdvancedLabel h.HTMLComponent
	Query         url.Values
}

type FilterTabsFunc func(ctx *web.EventContext) []*FilterTab

type Plugin interface {
	Install(pb *Builder) (err error)
}

type ModelPlugin interface {
	ModelInstall(pb *Builder, mb *ModelBuilder) (err error)
}

type ModelInstallFunc func(pb *Builder, mb *ModelBuilder) error
