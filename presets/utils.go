package presets

import (
	"fmt"
	"net/url"
	"time"

	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

func RecoverPrimaryColumnValuesBySlug(dec SlugDecoder, slug string) (r map[string]string, err error) {
	defer func() {
		if e := recover(); e != nil {
			r = nil
			err = fmt.Errorf("wrong slug: %v", slug)
		}
	}()
	r = dec.PrimaryColumnValuesBySlug(slug)
	return r, nil
}

func ShowMessage(r *web.EventResponse, msg string, color string) {
	if r == nil || msg == "" {
		return
	}

	if color == "" {
		color = "success"
	}

	msgJSON := h.JSONString(msg)
	colorJSON := h.JSONString(color)

	web.AppendRunScripts(r, fmt.Sprintf(
		`vars.presetsMessage = { show: true, message: %s, color: %s}`,
		msgJSON, colorJSON))
}

func EditDeleteRowMenuItemFuncs(mi *ModelInfo, url string, editExtraParams url.Values) []vx.RowMenuItemFunc {
	return []vx.RowMenuItemFunc{
		editRowMenuItemFunc(mi, url, editExtraParams),
		deleteRowMenuItemFunc(mi, url, editExtraParams),
	}
}

func editRowMenuItemFunc(mi *ModelInfo, url string, editExtraParams url.Values) vx.RowMenuItemFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		msgr := MustGetMessages(ctx.R)
		if mi.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		onclick := web.Plaid().
			EventFunc(actions.Edit).
			Queries(editExtraParams).
			Query(ParamID, id).
			URL(url)
		if IsInDialog(ctx) {
			onclick.URL(ctx.R.RequestURI).
				Query(ParamOverlay, actions.Dialog).
				Query(ParamInDialog, true).
				Query(ParamListingQueries, ctx.Queries().Encode())
		}
		return VListItem(
			web.Slot(
				VIcon("mdi-pencil"),
			).Name("prepend"),

			VListItemTitle(h.Text(msgr.Edit)),
		).Attr("@click", onclick.Go())
	}
}

func deleteRowMenuItemFunc(mi *ModelInfo, url string, editExtraParams url.Values) vx.RowMenuItemFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		msgr := MustGetMessages(ctx.R)
		if mi.mb.Info().Verifier().Do(PermDelete).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		onclick := web.Plaid().
			EventFunc(actions.DeleteConfirmation).
			Queries(editExtraParams).
			Query(ParamID, id).
			URL(url)
		if IsInDialog(ctx) {
			onclick.URL(ctx.R.RequestURI).
				Query(ParamOverlay, actions.Dialog).
				Query(ParamInDialog, true).
				Query(ParamListingQueries, ctx.Queries().Encode())
		}
		return VListItem(
			web.Slot(
				VIcon("mdi-delete"),
			).Name("prepend"),

			VListItemTitle(h.Text(msgr.Delete)),
		).Attr("@click", onclick.Go())
	}
}

func isInDialogFromQuery(ctx *web.EventContext) bool {
	return ctx.R.URL.Query().Get(ParamInDialog) == "true"
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
