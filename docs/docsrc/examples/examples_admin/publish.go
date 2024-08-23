package examples_admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/qor/oss"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// @snippet_begin(PublishInjectModules)
type WithPublishProduct struct {
	gorm.Model

	Name  string
	Price int

	publish.Status
	publish.Schedule
	publish.Version
}

// @snippet_end

// @snippet_begin(PublishImplementSlugInterfaces)
var (
	_ presets.SlugEncoder = (*WithPublishProduct)(nil)
	_ presets.SlugDecoder = (*WithPublishProduct)(nil)
)

func (p *WithPublishProduct) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *WithPublishProduct) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

// @snippet_end

// @snippet_begin(PublishImplementPublishInterfaces)
var (
	_ publish.PublishInterface   = (*WithPublishProduct)(nil)
	_ publish.UnPublishInterface = (*WithPublishProduct)(nil)
)

func (p *WithPublishProduct) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	// create publish actions
	return
}

func (p *WithPublishProduct) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	// create unpublish actions
	return
}

// @snippet_end
func PublishExample(b *presets.Builder, db *gorm.DB) http.Handler {
	err := db.AutoMigrate(&WithPublishProduct{})
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:presets:with_publish_products:*"),
		),
	)
	// @snippet_begin(PublishConfigureView)
	mb := b.Model(&WithPublishProduct{})
	dp := mb.Detailing(publish.VersionsPublishBar, "Details").Drawer(true)
	dp.Section("Details").
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			product := obj.(*WithPublishProduct)
			detail := vx.DetailInfo(
				vx.DetailColumn(
					vx.DetailField(vx.OptionalText(product.Name).ZeroLabel("No Name")).Label("Name"),
					vx.DetailField(vx.OptionalText(fmt.Sprint(product.Price)).ZeroLabel("No Price")).Label("Price"),
				).Header("PRODUCT INFORMATION"),
			)
			return detail
		}).
		Editing("Name", "Price")
	dp.AfterTitleCompFunc(publish.DefaultVersionBar(db))

	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:   "1",
			Name: "John",
		}, nil
	}).
		AutoMigrate()
	publisher := publish.New(db, nil).Activity(ab)
	b.Use(publisher)
	mb.Use(publisher)
	// run the publisher job if Schedule is used
	go publish.RunPublisher(db, nil, publisher)
	// @snippet_end
	return b
}
