package activity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type NoteCount struct {
	ModelName        string
	ModelKeys        string
	UnreadNotesCount int64
	TotalNotesCount  int64
}

func getNotesCounts(db *gorm.DB, tablePrefix string, creatorID string, modelName string, modelKeyses []string) ([]*NoteCount, error) {
	if creatorID == "" {
		return nil, errors.New("creatorID is required")
	}

	s, err := ParseSchemaWithDB(db, &ActivityLog{})
	if err != nil {
		return nil, err
	}
	tableName := tablePrefix + s.Table

	args := []any{
		ActionNote,
	}

	var explictWhere string
	if modelName != "" {
		explictWhere = ` AND model_name = ?`
		args = append(args, modelName)
	}
	if len(modelKeyses) > 0 {
		explictWhere += ` AND model_keys IN (?)`
		args = append(args, modelKeyses)
	}

	args = append(args, ActionLastView, creatorID)

	if modelName != "" {
		args = append(args, modelName)
	}
	if len(modelKeyses) > 0 {
		args = append(args, modelKeyses)
	}

	args = append(args, creatorID)

	raw := fmt.Sprintf(`
	WITH NoteRecords AS (
		SELECT model_name, model_keys, created_at, creator_id
		FROM %s
		WHERE action = ? AND deleted_at IS NULL
			%s
	),
	LastViewedAts AS (
		SELECT model_name, model_keys, MAX(updated_at) AS last_viewed_at
		FROM %s
		WHERE action = ? AND creator_id = ? AND deleted_at IS NULL
			%s
		GROUP BY model_name, model_keys
	)
	SELECT
		n.model_name,
		n.model_keys,
		COUNT(CASE WHEN n.creator_id <> ? AND (lva.last_viewed_at IS NULL OR n.created_at > lva.last_viewed_at) THEN 1 END) AS unread_notes_count,
		COUNT(*) AS total_notes_count
	FROM NoteRecords n
	LEFT JOIN LastViewedAts lva
		ON n.model_name = lva.model_name
		AND n.model_keys = lva.model_keys
	GROUP BY n.model_name, n.model_keys;`, tableName, explictWhere, tableName, explictWhere)

	counts := []*NoteCount{}
	if err := db.Raw(raw, args...).Scan(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func markAllNotesAsRead(db *gorm.DB, creatorID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var results []struct {
			ModelName    string
			ModelKeys    string
			ModelLabel   string
			ModelLink    string
			MaxCreatedAt time.Time
		}
		if err := tx.Model(&ActivityLog{}).
			Select("model_name, model_keys, MAX(model_label) AS model_label, MAX(model_link) AS model_link, MAX(created_at) AS max_created_at").
			Where("action = ?", ActionNote).
			Group("model_name, model_keys").Scan(&results).Error; err != nil {
			return errors.Wrap(err, "find created_at of last notes")
		}

		if len(results) <= 0 {
			return nil
		}

		if err := tx.Unscoped().Where("creator_id = ? AND action = ?", creatorID, ActionLastView).Delete(&ActivityLog{}).Error; err != nil {
			return errors.Wrap(err, "delete last views")
		}

		var logs []ActivityLog
		for _, v := range results {
			log := ActivityLog{
				CreatorID:  creatorID,
				Action:     ActionLastView,
				Hidden:     true,
				ModelName:  v.ModelName,
				ModelKeys:  v.ModelKeys,
				ModelLabel: v.ModelLabel,
				ModelLink:  v.ModelLink,
				Detail:     "null",
			}
			log.CreatedAt = v.MaxCreatedAt
			log.UpdatedAt = v.MaxCreatedAt
			logs = append(logs, log)
		}

		if err := tx.Create(logs).Error; err != nil {
			return errors.Wrap(err, "create new last views")
		}

		return nil
	})
}

func sqlConditionHasUnreadNotes(db *gorm.DB, tablePrefix string, creatorID string, modelName string, columns []string, sep string, columnPrefix string) (string, error) {
	a := strings.Join(lo.Map(columns, func(v string, _ int) string {
		return fmt.Sprintf("%s%s::text", columnPrefix, v)
	}), ",")
	b := strings.Join(lo.Map(columns, func(v string, i int) string {
		return fmt.Sprintf(`split_part(n.model_keys, '%s', %d) AS %s`, sep, i+1, v)
	}), ",\n")

	s, err := ParseSchemaWithDB(db, &ActivityLog{})
	if err != nil {
		return "", err
	}
	tableName := tablePrefix + s.Table

	return fmt.Sprintf(`
	(%s) IN (
	    WITH NoteRecords AS (
		    SELECT model_name, model_keys, created_at, creator_id
		    FROM %s
		    WHERE action = '%s' AND deleted_at IS NULL
		        AND model_name = '%s'
		),
		LastViewedAts AS (
		    SELECT model_name, model_keys, MAX(updated_at) AS last_viewed_at
		    FROM %s
		    WHERE action = '%s' AND creator_id = '%s' AND deleted_at IS NULL
		        AND model_name = '%s'
		    GROUP BY model_name, model_keys
		)
		
	    SELECT
%s
	    FROM NoteRecords n
	    LEFT JOIN LastViewedAts lva
	        ON n.model_name = lva.model_name
	        AND n.model_keys = lva.model_keys
	    WHERE n.creator_id <> '%s' 
	        AND (lva.last_viewed_at IS NULL OR n.created_at > lva.last_viewed_at)
	    GROUP BY n.model_keys
    )`, a, tableName, ActionNote, modelName, tableName, ActionLastView, creatorID, modelName, b, creatorID), nil
}

func (ab *Builder) GetNotesCounts(ctx context.Context, modelName string, modelKeyses []string) ([]*NoteCount, error) {
	user, err := ab.currentUserFunc(ctx)
	if err != nil {
		return nil, err
	}
	return getNotesCounts(ab.db, ab.tablePrefix, user.ID, modelName, modelKeyses)
}

func (ab *Builder) MarkAllNotesAsRead(ctx context.Context) error {
	user, err := ab.currentUserFunc(ctx)
	if err != nil {
		return err
	}
	return markAllNotesAsRead(ab.db, user.ID)
}

// SQLConditionHasUnreadNotes returns a SQL condition that can be used in a WHERE clause to filter records that have unread notes.
// Note that this method requires the applied db to be amb.ab.db, not any other db
func (amb *ModelBuilder) SQLConditionHasUnreadNotes(ctx context.Context, columnPrefix string) (string, error) {
	user, err := amb.ab.currentUserFunc(ctx)
	if err != nil {
		return "", err
	}
	return sqlConditionHasUnreadNotes(amb.ab.db, amb.ab.tablePrefix, user.ID, ParseModelName(amb.ref), amb.keyColumns, ModelKeysSeparator, columnPrefix)
}
