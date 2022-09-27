package login

import (
	"context"
	"net/http"
)

type hookContext int

const (
	hookOldSessionTokenBeforeExtendKey hookContext = iota
)

func withOldSessionTokenBeforeExtend(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), hookOldSessionTokenBeforeExtendKey, token))
}

func GetOldSessionTokenBeforeExtend(r *http.Request) string {
	v, ok := r.Context().Value(hookOldSessionTokenBeforeExtendKey).(string)
	if !ok {
		return ""
	}
	return v
}
