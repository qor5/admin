package inject

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
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
		results, err := inj.invoke(provider)
		if err != nil {
			return rv, err
		}

		inj.mu.Lock()
		for _, result := range results {
			resultType := result.Type()
			inj.values[resultType] = result

			// provider should not be called again
			delete(inj.providers, resultType)

			if resultType == rt {
				rv = result
			}
		}
		inj.mu.Unlock()

		if rv.IsValid() {
			return rv, nil
		}
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
