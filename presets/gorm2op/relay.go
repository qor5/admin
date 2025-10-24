package gorm2op

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/hook"
	"github.com/theplant/relay"
	"github.com/theplant/relay/cursor"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

func OffsetBasedPagination(skipTotalCount bool, cursorHooks ...func(next relay.ApplyCursorsFunc[any]) relay.ApplyCursorsFunc[any]) presets.RelayPagination {
	return relayPagination(gormrelay.NewOffsetAdapter[any], skipTotalCount, cursorHooks...)
}

func KeysetBasedPagination(skipTotalCount bool, cursorHooks ...func(next relay.ApplyCursorsFunc[any]) relay.ApplyCursorsFunc[any]) presets.RelayPagination {
	return relayPagination(gormrelay.NewKeysetAdapter[any], skipTotalCount, cursorHooks...)
}

func relayPagination(f func(db *gorm.DB, opts ...gormrelay.Option[any]) relay.ApplyCursorsFunc[any], skipTotalCount bool, cursorHooks ...func(next relay.ApplyCursorsFunc[any]) relay.ApplyCursorsFunc[any]) presets.RelayPagination {
	p := relay.New(
		func(ctx context.Context, req *relay.ApplyCursorsRequest) (*relay.ApplyCursorsResponse[any], error) {
			db, ok := ctx.Value(ctxKeyDBForRelay{}).(*gorm.DB)
			if !ok {
				return nil, errors.New("db not found in context")
			}
			opts, _ := ctx.Value(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
			opts = appendWithComputedIfHasHook(ctx, opts)
			return cursor.Base64(f(db, opts...))(ctx, req)
		},
		relay.EnsureLimits[any](presets.PerPageDefault, presets.PerPageMax),
		relay.PrependCursorHook(cursorHooks...),
	)
	return func(_ *web.EventContext) (relay.Paginator[any], error) {
		return relay.PaginatorFunc[any](func(ctx context.Context, req *relay.PaginateRequest[any]) (*relay.Connection[any], error) {
			ctx = relay.WithSkip(ctx, relay.Skip{
				Edges:      true,
				TotalCount: skipTotalCount,
			})
			return p.Paginate(ctx, req)
		}), nil
	}
}

type ctxKeyRelayOptions struct{}

// func AppendRelayOptions(ctx context.Context, opts ...gormrelay.Option[any]) context.Context {
// 	existingOpts, _ := ctx.Value(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
// 	opts = append(existingOpts, opts...)
// 	return context.WithValue(ctx, ctxKeyRelayOptions{}, opts)
// }

// func EventContextAppendRelayOptions(ctx *web.EventContext, opts ...gormrelay.Option[any]) *web.EventContext {
// 	existingOpts, _ := ctx.ContextValue(ctxKeyRelayOptions{}).([]gormrelay.Option[any])
// 	opts = append(existingOpts, opts...)
// 	return ctx.WithContextValue(ctxKeyRelayOptions{}, opts)
// }

type ctxKeyRelayPaginationHook struct{}

func WithRelayPaginationHooks(ctx context.Context, hooks ...hook.Hook[relay.Paginator[any]]) context.Context {
	previousHook, _ := ctx.Value(ctxKeyRelayPaginationHook{}).(hook.Hook[relay.Paginator[any]])
	hook := hook.Prepend(previousHook, hooks...)
	return context.WithValue(ctx, ctxKeyRelayPaginationHook{}, hook)
}

func EventContextWithRelayPaginationHooks(ctx *web.EventContext, hooks ...hook.Hook[relay.Paginator[any]]) *web.EventContext {
	previousHook, _ := ctx.ContextValue(ctxKeyRelayPaginationHook{}).(hook.Hook[relay.Paginator[any]])
	hook := hook.Prepend(previousHook, hooks...)
	return ctx.WithContextValue(ctxKeyRelayPaginationHook{}, hook)
}

type ctxKeyRelayComputedHook struct{}

func WithRelayComputedHook(ctx context.Context, hooks ...hook.Hook[*gormrelay.Computed[any]]) context.Context {
	previousHook, _ := ctx.Value(ctxKeyRelayComputedHook{}).(hook.Hook[*gormrelay.Computed[any]])
	hook := hook.Prepend(previousHook, hooks...)
	return context.WithValue(ctx, ctxKeyRelayComputedHook{}, hook)
}

func EventContextWithRelayComputedHook(ctx *web.EventContext, hooks ...hook.Hook[*gormrelay.Computed[any]]) *web.EventContext {
	previousHook, _ := ctx.ContextValue(ctxKeyRelayComputedHook{}).(hook.Hook[*gormrelay.Computed[any]])
	hook := hook.Prepend(previousHook, hooks...)
	return ctx.WithContextValue(ctxKeyRelayComputedHook{}, hook)
}

func appendWithComputedIfHasHook(ctx context.Context, opts []gormrelay.Option[any]) []gormrelay.Option[any] {
	computedHook, _ := ctx.Value(ctxKeyRelayComputedHook{}).(hook.Hook[*gormrelay.Computed[any]])
	if computedHook != nil {
		computed := &gormrelay.Computed[any]{
			Columns: gormrelay.NewComputedColumns(map[string]string{}),
			Scanner: gormrelay.NewComputedScanner[any],
		}
		computed = computedHook(computed)
		opts = append(opts, gormrelay.WithComputed(computed))
	}
	return opts
}
