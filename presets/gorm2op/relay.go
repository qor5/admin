package gorm2op

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/theplant/relay"
	"github.com/theplant/relay/cursor"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

func OffsetBasedPagination(skipTotalCount bool, cursorMiddlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	return relayPagination(gormrelay.NewOffsetAdapter[any], skipTotalCount, cursorMiddlewares...)
}

func KeysetBasedPagination(skipTotalCount bool, cursorMiddlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	return relayPagination(gormrelay.NewKeysetAdapter[any], skipTotalCount, cursorMiddlewares...)
}

func relayPagination(f func(db *gorm.DB, opts ...gormrelay.Option[any]) relay.ApplyCursorsFunc[any], skipTotalCount bool, cursorMiddlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	p := relay.New(
		func(ctx context.Context, req *relay.ApplyCursorsRequest) (*relay.ApplyCursorsResponse[any], error) {
			db, ok := ctx.Value(ctxKeyDBForRelay{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			opts, _ := ctx.Value(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
			return cursor.Base64(f(db, opts...))(ctx, req)
		},
		relay.EnsureLimits[any](presets.PerPageDefault, presets.PerPageMax),
		relay.AppendCursorMiddleware(cursorMiddlewares...),
	)
	return func(ctx *web.EventContext) (relay.Pagination[any], error) {
		return relay.PaginationFunc[any](func(ctx context.Context, req *relay.PaginateRequest[any]) (*relay.Connection[any], error) {
			ctx = relay.WithSkip(ctx, relay.Skip{
				Edges:      true,
				TotalCount: skipTotalCount,
			})
			return p.Paginate(ctx, req)
		}), nil
	}
}

type ctxKeyRelayOptions struct{}

func AppendRelayOptions(ctx context.Context, opts ...gormrelay.Option[any]) context.Context {
	existingOpts, _ := ctx.Value(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
	opts = append(existingOpts, opts...)
	return context.WithValue(ctx, ctxKeyRelayOptions{}, opts)
}

func EventContextAppendRelayOptions(ctx *web.EventContext, opts ...gormrelay.Option[any]) *web.EventContext {
	existingOpts, _ := ctx.ContextValue(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
	opts = append(existingOpts, opts...)
	return ctx.WithContextValue(ctxKeyRelayOptions{}, opts)
}

type ctxKeyRelayPaginationMiddlewares struct{}

func AppendRelayPaginationMiddlewares(ctx context.Context, mws ...relay.PaginationMiddleware[any]) context.Context {
	existingMws, _ := ctx.Value(ctxKeyRelayPaginationMiddlewares{}).([]relay.PaginationMiddleware[any])
	mws = append(existingMws, mws...)
	return context.WithValue(ctx, ctxKeyRelayPaginationMiddlewares{}, mws)
}

func EventContextAppendRelayPaginationMiddlewares(ctx *web.EventContext, mws ...relay.PaginationMiddleware[any]) *web.EventContext {
	existingMws, _ := ctx.ContextValue(ctxKeyRelayPaginationMiddlewares{}).([]relay.PaginationMiddleware[any])
	mws = append(existingMws, mws...)
	return ctx.WithContextValue(ctxKeyRelayPaginationMiddlewares{}, mws)
}
