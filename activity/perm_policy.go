package activity

import (
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
)

var PermPolicy = perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).
	ToDo(presets.PermUpdate, presets.PermDelete, presets.PermCreate).On("*:activity_logs").On("*:activity_logs:*")
