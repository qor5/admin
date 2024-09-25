package gorm2op

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	relay "github.com/theplant/gorelay"
	"github.com/theplant/gorelay/cursor"
	"github.com/theplant/gorelay/gormrelay"
	"gorm.io/gorm"
)

func OffsetBasedPagination(disableTotalCount bool, middlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	p := relay.New(
		true, // nodesOnly
		presets.PerPageMax, presets.PerPageDefault,
		func(ctx context.Context, req *relay.ApplyCursorsRequest) (*relay.ApplyCursorsResponse[any], error) {
			db, ok := ctx.Value(ctxKeyDB{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			var finer cursor.OffsetFinder[any]
			if disableTotalCount {
				finer = gormrelay.NewOffsetFinder[any](db)
			} else {
				finer = gormrelay.NewOffsetCounter[any](db)
			}
			f := cursor.Base64(cursor.NewOffsetAdapter(finer))
			for _, middleware := range middlewares {
				f = middleware(f)
			}
			return f(ctx, req)
		},
	)
	return func(ctx *web.EventContext) (relay.Pagination[any], error) {
		return p, nil
	}
}

func KeysetBasedPagination(disableTotalCount bool, middlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	p := relay.New(
		true, // nodesOnly
		presets.PerPageMax, presets.PerPageDefault,
		func(ctx context.Context, req *relay.ApplyCursorsRequest) (*relay.ApplyCursorsResponse[any], error) {
			db, ok := ctx.Value(ctxKeyDB{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			var finer cursor.KeysetFinder[any]
			if disableTotalCount {
				finer = gormrelay.NewKeysetFinder[any](db)
			} else {
				finer = gormrelay.NewKeysetCounter[any](db)
			}
			f := cursor.Base64(cursor.NewKeysetAdapter(finer))
			for _, middleware := range middlewares {
				f = middleware(f)
			}
			return f(ctx, req)
		},
	)
	return func(ctx *web.EventContext) (relay.Pagination[any], error) {
		return p, nil
	}
}
