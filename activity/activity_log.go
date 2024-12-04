package activity

import (
	"fmt"
	"strings"

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

var DefaultActions = []string{ActionCreate /* ActionView,*/, ActionEdit, ActionDelete, ActionNote}

func defaultActionLabels(msgr *Messages) map[string]string {
	return map[string]string{
		ActionCreate: msgr.ActionCreate,
		ActionView:   msgr.ActionView,
		ActionEdit:   msgr.ActionEdit,
		ActionDelete: msgr.ActionDelete,
		ActionNote:   msgr.ActionNote,
	}
}

type ActivityLog struct {
	gorm.Model

	UserID string `gorm:"index;not null;"`
	User   User   `gorm:"-"`

	Action     string `gorm:"index;not null;"`
	Hidden     bool   `gorm:"index;default:false;not null;"`
	ModelName  string `gorm:"index;not null;"`
	ModelKeys  string `gorm:"index;not null;"`
	ModelLabel string `gorm:"not null;"` // IMPROVE: need named resource sign
	ModelLink  string `gorm:"not null;"`
	Detail     string `gorm:"not null;"`
	Scope      string `gorm:"index;"`
}

func (v *ActivityLog) AfterMigrate(tx *gorm.DB, tablePrefix string) error {
	s, err := ParseSchemaWithDB(tx, v)
	if err != nil {
		return err
	}
	tableName := tablePrefix + s.Table

	tableBare := tableName
	if tables := strings.Split(tableName, "."); len(tables) == 2 {
		tableBare = tables[1]
	}
	uix := fmt.Sprintf(`uix_%s_user_id_model_name_keys_action_lastview`, tableBare)
	if err := tx.Exec(fmt.Sprintf(`
		CREATE UNIQUE INDEX IF NOT EXISTS %s
		ON %s (user_id, model_name, model_keys)
		WHERE action = '%s' AND deleted_at IS NULL
	`, uix, tableName, ActionLastView)).Error; err != nil {
		return errors.Wrapf(err, "failed to create index %s", uix)
	}
	return nil
}
