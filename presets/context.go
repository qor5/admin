package presets

import (
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type presetsCtx int

const (
	ctxInDialog presetsCtx = iota
	ctxActionsComponentTeleportToID
	ctxDetailingAfterTitleComponent
	ctxEventFuncAddonWrapper
	CtxPageTitleComponent
)

func IsInDialog(ctx *web.EventContext) bool {
	v, ok := ctx.ContextValue(ctxInDialog).(bool)
	if !ok {
		return false
	}
	return v
}

func GetActionsComponentTeleportToID(ctx *web.EventContext) string {
	v, _ := ctx.ContextValue(ctxActionsComponentTeleportToID).(string)
	return v
}

func GetComponentFromContext(ctx *web.EventContext, key presetsCtx) (h.HTMLComponent, bool) {
	v, ok := ctx.ContextValue(key).(h.HTMLComponent)
	return v, ok
}

type EventFuncAddon func(ctx *web.EventContext, r *web.EventResponse) (err error)

func getEventFuncAddonWrapper(ctx *web.EventContext) (func(in EventFuncAddon) EventFuncAddon, bool) {
	v, ok := ctx.ContextValue(ctxEventFuncAddonWrapper).(func(in EventFuncAddon) EventFuncAddon)
	return v, ok
}

func WrapEventFuncAddon(ctx *web.EventContext, w func(in EventFuncAddon) EventFuncAddon) {
	prev, ok := getEventFuncAddonWrapper(ctx)
	if !ok {
		ctx.WithContextValue(ctxEventFuncAddonWrapper, w)
	} else {
		ctx.WithContextValue(ctxEventFuncAddonWrapper, func(in EventFuncAddon) EventFuncAddon {
			return w(prev(in))
		})
	}
}
