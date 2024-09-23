package utils

type Messages struct {
	OK                string
	Cancel            string
	ModalTitleConfirm string
}

var Messages_en_US = &Messages{
	OK:                "OK",
	Cancel:            "Cancel",
	ModalTitleConfirm: "Confirm",
}

var Messages_zh_CN = &Messages{
	OK:                "确定",
	Cancel:            "取消",
	ModalTitleConfirm: "确认",
}

var Messages_ja_JP = &Messages{
	OK:                "OK",
	Cancel:            "キャンセル",
	ModalTitleConfirm: "確認",
}
