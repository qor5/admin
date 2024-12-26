package emailbuilder

import "gorm.io/gorm"

type MailTemplate struct {
	gorm.Model
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
}
