package presets

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets/actions"
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

func ShowSnackbarScript(msg, color string) string {
	if msg == "" {
		return ""
	}
	if color == "" {
		color = ColorSuccess
	}
	return fmt.Sprintf(`vars.presetsMessage = { show: true, message: %q, color: %q}`, msg, color)
}

func ShowMessage(r *web.EventResponse, msg, color string) {
	script := ShowSnackbarScript(msg, color)
	if script == "" {
		return
	}
	web.AppendRunScripts(r, script)
}

func EditDeleteRowMenuItemFuncs(mi *ModelInfo, url string, editExtraParams url.Values) []vx.RowMenuItemFunc {
	return []vx.RowMenuItemFunc{
		editRowMenuItemFunc(mi, url, editExtraParams),
		deleteRowMenuItemFunc(mi, url, editExtraParams),
	}
}

func editRowMenuItemFunc(mi *ModelInfo, url string, editExtraParams url.Values) vx.RowMenuItemFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		msgr := mi.mb.mustGetMessages(ctx.R)
		if mi.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		onclick := web.Plaid().
			EventFunc(actions.Edit).
			Queries(editExtraParams).
			Query(ParamID, id).
			URL(url)
		if IsInDialog(ctx) {
			onclick.URL(mi.ListingHref()).Query(ParamOverlay, actions.Dialog)
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
		msgr := mi.mb.mustGetMessages(ctx.R)
		if mi.mb.Info().Verifier().Do(PermDelete).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		onclick := web.Plaid().
			EventFunc(actions.DeleteConfirmation).
			Queries(editExtraParams).
			Query(ParamID, id).
			URL(url)
		if IsInDialog(ctx) {
			onclick.URL(mi.ListingHref()).Query(ParamOverlay, actions.Dialog)
		}
		return VListItem(
			web.Slot(
				VIcon("mdi-delete"),
			).Name("prepend"),

			VListItemTitle(h.Text(msgr.Delete)),
		).Attr("@click", onclick.Go())
	}
}

func copyURLWithQueriesRemoved(u *url.URL, qs ...string) *url.URL {
	newU, _ := url.Parse(u.String())
	newQuery := newU.Query()
	for _, k := range qs {
		newQuery.Del(k)
	}
	newU.RawQuery = newQuery.Encode()
	return newU
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func UpdateToPortal(update *web.PortalUpdate) *web.PortalBuilder {
	return web.Portal().Name(update.Name).Children(
		update.Body,
	)
}

func toValidationErrors(err error) *web.ValidationErrors {
	if vErr, ok := err.(*web.ValidationErrors); ok {
		return vErr
	}
	vErr := &web.ValidationErrors{}
	vErr.GlobalError(err.Error())
	return vErr
}

func ObjectID(obj any) string {
	var id string
	if slugger, ok := obj.(SlugEncoder); ok {
		id = slugger.PrimarySlug()
	} else {
		v, err := reflectutils.Get(obj, "ID")
		if err == nil {
			if v == "" {
				return ""
			}
			if reflect.ValueOf(v).IsZero() {
				return ""
			}
			id = fmt.Sprint(v)
		}
	}
	return id
}

func MustObjectID(obj any) string {
	id := ObjectID(obj)
	if id == "" {
		panic("empty object id")
	}
	return id
}

func JsonCopy(dst, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return errors.Wrap(err, "JsonCopy marshal")
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return errors.Wrap(err, "JsonCopy unmarshal")
	}
	return nil
}

func MustJsonCopy(dst, src any) {
	if err := JsonCopy(dst, src); err != nil {
		panic(err)
	}
}
