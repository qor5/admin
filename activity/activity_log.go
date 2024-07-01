package activity

import (
	"strings"

	"gorm.io/gorm"
)

const (
	ActivityView   = "View"
	ActivityEdit   = "Edit"
	ActivityCreate = "Create"
	ActivityDelete = "Delete"
)

type CreatorInterface interface {
	GetID() uint
	GetName() string
}

type ActivityLog struct {
	gorm.Model

	UserID     uint
	Creator    string
	Action     string
	ModelKeys  string `gorm:"index"`
	ModelName  string `gorm:"index"`
	ModelLabel string

	ModelLink  string
	ModelDiffs string `sql:"type:text;"`

	Content      string `gorm:"type:text;"`
	ResourceID   string `gorm:"index"`
	ResourceType string `gorm:"index"`

	Number int64

	TableNameOverride string `gorm:"-"`
}

func (a *ActivityLog) TableName() string {
	if a.TableNameOverride != "" {
		return a.TableNameOverride
	}
	return "activity_logs"
}

func (n *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(n.Content) == "" {
		//return errors.New("note cannot be empty")
	}
	return nil
}
