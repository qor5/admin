package presets

import (
	"context"

	h "github.com/theplant/htmlgo"
)

type presetsCtx int

const (
	ctxInDialog presetsCtx = iota
	ctxActionsComponent
)

func IsInDialog(ctx context.Context) bool {
	v, ok := ctx.Value(ctxInDialog).(bool)
	if !ok {
		return false
	}
	return v
}

func GetActionsComponent(ctx context.Context) h.HTMLComponent {
	v, _ := ctx.Value(ctxActionsComponent).(h.HTMLComponent)
	return v
}
