package compo

import (
	"context"
	"fmt"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type Dep interface {
	h.HTMLComponent
	EventDispatchActionHandler(ctx *web.EventContext) (r web.EventResponse, err error)
	_isCompoDep()
}

type injective interface {
	InjectDep(f Dep)
}

type CompoDep[T Named] struct {
	ref                     Dep
	initial                 T
	eventNameDispatchAction string
}

func NewDep[T Named](ref Dep, initial T) *CompoDep[T] {
	b := &CompoDep[T]{ref: ref, initial: initial}
	if b.ref == nil {
		b.ref = b
	}
	b.eventNameDispatchAction = fmt.Sprintf("%s%T__%s__", EventDispatchAction, b.ref, initial.CompoName())
	if c, ok := any(initial).(injective); ok {
		c.InjectDep(b.ref)
	}
	return b
}

func (b *CompoDep[T]) MarshalHTML(ctx context.Context) ([]byte, error) {
	web.MustGetPageBuilder(ctx).RegisterEventFunc(b.eventNameDispatchAction, b.EventDispatchActionHandler)
	ctx = WithDispatchActionEventName(ctx, b.eventNameDispatchAction)
	return b.initial.MarshalHTML(ctx)
}

func (b *CompoDep[T]) EventDispatchActionHandler(ctx *web.EventContext) (r web.EventResponse, err error) {
	ctx.R = ctx.R.WithContext(WithDispatchActionEventName(ctx.R.Context(), b.eventNameDispatchAction))
	return EventDispatchActionHandlerComplex(ctx, func(ctx context.Context, c any) (any, error) {
		if c, ok := c.(injective); ok {
			c.InjectDep(b.ref)
		}
		return c, nil
	})
}

func (b *CompoDep[T]) _isCompoDep() {}

type depCenter struct {
	m map[string]Dep
}

func (c *depCenter) Register(key string, f Dep) {
	c.m[key] = f
}

func (c *depCenter) Get(key string) Dep {
	return c.m[key]
}

var defaultCenter = &depCenter{m: map[string]Dep{}}

func RegisterDep(key string, f Dep) {
	defaultCenter.Register(key, f)
}

func GetDep(key string) Dep {
	return defaultCenter.Get(key)
}
