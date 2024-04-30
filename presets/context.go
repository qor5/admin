package presets

import (
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type presetsCtx int

const (
	ctxInDialog presetsCtx = iota
	ctxActionsComponent
	ctxDetailingAfterTitleComponent
)

func IsInDialog(ctx *web.EventContext) bool {
	v, ok := ctx.ContextValue(ctxInDialog).(bool)
	if !ok {
		return false
	}
	return v
}

func GetActionsComponent(ctx *web.EventContext) h.HTMLComponent {
	v, _ := ctx.ContextValue(ctxActionsComponent).(h.HTMLComponent)
	return v
}

func GetComponentFromContext(ctx *web.EventContext, key presetsCtx) (h.HTMLComponent, bool) {
	v, ok := ctx.ContextValue(key).(h.HTMLComponent)
	return v, ok
}
