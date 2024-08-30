package presets

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type RowMenuBuilder struct {
	mb              *ModelBuilder
	listings        []string
	defaultListings []string
	items           map[string]*RowMenuItemBuilder
}

func (b *ListingBuilder) RowMenu(listings ...string) *RowMenuBuilder {
	if b.rowMenu == nil {
		b.rowMenu = &RowMenuBuilder{
			mb:       b.mb,
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
	b.items = make(map[string]*RowMenuItemBuilder)
}

func (b *RowMenuBuilder) listingItemFuncs(ctx *web.EventContext) (fs []vx.RowMenuItemFunc) {
	listings := b.defaultListings
	if len(b.listings) > 0 {
		listings = b.listings
	}
	for _, li := range listings {
		if ib, ok := b.items[strcase.ToSnake(li)]; ok {
			comp := ib.getComponentFunc(ctx)
			if comp != nil {
				fs = append(fs, comp)
			}
		}
	}
	return fs
}

type RowMenuItemBuilder struct {
	rmb        *RowMenuBuilder
	name       string
	icon       string
	clickF     RowMenuItemClickFunc
	compF      vx.RowMenuItemFunc
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
		eventID: fmt.Sprintf("%s_rowMenuItemFunc_%s", b.mb.uriName, name),
	}
	b.items[strcase.ToSnake(name)] = ib
	b.defaultListings = append(b.defaultListings, name)

	b.mb.RegisterEventFunc(ib.eventID, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue(ParamID)
		if ib.permAction != "" {
			obj := b.mb.NewModel()
			obj, err = b.mb.editing.Fetcher(obj, id, ctx)
			if err != nil {
				return r, err
			}
			err = b.mb.Info().Verifier().Do(ib.permAction).ObjectOn(obj).WithReq(ctx.R).IsAllowed()
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

func (b *RowMenuItemBuilder) ComponentFunc(v vx.RowMenuItemFunc) *RowMenuItemBuilder {
	b.compF = v
	return b
}

func (b *RowMenuItemBuilder) PermAction(v string) *RowMenuItemBuilder {
	b.permAction = v
	return b
}

func (b *RowMenuItemBuilder) getComponentFunc(_ *web.EventContext) vx.RowMenuItemFunc {
	if b.compF != nil {
		return b.compF
	}

	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		if b.permAction != "" && b.rmb.mb.Info().Verifier().Do(b.permAction).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}
		return VListItem(
			h.If(b.icon != "", web.Slot(VIcon(b.icon)).Name("prepend")),
			VListItemTitle(h.Text(i18n.PT(ctx.R, ModelsI18nModuleKey, strcase.ToCamel(b.rmb.mb.label+" RowMenuItem"), b.name))),
		).Attr("@click", web.Plaid().
			EventFunc(b.eventID).
			Query(ParamID, id).
			Go())
	}
}
