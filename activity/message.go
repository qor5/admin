package activity

type Messages struct {
	ActivityAll     string
	ActivityView    string
	ActivityEdit    string
	ActivityCreate  string
	ActivityDelete  string
	Link            string
	Diffs           string
	FilterAction    string
	FilterCreatedAt string
	FilterCreator   string
	FilterModel     string
	DiffNew         string
	DiffDelete      string
	DiffChanges     string
	DiffField       string
	DiffOld         string
	DiffNow         string
	DiffValue       string
}

var Messages_en_US = &Messages{
	ActivityAll:     "All",
	ActivityView:    "View",
	ActivityEdit:    "Edit",
	ActivityCreate:  "Create",
	ActivityDelete:  "Delete",
	Link:            "Link",
	FilterAction:    "Action",
	FilterCreatedAt: "Create Time",
	FilterCreator:   "Creator",
	FilterModel:     "Model Name",
	DiffNew:         "New",
	DiffDelete:      "Delete",
	DiffChanges:     "Changes",
	DiffField:       "Filed",
	DiffOld:         "Old",
	DiffNow:         "Now",
	DiffValue:       "Value",
}

var Messages_zh_CN = &Messages{
	ActivityAll:     "全部",
	ActivityView:    "查看",
	ActivityEdit:    "编辑",
	ActivityCreate:  "创建",
	ActivityDelete:  "删除",
	Link:            "链接",
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
