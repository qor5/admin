package examples_admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/spf13/cast"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
)

// models
type (
	Campaign struct {
		gorm.Model
		SEO   seo.Setting
		Title string
		publish.Status
		publish.Schedule
		publish.Version
	}

	CampaignProduct struct {
		gorm.Model
		Name string
		publish.Status
		publish.Schedule
		publish.Version
	}

	PageProduct struct {
		gorm.Model
		Name string
		publish.Status
		publish.Schedule
		publish.Version
	}

	// Others

	CampaignWithStringID struct {
		ID        string `gorm:"primarykey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Name  string
		Price int

		publish.Status
		publish.Schedule
		publish.Version
	}
)

// containers
type (
	CampaignContent struct {
		ID     uint
		Title  string
		Banner string
		Image  media_library.MediaBox
	}
	MyContent struct {
		ID    uint
		Text  string
		Color string
	}

	PagesContent struct {
		ID    uint
		Text  string
		Color string
	}

	ProductContent struct {
		ID   uint
		Name string
	}
)

func (p *CampaignWithStringID) PrimarySlug() string {
	return fmt.Sprintf("%v", p.ID)
}

func (p *CampaignWithStringID) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 1 {
		panic(presets.ErrNotFound("wrong slug"))
	}
	return map[string]string{
		presets.ParamID: segs[0],
	}
}

func (b *CampaignWithStringID) PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	b.OnlineUrl = fmt.Sprintf("campaign-with-string-ids/%v/index.html", b.ID)
	return b.OnlineUrl
}

func (b *Campaign) GetTitle() string {
	return b.Title
}

func (b *Campaign) PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	b.OnlineUrl = fmt.Sprintf("campaigns/%v/index.html", b.ID)
	return b.OnlineUrl
}

func (b *Campaign) WrapPublishActions(in publish.PublishActionsFunc) publish.PublishActionsFunc {
	return func(ctx context.Context, db *gorm.DB, storage oss.StorageInterface, obj any) (actions []*publish.PublishAction, err error) {
		// default actions
		if actions, err = in(ctx, db, storage, obj); err != nil {
			return
		}
		actions = append(actions, &publish.PublishAction{
			Url:     "campaigns/index.html",
			Content: "Campaign List",
		})

		return
	}
}

func (p *Campaign) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Campaign) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	return map[string]string{
		presets.ParamID:     segs[0],
		publish.SlugVersion: segs[1],
	}
}

func (b *CampaignProduct) PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	b.OnlineUrl = fmt.Sprintf("campaign-products/%v/index.html", b.ID)
	return b.OnlineUrl
}

func (b *CampaignProduct) GetTitle() string {
	return b.Name
}

func (p *CampaignProduct) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *CampaignProduct) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	return map[string]string{
		presets.ParamID:     segs[0],
		publish.SlugVersion: segs[1],
	}
}

func (b *PageProduct) PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (s string) {
	b.OnlineUrl = fmt.Sprintf("page-products/%v/index.html", b.ID)
	return b.OnlineUrl
}

func (b *PageProduct) GetTitle() string {
	return b.Name
}

func (p *PageProduct) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *PageProduct) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	return map[string]string{
		presets.ParamID:     segs[0],
		publish.SlugVersion: segs[1],
	}
}

func TestHandler(pageBuilder *pagebuilder.Builder, b *presets.Builder) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(b.GetURIPrefix()+"/page_builder", pageBuilder)
	mux.Handle(b.GetURIPrefix()+"/page_builder/", pageBuilder)
	if b.GetURIPrefix() != "" {
		mux.Handle(b.GetURIPrefix(), b)
	}
	mux.Handle(b.GetURIPrefix()+"/", b)

	return mux
}

func PageBuilderExample(b *presets.Builder, db *gorm.DB) http.Handler {
	b.DataOperator(gorm2op.DataOperator(db))
	err := db.AutoMigrate(
		&Campaign{}, &CampaignProduct{}, &PageProduct{}, // models
		&MyContent{}, &CampaignContent{}, &ProductContent{}, &PagesContent{}, // containers
		&CampaignWithStringID{},
	)
	if err != nil {
		panic(err)
	}
	storage := filesystem.New("/tmp/publish")
	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:     "1",
			Name:   "John",
			Avatar: "https://i.pravatar.cc/300",
		}, nil
	}).AutoMigrate()

	puBuilder := publish.New(db, storage).DisablementCheckFunc(func(ctx *web.EventContext, obj any) *publish.Disablement {
		status := obj.(publish.StatusInterface).EmbedStatus().Status
		disabled := status == publish.StatusOnline
		return &publish.Disablement{
			DisabledRename: disabled,
			DisabledDelete: disabled,
		}
	})
	if b.GetPermission() == nil {
		b.Permission(
			perm.New().Policies(
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			),
		)
	}
	b.Use(puBuilder)
	seoBuilder := seo.New(db).AutoMigrate()

	pb := pagebuilder.New("/page_builder", db, b).
		Activity(ab).
		SEO(seoBuilder).
		Only("Title", "Slug").
		DisabledNormalContainersGroup(true).
		PreviewOpenNewTab(true).
		Publisher(puBuilder).
		DemoContainerActivityProcessor(func(ctx *web.EventContext, input *pagebuilder.DemoContainerLogInput) *pagebuilder.DemoContainerLogInput {
			return nil
		}).
		EditorActivityProcessor(func(ctx *web.EventContext, input *pagebuilder.EditorLogInput) *pagebuilder.EditorLogInput {
			return nil
		}).
		PreviewDevices(
			pagebuilder.Device{Name: pagebuilder.DeviceComputer, Width: "", Icon: "mdi-monitor", Disabled: true},
			pagebuilder.Device{Name: pagebuilder.DevicePhone, Width: "414px", Icon: "mdi-cellphone"},
			pagebuilder.Device{Name: pagebuilder.DeviceTablet, Width: "768px", Icon: "mdi-tablet", Disabled: true},
		).
		DefaultDevice(pagebuilder.DevicePhone).
		WrapPageLayout(func(v pagebuilder.PageLayoutFunc) pagebuilder.PageLayoutFunc {
			return func(body HTMLComponent, input *pagebuilder.PageLayoutInput, ctx *web.EventContext) HTMLComponent {
				input.WrapHead = func(comps HTMLComponents) HTMLComponents {
					comps = append(comps,
						Script("console.log('in head')"),
						Style(`.test-div { width: 200px;background-color:#E1E1E1; }`),
					)
					return comps
				}
				input.WrapBody = func(comps HTMLComponents) HTMLComponents {
					comps = append(comps, Script("console.log('in body')"),
						Style(`.test-div1 { width: 300px;background-color:blue; }`),
						Style(`.test-div2 { width: 400px;background-color:red; }`))
					return comps
				}
				return v(body, input, ctx)
			}
		})

	pb.WrapPageInstall(func(installFunc presets.ModelInstallFunc) presets.ModelInstallFunc {
		return func(innerPb *presets.Builder, mb *presets.ModelBuilder) (err error) {
			if err = installFunc(innerPb, mb); err != nil {
				return
			}
			mb.Detailing().Field("hide").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
				return Div(
					Iframe().Src(pb.GetPageModelBuilder().PreviewHTML(ctx.R.Context(), obj)),
				).Style("display:none").Id("display_preview")
			})
			return
		}
	})
	if err = pagebuilder.AutoMigrate(db); err != nil {
		panic(err)
	}

	// global containers
	header := pb.RegisterContainer("MyContent").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*MyContent)
			ctx.WithContextValue(pagebuilder.CtxKeyContainerToPageLayout{}, &pagebuilder.PageLayoutInput{
				FreeStyleCss: []string{`.test-ctx {
  color: red;
}`},
			})
			return Div(
				Div().Text(c.Text).Class("test-ctx"),
				Div().Text(c.Text).Class("test-div"),
			)
		}).Cover("https://qor5.com/img/qor-logo.png")

	ed := header.Model(&MyContent{}).Editing("Text", "Color")
	ed.Field("Color").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VTextField().
			Variant(vuetify.FieldVariantUnderlined).
			Label(field.Label).
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...)
	})

	// only pages view containers set OnlyPages true
	pc := pb.RegisterContainer("PagesContent").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*PagesContent)
			return Div().Text(c.Text).Class("test-div")
		}).Cover("https://qor5.com/img/qor-logo.png")

	pce := pc.Model(&PagesContent{}).Editing("Text", "Color")
	pce.Field("Color").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VTextField().
			Variant(vuetify.FieldVariantUnderlined).
			Label(field.Label).
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...)
	})

	// Campaigns Menu
	campaignModelBuilder := b.Model(&Campaign{})
	cmbCreating := campaignModelBuilder.Editing().Creating(pagebuilder.PageTemplateSelectionFiled, "Title")
	cmbCreating.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = in(obj, id, ctx); err != nil {
				return
			}
			if p, ok := obj.(presets.SlugEncoder); ok {
				ctx.R.Form.
					Set(presets.ParamOverlayAfterUpdateScript,
						web.Plaid().URL(campaignModelBuilder.Info().DetailingHref(p.PrimarySlug())).PushState(true).Go())
			}
			return
		}
	})
	campaignModelBuilder.Editing().ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Campaign)
		if c.Title == "" {
			err.GlobalError("title could not be empty")
		}
		return
	})
	campaignModelBuilder.Listing("Title")
	detail := campaignModelBuilder.Detailing(
		pagebuilder.PageBuilderPreviewCard,
		"CampaignDetail",
		seo.SeoDetailFieldName,
	)
	campaignDetailSection := presets.NewSectionBuilder(campaignModelBuilder, "CampaignDetail").Editing("Title")
	detail.Section(campaignDetailSection)
	pb.RegisterModelContainer("CampaignContent", campaignModelBuilder).Group("Campaign").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*CampaignContent)
			return Div(Text(c.Title)).Class("test-div1")
		}).Model(&CampaignContent{}).Editing("Title", "Banner", "Image")

	campaignModelBuilder.Use(pb)

	// Products Menu
	productModelBuilder := b.Model(&CampaignProduct{})
	productModelBuilder.Editing().Creating(pagebuilder.PageTemplateSelectionFiled, "Name")
	productModelBuilder.Listing("Name")

	detail2 := productModelBuilder.Detailing(
		pagebuilder.PageBuilderPreviewCard,
		"ProductDetail",
	)
	productDetail := presets.NewSectionBuilder(productModelBuilder, "ProductDetail").Editing("Name")
	detail2.Section(productDetail)

	pb.RegisterModelContainer("ProductContent", productModelBuilder).Group("CampaignProduct").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*ProductContent)
			return Div(Text(c.Name)).Class("test-div2")
		}).Model(&ProductContent{}).Editing("Name")

	productModelBuilder.Use(pb)

	// Page Product Menu
	pageProductModelBuilder := b.Model(&PageProduct{})
	pageProductModelBuilder.Editing().Creating(pagebuilder.PageTemplateSelectionFiled, "Name")
	// just use public containers
	pageProductModelBuilder.Listing("Name")
	detail3 := pageProductModelBuilder.Detailing(
		pagebuilder.PageBuilderPreviewCard,
		"ProductDetail",
	)

	productDetail3 := presets.NewSectionBuilder(pageProductModelBuilder, "ProductDetail").Editing("Name")
	detail3.Section(productDetail3)

	pageProductModelBuilder.Use(pb)

	mb := b.Model(&CampaignWithStringID{})
	mb.Editing().Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = reflectutils.Set(obj, "ID", fmt.Sprintf("ox%v", time.Now().Unix())); err != nil {
				return
			}
			return in(obj, id, ctx)
		}
	})
	dp := mb.Detailing(publish.VersionsPublishBar, "Details").Drawer(true)
	detailSection := presets.NewSectionBuilder(mb, "Details").
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			p := obj.(*CampaignWithStringID)
			return vx.DetailInfo(
				vx.DetailColumn(
					vx.DetailField(vx.OptionalText(p.Name).ZeroLabel("No Name")).Label("Name"),
					vx.DetailField(vx.OptionalText(fmt.Sprint(p.Price)).ZeroLabel("No Price")).Label("Price"),
				).Header("PRODUCT INFORMATION"),
			)
		}).
		Editing("Name", "Price")
	mb.Use(puBuilder)
	dp.Section(detailSection)

	// use demo container and media etc. plugins
	b.Use(pb)
	return TestHandler(pb, b)
}
