package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

// TODO:
var hasUnreadNotesQuery = `(%v.id in (
WITH subquery AS (
    SELECT al.model_keys as resource_id
    FROM (
             SELECT model_keys as resource_id, count(*) AS count
             FROM activity_logs
             WHERE model_name = '%v' AND action = 'create_note' AND deleted_at IS NULL
             GROUP BY model_keys
         ) AS al LEFT JOIN (
            SELECT model_keys as resource_id, number
            FROM activity_logs
            WHERE user_id = %v AND model_name = '%v' AND action LIKE '%%note%%' AND deleted_at IS NULL
         ) AS un ON al.resource_id = un.resource_id
    WHERE al.count > coalesce(un.number, 0)
)

SELECT DISTINCT split_part(resource_id, '_', 1)::integer
FROM subquery
))`

type UserUnreadNote struct {
	gorm.Model

	UserID  uint `gorm:"uniqueIndex"`
	Content string
}

type Result struct {
	ResourceType string
	ResourceID   string
	Count        int
}

func getUnreadNotesCount(ctx *web.EventContext, db *gorm.DB) (data map[string]int, err error) {
	data = make(map[string]int)

	user := getCurrentUser(ctx.R)
	if user == nil {
		return
	}

	var results []Result

	if err = db.Raw(fmt.Sprintf(`
WITH subquery AS (
    SELECT al.model_keys as resource_id, al.model_name as resource_type
    FROM (
        SELECT model_name, model_keys as resource_id, count(*) AS count
        FROM activity_logs
        WHERE action = 'create_note' AND deleted_at IS NULL
        GROUP BY model_keys, model_name
    ) AS al LEFT JOIN (
        SELECT model_name, model_keys as resource_id, number
        FROM activity_logs
        WHERE user_id = %v AND action LIKE '%%note%%' AND deleted_at IS NULL
    ) AS un ON al.resource_type = un.model_name AND al.resource_id = un.resource_id
    WHERE al.count > coalesce(un.number, 0)
)

SELECT count(*), resource_type
FROM (
         SELECT DISTINCT split_part(resource_id, '_', 1)::integer AS resource_id, resource_type
         FROM subquery
     ) AS sq
GROUP BY resource_type;
`, user.ID)).Scan(&results).Error; err != nil {
		return
	}

	for _, result := range results {
		data[result.ResourceType] = result.Count
	}

	var content []byte
	content, err = json.Marshal(data)
	if err != nil {
		return
	}

	var unreadNote activity.ActivityLog
	unreadNote.CreatorID = user.ID
	unreadNote.Detail = string(content) // TODO:
	unreadNote.Action = "unread_notes_count"
	if err = db.Save(&unreadNote).Error; err != nil {
		return
	}

	return
}

var noteMarkAllAsRead = "note_mark_all_as_read"

func markAllAsRead(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		u := getCurrentUser(ctx.R)
		if u == nil {
			return
		}

		if err = db.Transaction(func(tx *gorm.DB) (err1 error) {
			// Delete user-specific note read records
			if err1 = tx.Unscoped().Where("user_id = ? AND action LIKE '%%note%%'", u.ID).Delete(&activity.ActivityLog{}).Error; err1 != nil {
				return
			}

			// Fetch all notes count
			var results []Result
			if err = db.Raw(`
SELECT model_name as resource_type, model_keys as resource_id, count(*) AS count
FROM activity_logs
WHERE action = 'create_note' AND deleted_at IS NULL
GROUP BY model_name, model_keys;
`).Scan(&results).Error; err != nil {
				return
			}

			var userNotes []activity.ActivityLog
			for _, result := range results {
				un := activity.ActivityLog{
					Creator: activity.User{
						ID:   u.ID,
						Name: u.Name,
					},
					ModelName: result.ResourceType,
					ModelKeys: result.ResourceID,
					Action:    fmt.Sprintf("read_note:%d", result.Count),
					Detail:    fmt.Sprint(int64(result.Count)), // TODO:
				}
				userNotes = append(userNotes, un)
			}

			if err1 = tx.Create(&userNotes).Error; err1 != nil {
				return
			}

			// Update unread notes count
			var unreadNote activity.ActivityLog
			if err1 = db.Where("user_id = ? AND action = 'unread_notes_count'", u.ID).First(&unreadNote).Error; err1 != nil {
				return
			}
			unreadNote.Detail = "{}" // TODO:
			if err1 = db.Save(&unreadNote).Error; err1 != nil {
				return
			}

			return
		}); err != nil {
			return
		}

		r.Reload = true
		return
	}
}
