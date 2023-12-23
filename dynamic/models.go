package dynamic

import "encoding/json"

type TableSchema struct {
	Name   string          `gorm:"primaryKey"`
	Schema json.RawMessage `gorm:"type:jsonb"`
}

type Struct struct {
	Fields []Field
}

type Field struct {
	Name string
	Type interface{}
	Tag  string
}

func (ts TableSchema) TableName() string {
	return "dynamic_table_schemas"
}

type TableData struct {
	ID         uint            `gorm:"primaryKey"`
	SchemaName string          `gorm:"index"`
	Value      json.RawMessage `gorm:"type:jsonb"`
}

func (td TableData) TableName() string {
	return "dynamic_table_data"
}
