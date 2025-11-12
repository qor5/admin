package presets

import (
	"net/http"
	"net/url"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"
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
	ActionUpdateFunc    func(id string, ctx *web.EventContext, r *web.EventResponse) (err error)
)

type (
	BulkActionComponentFunc                  func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent
	BulkActionUpdateFunc                     func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error)
	BulkActionSelectedIdsProcessorFunc       func(selectedIds []string, ctx *web.EventContext) (processedSelectedIds []string, err error)
	BulkActionSelectedIdsProcessorNoticeFunc func(selectedIds []string, processedSelectedIds []string, unactionableIds []string) string
)

type MessagesFunc func(r *http.Request) *Messages

type DataOperator interface {
	Search(ctx *web.EventContext, params *SearchParams) (result *SearchResult, err error)
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
	SearchFunc func(ctx *web.EventContext, params *SearchParams) (result *SearchResult, err error)
	FetchFunc  func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error)
	SaveFunc   func(obj interface{}, id string, ctx *web.EventContext) (err error)
	DeleteFunc func(obj interface{}, id string, ctx *web.EventContext) (err error)
)

type SQLCondition struct {
	Query string
	Args  []interface{}
}

type (
	RelayPagination func(ctx *web.EventContext) (relay.Paginator[any], error)

	FilterOperator string
	FieldCondition struct {
		Field    string         `json:"field"`
		Operator FilterOperator `json:"operator"`
		Value    any            `json:"value"`
		Fold     bool           `json:"fold"`
	}
	Filter struct {
		And       []*Filter       `json:"and"`
		Or        []*Filter       `json:"or"`
		Not       *Filter         `json:"not"`
		Condition *FieldCondition `json:"condition"`
	}
	SearchParams struct {
		Model   any
		PageURL *url.URL

		KeywordColumns []string
		Keyword        string
		SQLConditions  []*SQLCondition
		Filter         *Filter

		Page    int64
		PerPage int64
		OrderBy []relay.Order

		// Both must exist simultaneously, and when they do, Page, PerPage, and OrderBy will be ignored
		// Or you can use the default pagination
		RelayPaginateRequest *relay.PaginateRequest[any]
		RelayPagination      RelayPagination
	}
)

const DummyCursor = "dummy"

type SearchResult struct {
	PageInfo   relay.PageInfo
	TotalCount *int
	Nodes      interface{}
}

type SlugDecoder interface {
	PrimaryColumnValuesBySlug(slug string) map[string]string
}

type SlugEncoder interface {
	PrimarySlug() string
}

type FilterDataFunc func(ctx *web.EventContext) vuetifyx.FilterData

type FilterNotificationFunc func(ctx *web.EventContext) h.HTMLComponent

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

type FieldPlugin interface {
	FieldInstall(fb *FieldBuilder) error
}

type (
	FieldInstallFunc func(fb *FieldBuilder) error
	ModelInstallFunc func(pb *Builder, mb *ModelBuilder) error
	InstallFunc      func(pb *Builder) error
)
