package utils

type Messages struct {
	OK     string
	Cancel string
}

var Messages_en_US = &Messages{
	OK:     "OK",
	Cancel: "Cancel",
}

var Messages_zh_CN = &Messages{
	OK:     "确定",
	Cancel: "取消",
}

var Messages_ja_JP = &Messages{
	OK:     "OK",
	Cancel: "キャンセル",
}
