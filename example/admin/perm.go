package admin

import (
	"net/http"

	"github.com/goplaid/x/perm"
	"github.com/qor/qor5/media/views"
)

// TODO use perm builder from persets builder instead
var tempPermBuilder = perm.New().Policies(
	perm.They("developer").Are(perm.Allowed).ToDo(perm.Anything).On("*"),
	perm.They("developer").Are(perm.Denied).ToDo(views.PermUpload).On("*"),
	perm.They("developer").Are(perm.Denied).ToDo(views.PermDelete).On("*:media_libraries:5"),
	perm.They("developer").Are(perm.Denied).ToDo(views.PermUpdateDesc).On("*:media_libraries:6"),
).
	SubjectsFunc(func(r *http.Request) []string {
		return []string{"developer"}
	}).
	ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
		return perm.Context{
			"owner": "1",
		}
	})
