package admin

import (
	"encoding/json"
	"fmt"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

var hasUnreadNotesQuery = `(%v.id in (
WITH subquery AS (
    SELECT qn.resource_id
    FROM (
             SELECT resource_id, count(*) AS count
             FROM qor_notes
             WHERE resource_type = '%v' AND deleted_at IS NULL
             GROUP BY resource_id
         ) AS qn LEFT JOIN (
            SELECT resource_id, number
            FROM user_notes
            WHERE user_id = %v AND resource_type = '%v' AND deleted_at IS NULL
         ) AS un ON qn.resource_id = un.resource_id
    WHERE qn.count > coalesce(un.number, 0)
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

	var unreadNote UserUnreadNote
	err = db.Where("user_id = ?", user.ID).First(&unreadNote).Error
	if err == nil {
		err = json.Unmarshal([]byte(unreadNote.Content), &data)
		return
	}

	if err == gorm.ErrRecordNotFound {
		var results []Result

		if err = db.Raw(fmt.Sprintf(`
WITH subquery AS (
    SELECT qn.resource_id, qn.resource_type
    FROM (
        SELECT resource_type, resource_id, count(*) AS count
        FROM qor_notes
        WHERE deleted_at IS NULL
        GROUP BY resource_id, resource_type
    ) AS qn LEFT JOIN (
        SELECT resource_type, resource_id, number
        FROM user_notes
        WHERE user_id = %v AND deleted_at IS NULL
    ) AS un ON qn.resource_type = un.resource_type AND qn.resource_id = un.resource_id
    WHERE qn.count > coalesce(un.number, 0)
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

		unreadNote.UserID = user.ID
		unreadNote.Content = string(content)
		if err = db.Save(&unreadNote).Error; err != nil {
			return
		}
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
			if err1 = tx.Unscoped().Where("user_id = ?", u.ID).Delete(&activity.UserNote{}).Error; err1 != nil {
				return
			}

			var results []Result

			if err = db.Raw(`
SELECT resource_type, resource_id, count(*) AS count
FROM qor_notes
WHERE deleted_at IS NULL
GROUP BY resource_type, resource_id;
`).Scan(&results).Error; err != nil {
				return
			}

			var userNotes []activity.UserNote
			for _, result := range results {
				un := activity.UserNote{
					UserID:       u.ID,
					ResourceType: result.ResourceType,
					ResourceID:   result.ResourceID,
					Number:       int64(result.Count),
				}
				userNotes = append(userNotes, un)
			}

			if err1 = tx.Create(userNotes).Error; err1 != nil {
				return
			}

			var unreadNote UserUnreadNote
			if err1 = db.Where("user_id = ?", u.ID).First(&unreadNote).Error; err1 != nil {
				return
			}
			unreadNote.Content = "{}"
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
