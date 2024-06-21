package stateful

import (
	"context"

	"github.com/go-playground/form/v4"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

var queryDecoder = func() *form.Decoder {
	decoder := form.NewDecoder()
	decoder.SetMode(form.ModeExplicit)
	decoder.SetTagName("query")
	return decoder
}()

var queryEncoder = func() *form.Encoder {
	encoder := form.NewEncoder()
	encoder.SetMode(form.ModeExplicit)
	encoder.SetTagName("query")
	return encoder
}()

type syncQueryCtxKey struct{}

func withSyncQuery(ctx context.Context) context.Context {
	return context.WithValue(ctx, syncQueryCtxKey{}, struct{}{})
}

func IsSyncQuery(ctx context.Context) bool {
	_, ok := ctx.Value(syncQueryCtxKey{}).(struct{})
	return ok
}

type querySyncer struct {
	h.HTMLComponent
}

func (c *querySyncer) MarshalHTML(ctx context.Context) ([]byte, error) {
	evCtx := web.MustGetEventContext(ctx)
	query := evCtx.R.URL.Query()

	// TODO: support prefix ?
	// const prefix = "main_"
	// r := url.Values{}
	// for k, v := range query {
	// 	if strings.HasPrefix(k, prefix) {
	// 		r[strings.TrimPrefix(k, prefix)] = v
	// 	}
	// }
	// query = r

	if err := queryDecoder.Decode(&c.HTMLComponent, query); err != nil {
		return nil, err
	}
	ctx = withSyncQuery(ctx)
	return c.HTMLComponent.MarshalHTML(ctx)
}

func (c *querySyncer) Unwrap() h.HTMLComponent {
	return c.HTMLComponent
}

func SyncQuery(c h.HTMLComponent) h.HTMLComponent {
	return &querySyncer{HTMLComponent: c}
}
