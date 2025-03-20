package publish

import (
	"fmt"
	"reflect"
	"time"

	"github.com/qor5/admin/v3/utils"
	"gorm.io/gorm"
)

func (*Version) GetNextVersion(t *time.Time) string {
	if t == nil {
		return ""
	}
	date := t.Local().Format("2006-01-02")
	return fmt.Sprintf("%s-v%02v", date, 1)
}

func (version *Version) CreateVersion(db *gorm.DB, paramID string, obj interface{}) (string, error) {
	date := db.NowFunc().Local().Format("2006-01-02")
	var count int64
	if err := utils.PrimarySluggerWhere(db.Unscoped(), obj, paramID, "version").
		Where("version like ?", date+"%").
		Order("version DESC").
		Count(&count).Error; err != nil {
		return "", err
	}

	versionName := fmt.Sprintf("%s-v%02v", date, count+1)
	version.Version = versionName
	version.VersionName = versionName
	return version.Version, nil
}

func IsVersion(obj interface{}) (IsVersion bool) {
	_, IsVersion = utils.GetStruct(reflect.TypeOf(obj)).(VersionInterface)
	return
}
