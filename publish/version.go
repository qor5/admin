package publish

import (
	"fmt"
	"reflect"
	"time"

	"github.com/qor5/admin/v3/utils"
	"gorm.io/gorm"
)

// @snippet_begin(PublishVersion)
type Version struct {
	Version       string `gorm:"primary_key;size:128;not null;default:null"`
	VersionName   string
	ParentVersion string
}

// @snippet_end

func (version Version) GetVersion() string {
	return version.Version
}

func (version *Version) SetVersion(v string) {
	version.Version = v
}

func (version Version) GetVersionName() string {
	return version.VersionName
}

func (version *Version) SetVersionName(v string) {
	version.VersionName = v
}

func (version *Version) GetNextVersion(t *time.Time) string {
	if t == nil {
		return ""
	}
	date := t.Format("2006-01-02")
	return fmt.Sprintf("%s-v%02v", date, 1)
}

func (version *Version) CreateVersion(db *gorm.DB, paramID string, obj interface{}) (string, error) {
	date := db.NowFunc().Format("2006-01-02")
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
