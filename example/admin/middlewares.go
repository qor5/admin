package admin

import (
	"context"
	"fmt"
	"github.com/qor5/admin/v3/activity"
	"net/http"

	"github.com/qor5/admin/v3/role"
	"github.com/qor5/x/v3/login"
	"gorm.io/gorm"
)

func withRoles(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := getCurrentUser(r)
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

func withNoteContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := getCurrentUser(r)
			if u == nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), activity.UserIDKey, u.ID)
			ctx = context.WithValue(ctx, activity.UserKey, fmt.Sprintf("%v (%v)", u.Name, u.Account))
			newR := r.WithContext(ctx)

			next.ServeHTTP(w, newR)
		})
	}
}

func validateSessionToken(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := getCurrentUser(r)
			if user == nil {
				next.ServeHTTP(w, r)
				return
			}
			if login.IsLoginWIP(r) {
				next.ServeHTTP(w, r)
				return
			}

			valid, err := checkIsTokenValidFromRequest(db, r, user.ID)
			if err != nil || !valid {
				if r.URL.Path == logoutURL {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, logoutURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
