package activity

import (
	"fmt"
	"strings"
)

type Messages struct {
	Activities   string
	ActionAll    string
	ActionView   string
	ActionEdit   string
	ActionCreate string
	ActionDelete string
	ActionNote   string

	ModelUserID    string
	ModelCreatedAt string
	ModelAction    string
	ModelUser      string
	ModelKeys      string
	ModelName      string
	ModelLabel     string
	ModelLink      string
	ModelDiffs     string

	FilterAction    string
	FilterCreatedAt string
	FilterUser      string
	FilterModel     string
	FilterModelKeys string

	DiffDetail  string
	DiffAdd     string
	DiffDelete  string
	DiffChanges string
	DiffField   string
	DiffOld     string
	DiffNew     string
	DiffValue   string

	AddedANote                    string
	LastEditedAtTemplate          string
	EditedNFieldsTemplate         string
	MoreInfo                      string
	Created                       string
	Viewed                        string
	Deleted                       string
	PerformActionNoDetailTemplate string
	PerformActionTemplate         string
	AddNote                       string
	UnknownUser                   string
	NoteCannotBeEmpty             string
	FailedToCreateNote            string
	SuccessfullyCreatedNote       string
	FailedToGetCurrentUser        string
	FailedToGetNote               string
	YouAreNotTheNoteUser          string
	FailedToUpdateNote            string
	SuccessfullyUpdatedNote       string
	FailedToDeleteNote            string
	SuccessfullyDeletedNote       string
	DeleteNoteDialogTitle         string
	DeleteNoteDialogText          string
	Cancel                        string
	Delete                        string
	NoActivitiesYet               string
	ViewAll                       string
	CannotShowMore                string

	HeaderNotes string

	ActivityLogs string
	ActivityLog  string

	FilterTabsHasUnreadNotes string
}

func (msgr *Messages) LastEditedAt(desc string) string {
	return strings.NewReplacer("{desc}", desc).
		Replace(msgr.LastEditedAtTemplate)
}

func (msgr *Messages) EditedNFields(n int) string {
	return strings.NewReplacer("{n}", fmt.Sprint(n)).
		Replace(msgr.EditedNFieldsTemplate)
}

func (msgr *Messages) PerformAction(action, detail string) string {
	if detail == "" || detail == "null" || detail == "{}" {
		return strings.NewReplacer(
			"{action}", action,
		).Replace(msgr.PerformActionNoDetailTemplate)
	}
	return strings.NewReplacer(
		"{action}", action,
		"{detail}", detail,
	).Replace(msgr.PerformActionTemplate)
}

var Messages_en_US = &Messages{
	Activities:   "Activity",
	ActionAll:    "All",
	ActionView:   "View",
	ActionEdit:   "Edit",
	ActionCreate: "Create",
	ActionDelete: "Delete",
	ActionNote:   "Note",

	ModelUserID:    "Creator ID",
	ModelCreatedAt: "Create Time",
	ModelAction:    "Action",
	ModelUser:      "Creator",
	ModelKeys:      "Model Keys",
	ModelName:      "Model Name",
	ModelLabel:     "Menu Name",
	ModelLink:      "Link",
	ModelDiffs:     "Diffs",

	FilterAction:    "Action",
	FilterCreatedAt: "Create Time",
	FilterUser:      "Creator",
	FilterModel:     "Model Name",
	FilterModelKeys: "Model Keys",

	DiffDetail:  "Detail",
	DiffAdd:     "New",
	DiffDelete:  "Delete",
	DiffChanges: "Changes",
	DiffField:   "Field",
	DiffOld:     "Old",
	DiffNew:     "Now",
	DiffValue:   "Value",

	AddedANote:                    "Added a note :",
	LastEditedAtTemplate:          "(edited at {desc})",
	EditedNFieldsTemplate:         "Edited {n} fields",
	MoreInfo:                      "More Info",
	Created:                       "Created",
	Viewed:                        "Viewed",
	Deleted:                       "Deleted",
	PerformActionNoDetailTemplate: "Perform {action}",
	PerformActionTemplate:         "Perform {action} with {detail}",
	AddNote:                       "Add Note",
	UnknownUser:                   "Unknown",
	NoteCannotBeEmpty:             "Note cannot be empty",
	FailedToCreateNote:            "Failed to create note",
	SuccessfullyCreatedNote:       "Successfully created note",
	FailedToGetCurrentUser:        "Failed to get current user",
	FailedToGetNote:               "Failed to get note",
	YouAreNotTheNoteUser:          "You are not the creator of this note",
	FailedToUpdateNote:            "Failed to update note",
	SuccessfullyUpdatedNote:       "Successfully updated note",
	FailedToDeleteNote:            "Failed to delete note",
	SuccessfullyDeletedNote:       "Successfully deleted note",
	DeleteNoteDialogTitle:         "Delete Note",
	DeleteNoteDialogText:          "Are you sure you want to delete this note?",
	Cancel:                        "Cancel",
	Delete:                        "Delete",
	NoActivitiesYet:               "No activities yet",
	ViewAll:                       "View All",
	CannotShowMore:                "Reached the display limit, unable to load more.",

	HeaderNotes: "Notes",

	ActivityLogs: "Activity Logs",
	ActivityLog:  "Activity Log",

	FilterTabsHasUnreadNotes: "Has Unread Notes",
}

var Messages_zh_CN = &Messages{
	Activities:   "活动",
	ActionAll:    "全部",
	ActionView:   "查看",
	ActionEdit:   "编辑",
	ActionCreate: "创建",
	ActionDelete: "删除",
	ActionNote:   "备注",

	ModelUserID:    "操作者ID",
	ModelCreatedAt: "日期时间",
	ModelAction:    "操作",
	ModelUser:      "操作者",
	ModelKeys:      "主键",
	ModelName:      "对象",
	ModelLabel:     "菜单名",
	ModelLink:      "链接",
	ModelDiffs:     "差异",

	FilterAction:    "操作类型",
	FilterCreatedAt: "操作时间",
	FilterUser:      "操作者",
	FilterModel:     "对象",
	FilterModelKeys: "主键",

	DiffDetail:  "详情",
	DiffAdd:     "新加",
	DiffDelete:  "删除",
	DiffChanges: "修改",
	DiffField:   "字段",
	DiffOld:     "之前的值",
	DiffNew:     "当前的值",
	DiffValue:   "值",

	AddedANote:                    "添加了一个备注：",
	LastEditedAtTemplate:          "编辑于 {desc}",
	EditedNFieldsTemplate:         "编辑了 {n} 个字段",
	MoreInfo:                      "更多信息",
	Created:                       "创建",
	Viewed:                        "查看",
	Deleted:                       "删除",
	PerformActionNoDetailTemplate: "执行 {action}",
	PerformActionTemplate:         "执行 {action} 操作，详情为 {detail}",
	AddNote:                       "添加备注",
	UnknownUser:                   "未知",
	NoteCannotBeEmpty:             "备注不能为空",
	FailedToCreateNote:            "创建备注失败",
	SuccessfullyCreatedNote:       "成功创建备注",
	FailedToGetCurrentUser:        "获取当前用户失败",
	FailedToGetNote:               "获取备注失败",
	YouAreNotTheNoteUser:          "您不是备注的创建者",
	FailedToUpdateNote:            "更新备注失败",
	SuccessfullyUpdatedNote:       "成功更新备注",
	FailedToDeleteNote:            "删除备注失败",
	SuccessfullyDeletedNote:       "成功删除备注",
	DeleteNoteDialogTitle:         "删除备注",
	DeleteNoteDialogText:          "确定要删除此备注吗？",
	Cancel:                        "取消",
	Delete:                        "删除",
	NoActivitiesYet:               "暂无活动",
	ViewAll:                       "查看全部",
	CannotShowMore:                "已达到显示上限，无法加载更多。",

	HeaderNotes: "备注",

	ActivityLogs: "操作日志",
	ActivityLog:  "操作日志",

	FilterTabsHasUnreadNotes: "未读备注",
}

var Messages_ja_JP = &Messages{
	Activities:   "作業履歴",
	ActionAll:    "全て",
	ActionView:   "表示",
	ActionEdit:   "編集",
	ActionCreate: "作成する",
	ActionDelete: "削除",
	ActionNote:   "ノート",

	ModelUserID:    "作成者ID",
	ModelCreatedAt: "日時",
	ModelAction:    "アクション",
	ModelUser:      "作成者",
	ModelKeys:      "キー",
	ModelName:      "モデル名",
	ModelLabel:     "メニュー名",
	ModelLink:      "リンク",
	ModelDiffs:     "差分",

	FilterAction:    "アクション",
	FilterCreatedAt: "作成日時",
	FilterUser:      "作成者",
	FilterModel:     "モデル名",
	FilterModelKeys: "キー",

	DiffDetail:  "詳細",
	DiffAdd:     "追加",
	DiffDelete:  "削除",
	DiffChanges: "変更",
	DiffField:   "フィールド",
	DiffOld:     "以前",
	DiffNew:     "現在",
	DiffValue:   "値",

	AddedANote:                    "ノートを追加しました：",
	LastEditedAtTemplate:          "{desc} に編集",
	EditedNFieldsTemplate:         "{n}つのフィールドを編集しました",
	MoreInfo:                      "詳細情報",
	Created:                       "作成する",
	Viewed:                        "表示",
	Deleted:                       "削除",
	PerformActionNoDetailTemplate: "{action} を実行",
	PerformActionTemplate:         "{action} を実行し、{detail} を使用",
	AddNote:                       "ノートを追加",
	UnknownUser:                   "不明",
	NoteCannotBeEmpty:             "ノートは必須です",
	FailedToCreateNote:            "ノートの作成に失敗しました",
	SuccessfullyCreatedNote:       "ノートの作成に成功しました",
	FailedToGetCurrentUser:        "現在のユーザーの取得に失敗しました",
	FailedToGetNote:               "ノートの取得に失敗しました",
	YouAreNotTheNoteUser:          "このノートの作成者ではありません",
	FailedToUpdateNote:            "ノートの更新に失敗しました",
	SuccessfullyUpdatedNote:       "ノートの更新に成功しました",
	FailedToDeleteNote:            "ノートの削除に失敗しました",
	SuccessfullyDeletedNote:       "ノートの削除に成功しました",
	DeleteNoteDialogTitle:         "ノートを削除",
	DeleteNoteDialogText:          "このノートを削除してもよろしいですか？",
	Cancel:                        "キャンセル",
	Delete:                        "削除",
	NoActivitiesYet:               "まだアクティビティはありません",
	ViewAll:                       "全て表示",
	CannotShowMore:                "表示上限に達しました。これ以上読み込むことができません。",

	HeaderNotes: "ノート",

	ActivityLogs: "作業履歴",
	ActivityLog:  "作業履歴",

	FilterTabsHasUnreadNotes: "未読ノート",
}
