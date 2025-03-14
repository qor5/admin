package emailbuilder

type (
	EmailDetailInterface interface {
		EmbedEmailDetail() *EmailDetail
	}
	EmailDetail struct {
		Name     string
		Subject  string `json:"subject"`
		JSONBody string `json:"json_body"`
		HTMLBody string `json:"html_body"`
	}
)

func (ed *EmailDetail) EmbedEmailDetail() *EmailDetail {
	return ed
}
