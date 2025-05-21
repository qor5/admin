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

func relayPagination(f func(db *gorm.DB) relay.ApplyCursorsFunc[any], skipTotalCount bool, cursorMiddlewares ...relay.CursorMiddleware[any]) presets.RelayPagination {
	p := relay.New(
		func(ctx context.Context, req *relay.ApplyCursorsRequest) (*relay.ApplyCursorsResponse[any], error) {
			db, ok := ctx.Value(ctxKeyDBForRelay{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			return cursor.Base64(f(db))(ctx, req)
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
