package starter

import (
	"net/http"

	"github.com/qor5/admin/v3/role"
	"gorm.io/gorm"
)

func withRoles(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := GetCurrentUser(r)
			if u == nil {
				next.ServeHTTP(w, r)
				return
			}

			var roleIDs []uint
			if err := db.Table("user_role_join").Select("role_id").Where("user_id=?", u.ID).Scan(&roleIDs).Error; err != nil {
				panic(err)
			}
			if len(roleIDs) > 0 {
				var roles []role.Role
				if err := db.Where("id in (?)", roleIDs).Find(&roles).Error; err != nil {
					panic(err)
				}
				u.Roles = roles
			}
			next.ServeHTTP(w, r)
		})
	}
}

func securityMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Add("Cache-control", "no-cache, no-store, max-age=0, must-revalidate")
			w.Header().Add("Pragma", "no-cache")

			next.ServeHTTP(w, req)
		})
	}
}
