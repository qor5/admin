package activity

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

func fetchOldWithSlug(ref any, slug string, db *gorm.DB) (any, bool) {
	if slug == "" {
		return fetchOld(ref, db)
	}

	var (
		rt  = reflect.Indirect(reflect.ValueOf(ref)).Type()
		old = reflect.New(rt).Interface()
	)

	if slugger, ok := ref.(presets.SlugDecoder); ok {
		cs := slugger.PrimaryColumnValuesBySlug(slug)
		for key, value := range cs {
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
		}
	} else {
		db = db.Where("id = ?", slug)
	}

	if db.First(old).Error != nil {
		return nil, false
	}

	return old, true
}

func fetchOld(ref any, db *gorm.DB) (any, bool) {
	var (
		rtRef = reflect.Indirect(reflect.ValueOf(ref))
		old   = reflect.New(rtRef.Type()).Interface()
		sqls  []string
		vars  []any
	)

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(ref); err != nil {
		return nil, false
	}

	for _, dbName := range stmt.Schema.DBNames {
		if field := stmt.Schema.LookUpField(dbName); field != nil && field.PrimaryKey {
			if value, isZero := field.ValueOf(db.Statement.Context, rtRef); !isZero {
				sqls = append(sqls, fmt.Sprintf("%v = ?", dbName))
				vars = append(vars, value)
			}
		}
	}

	if len(sqls) == 0 || len(vars) == 0 || len(sqls) != len(vars) {
		return nil, false
	}

	if db.Where(strings.Join(sqls, " AND "), vars...).First(old).Error != nil {
		return nil, false
	}

	return old, true
}

// TODO: 这个其实也很好奇，为什么不使用 gorm 的 scheme parse 来搞？
// getPrimaryKey get primary keys from a model
func getPrimaryKey(t reflect.Type) (keys []string) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primary") {
			keys = append(keys, t.Field(i).Name)
			continue
		}

		if t.Field(i).Type.Kind() == reflect.Ptr && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type.Elem())...)
		}

		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type)...)
		}
	}
	return
}

func getBasicModel(m any) any {
	if preset, ok := m.(*presets.ModelBuilder); ok {
		return preset.NewModel()
	}

	return m
}

// func GetUnreadNotesCount(db *gorm.DB, userID uint, resourceType, resourceID string) (int64, error) {
// 	var total int64
// 	if err := db.Model(&ActivityLog{}).Where("model_name = ? AND model_keys = ? AND action = ?", resourceType, resourceID, ActionCreateNote).Count(&total).Error; err != nil {
// 		return 0, err
// 	}

// 	if total == 0 {
// 		return 0, nil
// 	}

// 	// TODO: 这个逻辑貌似不太对
// 	var userNote ActivityLog
// 	if err := db.Where("creator_id = ? AND model_name = ? AND model_keys = ?", userID, resourceType, resourceID).First(&userNote).Error; err != nil {
// 		if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			return 0, err
// 		}
// 	}

// 	return total - userNote.Number, nil
// }

// func handleError(err error, r *web.EventResponse, errorMessage string) {
// 	if err != nil {
// 		log.Println(errorMessage, err) // TODO:
// 		presets.ShowMessage(r, errorMessage, "error")
// 	}
// }