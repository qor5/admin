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
	SetID(uint)
	SetCreatedAt(time.Time)
	SetCreator(string)
	SetAction(string)
	SetModelKeys(string)
	SetModelName(string)
	SetModelLink(string)
	SetModelDiff(string)
	SetLocale(string)
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

func (al *ActivityLog) SetID(id uint) {
	al.ID = id
}

func (al *ActivityLog) SetCreatedAt(t time.Time) {
	al.CreatedAt = t
}

func (al *ActivityLog) SetCreator(s string) {
	al.Creator = s
}

func (al *ActivityLog) SetAction(s string) {
	al.Action = s
}

func (al *ActivityLog) SetModelKeys(s string) {
	al.ModelKeys = s
}

func (al *ActivityLog) SetModelName(s string) {
	al.ModelName = s
}

func (al *ActivityLog) SetModelLink(s string) {
	al.ModelLink = s
}

func (al *ActivityLog) SetModelDiff(s string) {
	al.ModelDiffs = s
}

func (al *ActivityLog) SetLocale(s string) {
}

type ActivityLogWithLocale struct {
	ActivityLog
	// l10n.LocaleCreatable
}

func (al *ActivityLogWithLocale) SetLocale(s string) {
}

type ModelDiffs struct {
	Diffs []Diff
}

type Diff struct {
	Type string
	Path []string
	From string
	To   string
}
