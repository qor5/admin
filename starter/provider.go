package starter

import (
	"context"
	"net/http"

	"github.com/theplant/inject"
)

func setupHandlerFactory(build func(ctx context.Context, handler *Handler, ctors ...any) error, ctors ...any) func(ctx context.Context, inj *inject.Injector, conf *Config, mux *http.ServeMux) (*Handler, error) {
	return func(ctx context.Context, inj *inject.Injector, conf *Config, mux *http.ServeMux) (*Handler, error) {
		handler := NewHandler(conf)
		if err := handler.SetParent(inj); err != nil {
			return nil, err
		}
		if err := build(ctx, handler, ctors...); err != nil {
			return nil, err
		}
		mux.Handle("/", handler)
		return handler, nil
	}
}

func SetupHandlerFactory(ctors ...any) func(ctx context.Context, inj *inject.Injector, conf *Config, mux *http.ServeMux) (*Handler, error) {
	return setupHandlerFactory(func(ctx context.Context, handler *Handler, ctors ...any) error {
		return handler.Build(ctx, ctors...)
	}, ctors...)
}

func SetupTestHandlerFactory(ctors ...any) func(ctx context.Context, inj *inject.Injector, conf *Config, mux *http.ServeMux) (*Handler, error) {
	return setupHandlerFactory(func(ctx context.Context, handler *Handler, ctors ...any) error {
		return handler.BuildForTest(ctx, ctors...)
	}, ctors...)
}
