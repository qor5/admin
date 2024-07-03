package publish

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
)

func Notification(payload any) string {
	return reflect.TypeOf(payload).String()
}

func Notify(payload any) string {
	notif := Notification(payload)
	return web.Emit(notif, payload)
}

type PayloadItem struct {
	Model    string    `json:"model"`
	Slug     string    `json:"slug"`
	Status   *Status   `json:"status"`
	Version  *Version  `json:"version"`
	Schedule *Schedule `json:"schedule"`
}

func ToPayloadItem(obj any, label string) *PayloadItem {
	if lo.IsNil(obj) {
		return nil
	}
	return &PayloadItem{
		Model:    label,
		Slug:     obj.(presets.SlugEncoder).PrimarySlug(),
		Status:   EmbedStatus(obj),
		Version:  EmbedVersion(obj),
		Schedule: EmbedSchedule(obj),
	}
}

type PayloadItemDeleted struct {
	Model       string       `json:"model"`
	Slug        string       `json:"slug"`
	NextVersion *PayloadItem `json:"next_version"`
}

type PayloadVersionSelected struct {
	Model string `json:"model"`
	Slug  string `json:"slug"`
}

func NewListenerVersionSelected(mb *presets.ModelBuilder, ownerSlug string) h.HTMLComponent {
	return web.Listen(Notification(PayloadVersionSelected{}), fmt.Sprintf(`
	if (payload.model != %q || payload.slug === %q) {
		return
	}
	
	%s = payload.slug
	
	if (vars.presetsRightDrawer) {
		%s
		return
	}
	%s
	`,
		mb.Info().Label(),
		ownerSlug,
		VarCurrentDisplaySlug,
		strings.Join([]string{
			presets.CloseRightDrawerVarScript,
			web.Plaid().EventFunc(actions.DetailingDrawer).Query(presets.ParamID, web.Var("payload.slug")).Go(),
		}, ";"),
		web.Plaid().PushState(true).URL(web.Var(fmt.Sprintf(`%q + '/' + payload.slug`, mb.Info().ListingHref()))).Go(),
	))
}

func NewListenerItemDeleted(mb *presets.ModelBuilder, ownerSlug string) h.HTMLComponent {
	return web.Listen(Notification(PayloadItemDeleted{}), fmt.Sprintf(`
	if (payload.model != %q || payload.slug != %q) {
		return
	}
	
	if (!payload.next_version) {
		%s
	}
	`,
		mb.Info().Label(),
		ownerSlug,
		web.Plaid().PushState(true).URL(mb.Info().ListingHref()).Go(),
	))
}
