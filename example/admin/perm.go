package admin

import (
	"fmt"
	"net/http"

	"github.com/ory/ladon"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

func initPermission(b *presets.Builder, db *gorm.DB) {
	perm.Verbose = true
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:orders:*"),
			perm.PolicyFor(
				models.RoleViewer,
				models.RoleEditor,
				models.RoleManager,
			).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On("*:roles:*", "*:users:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On(perm.Anything),
			perm.PolicyFor(models.RoleManager).WhoAre(perm.Denied).ToDo(perm.Anything).
				On("*:activity_logs").On("*:activity_logs:*").
				Given(perm.Conditions{
					"is_authorized": &ladon.BooleanCondition{},
				}),
		).SubjectsFunc(func(r *http.Request) []string {
			u := getCurrentUser(r)
			if u == nil {
				return nil
			}
			return u.GetRoles()
		}).ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
			c := make(perm.Context)
			for _, obj := range objs {
				switch v := obj.(type) {
				case *activity.ActivityLog:
					u := getCurrentUser(r)
					if fmt.Sprint(u.GetID()) == v.UserID {
						c["is_authorized"] = true
					} else {
						c["is_authorized"] = false
					}
				}
			}
			return c
		}).DBPolicy(perm.NewDBPolicy(db)),
	)
}
