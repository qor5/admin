package gorm2op

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/theplant/relay"
	"github.com/theplant/relay/cursor"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/gorm"
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
			db, ok := ctx.Value(ctxKeyDB{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			return cursor.Base64(f(db))(ctx, req)
		},
		relay.EnsureLimits[any](presets.PerPageMax, presets.PerPageDefault),
		relay.AppendCursorMiddleware(cursorMiddlewares...),
	)
	return func(ctx *web.EventContext) (relay.Pagination[any], error) {
		return relay.PaginationFunc[any](func(ctx context.Context, req *relay.PaginateRequest[any]) (*relay.PaginateResponse[any], error) {
			ctx = relay.WithSkipEdges(ctx)
			if skipTotalCount {
				ctx = relay.WithSkipTotalCount(ctx)
			}
			return p.Paginate(ctx, req)
		}), nil
	}
}
