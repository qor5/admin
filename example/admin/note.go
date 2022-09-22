package admin

import (
	"encoding/json"
	"fmt"

	"github.com/goplaid/web"
	"gorm.io/gorm"
)

var hasUnreadNotesQuery = `(%v.id in (
WITH subquery AS (
    SELECT qn.resource_id, qn.resource_type
    FROM (SELECT resource_id, resource_type, count(*) AS count
          FROM qor_notes
          WHERE deleted_at IS NULL
          GROUP BY resource_id, resource_type) AS qn
             LEFT JOIN (SELECT resource_id, number
                        FROM user_notes
                        WHERE user_id = %v) AS un
                       ON qn.resource_id = un.resource_id
    WHERE qn.resource_type = '%v' AND qn.count > coalesce(un.number, 0)
)

SELECT DISTINCT split_part(resource_id, '_', 1)::integer
FROM subquery
))`

var NoteAfterCreateFunc = func(db *gorm.DB) (err error) {
	return db.Exec(`DELETE FROM "user_unread_notes";`).Error
}

type UserUnreadNote struct {
	gorm.Model

	UserID  uint `gorm:"uniqueIndex"`
	Content string
}

type Result struct {
	ResourceType string
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
    FROM (SELECT resource_id, resource_type, count(*) AS count
          FROM qor_notes
          WHERE deleted_at IS NULL
          GROUP BY resource_id, resource_type) AS qn
             LEFT JOIN (SELECT resource_id, number
                        FROM user_notes
                        WHERE user_id = %v) AS un
                       ON qn.resource_id = un.resource_id
    WHERE qn.count > coalesce(un.number, 0)
)

SELECT count(*), resource_type
FROM (SELECT DISTINCT split_part(resource_id, '_', 1)::integer AS resource_id, resource_type
      FROM subquery) AS sq
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
