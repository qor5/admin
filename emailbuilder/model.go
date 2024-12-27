package emailbuilder

import "gorm.io/gorm"

type MailTemplate struct {
	gorm.Model
	Subject  string `json:"subject"`
	JSONBody string `json:"json_body"`
	HTMLBody string `json:"html_body"`
}
