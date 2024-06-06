package basics

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Permissions = Doc(
	Markdown(`
QOR5 permission is based on https://github.com/ory/ladon.  
A piece of policy looks like this:  
**Who** is **able** to do **what** on **something** (with given some **context**)  
    `),
	ch.Code(generated.PermissionSyntax).Language("go"),
	Markdown(fmt.Sprintf(`
## Who - Subject
Typically in admin system, they are roles like %s, %s.  
Use %s to fetch current subjects:
    `, "`Admin`", "`Super Admin`", "`SubjectsFunc`")),
	ch.Code(generated.PermissionSubjectsFunc).Language("go"),
	Markdown(fmt.Sprintf(`
## Able - Effect
- %s
- %s

## What - Action
presets has a list of actions:
- %s
- %s
- %s
- %s
- %s

And you can define other specific actions if needed.
## Something - Resource
An arbitrary unique resource name.  
The presets builtin resource format is %s.  
For example %s represents the user record with id 1 under uri user_management.  
Use %s as wildcard.
## Context - Condition
Optional.  
The current context that containing condition information about the resource.  
Use %s to set the context:
    `,
		strings.TrimRight(generated.PermissionAllowed, ","),
		strings.TrimRight(generated.PermissionDenied, ","),
		strings.TrimRight(generated.PermissionPermList, ","),
		strings.TrimRight(generated.PermissionPermGet, ","),
		strings.TrimRight(generated.PermissionPermCreate, ","),
		strings.TrimRight(generated.PermissionPermUpdate, ","),
		strings.TrimRight(generated.PermissionPermDelete, ","),
		"`:presets:mg_menu_group:uri:resource_rn:f_field:`",
		"`:presets:user_management:users:1:`",
		"`*`",
		"`ContextFunc`",
	)),
	ch.Code(generated.PermissionContextFunc).Language("go"),
	Markdown(fmt.Sprintf(`
Policy uses %s to set conditions:  
    `, "`Given`")),
	ch.Code(generated.PermissionGivenFunc).Language("go"),
	Markdown(fmt.Sprintf(`
## Custom Action
Let's say there is a button on User detailing page used to ban the user. And only %s users have permission to execute this action.  
First, create a verifier
    `, "`super_admin`")),
	ch.Code(generated.PermissionNewVerifier).Language("go"),
	Markdown(fmt.Sprintf(`
Then inject this verifier to relevant logic, such as
- whether to show the ban button.
- validate permission before execute the ban action.
    `)),
	ch.Code(generated.PermissionVerifierCheck).Language("go"),
	Markdown(`
Finally, add policy
    `),
	ch.Code(generated.PermissionAddCustomPolicy).Language("go"),
	Markdown(`
## Example
    `),
	ch.Code(generated.PermissionExample).Language("go"),
	Markdown(`
## Debug
    `),
	ch.Code(generated.PermissionVerbose).Language("go"),
	Markdown(`
prints permission logs which is very helpful for debugging the permission policies:
    `),
	ch.Code(`
have permission: true, req: &ladon.Request{Resource:":presets:articles:", Action:"presets:list", Subject:"viewer", Context:ladon.Context(nil)}
have permission: true, req: &ladon.Request{Resource:":presets:articles:articles:1:", Action:"presets:update", Subject:"viewer", Context:ladon.Context(nil)}
have permission: false, req: &ladon.Request{Resource:":presets:articles:articles:2:", Action:"presets:update", Subject:"viewer", Context:ladon.Context(nil)}
    `).Language("plain"),
).Title("Permissions").
	Slug("presets-guide/permissions")
