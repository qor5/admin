package emailbuilder

import (
	"fmt"

	"gorm.io/gorm"
)

type MailTemplate struct {
	gorm.Model
	Subject  string `json:"subject"`
	JSONBody string `json:"json_body"`
	HTMLBody string `json:"html_body"`
}

func (m *MailTemplate) TableName() string {
	return "mail_templates"
}

func (m *MailTemplate) PrimarySlug() string {
	return fmt.Sprintf("%d", m.ID)
}

func (m *MailTemplate) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"id": slug,
	}
}
