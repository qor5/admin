package admin

import (
	"net/http"
	"strings"

	"github.com/goplaid/x/perm"
	"github.com/ory/ladon"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/media/views"
)

// TODO use perm builder from persets builder instead
var tempPermBuilder = perm.New().Policies(
	perm.They("developer").Are(perm.Allowed).ToDo(perm.Anything).On("*"),
	perm.They("s1").Are(perm.Allowed).ToDo(views.PermUpload, views.PermDelete, views.PermUpdateDesc).On("*"),
	perm.They("s2").Are(perm.Allowed).ToDo(views.PermUpload).On("*"),
	perm.They("s3").Are(perm.Allowed).ToDo(views.PermDelete).On("*"),
	perm.They("s4").Are(perm.Allowed).ToDo(views.PermUpdateDesc).On("*"),
	perm.They("s5").Are(perm.Allowed).ToDo(views.PermUpdateDesc).On("*:media_libraries:6"),
	perm.They("s6").Are(perm.Allowed).ToDo(views.PermUpload, views.PermDelete, views.PermUpdateDesc).On("*"),
	perm.They("s6").Are(perm.Denied).ToDo(views.PermDelete, views.PermUpdateDesc).On("*").Given(
		perm.Conditions{
			"is_vip_file": &ladon.BooleanCondition{
				BooleanValue: true,
			},
		},
	),
).
	SubjectsFunc(func(r *http.Request) []string {
		c, err := r.Cookie("subjects")
		if err != nil {
			if err == http.ErrNoCookie {
				return []string{"developer"}
			}
			panic(err)
		}
		subjects := strings.Split(c.Value, ",")
		if len(subjects) == 0 {
			return []string{"developer"}
		}

		return subjects
	}).
	ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
		c := make(perm.Context)
		for _, obj := range objs {
			switch v := obj.(type) {
			case *media_library.MediaLibrary:
				if strings.HasPrefix(v.File.GetFileName(), "vip") {
					c["is_vip_file"] = true
				}
			}
		}

		return c
	})
