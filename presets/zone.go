package presets

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"

	"github.com/qor5/web/v3"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
)

const QueryKeyZone = "_zone_"

type zoneCtxKey struct{}

func Zone[T any](ctx *web.EventContext) T {
	zone, ok := ctx.ContextValue(zoneCtxKey{}).(T)
	if ok {
		return zone
	}

	z := ctx.R.FormValue(QueryKeyZone)
	if z != "" {
		if err := json.Unmarshal([]byte(z), &zone); err != nil {
			panic(err)
		}
		ctx.WithContextValue(zoneCtxKey{}, zone)
	}
	return zone
}

func ZoneOrCreate[T any](ctx *web.EventContext) T {
	zone := Zone[T](ctx)
	if lo.IsNil(zone) {
		zone = reflect.New(reflect.TypeOf(zone).Elem()).Interface().(T)
		ctx.WithContextValue(zoneCtxKey{}, zone)
	}
	return zone
}

type ListingZone struct {
	ID       string                `json:"id,omitempty"`
	Style    ListingComponentStyle `json:"style,omitempty"`
	ParentID string                `json:"parent_id,omitempty"`
}

func (z *ListingZone) Plaid() *web.VueEventTagBuilder {
	if z == nil {
		return web.Plaid()
	}
	if z.ID == "" && z.Style == ListingComponentStylePage {
		return web.Plaid()
	}

	// If this is used, the pass is lost in many places. eg: EditingBuilder.doDelete
	// return web.Plaid().FieldValue(QueryKeyZone, h.JSONString(z))

	// StringQuery will be not overwritten by Queries.
	vs := url.Values{
		QueryKeyZone: []string{
			h.JSONString(z),
		},
	}
	return web.Plaid().StringQuery(vs.Encode())
}

func (z *ListingZone) Portal(name string) string {
	if z == nil || z.ID == "" {
		return name
	}
	return z.ID + "." + name
}

func (z *ListingZone) ApplyToRequest(r *http.Request) {
	query := r.URL.Query()
	query.Set(QueryKeyZone, h.JSONString(z))
	r.URL.RawQuery = query.Encode()
	r.RequestURI = r.URL.String()
}

func ListingZoneFromContext(ctx *web.EventContext) *ListingZone {
	zone := Zone[*ListingZone](ctx)
	if zone != nil {
		return nil
	}
	listingQueries := ctx.R.FormValue(ParamListingQueries)
	if listingQueries == "" {
		return nil
	}
	qs, err := url.ParseQuery(listingQueries)
	if err != nil {
		return nil
	}
	z := qs.Get(QueryKeyZone)
	if z == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(z), &zone); err != nil {
		panic(err)
	}
	ctx.WithContextValue(zoneCtxKey{}, zone)
	return zone
}
