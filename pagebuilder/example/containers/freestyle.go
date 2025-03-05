package containers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type Freestyle struct {
	ID          uint
	TopSpace    int
	BottomSpace int
	AnchorID    string

	HTML     string
	CSS      string
	TopJS    string
	BottomJS string
}

func (*Freestyle) TableName() string {
	return "container_freestyles"
}

func RegisterFreestyleContainer(pb *pagebuilder.Builder, db *gorm.DB, ab *activity.Builder) {
	vb := pb.RegisterContainer("Freestyle").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) h.HTMLComponent {
			fs := obj.(*Freestyle)
			var pi *pagebuilder.PageLayoutInput
			if f := ctx.R.Context().Value(pagebuilder.CtxKeyContainerToPageLayout{}); f != nil {
				pi = f.(*pagebuilder.PageLayoutInput)
			} else {
				pi = &pagebuilder.PageLayoutInput{}
				ctx.R = ctx.R.WithContext(context.WithValue(ctx.R.Context(), pagebuilder.CtxKeyContainerToPageLayout{}, pi))
			}
			if strings.TrimSpace(fs.CSS) != "" {
				pi.FreeStyleCss = append(pi.FreeStyleCss, strings.TrimSpace(fs.CSS))
			}
			if strings.TrimSpace(fs.TopJS) != "" {
				pi.FreeStyleTopJs = append(pi.FreeStyleTopJs, strings.TrimSpace(fs.TopJS))
			}
			if strings.TrimSpace(fs.BottomJS) != "" {
				pi.FreeStyleBottomJs = append(pi.FreeStyleBottomJs, strings.TrimSpace(fs.BottomJS))
			}
			return FreestyleBody(fs, input)
		})
	mb := vb.Model(&Freestyle{})
	eb := mb.Editing("TopSpace", "BottomSpace", "AnchorID", "HTML", "CSS", "TopJS", "BottomJS")
	eb.Field("HTML").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Type("textarea").Attr(web.VField(field.FormKey, field.Value(obj))...).Label(field.Label).Disabled(field.Disabled)
	})
	eb.Field("CSS").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Type("textarea").Attr(web.VField(field.FormKey, field.Value(obj))...).Label(field.Label).Disabled(field.Disabled)
	})
	eb.Field("TopJS").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Type("textarea").Attr(web.VField(field.FormKey, field.Value(obj))...).Label(field.Label).Disabled(field.Disabled)
	})
	eb.Field("BottomJS").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Type("textarea").Attr(web.VField(field.FormKey, field.Value(obj))...).Label(field.Label).Disabled(field.Disabled)
	})
	ab.RegisterModel(mb.GetModelBuilder())

	eb.Field("TopSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXSelect().
			Items([]int{0, 8, 12, 16, 20, 24, 32, 40}).
			Label(field.Label).
			Attr(web.VField(field.FormKey, field.Value(obj))...).
			Disabled(field.Disabled)
	})
	eb.Field("BottomSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXSelect().
			Items([]int{0, 8, 12, 16, 20, 24, 32, 40}).
			Label(field.Label).
			Attr(web.VField(field.FormKey, field.Value(obj))...).
			Disabled(field.Disabled)
	})

	eb.Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			anchorID, err := reflectutils.Get(obj, "AnchorID")
			if err != nil {
				return
			}
			if anchorID == "" {
				if err = reflectutils.Set(obj, "AnchorID", generateAnchorID(obj)); err != nil {
					return
				}
			}
			if err = in(obj, id, ctx); err != nil {
				return
			}
			return
		}
	})

	eb.EditingTitleFunc(func(obj interface{}, defaultTitle string, ctx *web.EventContext) h.HTMLComponent {
		id, err := reflectutils.Get(obj, "ID")
		if err != nil {
			panic(err)
		}

		modelName := reflect.TypeOf(obj).Elem().Name()

		var displayName string
		if err := db.Table("page_builder_containers").
			Where("model_name = ? AND model_id = ?", modelName, id).
			Pluck("display_name", &displayName).
			Error; err != nil {
			panic(err)
		}

		return h.Text(displayName)
	})
}

func generateAnchorID(obj interface{}) string {
	return fmt.Sprintf("%s-%s", strcase.ToKebab(reflect.TypeOf(obj).Elem().Name()),
		strings.ReplaceAll(uuid.New().String(), "-", ""),
	)
}

func FreestyleBody(data *Freestyle, input *pagebuilder.RenderInput) (body h.HTMLComponent) {
	body = ContainerWrapper1(
		data.AnchorID, data.TopSpace, data.BottomSpace, "padding-left:16px;padding-right:16px;",
		h.RawHTML(data.HTML),
	)
	return
}

func ContainerWrapper1(
	anchorID string, TopSpace, BottomSpace int, style string,
	comp ...h.HTMLComponent,
) h.HTMLComponent {
	return h.Div(comp...).
		Id(anchorID).
		Class("container-instance").
		StyleIf(fmt.Sprintf("padding-top:%dpx", TopSpace), TopSpace != 0).
		StyleIf(fmt.Sprintf("padding-bottom:%dpx", BottomSpace), BottomSpace != 0).
		Style("position:relative;").StyleIf(style, style != "")
}
