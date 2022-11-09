package activity

import (
	"github.com/qor5/x/perm"
	"github.com/qor5/admin/presets"
)

var PermPolicy = perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
	ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).On("*:activity_logs").On("*:activity_logs:*")
