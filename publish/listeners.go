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
	ModelLabel string
	Slug       string
	Status     *Status
	Version    *Version
	Schedule   *Schedule
}

func ToPayloadItem(obj any, label string) *PayloadItem {
	if lo.IsNil(obj) {
		return nil
	}
	return &PayloadItem{
		ModelLabel: label,
		Slug:       obj.(presets.SlugEncoder).PrimarySlug(),
		Status:     EmbedStatus(obj),
		Version:    EmbedVersion(obj),
		Schedule:   EmbedSchedule(obj),
	}
}

type PayloadItemUpdated struct {
	*PayloadItem
}

type PayloadItemDeleted struct {
	ModelLabel  string
	Slug        string
	NextVersion *PayloadItem
}

type PayloadVersionSelected struct {
	*PayloadItem
}

func NewListenerVersionSelected(mb *presets.ModelBuilder, ownerSlug string) h.HTMLComponent {
	return web.Listen(Notification(PayloadVersionSelected{}), fmt.Sprintf(`
	if (payload.ModelLabel != %q || payload.Slug === %q) {
		return
	}
	
	%s = payload.Slug
	
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
			web.Plaid().EventFunc(actions.DetailingDrawer).Query(presets.ParamID, web.Var("payload.Slug")).Go(),
		}, ";"),
		web.Plaid().PushState(true).URL(web.Var(fmt.Sprintf(`%q + '/' + payload.Slug`, mb.Info().ListingHref()))).Go(),
	))
}

func NewListenerItemDeleted(mb *presets.ModelBuilder, ownerSlug string) h.HTMLComponent {
	return web.Listen(Notification(PayloadItemDeleted{}), fmt.Sprintf(`
	if (payload.ModelLabel != %q || payload.Slug != %q) {
		return
	}
	
	if (!payload.NextVersion) {
		%s
	}
	`,
		mb.Info().Label(),
		ownerSlug,
		web.Plaid().PushState(true).URL(mb.Info().ListingHref()).Go(),
	))
}
