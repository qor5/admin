package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/qor5/admin/note"
	"github.com/qor5/admin/role"
	"github.com/qor5/x/login"
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

func withNoteContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := getCurrentUser(r)
			if u == nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), note.UserIDKey, u.ID)
			ctx = context.WithValue(ctx, note.UserKey, fmt.Sprintf("%v (%v)", u.Name, u.Account))
			newR := r.WithContext(ctx)

			next.ServeHTTP(w, newR)
		})
	}
}

func validateSessionToken() func(next http.Handler) http.Handler {
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

			valid, err := checkIsTokenValidFromRequest(r, user.ID)
			if err != nil || !valid {
				logoutURL := "/auth/logout"
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

func isOAuthInfoCompleted() func(next http.Handler) http.Handler {
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

			if user.OAuthUserID != "" && !user.IsInfoCompleted {
				logoutURL := "/auth/logout"
				if r.URL.Path == logoutURL {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, oauthCompleteInfoPageURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
