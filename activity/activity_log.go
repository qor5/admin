package activity

import (
	"time"
)

const (
	ActivityView   = "View"
	ActivityEdit   = "Edit"
	ActivityCreate = "Create"
	ActivityDelete = "Delete"
)

type ActivityLogInterface interface {
	SetCreatedAt(time.Time)
	SetCreator(string)
	GetCreator() string
	SetAction(string)
	GetAction() string
	SetModelKeys(string)
	GetModelKeys() string
	SetModelName(string)
	GetModelName() string
	SetModelLink(string)
	GetModelLink() string
	SetModelDiffs(string)
	GetModelDiffs() string
	SetLocale(string)
	GetLocale() string
}

type ActivityLog struct {
	ID         uint `gorm:"primary_key"`
	CreatedAt  time.Time
	Creator    string
	Action     string
	ModelKeys  string
	ModelName  string
	ModelLink  string
	ModelDiffs string `sql:"type:text;"`
}

func (al *ActivityLog) SetCreatedAt(t time.Time) {
	al.CreatedAt = t
}

func (al *ActivityLog) SetCreator(s string) {
	al.Creator = s
}

func (al *ActivityLog) GetCreator() string {
	return al.Creator
}

func (al *ActivityLog) SetAction(s string) {
	al.Action = s
}

func (al *ActivityLog) GetAction() string {
	return al.Action
}

func (al *ActivityLog) SetModelKeys(s string) {
	al.ModelKeys = s
}

func (al *ActivityLog) GetModelKeys() string {
	return al.ModelKeys
}

func (al *ActivityLog) SetModelName(s string) {
	al.ModelName = s
}

func (al *ActivityLog) GetModelName() string {
	return al.ModelName
}

func (al *ActivityLog) SetModelLink(s string) {
	al.ModelLink = s
}

func (al *ActivityLog) GetModelLink() string {
	return al.ModelLink
}

func (al *ActivityLog) SetModelDiffs(s string) {
	al.ModelDiffs = s
}

func (al *ActivityLog) GetModelDiffs() string {
	return al.ModelDiffs
}

func (al *ActivityLog) SetLocale(s string) {
}

func (al *ActivityLog) GetLocale() string {
	return ""
}

type ActivityLocaleLog struct {
	ActivityLog
	// l10n.LocaleCreatable
}
