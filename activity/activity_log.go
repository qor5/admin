package activity

import (
	"gorm.io/gorm"
)

const (
	ActionView       = "View"
	ActionEdit       = "Edit"
	ActionCreate     = "Create"
	ActionDelete     = "Delete"
	ActionCreateNote = "CreateNote"
)

type ActivityLog struct {
	gorm.Model

	UserID     uint
	Creator    User   `gorm:"serializer:json"`
	Action     string `gorm:"index"`
	ModelKeys  string `gorm:"index"`
	ModelName  string `gorm:"index"`
	ModelLabel string
	ModelLink  string
	ModelDiffs string `sql:"type:text;"`
	Comment    string `gorm:"type:text;"`
	Number     int64
}
