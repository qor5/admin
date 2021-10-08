package pagebuilder

type Page struct {
	ID    uint
	Title string
	Slug  string
}

func (*Page) TableName() string {
	return "page_builder_pages"
}

type Container struct {
	ID           uint
	PageID       uint
	Name         string
	ModelID      uint
	DisplayOrder float64
}

func (*Container) TableName() string {
	return "page_builder_containers"
}
