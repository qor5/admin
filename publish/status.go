package publish

const (
	StatusDraft   = "draft"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Status struct {
	Status    string
	OnlineUrl string
}

func (status Status) GeStatus() string {
	return status.Status
}

func (status *Status) SetStatus(s string) {
	status.Status = s
}

func (status *Status) GetOnlineUrl() string {
	return status.OnlineUrl
}

func (status *Status) SetOnlineUrl(onlineUrl string) {
	status.OnlineUrl = onlineUrl
}
