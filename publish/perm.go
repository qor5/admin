package publish

import (
	"net/http"

	"github.com/qor5/x/v3/perm"
)

const (
	PermAll       = "publish:*"
	PermPublish   = "publish:publish"
	PermUnpublish = "publish:unpublish"
	PermSchedule  = "publish:schedule"  // Prerequisite: PermPublish/PermUnpublish
	PermDuplicate = "publish:duplicate" // Prerequisite: presets.PermUpdate
)

func DeniedDo(verifier *perm.Verifier, obj any, r *http.Request, actions ...string) bool {
	for _, action := range actions {
		b := verifier.Do(action).WithReq(r)
		if obj != nil {
			b.ObjectOn(obj)
		}
		if b.IsAllowed() != nil {
			return true
		}
	}
	return false
}
