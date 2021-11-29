package activity

type Messages struct {
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
	ModelLink      string
	ModelDiffs     string

	FilterAction    string
	FilterCreatedAt string
	FilterCreator   string
	FilterModel     string

	DiffNew     string
	DiffDelete  string
	DiffChanges string
	DiffField   string
	DiffOld     string
	DiffNow     string
	DiffValue   string
}

var Messages_en_US = &Messages{
	ActionAll:    "All",
	ActionView:   "View",
	ActionEdit:   "Edit",
	ActionCreate: "Create",
	ActionDelete: "Delete",

	ModelUserID:    "User ID",
	ModelCreatedAt: "Created At",
	ModelAction:    "Action",
	ModelCreator:   "Creator",
	ModelKeys:      "Keys",
	ModelName:      "Model",
	ModelLink:      "Link",
	ModelDiffs:     "Diffs",

	FilterAction:    "Action",
	FilterCreatedAt: "Create Time",
	FilterCreator:   "Creator",
	FilterModel:     "Model Name",

	DiffNew:     "New",
	DiffDelete:  "Delete",
	DiffChanges: "Changes",
	DiffField:   "Filed",
	DiffOld:     "Old",
	DiffNow:     "Now",
	DiffValue:   "Value",
}

var Messages_zh_CN = &Messages{
	ActionAll:    "全部",
	ActionView:   "查看",
	ActionEdit:   "编辑",
	ActionCreate: "创建",
	ActionDelete: "删除",

	ModelUserID:    "用户ID",
	ModelCreatedAt: "创建时间",
	ModelAction:    "操作",
	ModelCreator:   "创建者",
	ModelKeys:      "对象的主键值",
	ModelName:      "对象",
	ModelLink:      "链接",
	ModelDiffs:     "差异",

	FilterAction:    "操作类型",
	FilterCreatedAt: "操作时间",
	FilterCreator:   "操作人",
	FilterModel:     "操作对象",

	DiffNew:     "新加",
	DiffDelete:  "删除",
	DiffChanges: "修改",
	DiffField:   "字段",
	DiffOld:     "之前的值",
	DiffNow:     "当前的值",
	DiffValue:   "值",
}
