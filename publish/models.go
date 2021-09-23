package publish

import "time"

type Status struct {
	PublishStatus string
}

type Schedule struct {
	ScheduledStartAt *time.Time `gorm:"index"`
	ScheduledEndAt   *time.Time `gorm:"index"`
}

type Version struct {
	VersionName     string `gorm:"primary_key;size:128"`
	VersionPriority string `gorm:"index"`
}

//
//type PublishStatus struct {
//	ModelName        string     `gorm:"primary_key"`
//	ID               uint       `gorm:"primary_key"`
//	Locale           string     `gorm:"primary_key"`
//	VersionName      string     `gorm:"primary_key;size:128"`
//	ScheduledStartAt *time.Time `gorm:"index"`
//	ScheduledEndAt   *time.Time `gorm:"index"`
//	Status           string
//	Slug             string
//	UpdatedAt        time.Time
//	LocaleSlugs      string
//}
