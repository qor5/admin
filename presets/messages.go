package presets

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Messages struct {
	DialogTitleDefault                         string
	SuccessfullyUpdated                        string
	SuccessfullyCreated                        string
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
	SelectedTemplate                           func(v any) string
	DeleteConfirmationText                     string
	DeleteObjectsConfirmationText              func(v int) string
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
	FiltersDateStartAt                         string
	FiltersDateEndAt                           string
	FiltersDateTo                              string
	FiltersDateClear                           string
	FiltersDateOK                              string
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
	ConfirmDialogTitleText                     string
	ConfirmDialogPromptText                    string
	Language                                   string
	Colon                                      string
	NotFoundPageNotice                         string
	ButtonLabelActionsMenu                     string
	Save                                       string
	AddRow                                     string
	AddCard                                    string
	AddButton                                  string
	CheckboxTrueLabel                          string
	CheckboxFalseLabel                         string

	HumanizeTimeAgo       string
	HumanizeTimeFromNow   string
	HumanizeTimeNow       string
	HumanizeTime1Second   string
	HumanizeTimeSeconds   string
	HumanizeTime1Minute   string
	HumanizeTimeMinutes   string
	HumanizeTime1Hour     string
	HumanizeTimeHours     string
	HumanizeTime1Day      string
	HumanizeTimeDays      string
	HumanizeTime1Week     string
	HumanizeTimeWeeks     string
	HumanizeTime1Month    string
	HumanizeTimeMonths    string
	HumanizeTime1Year     string
	HumanizeTime2Years    string
	HumanizeTimeYears     string
	HumanizeTimeLongWhile string

	LeaveBeforeUnsubmit string

	RecordNotFound string
}

func (msgr *Messages) CreatingObjectTitle(modelName string) string {
	return strings.NewReplacer("{modelName}", modelName).
		Replace(msgr.CreatingObjectTitleTemplate)
}

func (msgr *Messages) EditingObjectTitle(label, name string) string {
	return strings.NewReplacer("{id}", name, "{modelName}", label).
		Replace(msgr.EditingObjectTitleTemplate)
}

func (msgr *Messages) ListingObjectTitle(label string) string {
	return strings.NewReplacer("{modelName}", label).
		Replace(msgr.ListingObjectTitleTemplate)
}

func (msgr *Messages) DetailingObjectTitle(label, name string) string {
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

func (msgr *Messages) HumanizeTime(then time.Time) string {
	return humanize.CustomRelTime(then, time.Now(),
		msgr.HumanizeTimeAgo, msgr.HumanizeTimeFromNow,
		[]humanize.RelTimeMagnitude{
			{D: time.Second, Format: msgr.HumanizeTimeNow, DivBy: time.Second},
			{D: 2 * time.Second, Format: msgr.HumanizeTime1Second, DivBy: 1},
			{D: time.Minute, Format: msgr.HumanizeTimeSeconds, DivBy: time.Second},
			{D: 2 * time.Minute, Format: msgr.HumanizeTime1Minute, DivBy: 1},
			{D: time.Hour, Format: msgr.HumanizeTimeMinutes, DivBy: time.Minute},
			{D: 2 * time.Hour, Format: msgr.HumanizeTime1Hour, DivBy: 1},
			{D: humanize.Day, Format: msgr.HumanizeTimeHours, DivBy: time.Hour},
			{D: 2 * humanize.Day, Format: msgr.HumanizeTime1Day, DivBy: 1},
			{D: humanize.Week, Format: msgr.HumanizeTimeDays, DivBy: humanize.Day},
			{D: 2 * humanize.Week, Format: msgr.HumanizeTime1Week, DivBy: 1},
			{D: humanize.Month, Format: msgr.HumanizeTimeWeeks, DivBy: humanize.Week},
			{D: 2 * humanize.Month, Format: msgr.HumanizeTime1Month, DivBy: 1},
			{D: humanize.Year, Format: msgr.HumanizeTimeMonths, DivBy: humanize.Month},
			{D: 18 * humanize.Month, Format: msgr.HumanizeTime1Year, DivBy: 1},
			{D: 2 * humanize.Year, Format: msgr.HumanizeTime2Years, DivBy: 1},
			{D: humanize.LongTime, Format: msgr.HumanizeTimeYears, DivBy: humanize.Year},
			{D: math.MaxInt64, Format: msgr.HumanizeTimeLongWhile, DivBy: 1},
		})
}

var Messages_en_US = &Messages{
	DialogTitleDefault:  "Confirm",
	SuccessfullyUpdated: "Successfully Updated",
	SuccessfullyCreated: "Successfully Created",
	Search:              "Search",
	New:                 "New",
	Update:              "Update",
	Delete:              "Delete",
	Edit:                "Edit",
	FormTitle:           "Form",
	OK:                  "OK",
	Cancel:              "Cancel",
	Clear:               "Clear",
	Create:              "Create",
	SelectedTemplate: func(v any) string {
		return fmt.Sprintf("%v Selected", v)
	},
	DeleteConfirmationText: "Are you sure you want to delete this object?",
	DeleteObjectsConfirmationText: func(v int) string {
		return fmt.Sprintf(`Are you sure you want to delete %v objects`, v)
	},
	CreatingObjectTitleTemplate:                "New {modelName}",
	EditingObjectTitleTemplate:                 "Editing {modelName} {id}",
	ListingObjectTitleTemplate:                 "{modelName}",
	DetailingObjectTitleTemplate:               "{modelName} {id}",
	FiltersClear:                               "Reset",
	FiltersAdd:                                 "Add Filters",
	FilterApply:                                "Apply",
	FilterByTemplate:                           "Filter by {filter}",
	FiltersDateInTheLast:                       "is in the last",
	FiltersDateEquals:                          "is equal to",
	FiltersDateBetween:                         "is between",
	FiltersDateIsAfter:                         "is after",
	FiltersDateIsAfterOrOn:                     "is on or after",
	FiltersDateIsBefore:                        "is before",
	FiltersDateIsBeforeOrOn:                    "is before or on",
	FiltersDateDays:                            "days",
	FiltersDateMonths:                          "months",
	FiltersDateAnd:                             "and",
	FiltersDateStartAt:                         "Start at",
	FiltersDateEndAt:                           "End at",
	FiltersDateTo:                              "to",
	FiltersDateClear:                           "Clear",
	FiltersDateOK:                              "OK",
	FiltersNumberEquals:                        "is equal to",
	FiltersNumberBetween:                       "between",
	FiltersNumberGreaterThan:                   "is greater than",
	FiltersNumberLessThan:                      "is less than",
	FiltersNumberAnd:                           "and",
	FiltersStringEquals:                        "is equal to",
	FiltersStringContains:                      "contains",
	FiltersMultipleSelectIn:                    "in",
	FiltersMultipleSelectNotIn:                 "not in",
	PaginationRowsPerPage:                      "Rows per page: ",
	ListingNoRecordToShow:                      "No records to show",
	ListingSelectedCountNotice:                 "{count} records are selected. ",
	ListingClearSelection:                      "clear selection",
	BulkActionNoRecordsSelected:                "No records selected",
	BulkActionNoAvailableRecords:               "None of the selected records can be executed with this action.",
	BulkActionSelectedIdsProcessNoticeTemplate: "Partially selected records cannot be executed with this action: {ids}.",
	ConfirmDialogTitleText:                     "Confirm",
	ConfirmDialogPromptText:                    "Are you sure?",
	Language:                                   "Language",
	Colon:                                      ":",
	NotFoundPageNotice:                         "Sorry, the requested page cannot be found. Please check the URL.",
	ButtonLabelActionsMenu:                     "Actions",
	Save:                                       "Save",
	AddRow:                                     "Add Item",
	AddCard:                                    "Add Card",
	AddButton:                                  "Add Button",
	CheckboxTrueLabel:                          "YES",
	CheckboxFalseLabel:                         "NO",

	HumanizeTimeAgo:       "ago",
	HumanizeTimeFromNow:   "from now",
	HumanizeTimeNow:       "now",
	HumanizeTime1Second:   "1 second %s",
	HumanizeTimeSeconds:   "%d seconds %s",
	HumanizeTime1Minute:   "1 minute %s",
	HumanizeTimeMinutes:   "%d minutes %s",
	HumanizeTime1Hour:     "1 hour %s",
	HumanizeTimeHours:     "%d hours %s",
	HumanizeTime1Day:      "1 day %s",
	HumanizeTimeDays:      "%d days %s",
	HumanizeTime1Week:     "1 week %s",
	HumanizeTimeWeeks:     "%d weeks %s",
	HumanizeTime1Month:    "1 month %s",
	HumanizeTimeMonths:    "%d months %s",
	HumanizeTime1Year:     "1 year %s",
	HumanizeTime2Years:    "2 years %s",
	HumanizeTimeYears:     "%d years %s",
	HumanizeTimeLongWhile: "a long while %s",

	LeaveBeforeUnsubmit: "If you leave before submitting the form, you will lose all the unsaved input.",

	RecordNotFound: "record not found",
}

var Messages_zh_CN = &Messages{
	DialogTitleDefault:  "确认",
	SuccessfullyUpdated: "成功更新了",
	SuccessfullyCreated: "成功创建了",
	Search:              "搜索",
	New:                 "新建",
	Update:              "更新",
	Delete:              "删除",
	Edit:                "编辑",
	FormTitle:           "表单",
	OK:                  "确定",
	Cancel:              "取消",
	Clear:               "清空",
	Create:              "创建",
	SelectedTemplate: func(v any) string {
		return fmt.Sprintf("%v 已选择", v)
	},
	DeleteConfirmationText: "你确定你要删除这个对象吗？",
	DeleteObjectsConfirmationText: func(v int) string {
		return fmt.Sprintf(`是否删除 %v 个条目`, v)
	},
	CreatingObjectTitleTemplate:                "新建{modelName}",
	EditingObjectTitleTemplate:                 "编辑{modelName} {id}",
	ListingObjectTitleTemplate:                 "{modelName}",
	DetailingObjectTitleTemplate:               "{modelName} {id}",
	FiltersClear:                               "清空筛选器",
	FiltersAdd:                                 "添加筛选器",
	FilterApply:                                "应用",
	FilterByTemplate:                           "按{filter}筛选",
	FiltersDateInTheLast:                       "过去",
	FiltersDateEquals:                          "等于",
	FiltersDateBetween:                         "之间",
	FiltersDateIsAfter:                         "之后",
	FiltersDateIsAfterOrOn:                     "当天或之后",
	FiltersDateIsBefore:                        "之前",
	FiltersDateIsBeforeOrOn:                    "当天或之前",
	FiltersDateDays:                            "天",
	FiltersDateMonths:                          "月",
	FiltersDateAnd:                             "和",
	FiltersDateStartAt:                         "开始于",
	FiltersDateEndAt:                           "结束于",
	FiltersDateTo:                              "至",
	FiltersDateClear:                           "清空",
	FiltersDateOK:                              "确定",
	FiltersNumberEquals:                        "等于",
	FiltersNumberBetween:                       "之间",
	FiltersNumberGreaterThan:                   "大于",
	FiltersNumberLessThan:                      "小于",
	FiltersNumberAnd:                           "和",
	FiltersStringEquals:                        "等于",
	FiltersStringContains:                      "包含",
	FiltersMultipleSelectIn:                    "包含",
	FiltersMultipleSelectNotIn:                 "不包含",
	PaginationRowsPerPage:                      "每页: ",
	ListingNoRecordToShow:                      "没有可显示的记录",
	ListingSelectedCountNotice:                 "{count}条记录被选中。",
	ListingClearSelection:                      "清除选择",
	BulkActionNoRecordsSelected:                "没有选中的记录",
	BulkActionNoAvailableRecords:               "所有选中的记录均无法执行这个操作。",
	BulkActionSelectedIdsProcessNoticeTemplate: "部分选中的记录无法被执行这个操作: {ids}。",
	ConfirmDialogTitleText:                     "确认",
	ConfirmDialogPromptText:                    "你确定吗?",
	Language:                                   "语言",
	Colon:                                      "：",
	NotFoundPageNotice:                         "很抱歉，所请求的页面不存在，请检查URL。",
	ButtonLabelActionsMenu:                     "菜单",
	Save:                                       "保存",
	AddRow:                                     "新增",
	AddCard:                                    "新增卡片",
	AddButton:                                  "新增按钮",
	CheckboxTrueLabel:                          "是",
	CheckboxFalseLabel:                         "否",

	HumanizeTimeAgo:       "前",
	HumanizeTimeFromNow:   "后",
	HumanizeTimeNow:       "刚刚",
	HumanizeTime1Second:   "1 秒%s",
	HumanizeTimeSeconds:   "%d 秒%s",
	HumanizeTime1Minute:   "1 分钟%s",
	HumanizeTimeMinutes:   "%d 分钟%s",
	HumanizeTime1Hour:     "1 小时%s",
	HumanizeTimeHours:     "%d 小时%s",
	HumanizeTime1Day:      "1 天%s",
	HumanizeTimeDays:      "%d 天%s",
	HumanizeTime1Week:     "1 星期%s",
	HumanizeTimeWeeks:     "%d 星期%s",
	HumanizeTime1Month:    "1 个月%s",
	HumanizeTimeMonths:    "%d 个月%s",
	HumanizeTime1Year:     "1 年%s",
	HumanizeTime2Years:    "2 年%s",
	HumanizeTimeYears:     "%d 年%s",
	HumanizeTimeLongWhile: "很久之%s",

	LeaveBeforeUnsubmit: "如果您在提交表单之前离开，您将丢失所有未保存的输入。",

	RecordNotFound: "记录未找到",
}

var Messages_ja_JP = &Messages{
	DialogTitleDefault:  "確認",
	SuccessfullyUpdated: "更新に成功しました",
	SuccessfullyCreated: "作成に成功しました",
	Search:              "検索",
	New:                 "作成する",
	Update:              "更新",
	Delete:              "削除",
	Edit:                "編集",
	FormTitle:           "フォーム",
	OK:                  "OK",
	Cancel:              "キャンセル",
	Clear:               "消去する",
	Create:              "作成する",
	SelectedTemplate: func(v any) string {
		return fmt.Sprintf("%v 選択済み", v)
	},
	DeleteConfirmationText: "削除してもよろしいですか？",
	DeleteObjectsConfirmationText: func(v int) string {
		return fmt.Sprintf(`Are you sure you want to delete %v objects`, v)
	},
	CreatingObjectTitleTemplate:                "新規{modelName}作成",
	EditingObjectTitleTemplate:                 "{modelName} {id}を編集する",
	ListingObjectTitleTemplate:                 "{modelName}",
	DetailingObjectTitleTemplate:               "{modelName} {id}",
	FiltersClear:                               "フィルタのクリア",
	FiltersAdd:                                 "フィルタの追加",
	FilterApply:                                "適用",
	FilterByTemplate:                           "フィルタ",
	FiltersDateInTheLast:                       "最後の",
	FiltersDateEquals:                          "と同様",
	FiltersDateBetween:                         "の間にある",
	FiltersDateIsAfter:                         "の後である",
	FiltersDateIsAfterOrOn:                     "の上または後である",
	FiltersDateIsBefore:                        "は前",
	FiltersDateIsBeforeOrOn:                    "が前か後か",
	FiltersDateDays:                            "日",
	FiltersDateMonths:                          "ヶ月",
	FiltersDateAnd:                             "と",
	FiltersDateStartAt:                         "開始",
	FiltersDateEndAt:                           "終了",
	FiltersDateTo:                              "に",
	FiltersDateClear:                           "消去する",
	FiltersDateOK:                              "OK",
	FiltersNumberEquals:                        "と同様",
	FiltersNumberBetween:                       "の間",
	FiltersNumberGreaterThan:                   "より大きい",
	FiltersNumberLessThan:                      "より小さい",
	FiltersNumberAnd:                           "と",
	FiltersStringEquals:                        "と同様",
	FiltersStringContains:                      "を含む",
	FiltersMultipleSelectIn:                    "で",
	FiltersMultipleSelectNotIn:                 "でない",
	PaginationRowsPerPage:                      "ページあたりの行数 ",
	ListingNoRecordToShow:                      "表示できるレコードがありません",
	ListingSelectedCountNotice:                 "{count}レコードが選択されています。",
	ListingClearSelection:                      "クリア選択",
	BulkActionNoRecordsSelected:                "レコードが選択されていません",
	BulkActionNoAvailableRecords:               "選択されたレコードのいずれも、このアクションでは実行できません。",
	BulkActionSelectedIdsProcessNoticeTemplate: "部分的に選択されたレコードは、このアクションでは実行できません：{ids}。",
	ConfirmDialogTitleText:                     "確認",
	ConfirmDialogPromptText:                    "実行してもよろしいですか?",
	Language:                                   "言語",
	Colon:                                      ":",
	NotFoundPageNotice:                         "リクエストされたページが見つかりません。URLを確認してください。",
	ButtonLabelActionsMenu:                     "アクション",
	Save:                                       "保存する",
	AddRow:                                     "アイテムを追加",
	AddCard:                                    "カードを追加",
	AddButton:                                  "ボタンを追加",
	CheckboxTrueLabel:                          "選択済み",
	CheckboxFalseLabel:                         "未選択",

	HumanizeTimeAgo:       "前",
	HumanizeTimeFromNow:   "今後",
	HumanizeTimeNow:       "現在",
	HumanizeTime1Second:   "1 秒%s",
	HumanizeTimeSeconds:   "%d 秒%s",
	HumanizeTime1Minute:   "1 分%s",
	HumanizeTimeMinutes:   "%d 分%s",
	HumanizeTime1Hour:     "1 時間%s",
	HumanizeTimeHours:     "%d 時間%s",
	HumanizeTime1Day:      "1 日%s",
	HumanizeTimeDays:      "%d 日%s",
	HumanizeTime1Week:     "1 週間%s",
	HumanizeTimeWeeks:     "%d 週間%s",
	HumanizeTime1Month:    "1 ヶ月%s",
	HumanizeTimeMonths:    "%d ヶ月%s",
	HumanizeTime1Year:     "1 年間%s",
	HumanizeTime2Years:    "2 年間%s",
	HumanizeTimeYears:     "%d 年間%s",
	HumanizeTimeLongWhile: "長い間%s",

	LeaveBeforeUnsubmit: "フォームを送信する前に離れると、すべての未保存の入力が失われます。",

	RecordNotFound: "レコードが見つかりません",
}
