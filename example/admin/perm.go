package admin

import (
	"net/http"
	"time"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/presets"
	"github.com/qor5/x/perm"
	"gorm.io/gorm"
)

func initPermission(b *presets.Builder, db *gorm.DB) {
	// perm.Verbose = true

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
			activity.PermPolicy,
		).SubjectsFunc(func(r *http.Request) []string {
			u := getCurrentUser(r)
			if u == nil {
				return nil
			}
			return u.GetRoles()
		}).EnableDBPolicy(db, perm.DefaultDBPolicy{}, time.Minute),
	)
}
