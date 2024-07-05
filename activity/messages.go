package activity

type Messages struct {
	Activities   string
	ActionAll    string
	ActionView   string
	ActionEdit   string
	ActionCreate string
	ActionDelete string

	ModelUserID    string
	ModelCreatedAt string
	ModelAction    string
	ModelCreator   string
	ModelKeys      string
	ModelName      string
	ModelLabel     string
	ModelLink      string
	ModelDiffs     string

	FilterAction    string
	FilterCreatedAt string
	FilterCreator   string
	FilterModel     string

	DiffDetail  string
	DiffNew     string
	DiffDelete  string
	DiffChanges string
	DiffField   string
	DiffOld     string
	DiffNow     string
	DiffValue   string

	SuccessfullyCreated string
	Item                string
	Notes               string
	NewNote             string
}

var Messages_en_US = &Messages{
	Activities:   "Activities",
	ActionAll:    "All",
	ActionView:   "View",
	ActionEdit:   "Edit",
	ActionCreate: "Create",
	ActionDelete: "Delete",

	ModelUserID:    "Creator ID",
	ModelCreatedAt: "Date Time",
	ModelAction:    "Action",
	ModelCreator:   "Creator",
	ModelKeys:      "Keys",
	ModelName:      "Table Name",
	ModelLabel:     "Menu Name",
	ModelLink:      "Link",
	ModelDiffs:     "Diffs",

	FilterAction:    "Action",
	FilterCreatedAt: "Create Time",
	FilterCreator:   "Creator",
	FilterModel:     "Model Name",

	DiffDetail:  "Detail",
	DiffNew:     "New",
	DiffDelete:  "Delete",
	DiffChanges: "Changes",
	DiffField:   "Filed",
	DiffOld:     "Old",
	DiffNow:     "Now",
	DiffValue:   "Value",

	SuccessfullyCreated: "Successfully Created",
	Item:                "Item",
	Notes:               "Notes",
	NewNote:             "New Note",
}

var Messages_zh_CN = &Messages{
	Activities:   "活动",
	ActionAll:    "全部",
	ActionView:   "查看",
	ActionEdit:   "编辑",
	ActionCreate: "创建",
	ActionDelete: "删除",

	ModelUserID:    "操作者ID",
	ModelCreatedAt: "日期时间",
	ModelAction:    "操作",
	ModelCreator:   "操作者",
	ModelKeys:      "表的主键值",
	ModelName:      "表名",
	ModelLabel:     "菜单名",
	ModelLink:      "链接",
	ModelDiffs:     "差异",

	FilterAction:    "操作类型",
	FilterCreatedAt: "操作时间",
	FilterCreator:   "操作人",
	FilterModel:     "操作对象",
	DiffDetail:      "详情",
	DiffNew:         "新加",
	DiffDelete:      "删除",
	DiffChanges:     "修改",
	DiffField:       "字段",
	DiffOld:         "之前的值",
	DiffNow:         "当前的值",
	DiffValue:       "值",

	SuccessfullyCreated: "成功创建",
	Item:                "记录",
	Notes:               "备注",
	NewNote:             "新建备注",
}

var Messages_ja_JP = &Messages{
	SuccessfullyCreated: "作成に成功しました",
	Item:                "アイテム",
	Notes:               "ノート",
	NewNote:             "新規ノート",
}
