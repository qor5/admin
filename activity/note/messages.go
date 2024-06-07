package note

type Messages struct {
	SuccessfullyCreated string
	Item                string
	Notes               string
	NewNote             string
}

var Messages_en_US = &Messages{
	SuccessfullyCreated: "Successfully Created",
	Item:                "Item",
	Notes:               "Notes",
	NewNote:             "New Note",
}

var Messages_zh_CN = &Messages{
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
