package presets

import (
	"context"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type InlineListingBuilder struct {
	*ModelBuilder
	parent     *ModelBuilder
	foreignKey string
}

func (parent *ModelBuilder) InlineListing(elementModel any, foreignKey string) *InlineListingBuilder {
	rtElement := reflect.TypeOf(elementModel)
	for rtElement.Kind() != reflect.Ptr {
		panic(errors.Errorf("element model %T is not a pointer", elementModel))
	}
	if !hasField(rtElement, foreignKey) {
		panic(errors.Errorf("field %s not found in element model %T", foreignKey, elementModel))
	}

	mb := parent.p.Model(elementModel).InMenu(false)

	foreignQuery := strcase.ToSnake(foreignKey) + " = ?"
	mb.Listing().PerPage(10).Except(foreignKey).WrapSearchFunc(func(in SearchFunc) SearchFunc {
		return func(model any, params *SearchParams, ctx *web.EventContext) (r any, totalCount int, err error) {
			z := Zone[*ListingZone](ctx)
			if z == nil || z.ParentID == "" {
				return nil, 0, perm.PermissionDenied
			}
			params.SQLConditions = append(params.SQLConditions, &SQLCondition{
				Query: foreignQuery,
				Args:  []any{z.ParentID},
			})
			return in(model, params, ctx)
		}
	})
	mb.Detailing().Except(foreignKey).Drawer(true)
	mb.Editing().Except(foreignKey).WrapSaveFunc(func(in SaveFunc) SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			z := ListingZoneFromContext(ctx)
			if z == nil || z.ParentID == "" {
				return perm.PermissionDenied
			}
			if err := reflectutils.Set(obj, foreignKey, z.ParentID); err != nil {
				return err
			}
			return in(obj, id, ctx)
		}
	})

	return &InlineListingBuilder{
		ModelBuilder: mb,
		parent:       parent,
		foreignKey:   foreignKey,
	}
}

func (mb *InlineListingBuilder) Install(fb *FieldBuilder) (err error) {
	mb.URIName(fmt.Sprintf("%s-inline-%s", mb.parent.Info().URIName(), inflection.Plural(strcase.ToKebab(fb.name))))

	fb.ComponentFunc(func(obj any, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var pid string
		if slugger, ok := obj.(SlugEncoder); ok {
			pid = slugger.PrimarySlug()
		} else {
			pid = fmt.Sprint(reflectutils.MustGet(obj, "ID"))
		}

		zoneID := fmt.Sprintf("[%s:%s:%s]", mb.Info().ListingHref(), pid, fb.name)
		compo, err := mb.Listing().inlineListingComponent(ctx, zoneID, pid)
		if err != nil {
			return h.Div(h.Text(fmt.Sprintf("Error: %s", err.Error())))
		}
		return compo
	})

	return nil
}

func hasField(rt reflect.Type, name string) bool {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if _, ok := rt.FieldByName(name); ok {
		return true
	}
	return false
}

// func InspectInlineListingPluginParams(rtParant reflect.Type, sliceFieldName string) (elementModel any, foreignKey string, err error) {
// 	for rtParant.Kind() == reflect.Ptr {
// 		rtParant = rtParant.Elem()
// 	}
// 	field, ok := rtParant.FieldByName(sliceFieldName)
// 	if !ok {
// 		return nil, "", errors.Errorf("field %s not found", sliceFieldName)
// 	}
// 	if field.Type.Kind() != reflect.Slice {
// 		return nil, "", errors.Errorf("field %s is not a slice", sliceFieldName)
// 	}
// 	rtElement := field.Type.Elem()
// 	if rtElement.Kind() == reflect.Ptr {
// 		elementModel = reflect.New(rtElement.Elem()).Interface()
// 	} else {
// 		elementModel = reflect.New(rtElement).Elem().Interface()
// 	}

// 	foreignKey = rtParant.Name() + "ID"
// 	if !hasField(rtElement, foreignKey) {
// 		return nil, "", errors.Errorf("field %s not found in element model %T", foreignKey, rtElement.Elem())
// 	}

// 	return elementModel, foreignKey, nil
// }

const (
	portalListingInlineActions = "presets_portalListingInlineActions"
	portalListingInlineList    = "presets_portalListingInlineList"
)

func (b *ListingBuilder) inlinePortalUpdates(ctx *web.EventContext) (r []*web.PortalUpdate) {
	z := Zone[*ListingZone](ctx)
	listingCompo := b.listingComponent(ctx, true)
	return []*web.PortalUpdate{
		{
			Name: z.Portal(portalListingInlineActions),
			Body: GetActionsComponent(ctx), // TODO: ?
		},
		{
			Name: z.Portal(portalListingInlineList),
			Body: listingCompo,
		},
	}
}

func (b *ListingBuilder) inlineListingComponent(ctx *web.EventContext, id, parentID string) (r h.HTMLComponent, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(ctx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	// hijack
	hijackEventCtx := ShadowCloneEventContext(ctx)
	req := hijackEventCtx.R.Clone(hijackEventCtx.R.Context())
	req.URL.Path = b.mb.Info().ListingHref()
	req.URL.RawQuery = ""
	req.RequestURI = req.URL.RequestURI()
	hijackEventCtx.R = req
	defer func() {
		if err == nil {
			compo := r
			r = h.ComponentFunc(func(ctx context.Context) (r []byte, err error) {
				ctx = web.WrapEventContext(ctx, hijackEventCtx)
				return compo.MarshalHTML(ctx)
			})
		}
	}()
	ctx = hijackEventCtx

	// ensure zone
	z := ZoneOrCreate[*ListingZone](ctx)
	z.Style = ListingComponentStyleInline
	z.ID = id
	z.ParentID = parentID
	z.ApplyToRequest(ctx.R)

	// compo
	title := b.title
	if title == "" {
		msgr := MustGetMessages(ctx.R)
		title = msgr.ListingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, b.mb.label))
	}
	portals := b.inlinePortalUpdates(ctx)
	r = web.Scope().VSlot("{ form }").Children(
		VCard().Elevation(0).Class("ma-n2").Children(
			VCardTitle().Class("d-flex align-center").Children(
				h.Text(title),
				VSpacer(),
				UpdateToPortal(portals[0]),
			),
			VCardText().Class("pa-0").Children(
				UpdateToPortal(portals[1]),
			),
		),
	)
	return
}
