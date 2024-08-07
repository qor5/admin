package presets

import (
	"strings"
)

type Messages struct {
	SuccessfullyUpdated                        string
	Search                                     string
	New                                        string
	Update                                     string
	Delete                                     string
	Edit                                       string
	FormTitle                                  string
	OK                                         string
	Cancel                                     string
	Clear                                      string
	Create                                     string
	DeleteConfirmationTextTemplate             string
	CreatingObjectTitleTemplate                string
	EditingObjectTitleTemplate                 string
	ListingObjectTitleTemplate                 string
	DetailingObjectTitleTemplate               string
	FiltersClear                               string
	FiltersAdd                                 string
	FilterApply                                string
	FilterByTemplate                           string
	FiltersDateInTheLast                       string
	FiltersDateEquals                          string
	FiltersDateBetween                         string
	FiltersDateIsAfter                         string
	FiltersDateIsAfterOrOn                     string
	FiltersDateIsBefore                        string
	FiltersDateIsBeforeOrOn                    string
	FiltersDateDays                            string
	FiltersDateMonths                          string
	FiltersDateAnd                             string
	FiltersDateTo                              string
	FiltersNumberEquals                        string
	FiltersNumberBetween                       string
	FiltersNumberGreaterThan                   string
	FiltersNumberLessThan                      string
	FiltersNumberAnd                           string
	FiltersStringEquals                        string
	FiltersStringContains                      string
	FiltersMultipleSelectIn                    string
	FiltersMultipleSelectNotIn                 string
	PaginationRowsPerPage                      string
	ListingNoRecordToShow                      string
	ListingSelectedCountNotice                 string
	ListingClearSelection                      string
	BulkActionNoRecordsSelected                string
	BulkActionNoAvailableRecords               string
	BulkActionSelectedIdsProcessNoticeTemplate string
	ConfirmDialogPromptText                    string
	Language                                   string
	Colon                                      string
	NotFoundPageNotice                         string
	ButtonLabelActionsMenu                     string
	Save                                       string
	AddRow                                     string
}

func (msgr *Messages) DeleteConfirmationText(id string) string {
	return strings.NewReplacer("{id}", id).
		Replace(msgr.DeleteConfirmationTextTemplate)
}

func (msgr *Messages) CreatingObjectTitle(modelName string) string {
	return strings.NewReplacer("{modelName}", modelName).
		Replace(msgr.CreatingObjectTitleTemplate)
}

func (msgr *Messages) EditingObjectTitle(label string, name string) string {
	return strings.NewReplacer("{id}", name, "{modelName}", label).
		Replace(msgr.EditingObjectTitleTemplate)
}

func (msgr *Messages) ListingObjectTitle(label string) string {
	return strings.NewReplacer("{modelName}", label).
		Replace(msgr.ListingObjectTitleTemplate)
}

func (msgr *Messages) DetailingObjectTitle(label string, name string) string {
	return strings.NewReplacer("{id}", name, "{modelName}", label).
		Replace(msgr.DetailingObjectTitleTemplate)
}

func (msgr *Messages) BulkActionSelectedIdsProcessNotice(ids string) string {
	return strings.NewReplacer("{ids}", ids).
		Replace(msgr.BulkActionSelectedIdsProcessNoticeTemplate)
}

func (msgr *Messages) FilterBy(filter string) string {
	return strings.NewReplacer("{filter}", filter).
		Replace(msgr.FilterByTemplate)
}

var Messages_en_US = &Messages{
	SuccessfullyUpdated:            "Successfully Updated",
	Search:                         "Search",
	New:                            "New",
	Update:                         "Update",
	Delete:                         "Delete",
	Edit:                           "Edit",
	FormTitle:                      "Form",
	OK:                             "OK",
	Cancel:                         "Cancel",
	Clear:                          "Clear",
	Create:                         "Create",
	DeleteConfirmationTextTemplate: "Are you sure you want to delete object with id: {id}?",
	CreatingObjectTitleTemplate:    "New {modelName}",
	EditingObjectTitleTemplate:     "Editing {modelName} {id}",
	ListingObjectTitleTemplate:     "Listing {modelName}",
	DetailingObjectTitleTemplate:   "{modelName} {id}",
	FiltersClear:                   "Clear Filters",
	FiltersAdd:                     "Add Filters",
	FilterApply:                    "Apply",
	FilterByTemplate:               "Filter by {filter}",
	FiltersDateInTheLast:           "is in the last",
	FiltersDateEquals:              "is equal to",
	FiltersDateBetween:             "is between",
	FiltersDateIsAfter:             "is after",
	FiltersDateIsAfterOrOn:         "is on or after",
	FiltersDateIsBefore:            "is before",
	FiltersDateIsBeforeOrOn:        "is before or on",
	FiltersDateDays:                "days",
	FiltersDateMonths:              "months",
	FiltersDateAnd:                 "and",
	FiltersDateTo:                  "to",
	FiltersNumberEquals:            "is equal to",
	FiltersNumberBetween:           "between",
	FiltersNumberGreaterThan:       "is greater than",
	FiltersNumberLessThan:          "is less than",
	FiltersNumberAnd:               "and",
	FiltersStringEquals:            "is equal to",
	FiltersStringContains:          "contains",
	FiltersMultipleSelectIn:        "in",
	FiltersMultipleSelectNotIn:     "not in",
	PaginationRowsPerPage:          "Rows per page: ",
	ListingNoRecordToShow:          "No records to show",
	ListingSelectedCountNotice:     "{count} records are selected. ",
	ListingClearSelection:          "clear selection",
	BulkActionNoRecordsSelected:    "No records selected",
	BulkActionNoAvailableRecords:   "None of the selected records can be executed with this action.",
	BulkActionSelectedIdsProcessNoticeTemplate: "Partially selected records cannot be executed with this action: {ids}.",
	ConfirmDialogPromptText:                    "Are you sure?",
	Language:                                   "Language",
	Colon:                                      ":",
	NotFoundPageNotice:                         "Sorry, the requested page cannot be found. Please check the URL.",
	ButtonLabelActionsMenu:                     "Actions",
	Save:                                       "Save",
	AddRow:                                     "Add Row",
}

var Messages_zh_CN = &Messages{
	SuccessfullyUpdated:            "成功更新了",
	Search:                         "搜索",
	New:                            "新建",
	Update:                         "更新",
	Delete:                         "删除",
	Edit:                           "编辑",
	FormTitle:                      "表单",
	OK:                             "确定",
	Cancel:                         "取消",
	Clear:                          "清空",
	Create:                         "创建",
	DeleteConfirmationTextTemplate: "你确定你要删除这个对象吗，对象ID: {id}?",
	CreatingObjectTitleTemplate:    "新建{modelName}",
	EditingObjectTitleTemplate:     "编辑{modelName} {id}",
	ListingObjectTitleTemplate:     "{modelName}列表",
	DetailingObjectTitleTemplate:   "{modelName} {id}",
	FiltersClear:                   "清空筛选器",
	FiltersAdd:                     "添加筛选器",
	FilterApply:                    "应用",
	FilterByTemplate:               "按{filter}筛选",
	FiltersDateInTheLast:           "过去",
	FiltersDateEquals:              "等于",
	FiltersDateBetween:             "之间",
	FiltersDateIsAfter:             "之后",
	FiltersDateIsAfterOrOn:         "当天或之后",
	FiltersDateIsBefore:            "之前",
	FiltersDateIsBeforeOrOn:        "当天或之前",
	FiltersDateDays:                "天",
	FiltersDateMonths:              "月",
	FiltersDateAnd:                 "和",
	FiltersDateTo:                  "至",
	FiltersNumberEquals:            "等于",
	FiltersNumberBetween:           "之间",
	FiltersNumberGreaterThan:       "大于",
	FiltersNumberLessThan:          "小于",
	FiltersNumberAnd:               "和",
	FiltersStringEquals:            "等于",
	FiltersStringContains:          "包含",
	FiltersMultipleSelectIn:        "包含",
	FiltersMultipleSelectNotIn:     "不包含",
	PaginationRowsPerPage:          "每页: ",
	ListingNoRecordToShow:          "没有可显示的记录",
	ListingSelectedCountNotice:     "{count}条记录被选中。",
	ListingClearSelection:          "清除选择",
	BulkActionNoRecordsSelected:    "没有选中的记录",
	BulkActionNoAvailableRecords:   "所有选中的记录均无法执行这个操作。",
	BulkActionSelectedIdsProcessNoticeTemplate: "部分选中的记录无法被执行这个操作: {ids}。",
	ConfirmDialogPromptText:                    "你确定吗?",
	Language:                                   "语言",
	Colon:                                      "：",
	NotFoundPageNotice:                         "很抱歉，所请求的页面不存在，请检查URL。",
	ButtonLabelActionsMenu:                     "菜单",
	Save:                                       "保存",
	AddRow:                                     "新增",
}

var Messages_ja_JP = &Messages{
	SuccessfullyUpdated:            "更新に成功しました",
	Search:                         "検索",
	New:                            "新規",
	Update:                         "更新",
	Delete:                         "削除",
	Edit:                           "編集",
	FormTitle:                      "フォーム",
	OK:                             "OK",
	Cancel:                         "キャンセル",
	Create:                         "新規作成",
	DeleteConfirmationTextTemplate: "id: {id} のオブジェクトを削除してもよろしいですか？",
	CreatingObjectTitleTemplate:    "新規{modelName}作成",
	EditingObjectTitleTemplate:     "{modelName} {id}を編集する",
	ListingObjectTitleTemplate:     "{modelName}のリスト",
	DetailingObjectTitleTemplate:   "{modelName} {id}",
	FiltersClear:                   "フィルタのクリア",
	FiltersAdd:                     "フィルタの追加",
	FilterApply:                    "適用",
	FilterByTemplate:               "フィルタ",
	FiltersDateInTheLast:           "最後の",
	FiltersDateEquals:              "と同様",
	FiltersDateBetween:             "の間にある",
	FiltersDateIsAfter:             "の後である",
	FiltersDateIsAfterOrOn:         "の上または後である",
	FiltersDateIsBefore:            "は前",
	FiltersDateIsBeforeOrOn:        "が前か後か",
	FiltersDateDays:                "日",
	FiltersDateMonths:              "ヶ月",
	FiltersDateAnd:                 "と",
	FiltersDateTo:                  "に",
	FiltersNumberEquals:            "と同様",
	FiltersNumberBetween:           "の間",
	FiltersNumberGreaterThan:       "より大きい",
	FiltersNumberLessThan:          "より小さい",
	FiltersNumberAnd:               "と",
	FiltersStringEquals:            "と同様",
	FiltersStringContains:          "を含む",
	FiltersMultipleSelectIn:        "で",
	FiltersMultipleSelectNotIn:     "でない",
	PaginationRowsPerPage:          "ページあたりの行数 ",
	ListingNoRecordToShow:          "表示するレコードがない",
	ListingSelectedCountNotice:     "{count}レコードが選択されています。",
	ListingClearSelection:          "クリア選択",
	BulkActionNoRecordsSelected:    "レコードが選択されていません",
	BulkActionNoAvailableRecords:   "選択されたレコードのいずれも、このアクションでは実行できません。",
	BulkActionSelectedIdsProcessNoticeTemplate: "部分的に選択されたレコードは、このアクションでは実行できません：{ids}。",
	ConfirmDialogPromptText:                    "実行してもよろしいですか?",
	Language:                                   "言語",
	Colon:                                      ":",
	NotFoundPageNotice:                         "リクエストされたページが見つかりません。URLを確認してください。",
	ButtonLabelActionsMenu:                     "アクション",
	Save:                                       "保存",
	AddRow:                                     "行の追加",
}
