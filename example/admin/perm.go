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

	InitDefaultRolesToDB(db)

	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:orders:*"),
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

func InitDefaultRolesToDB(db *gorm.DB) {
	var cnt int64
	if err := db.Table("roles").Count(&cnt).Error; err != nil {
		panic(err)
	}

	if cnt == 0 {
		if err := db.Table("roles").Create(
			&[]map[string]interface{}{
				{"id": models.RoleAdminID, "name": models.RoleAdmin},
				{"id": models.RoleManagerID, "name": models.RoleManager},
				{"id": models.RoleEditorID, "name": models.RoleEditor},
				{"id": models.RoleViewerID, "name": models.RoleViewer},
			}).Error; err != nil {
			panic(err)
		}
	}
}
