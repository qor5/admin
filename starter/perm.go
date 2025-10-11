package starter

import (
	"fmt"
	"net/http"

	"github.com/ory/ladon"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
)

func (a *Handler) configurePermission(b *presets.Builder) {
	perm.Verbose = true

	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			perm.PolicyFor(
				RoleViewer,
				RoleEditor,
				RoleManager,
			).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On("*:roles:*", "*:users:*"),
			perm.PolicyFor(RoleViewer).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On(perm.Anything),

			perm.PolicyFor(RoleManager).WhoAre(perm.Denied).ToDo(perm.Anything).
				On("*:activity_logs").On("*:activity_logs:*").
				Given(perm.Conditions{
					"is_authorized": &ladon.BooleanCondition{},
				}),
		).SubjectsFunc(func(r *http.Request) []string {
			u := GetCurrentUser(r)
			if u == nil {
				return nil
			}
			return u.GetRoles()
		}).ContextFunc(func(r *http.Request, objs []any) perm.Context {
			c := make(perm.Context)
			for _, obj := range objs {
				// nolint:gocritic
				switch v := obj.(type) {
				case *activity.ActivityLog:
					u := GetCurrentUser(r)
					if fmt.Sprint(u.GetID()) == v.UserID {
						c["is_authorized"] = true
					} else {
						c["is_authorized"] = false
					}
				}
			}
			return c
		}).DBPolicy(perm.NewDBPolicy(a.DB)),
	)
}
