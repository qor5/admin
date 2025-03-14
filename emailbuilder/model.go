package emailbuilder

import (
	"fmt"

	"gorm.io/gorm"
)

type (
	EmailTemplate struct {
		gorm.Model
		Name string
		EmailDetail
	}
)

func (m *EmailTemplate) TableName() string {
	return "email_templates"
}

func (m *EmailTemplate) PrimarySlug() string {
	return fmt.Sprintf("%d", m.ID)
}

func (m *EmailTemplate) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"id": slug,
	}
}
