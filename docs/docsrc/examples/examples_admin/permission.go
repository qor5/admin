package examples_admin

import (
	"net/http"

	"github.com/ory/ladon"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
)

func permissionPieces() {
	_ = []interface{}{
		// @snippet_begin(PermissionAllowed)
		perm.Allowed,
		// @snippet_end
		// @snippet_begin(PermissionDenied)
		perm.Denied,
		// @snippet_end
		// @snippet_begin(PermissionPermList)
		presets.PermList,
		// @snippet_end
		// @snippet_begin(PermissionPermGet)
		presets.PermGet,
		// @snippet_end
		// @snippet_begin(PermissionPermCreate)
		presets.PermCreate,
		// @snippet_end
		// @snippet_begin(PermissionPermUpdate)
		presets.PermUpdate,
		// @snippet_end
		// @snippet_begin(PermissionPermDelete)
		presets.PermDelete,
		// @snippet_end
	}

	var Who, Able, What, Something string
	var Context perm.Conditions
	// @snippet_begin(PermissionSyntax)
	perm.PolicyFor(Who).WhoAre(Able).ToDo(What).On(Something).Given(Context)
	// @snippet_end

	var permBuilder perm.Builder
	var subjects_like_user_roles []string
	// @snippet_begin(PermissionSubjectsFunc)
	permBuilder.SubjectsFunc(func(r *http.Request) []string {
		return subjects_like_user_roles
	})
	// @snippet_end

	type resource1 struct {
		Owner string
	}
	// @snippet_begin(PermissionContextFunc)
	permBuilder.ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
		c := make(perm.Context)
		for _, obj := range objs {
			switch v := obj.(type) {
			case resource1:
				c["owner"] = v.Owner
				// ...other resource cases
			}
		}
		return c
	})
	// @snippet_end

	// @snippet_begin(PermissionGivenFunc)
	perm.PolicyFor(Who).WhoAre(Able).ToDo(What).On("*:resource1:*").Given(perm.Conditions{
		"owner": &ladon.EqualsSubjectCondition{},
	})
	// @snippet_end

	var presetsBuilder *presets.Builder
	type User struct {
		ID    string
		Roles []string
	}
	var getCurrentUser func(r *http.Request) *User
	type Article struct {
		OwnerID string
	}
	// @snippet_begin(PermissionExample)
	presetsBuilder.Permission(
		perm.New().Policies(
			// admin can do anything
			perm.PolicyFor("admin").WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			// viewer can view anything except users
			perm.PolicyFor("viewer").WhoAre(perm.Allowed).ToDo(presets.PermRead...).On(perm.Anything),
			perm.PolicyFor("viewer").WhoAre(perm.Denied).ToDo(perm.Anything).On("*:users:*"),
			// editor can edit their own articles
			perm.PolicyFor("editor").WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:articles:*").Given(perm.Conditions{
				"owner_id": &ladon.EqualsSubjectCondition{},
			}),
		).SubjectsFunc(func(r *http.Request) (ss []string) {
			user := getCurrentUser(r)
			ss = append(ss, user.ID)
			ss = append(ss, user.Roles...)
			return ss
		}).ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
			c := make(perm.Context)
			for _, obj := range objs {
				switch v := obj.(type) {
				case *Article:
					c["owner_id"] = v.OwnerID
				}
			}
			return c
		}),
	)
	// @snippet_end

	// @snippet_begin(PermissionVerbose)
	perm.Verbose = true
	// @snippet_end

	var r *http.Request
	var user interface{}
	// @snippet_begin(PermissionNewVerifier)
	verifier := perm.NewVerifier("module_users", presetsBuilder.GetPermission())
	// @snippet_end
	// @snippet_begin(PermissionVerifierCheck)
	if verifier.Do("ban").ObjectOn(user).WithReq(r).IsAllowed() == nil {
		// ui: show the ban button
		// action: can execute the ban action
	}
	// @snippet_end
	// @snippet_begin(PermissionAddCustomPolicy)
	perm.PolicyFor("super_admin").WhoAre(perm.Allowed).ToDo("ban").On(":module_users:*")
	// @snippet_end
}
