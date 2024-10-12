package presets

import (
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type NestedManyBuilder struct {
	*ModelBuilder
	parent     *ModelBuilder
	foreignKey string

	initialListingCompoProcessor func(evCtx *web.EventContext, lb *ListingBuilder, compo *ListingCompo) error
}

const ParamParentID = "parent_id"

func (parent *ModelBuilder) NestedMany(elementModel any, foreignKey string) *NestedManyBuilder {
	rtElement := reflect.TypeOf(elementModel)
	for rtElement.Kind() != reflect.Ptr {
		panic(errors.Errorf("element model %T is not a pointer", elementModel))
	}
	if !hasField(rtElement, foreignKey) {
		panic(errors.Errorf("field %s not found in element model %T", foreignKey, elementModel))
	}

	mb := parent.p.Model(elementModel).InMenu(false)
	mb.Listing().PerPage(10).Except(foreignKey)
	mb.Editing().Except(foreignKey)

	return &NestedManyBuilder{
		ModelBuilder: mb,
		parent:       parent,
		foreignKey:   foreignKey,
	}
}

func (mb *NestedManyBuilder) InitialListingCompoProcessor(f func(evCtx *web.EventContext, lb *ListingBuilder, compo *ListingCompo) error) *NestedManyBuilder {
	mb.initialListingCompoProcessor = f
	return mb
}

func (mb *NestedManyBuilder) FieldInstall(fb *FieldBuilder) error {
	mb.URIName(fmt.Sprintf("%s-nested-%s", mb.parent.Info().URIName(), inflection.Plural(strcase.ToKebab(fb.name))))

	foreignQuery := strcase.ToSnake(mb.foreignKey) + " = ?"
	mb.Listing().WrapSearchFunc(func(in SearchFunc) SearchFunc {
		return func(ctx *web.EventContext, params *SearchParams) (result *SearchResult, err error) {
			compo := ListingCompoFromContext(ctx.R.Context())
			if compo == nil || compo.ParentID == "" {
				err = perm.PermissionDenied
				return
			}
			params.SQLConditions = append(params.SQLConditions, &SQLCondition{
				Query: foreignQuery,
				Args:  []any{compo.ParentID},
			})
			return in(ctx, params)
		}
	})
	mb.Editing().WrapSaveFunc(func(in SaveFunc) SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			parentID := ctx.R.FormValue(ParamParentID)
			if parentID == "" {
				return perm.PermissionDenied
			}
			if err := reflectutils.Set(obj, mb.foreignKey, parentID); err != nil {
				return err
			}
			return in(obj, id, ctx)
		}
	})

	fb.ComponentFunc(func(obj any, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		pid := MustObjectID(obj)

		compo, err := mb.Listing().nestedManyComponent(ctx,
			pid, fb.name,
			mb.initialListingCompoProcessor,
		)
		if err != nil {
			panic(err)
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

func (b *ListingBuilder) nestedManyComponent(evCtx *web.EventContext,
	parentID, unique string,
	initialCompoPreprocessor func(evCtx *web.EventContext, lb *ListingBuilder, compo *ListingCompo) error,
) (r h.HTMLComponent, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
	}

	title, titleCompo, err := b.getTitle(evCtx, ListingStyleNested)
	if err != nil {
		return r, err
	}
	if titleCompo == nil {
		titleCompo = h.H2(title).Attr("v-pre", true).Class("section-title")
	}

	evCtx.WithContextValue(ctxInDialog, true)

	injectorName := b.injectorName()
	compo := &ListingCompo{
		ID:                 fmt.Sprintf("%s_inline_%s_%s", injectorName, parentID, unique),
		Popup:              true,
		LongStyleSearchBox: true,
		ParentID:           parentID,
	}
	if initialCompoPreprocessor != nil {
		if err := initialCompoPreprocessor(evCtx, b, compo); err != nil {
			return r, err
		}
	}
	return web.Scope().VSlot("{ form }").Children(
		h.Div(
			h.Div().Class("d-flex align-center section-title-wrap mb-0").Children(
				titleCompo,
				VSpacer(),
				h.Div().Id(compo.ActionsComponentTeleportToID()),
			),
			VCardText().Class("pa-0").Children(
				b.mb.p.dc.MustInject(injectorName, compo),
			),
		).Class("section-wrap"),
	), nil
}
