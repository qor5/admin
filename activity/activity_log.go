package activity

import (
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	ActionView     = "View"
	ActionEdit     = "Edit"
	ActionCreate   = "Create"
	ActionDelete   = "Delete"
	ActionNote     = "Note"
	ActionLastView = "LastView" // hidden and only for internal use
)

var DefaultActions = []string{ActionView, ActionEdit, ActionCreate, ActionDelete, ActionNote}

type ActivityLog struct {
	gorm.Model

	CreatorID string `gorm:"index;not null;"`
	Creator   User   `gorm:"-"`

	Action     string `gorm:"index;not null;"`
	Hidden     bool   `gorm:"index;default:false;not null;"`
	ModelName  string `gorm:"index;not null;"`
	ModelKeys  string `gorm:"index;not null;"`
	ModelLabel string `gorm:"not null;"`
	ModelLink  string `gorm:"not null;"`
	Detail     string `gorm:"not null;"`
}

func (*ActivityLog) AfterMigrate(tx *gorm.DB) error {
	if err := tx.Exec(fmt.Sprintf(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_model_name_keys_action_lastview
		ON activity_logs (model_name, model_keys)
		WHERE action = '%s' AND deleted_at IS NULL
	`, ActionLastView)).Error; err != nil {
		return errors.Wrap(err, "failed to create index idx_model_name_keys_action_lastview")
	}
	return nil
}
