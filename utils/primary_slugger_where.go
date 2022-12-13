package utils

import (
	"fmt"

	"github.com/qor5/admin/presets"
	"gorm.io/gorm"
)

func PrimarySluggerWhere(db *gorm.DB, obj interface{}, id string, withoutKeys ...string) *gorm.DB {
	wh := db.Model(obj)

	if id == "" {
		return wh
	}

	if slugger, ok := obj.(presets.SlugDecoder); ok {
		cs := slugger.PrimaryColumnValuesBySlug(id)
		for key, value := range cs {
			if !Contains(withoutKeys, key) {
				wh = wh.Where(fmt.Sprintf("%s = ?", key), value)
			}
		}
	} else {
		wh = wh.Where("id =  ?", id)
	}

	return wh
}
