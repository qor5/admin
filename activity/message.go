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
}

var Messages_en_US = &Messages{
	ActivityAll:     "All",
	ActivityView:    "View",
	ActivityEdit:    "Edit",
	ActivityCreate:  "Create",
	ActivityDelete:  "Delete",
	Link:            "Link",
	Diffs:           "Diffs",
	FilterAction:    "Action",
	FilterCreatedAt: "Create Time",
	FilterCreator:   "Creator",
	FilterModel:     "Model Name",
}

var Messages_zh_CN = &Messages{
	ActivityAll:     "全部",
	ActivityView:    "查看",
	ActivityEdit:    "编辑",
	ActivityCreate:  "创建",
	ActivityDelete:  "删除",
	Link:            "链接",
	Diffs:           "改动",
	FilterAction:    "操作类型",
	FilterCreatedAt: "操作时间",
	FilterCreator:   "操作人",
	FilterModel:     "操作对象",
}
