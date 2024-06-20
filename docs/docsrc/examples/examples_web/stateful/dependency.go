package stateful

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_web/inject"

	h "github.com/theplant/htmlgo"
)

var ErrInjectorNotFound = errors.New("injector not found")

var defaultDependencyCenter = NewDependencyCenter()

// just to make it easier to get the name of the currently applied injector
type InjectorName string

type DependencyCenter struct {
	mu        sync.RWMutex
	injectors map[string]*inject.Injector
}

func NewDependencyCenter() *DependencyCenter {
	return &DependencyCenter{
		injectors: map[string]*inject.Injector{},
	}
}

func (r *DependencyCenter) RegisterInjector(name string, parent string) {
	name = strings.TrimSpace(name)
	parent = strings.TrimSpace(parent)
	if name == "" {
		panic(fmt.Errorf("name is required"))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.injectors[name]; ok {
		panic(fmt.Errorf("injector %q already exists", name))
	}

	var parentInjector *inject.Injector
	if parent != "" {
		var ok bool
		parentInjector, ok = r.injectors[parent]
		if !ok {
			panic(fmt.Errorf("parent injector %q not found", parent))
		}
	}

	inj := inject.New()
	inj.Provide(func() InjectorName { return InjectorName(name) })
	if parentInjector != nil {
		inj.SetParent(parentInjector)
	}
	r.injectors[name] = inj
}

func (r *DependencyCenter) Injector(name string) (*inject.Injector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inj, ok := r.injectors[name]
	if !ok {
		return nil, errors.Wrap(ErrInjectorNotFound, name)
	}
	return inj, nil
}

func RegisterInjector(name string, parent string) {
	defaultDependencyCenter.RegisterInjector(name, parent)
}

func Injector(name string) (*inject.Injector, error) {
	return defaultDependencyCenter.Injector(name)
}

func MustInjector(name string) *inject.Injector {
	inj, err := Injector(name)
	if err != nil {
		panic(err)
	}
	return inj
}

type injectorNameCtxKey struct{}

func withInjectorName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, injectorNameCtxKey{}, name)
}

func injectorNameFromContext(ctx context.Context) string {
	name, _ := ctx.Value(injectorNameCtxKey{}).(string)
	return name
}

func Apply(ctx context.Context, target any) error {
	name := injectorNameFromContext(ctx)
	if name == "" {
		return nil
	}
	inj, err := defaultDependencyCenter.Injector(name)
	if err != nil {
		return err
	}
	return inj.Apply(target)
}

func MustApply[T any](ctx context.Context, target T) T {
	err := Apply(ctx, target)
	if err != nil {
		panic(err)
	}
	return target
}

func Inject(injectorName string, c h.HTMLComponent) (h.HTMLComponent, error) {
	inj, err := defaultDependencyCenter.Injector(injectorName)
	if err != nil {
		return nil, err
	}
	if err := inj.Apply(c); err != nil {
		return nil, err
	}
	return h.ComponentFunc(func(ctx context.Context) ([]byte, error) {
		ctx = withInjectorName(ctx, injectorName)
		return c.MarshalHTML(ctx)
	}), nil
}

func MustInject(injectorName string, c h.HTMLComponent) h.HTMLComponent {
	c, err := Inject(injectorName, c)
	if err != nil {
		panic(err)
	}
	return c
}
