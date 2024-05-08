package microsite

type Messages struct {
	CurrentPackage string
}

var Messages_en_US = &Messages{
	CurrentPackage: "Current Package",
}

var Messages_zh_CN = &Messages{
	CurrentPackage: "当前压缩包",
}
