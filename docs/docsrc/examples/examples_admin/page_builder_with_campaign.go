package examples_admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// models
type (
	Campaign struct {
		gorm.Model
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
)

// containers
type (
	CampaignContent struct {
		ID     uint
		Title  string
		Banner string
	}
	MyContent struct {
		ID    uint
		Text  string
		Color string
	}
	ProductContent struct {
		ID   uint
		Name string
	}
)

func (b *Campaign) GetTitle() string {
	return b.Title
}

func (p *Campaign) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Campaign) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		presets.ParamID:     segs[0],
		publish.SlugVersion: segs[1],
	}
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
		panic("wrong slug")
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
		&Campaign{}, &CampaignProduct{}, // models
		&MyContent{}, &CampaignContent{}, &ProductContent{}, // containers

	)
	if err != nil {
		panic(err)
	}
	puBuilder := publish.New(db, nil)
	if b.GetPermission() == nil {
		b.Permission(
			perm.New().Policies(
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			),
		)
	}
	b.Use(puBuilder)

	pb := pagebuilder.New(b.GetURIPrefix()+"/page_builder", db, b.I18n()).
		Publisher(puBuilder)

	header := pb.RegisterContainer("MyContent").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*MyContent)
			return Div().Text(c.Text).Style("height:200px")
		})

	ed := header.Model(&MyContent{}).Editing("Text", "Color")
	ed.Field("Color").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VTextField().
			Variant(vuetify.FieldVariantUnderlined).
			Label(field.Label).
			Attr(web.VField(field.FormKey, field.Value(obj))...)
	})

	// Campaigns Menu
	campaignModelBuilder := b.Model(&Campaign{})
	campaignModelBuilder.Listing("Title")
	detail := campaignModelBuilder.Detailing(
		pagebuilder.PageBuilderPreviewCard,
		"CampaignDetail",
	)
	detail.Section("CampaignDetail").Editing("Title")

	pb.RegisterModelContainer("CampaignContent", campaignModelBuilder).Group("Campaign").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*CampaignContent)
			return Div(Text(c.Title)).Style("height:200px")
		}).Model(&CampaignContent{}).Editing("Title", "Banner")

	campaignModelBuilder.Use(pb)

	// Products Menu
	productModelBuilder := b.Model(&CampaignProduct{})
	productModelBuilder.Listing("Name")

	detail2 := productModelBuilder.Detailing(
		pagebuilder.PageBuilderPreviewCard,
		"ProductDetail",
	)

	detail2.Section("ProductDetail").Editing("Name")

	pb.RegisterModelContainer("ProductContent", productModelBuilder).Group("CampaignProduct").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*ProductContent)
			return Div(Text(c.Name)).Style("height:200px")
		}).Model(&ProductContent{}).Editing("Name")

	productModelBuilder.Use(pb)

	return TestHandler(pb, b)
}
