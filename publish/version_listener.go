package publish

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type PayloadModelsDeletedAddon struct {
	NextVersionSlug string `json:"next_version_slug"`
}

func NotifVersionSelected(mb *presets.ModelBuilder) string {
	return fmt.Sprintf("publish_NotifVersionSelected_%T", mb.NewModel())
}

type PayloadVersionSelected struct {
	Slug string `json:"slug"`
}

func NewListenerVersionSelected(mb *presets.ModelBuilder, slug string) h.HTMLComponent {
	event := actions.Edit
	if mb.HasDetailing() {
		event = actions.DetailingDrawer
	}
	drawerToSlug := web.Plaid().URL(mb.Info().ListingHref()).EventFunc(event).Query(presets.ParamID, web.Var("payload.slug")).Go()
	return web.Listen(NotifVersionSelected(mb), fmt.Sprintf(`
		if (payload.slug === %q) {
			return
		}
		if (vars.presetsRightDrawer) {
			%s
			return
		}
		%s
	`,
		slug,
		strings.Join([]string{
			presets.CloseRightDrawerVarScript,
			drawerToSlug,
		}, ";"),
		web.Plaid().PushState(true).URL(web.Var(fmt.Sprintf(`%q + "/" + payload.slug`, mb.Info().ListingHref()))).Go(),
	))
}

func NewListenerModelsDeleted(mb *presets.ModelBuilder, slug string) h.HTMLComponent {
	return web.Listen(mb.NotifModelsDeleted(), fmt.Sprintf(`(payload, addon) => { 
		if (!payload.ids.includes(%q)) {
			return
		}		
		if (addon.next_version_slug) {
			%s
		} else {
			%s
		}
	}`,
		slug,
		web.Emit(
			NotifVersionSelected(mb),
			web.Var(fmt.Sprintf(`{ ...%s, slug: addon.next_version_slug }`, h.JSONString(PayloadVersionSelected{}))),
		),
		web.Plaid().PushState(true).URL(mb.Info().ListingHref()).Go(), // directly to resource list , not version list
	))
}
