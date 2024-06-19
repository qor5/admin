package compo

import (
	"context"
	"sync"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/inject"

	h "github.com/theplant/htmlgo"
)

type Scope string

const (
	ScopeTop Scope = ""
)

type Registry struct {
	mu        sync.RWMutex
	injectors map[Scope]*inject.Injector
}

func NewRegistry() *Registry {
	return &Registry{
		injectors: map[Scope]*inject.Injector{},
	}
}

func (r *Registry) injectorUnlocked(scope Scope) *inject.Injector {
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

func (r *Registry) injector(scope Scope) *inject.Injector {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.injectorUnlocked(scope)
}

func (r *Registry) Provide(scope Scope, constructor any) error {
	inj := r.injector(scope)
	return inj.Provide(constructor)
}

func (r *Registry) Apply(scope Scope, target any) error {
	inj := r.injector(scope)
	return inj.Apply(target)
}

// TODO: 这个 Default 应该通过 ctx 向下传递？ 这样就有改变 registry 的能力了？ 貌似不是很有必要
var DefaultRegistry = NewRegistry()

func Provide(scope Scope, constructor any) error {
	return DefaultRegistry.Provide(scope, constructor)
}

func Apply[T any](scope Scope, target T) (T, error) {
	err := DefaultRegistry.Apply(scope, target)
	if err != nil {
		return target, err
	}
	return target, nil
}

func MustProvide(scope Scope, constructor any) {
	err := DefaultRegistry.Provide(scope, constructor)
	if err != nil {
		panic(err)
	}
}

func MustApply[T any](scope Scope, target T) T {
	err := DefaultRegistry.Apply(scope, target)
	if err != nil {
		panic(err)
	}
	return target
}

func Scoped(scope Scope, target h.HTMLComponent) (h.HTMLComponent, error) {
	c, err := Apply(scope, target)
	if err != nil {
		return nil, err
	}
	return &scopedCompo{HTMLComponent: c, Scope: scope}, nil
}

func MustScoped(scope Scope, target h.HTMLComponent) h.HTMLComponent {
	c, err := Scoped(scope, target)
	if err != nil {
		panic(err)
	}
	return c
}

type scopedCompo struct {
	h.HTMLComponent
	Scope Scope
}

type scopeCtxKey struct{}

func (c *scopedCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	ctx = context.WithValue(ctx, scopeCtxKey{}, c.Scope)
	return c.HTMLComponent.MarshalHTML(ctx)
}
