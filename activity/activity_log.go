package activity

import (
	"gorm.io/gorm"
)

const (
	ActionView   = "View"
	ActionEdit   = "Edit"
	ActionCreate = "Create"
	ActionDelete = "Delete"
	ActionNote   = "Note"
)

var DefaultActions = []string{ActionView, ActionEdit, ActionCreate, ActionDelete, ActionNote}

type ActivityLog struct {
	gorm.Model

	CreatorID uint `gorm:"index"`
	Creator   User `gorm:"-"`

	Action     string `gorm:"index"`
	ModelKeys  string `gorm:"index"`
	ModelName  string `gorm:"index"`
	ModelLabel string
	ModelLink  string
	Detail     string `gorm:"type:text;"`
}
