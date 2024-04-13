package activity

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/x/v3/perm"
)

var permPolicy = perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
	ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).On("*:activity_logs").On("*:activity_logs:*")
