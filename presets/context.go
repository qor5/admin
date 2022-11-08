package presets

import (
	"context"
)

type presetsCtx int

const (
	ctxInDialog presetsCtx = iota
)

func IsInDialog(ctx context.Context) bool {
	v, ok := ctx.Value(ctxInDialog).(bool)
	if !ok {
		return false
	}
	return v
}
