package publish

const (
	StatusDraft   = "draft"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Status struct {
	Status   string
	OnlineID string
}

func (status Status) GeStatus() string {
	return status.Status
}

func (status *Status) SetStatus(s string) {
	status.Status = s
}

func (status *Status) GetOnlineID() string {
	return status.OnlineID
}

func (status *Status) SetOnlineID(onlineID string) {
	status.OnlineID = onlineID
}
