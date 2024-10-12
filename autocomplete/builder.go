package autocomplete

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/samber/lo"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"go.uber.org/zap"
	"gorm.io/gorm"

	vx "github.com/qor5/x/v3/ui/vuetifyx"
)

type (
	Builder struct {
		prefix           string
		db               *gorm.DB
		logger           *zap.Logger
		handler          http.Handler
		models           []*ModelBuilder
		allowCrossOrigin bool
	}

	Handler interface {
		Handle(pattern string, handler http.Handler)
	}
)

func New() *Builder {
	l, _ := zap.NewDevelopment()
	return &Builder{
		prefix:           "",
		logger:           l,
		allowCrossOrigin: true,
	}
}

func (b *Builder) NewModelBuilder(model interface{}) (mb *ModelBuilder) {
	mb = &ModelBuilder{p: b, model: model}
	mb.modelType = reflect.TypeOf(model)
	if mb.modelType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("model %#+v must be pointer", model))
	}
	modelstr := mb.modelType.String()
	modelName := modelstr[strings.LastIndex(modelstr, ".")+1:]
	mb.uriName = inflection.Plural(strcase.ToKebab(modelName))
	mb.columns = []string{"id"}
	return
}

func (b *Builder) Model(v interface{}) (r *ModelBuilder) {
	r = b.NewModelBuilder(v)
	b.models = append(b.models, r)
	return r
}

func (b *Builder) modelNames() (r []string) {
	for _, m := range b.models {
		r = append(r, m.uriName)
	}
	return
}

func (b *Builder) DB(v *gorm.DB) *Builder {
	b.db = v
	return b
}

func (b *Builder) Prefix(v string) *Builder {
	b.prefix = v
	return b
}

func (b *Builder) AllowCrossOrigin(v bool) *Builder {
	b.allowCrossOrigin = v
	return b
}

func (b *Builder) Logger(v *zap.Logger) *Builder {
	b.logger = v
	return b
}

func (b *Builder) Mux(mux Handler) {
	mns := b.modelNames()
	if len(lo.Uniq(mns)) != len(mns) {
		panic(fmt.Sprintf("Duplicated model names registered %v", mns))
	}
	b.initMux(mux)
}

func (b *Builder) initMux(mux Handler) {
	b.logger.Info("initializing mux for", zap.Reflect("models", b.modelNames()), zap.String("prefix", b.prefix))
	for _, m := range b.models {
		path := m.JsonHref()
		mux.Handle(
			path,
			m,
		)
		b.logger.Info(fmt.Sprintf("mounted url: %s\n", path))
	}
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b.handler == nil {
		mux := http.NewServeMux()
		b.Mux(mux)
		b.handler = mux
	}
	b.handler.ServeHTTP(w, r)
}

func NewDefaultAutocompleteDataSource(v string) *vx.AutocompleteDataSource {
	return &vx.AutocompleteDataSource{
		RemoteURL:     v,
		IsPaging:      true,
		ItemTitle:     "title",
		ItemValue:     "id",
		PageField:     ParamPage,
		PageSizeField: ParamPageSize,
		CurrentField:  ResponseCurrent,
		PagesField:    ResponsePages,
		TotalField:    ResponseTotal,
		SearchField:   ParamSearch,
		ItemsField:    ResponseItems,
		Page:          1,
		PageSize:      5,
		Separator:     "__",
	}
}
