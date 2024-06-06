package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Role = Doc(
	Markdown(`
**Role** provides a UI interface to manage roles(subjects) and their permissions.  

1\.  enable permission DBPolicy
    `),
	ch.Code(generated.RolePermEnableDBPolicy).Language("go"),
	Markdown(`
2\. configure role  
set resources that you want to manage on interface
    `),
	ch.Code(generated.RoleSetResources).Language("go"),
	Markdown(`
(optional) set actions, the default value is the following 
    `),
	ch.Code(generated.RoleSetActions).Language("go"),
	Markdown(`
(optional) set editor subject to set who can edit **Role**
    `),
	ch.Code(generated.RoleSetEditorSubject).Language("go"),
	Markdown(`
attach role to presets builder
    `),
	ch.Code(generated.RoleAttachToPresetsBuilder).Language("go"),
).Title("Role").
	Slug("presets-guide/role")
