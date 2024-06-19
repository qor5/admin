package inject

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

var (
	ErrTypeNotProvided     = errors.New("type not provided")
	ErrTypeAlreadyProvided = errors.New("type already provided")
	ErrParentAlreadySet    = errors.New("parent already set")
)

type Injector struct {
	mu sync.RWMutex

	values    map[reflect.Type]reflect.Value
	providers map[reflect.Type]any // value func
	parent    *Injector

	sfg singleflight.Group
}

func New() *Injector {
	return &Injector{
		values:    map[reflect.Type]reflect.Value{},
		providers: map[reflect.Type]any{},
	}
}

func (inj *Injector) SetParent(parent *Injector) error {
	inj.mu.RLock()
	defer inj.mu.RUnlock()
	if inj.parent != nil {
		return ErrParentAlreadySet
	}
	inj.parent = parent
	return nil
}

// TODO: 如果 func 的最后一个返回值是 error 的话，貌似不应该作为 dep 提供
// TODO: 如果 func 的第一个参数是 ctx 的话，是否应该特殊处理呢？若需要，这个也需要 invoke 和 resolve 都处理
func (inj *Injector) provide(f any) (err error) {
	rv := reflect.ValueOf(f)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic("Provide only accepts a function")
	}

	inj.mu.Lock()
	defer inj.mu.Unlock()

	setted := []reflect.Type{}
	defer func() {
		if err != nil {
			for _, t := range setted {
				delete(inj.providers, t)
			}
		}
	}()

	for i := 0; i < rt.NumOut(); i++ {
		outType := rt.Out(i)

		if _, ok := inj.values[outType]; ok {
			return errors.Wrap(ErrTypeAlreadyProvided, outType.String())
		}

		if _, ok := inj.providers[outType]; ok {
			return errors.Wrap(ErrTypeAlreadyProvided, outType.String())
		}

		inj.providers[outType] = f
		setted = append(setted, outType)
	}
	return nil
}

func (inj *Injector) invoke(f any) ([]reflect.Value, error) {
	rt := reflect.TypeOf(f)
	if rt.Kind() != reflect.Func {
		panic("Invoke only accepts a function")
	}

	numIn := rt.NumIn()
	in := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		argType := rt.In(i)
		argValue, err := inj.resolve(argType)
		if err != nil {
			return nil, err
		}
		in[i] = argValue
	}

	return reflect.ValueOf(f).Call(in), nil
}

func (inj *Injector) resolve(rt reflect.Type) (reflect.Value, error) {
	inj.mu.RLock()
	rv := inj.values[rt]
	if rv.IsValid() {
		inj.mu.RUnlock()
		return rv, nil
	}
	provider, ok := inj.providers[rt]
	parent := inj.parent
	inj.mu.RUnlock()

	if ok {
		// ensure that the provider is only executed once same time
		_, err, _ := inj.sfg.Do(fmt.Sprintf("%p", provider), func() (any, error) {
			// must recheck the provider, because it may be deleted by prev inj.sfg.Do
			inj.mu.RLock()
			_, ok := inj.providers[rt]
			inj.mu.RUnlock()
			if !ok {
				return nil, nil
			}

			results, err := inj.invoke(provider)
			if err != nil {
				return nil, err
			}

			inj.mu.Lock()
			for _, result := range results {
				// TODO: Does need to Apply ?
				resultType := result.Type()
				inj.values[resultType] = result
				delete(inj.providers, resultType)
			}
			inj.mu.Unlock()

			return nil, nil
		})
		if err != nil {
			return rv, err
		}
		return inj.resolve(rt)
	}

	if parent != nil {
		return parent.resolve(rt)
	}

	return rv, errors.Wrap(ErrTypeNotProvided, rt.String())
}

func (inj *Injector) Apply(val any) error {
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("Apply only accepts a struct")
	}

	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)
		if _, ok := structField.Tag.Lookup("inject"); ok {
			if !field.CanSet() {
				// If the field is unexported, we need to create a new field that is settable.
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}
			dep, err := inj.resolve(field.Type())
			if err != nil {
				return err
			}
			field.Set(dep)
		}
		// TODO: 如果是 embed *Injector 的话，应该把自身塞进去
	}

	return nil
}

func (inj *Injector) Provide(fs ...any) error {
	for _, f := range fs {
		if err := inj.provide(f); err != nil {
			return err
		}
	}
	return nil
}

func (inj *Injector) Invoke(f any) ([]any, error) {
	results, err := inj.invoke(f)
	if err != nil {
		return nil, err
	}
	out := make([]any, len(results))
	for i, result := range results {
		out[i] = result.Interface()
	}
	return out, nil
}

func (inj *Injector) Resolve(ref any) error {
	rv, err := inj.resolve(reflect.TypeOf(ref).Elem())
	if err != nil {
		return err
	}
	reflect.ValueOf(ref).Elem().Set(rv)
	return nil
}
