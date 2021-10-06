package seo

import "net/http"

const (
	PermEdit = "perm_seo_edit"
)

func editIsAllowed(r *http.Request) error {
	return permVerifier.Do(PermEdit).On("qor_seo_settings").WithReq(r).IsAllowed()
}
