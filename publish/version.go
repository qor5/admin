package publish

type Version struct {
	VersionName       string `gorm:"primary_key;size:128"`
	ParentVersionName string
	VersionPriority   string `gorm:"index"`
}

func (version Version) GetVersionName() string {
	return version.VersionName
}

func (version *Version) SetVersionName(v string) {
	version.VersionName = v
	return
}
