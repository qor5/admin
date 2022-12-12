package publish

import (
	"fmt"
	"reflect"

	"github.com/qor5/admin/utils"
	"gorm.io/gorm"
)

type Version struct {
	Version       string `gorm:"primary_key;size:128"`
	VersionName   string
	ParentVersion string
	OnlineVersion bool `gorm:"index;default:false"`
}

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

func (version *Version) CreateVersion(db *gorm.DB, paramID string, obj interface{}) string {
	date := db.NowFunc().Format("2006-01-02")
	var count int64
	utils.PrimarySluggerWhere(db.Unscoped(), obj, paramID, "version").
		Where("version like ?", date+"%").
		Order("version DESC").
		Count(&count)

	versionName := fmt.Sprintf("%s-v%02v", date, count+1)
	version.Version = versionName
	version.VersionName = versionName
	return version.Version
}

func IsVersion(obj interface{}) (IsVersion bool) {
	_, IsVersion = utils.GetStruct(reflect.TypeOf(obj)).(VersionInterface)
	return
}
