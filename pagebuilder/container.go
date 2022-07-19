package pagebuilder

import "gorm.io/gorm"

type Container struct {
	gorm.Model
	PageID       uint
	PageVersion  string
	Name         string
	ModelID      uint
	DisplayOrder float64
	Shared       bool
	DisplayName  string
}

func (*Container) TableName() string {
	return "page_builder_containers"
}

type DemoContainer struct {
	gorm.Model
	ModelName string
	ModelID   uint
}

func (*DemoContainer) TableName() string {
	return "page_builder_demo_containers"
}
