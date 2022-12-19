package presets

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/qor5/admin/presets/actions"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
)

func ShowMessage(r *web.EventResponse, msg string, color string) {
	if msg == "" {
		return
	}

	if color == "" {
		color = "success"
	}

	web.AppendVarsScripts(r, fmt.Sprintf(
		`vars.presetsMessage = { show: true, message: %s, color: %s}`,
		h.JSONString(msg), h.JSONString(color)))
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
		if IsInDialog(ctx.R.Context()) {
			onclick.URL(ctx.R.RequestURI).
				Query(ParamOverlay, actions.Dialog).
				Query(ParamInDialog, true).
				Query(ParamListingQueries, ctx.Queries().Encode())
		}
		return VListItem(
			VListItemIcon(VIcon("edit")),
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
		if IsInDialog(ctx.R.Context()) {
			onclick.URL(ctx.R.RequestURI).
				Query(ParamOverlay, actions.Dialog).
				Query(ParamInDialog, true).
				Query(ParamListingQueries, ctx.Queries().Encode())
		}
		return VListItem(
			VListItemIcon(VIcon("delete")),
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

func isInDialogFromQuery(ctx *web.EventContext) bool {
	return ctx.R.URL.Query().Get(ParamInDialog) == "true"
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

type inputElem interface {
	*h.HTMLTagBuilder | *VSelectBuilder | *VTextFieldBuilder
}

var (
	htmlTagBuilderType = reflect.TypeOf(&h.HTMLTagBuilder{})
)

// InputWithDefaults fills input element with name, value and label.
// For now, input must be h.Input, v.VSelect or v.VTextField.
// This function will be implemented using the FieldContext method
// when golang supports generic method and the current function will be deprecated.
func InputWithDefaults[T inputElem](input T, obj any, field *FieldContext) T {
	rv := reflect.ValueOf(input)
	rt := rv.Type()
	if rt == htmlTagBuilderType {
		kvs := web.VFieldName(field.FormKey)
		in := make([]reflect.Value, 0, len(kvs))
		for i, _ := range kvs {
			v := kvs[i]
			in = append(in, reflect.ValueOf(v))
		}
		rv.MethodByName("Attr").Call(in)
		rv.MethodByName("Value").Call([]reflect.Value{reflect.ValueOf(field.StringValue(obj))})
	} else {
		rv.MethodByName("FieldName").Call([]reflect.Value{reflect.ValueOf(field.FormKey)})
		rv.MethodByName("Value").Call([]reflect.Value{reflect.ValueOf(field.Value(obj))})
		rv.MethodByName("Label").Call([]reflect.Value{reflect.ValueOf(field.Label)})
	}

	return input
}
