package stateful

import (
	"context"

	"github.com/go-playground/form/v4"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type querySyncer struct {
	h.HTMLComponent
}

var formDecoder = func() *form.Decoder {
	decoder := form.NewDecoder()
	decoder.SetMode(form.ModeExplicit)
	decoder.SetTagName("query")
	// decoder.RegisterTagNameFunc(func(field reflect.StructField) string {
	// 	tag := field.Tag.Get("query")
	// 	if tag == "" {
	// 		return strcase.ToLowerCamel(field.Name)
	// 	}
	// 	return tag
	// })
	return decoder
}()

func (c *querySyncer) MarshalHTML(ctx context.Context) ([]byte, error) {
	// TODO: 然后需要在各个 reload action 中，需要将其通过 pushState 同步到 query 中
	// TODO: 如何保证页面中只会有一个呢？貌似没什么办法？
	evCtx := web.MustGetEventContext(ctx)
	query := evCtx.R.URL.Query()
	if err := formDecoder.Decode(&c.HTMLComponent, query); err != nil {
		return nil, err
	}
	return c.HTMLComponent.MarshalHTML(ctx)
}

func (c *querySyncer) Unwrap() h.HTMLComponent {
	return c.HTMLComponent
}

func SyncQuery(c h.HTMLComponent) h.HTMLComponent {
	return &querySyncer{HTMLComponent: c}
}
