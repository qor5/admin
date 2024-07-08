package activity

import (
	"gorm.io/gorm"
)

const (
	ActionView       = "View"
	ActionEdit       = "Edit"
	ActionCreate     = "Create"
	ActionDelete     = "Delete"
	ActionCreateNote = "CreateNote" // TODO:
)

type ActivityLog struct {
	gorm.Model

	UserID     uint
	Creator    User   `gorm:"serializer:json"`
	Action     string `gorm:"index"`
	ModelKeys  string `gorm:"index"`
	ModelName  string `gorm:"index"`
	ModelLabel string
	ModelLink  string
	ModelDiffs string `sql:"type:text;"`
	// TODO: 这个貌似只是为了给 note 使用的，或许直接和 ModelDiffs 合并为一个字段，例如叫 Description 或者 Content 或者 Detail ?
	Comment string `gorm:"type:text;"`
	// TODO: 这个字段貌似只是为了记录已读数量，改成单独一个表来存储这个信息是否更合适？或者也应该是一个单独的 action ，而信息存储到 Detail 里？
	Number int64
	// TODO: 对于展示 timeline 来说，不同的 action 应该有不同的展示样式，例如是要对 action 指定 detail 的 parse 成 compnent 的 func，甚至是列表详情中和timeline要有不同的样式？
}
