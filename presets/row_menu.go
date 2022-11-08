package presets

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/ui/stripeui"
	s "github.com/goplaid/ui/stripeui"
	. "github.com/goplaid/ui/vuetify"
	"github.com/iancoleman/strcase"
	h "github.com/theplant/htmlgo"
)

type RowMenuBuilder struct {
	lb              *ListingBuilder
	listings        []string
	defaultListings []string
	items           map[string]*RowMenuItemBuilder
}

func (b *ListingBuilder) RowMenu(listings ...string) *RowMenuBuilder {
	if b.rowMenu == nil {
		b.rowMenu = &RowMenuBuilder{
			lb:       b,
			listings: listings,
			items:    make(map[string]*RowMenuItemBuilder),
		}
	}

	rmb := b.rowMenu
	if len(listings) == 0 {
		return rmb
	}
	rmb.listings = listings
	for _, li := range rmb.listings {
		rmb.RowMenuItem(li)
	}

	return rmb
}

func (b *RowMenuBuilder) Empty() {
	b.listings = nil
	b.items = nil
}

func (b *RowMenuBuilder) listingItemFuncs(ctx *web.EventContext) (fs []s.RowMenuItemFunc) {
	listings := b.defaultListings
	if len(b.listings) > 0 {
		listings = b.listings
	}
	for _, li := range listings {
		if ib, ok := b.items[strcase.ToSnake(li)]; ok {
			fs = append(fs, ib.getComponentFunc(ctx))
		}
	}
	return fs
}

type RowMenuItemBuilder struct {
	rmb        *RowMenuBuilder
	name       string
	icon       string
	clickF     RowMenuItemClickFunc
	compF      stripeui.RowMenuItemFunc
	permAction string
	eventID    string
}

func (b *RowMenuBuilder) RowMenuItem(name string) *RowMenuItemBuilder {
	if v, ok := b.items[strcase.ToSnake(name)]; ok {
		return v
	}

	ib := &RowMenuItemBuilder{
		rmb:     b,
		name:    name,
		eventID: fmt.Sprintf("%s_rowMenuItemFunc_%s", b.lb.mb.uriName, name),
	}
	b.items[strcase.ToSnake(name)] = ib
	b.defaultListings = append(b.defaultListings, name)

	b.lb.mb.RegisterEventFunc(ib.eventID, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue(ParamID)
		if ib.permAction != "" {
			var obj = b.lb.mb.NewModel()
			obj, err = b.lb.mb.editing.Fetcher(obj, id, ctx)
			if err != nil {
				return r, err
			}
			err = b.lb.mb.Info().Verifier().Do(ib.permAction).ObjectOn(obj).WithReq(ctx.R).IsAllowed()
			if err != nil {
				return r, err
			}
		}
		if ib.clickF == nil {
			return r, nil
		}
		return ib.clickF(ctx, id)
	})

	return ib
}

func (b *RowMenuItemBuilder) Icon(v string) *RowMenuItemBuilder {
	b.icon = v
	return b
}

type RowMenuItemClickFunc func(ctx *web.EventContext, id string) (r web.EventResponse, err error)

func (b *RowMenuItemBuilder) OnClick(v RowMenuItemClickFunc) *RowMenuItemBuilder {
	b.clickF = v
	return b
}

func (b *RowMenuItemBuilder) ComponentFunc(v stripeui.RowMenuItemFunc) *RowMenuItemBuilder {
	b.compF = v
	return b
}

func (b *RowMenuItemBuilder) PermAction(v string) *RowMenuItemBuilder {
	b.permAction = v
	return b
}

func (b *RowMenuItemBuilder) getComponentFunc(ctx *web.EventContext) stripeui.RowMenuItemFunc {
	if b.compF != nil {
		return b.compF
	}

	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		if b.permAction != "" && b.rmb.lb.mb.Info().Verifier().Do(b.permAction).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}
		return VListItem(
			VListItemIcon(VIcon(b.icon)),
			VListItemTitle(h.Text(i18n.PT(ctx.R, ModelsI18nModuleKey, strcase.ToCamel(b.rmb.lb.mb.label+" RowMenuItem"), b.name))),
		).Attr("@click", web.Plaid().
			EventFunc(b.eventID).
			Query(ParamID, id).
			Go())
	}
}
