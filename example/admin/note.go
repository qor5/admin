package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor5/example/models"
	"gorm.io/gorm"
)

var NoteAfterCreateFunc = func(db *gorm.DB, userID uint) (err error) {
	if err = db.AutoMigrate(&UserUnreadNote{}); err != nil {
		return
	}

	var users []models.User
	if userID != 0 {
		db.Where("id = ?", userID).Select("id").Find(&users)
	} else {
		db.Order("id ASC").Select("id").Find(&users)
	}

	for _, user := range users {
		var unreadNote UserUnreadNote
		db.FirstOrInit(&unreadNote, UserUnreadNote{UserID: user.ID})

		var data map[string]int
		data, err = getUnreadNotesCount(db, user)
		if err != nil {
			return
		}

		var content []byte
		content, err = json.Marshal(data)
		if err != nil {
			return
		}

		unreadNote.Content = string(content)
		if err = db.Save(&unreadNote).Error; err != nil {
			return
		}
	}

	return
}

type UserUnreadNote struct {
	gorm.Model

	UserID  uint   `sql:"not null"`
	Content string `sql:"not null"`
}

type Result struct {
	ResourceType string
	Count        int
}

func getUnreadNotesCount(db *gorm.DB, u models.User) (data map[string]int, err error) {
	data = make(map[string]int)

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
`, u.ID)).Scan(&results).Error; err != nil {
		return
	}

	for _, result := range results {
		data[result.ResourceType] = result.Count
	}

	return
}

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
