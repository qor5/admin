package publish

// @snippet_begin(PublishStatus)
const (
	StatusDraft   = "draft"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Status struct {
	Status    string `gorm:"default:'draft'"`
	OnlineUrl string
}

// @snippet_end

func (status Status) GetStatus() string {
	return status.Status
}

func (status Status) GetOnlineUrl() string {
	return status.OnlineUrl
}

func (status *Status) SetStatus(s string) {
	status.Status = s
}

func (status *Status) SetOnlineUrl(onlineUrl string) {
	status.OnlineUrl = onlineUrl
}
