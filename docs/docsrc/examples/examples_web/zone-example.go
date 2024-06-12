package examples_web

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

const (
	updatePortalEvent = "updatePortal"
	portalName1       = "portal1"
)

type ListZone struct {
	Id     string
	Number int
}

func (lz *ListZone) PortalName(n string) string {
	return fmt.Sprintf("ListZone_%s:%s", lz.Id, n)
}

func ZoneExample(cx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = comp1(&ListZone{Id: "123", Number: 500})
	return
}

func comp1(z *ListZone) HTMLComponent {
	return web.ComponentWithZone(z, ComponentFunc(func(ctx context.Context) (r []byte, err error) {
		return Div(
			web.Portal().Name(z.PortalName(portalName1)),
			Button("Send").Attr(
				"@click",
				web.ZonedPlaid(ctx).
					Query("abc", "1").
					EventFunc(updatePortalEvent).
					Go(),
			),
			comp2(),
		).MarshalHTML(ctx)
	}))
}

func comp2() HTMLComponent {
	return ComponentFunc(func(ctx context.Context) (r []byte, err error) {
		return Button("Send2").
			Attr("@click",
				web.ZonedPlaid(ctx).
					Query("abc", "1").
					EventFunc(updatePortalEvent).
					Go(),
			).MarshalHTML(ctx)
	})
}

func updatePortal(ctx *web.EventContext) (r web.EventResponse, err error) {
	z := web.MustGetZone[*ListZone](ctx.R)
	// z.UpdatePortal(&r, Div().Text("test"))
	r.UpdatePortals = append(r.UpdatePortals, web.ZonePortalUpdate(z, portalName1, Div(
		Pre(JSONString(z)),
		comp2(),
	)))
	return
}

var ZoneExamplePB = web.Page(ZoneExample).
	EventFunc(updatePortalEvent, updatePortal)

var ZoneExamplePath = examples.URLPathByFunc(ZoneExample)
