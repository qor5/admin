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
	unreadNote.CreatorID = fmt.Sprint(user.ID)
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

		if err = activity.MarkAllNotesAsRead(db, fmt.Sprint(u.ID)); err != nil {
			return r, err
		}

		r.Reload = true
		return
	}
}
