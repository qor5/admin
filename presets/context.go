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
