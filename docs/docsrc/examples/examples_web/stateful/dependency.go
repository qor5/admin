package stateful

import (
	"context"
	"sync"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/inject"

	h "github.com/theplant/htmlgo"
)

type Scope string

const (
	// ScopeTop is the top level scope.
	// It must be empty because top level dependencies are typically required,
	// and even if the eventDispatchActionHandler side detects an empty scope, it should still be applied.
	ScopeTop Scope = ""
)

var defaultDependencyCenter = NewDependencyCenter()

type DependencyCenter struct {
	mu        sync.RWMutex
	injectors map[Scope]*inject.Injector
}

func NewDependencyCenter() *DependencyCenter {
	return &DependencyCenter{
		injectors: map[Scope]*inject.Injector{},
	}
}

func (r *DependencyCenter) injectorUnlocked(scope Scope) *inject.Injector {
	inj, ok := r.injectors[scope]
	if !ok {
		inj = inject.New()
		if scope != ScopeTop {
			// TODO: 这样设计的话只能支持两层，是否有必要支持多层？
			inj.SetParent(r.injectorUnlocked(ScopeTop))
		}
		inj.Provide(func() Scope { return scope })
		r.injectors[scope] = inj
	}
	return inj
}

func (r *DependencyCenter) injector(scope Scope) *inject.Injector {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.injectorUnlocked(scope)
}

func (r *DependencyCenter) Provide(scope Scope, ctors ...any) error {
	inj := r.injector(scope)
	return inj.Provide(ctors...)
}

func (r *DependencyCenter) Apply(scope Scope, target any) error {
	inj := r.injector(scope)
	return inj.Apply(target)
}

func Provide(scope Scope, ctors ...any) error {
	return defaultDependencyCenter.Provide(scope, ctors...)
}

func MustProvide(scope Scope, ctors ...any) {
	err := Provide(scope, ctors...)
	if err != nil {
		panic(err)
	}
}

type scopeCtxKey struct{}

func withScope(ctx context.Context, scope Scope) context.Context {
	return context.WithValue(ctx, scopeCtxKey{}, scope)
}

func scopeFromContext(ctx context.Context) Scope {
	scope, _ := ctx.Value(scopeCtxKey{}).(Scope)
	return scope
}

func Apply(ctx context.Context, target any) error {
	scope := scopeFromContext(ctx)
	return defaultDependencyCenter.Apply(scope, target)
}

func MustApply[T any](ctx context.Context, target T) T {
	err := Apply(ctx, target)
	if err != nil {
		panic(err)
	}
	return target
}

func Scoped(scope Scope, c h.HTMLComponent) (h.HTMLComponent, error) {
	if err := defaultDependencyCenter.Apply(scope, c); err != nil {
		return nil, err
	}
	return h.ComponentFunc(func(ctx context.Context) ([]byte, error) {
		ctx = withScope(ctx, scope)
		return c.MarshalHTML(ctx)
	}), nil
}

func MustScoped(scope Scope, c h.HTMLComponent) h.HTMLComponent {
	c, err := Scoped(scope, c)
	if err != nil {
		panic(err)
	}
	return c
}
