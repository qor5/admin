package examples_admin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

func (p *WithPublishProduct) PermissionRN() []string {
	return []string{"a"}
}

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

type WithPublishMenuProduct struct {
	gorm.Model

	Name  string
	Price int

	publish.Status
	publish.Schedule
	publish.Version
}

func (p *WithPublishMenuProduct) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *WithPublishMenuProduct) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (p *WithPublishMenuProduct) PermissionRN() []string {
	return []string{p.Name, strconv.Itoa(int(p.ID))}
}

func (p *WithPublishMenuProduct) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	// create publish actions
	return
}

func (p *WithPublishMenuProduct) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	// create unpublish actions
	return
}

func configWithPublishMenuProduct(b *presets.Builder) (mb *presets.ModelBuilder) {
	mb = b.Model(&WithPublishMenuProduct{})

	b.MenuGroup("permissionRN").SubItems("WithPublishMenuProduct")
	b.GetPermission().CreatePolicies(
		perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(presets.PermList).On(":presets:mg_permission_rn:*"),
		perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:presets:mg_permission_rn:with_publish_menu_products:*"),
	)
	// preset:model_name:{id}
	mb.Detailing(publish.VersionsPublishBar, "nm").Drawer(true)
	mb.Detailing().Section("nm").Viewing("name", "price").Editing("name", "price")
	return
}

func PublishExample(b *presets.Builder, db *gorm.DB) http.Handler {
	err := db.AutoMigrate(
		&WithPublishProduct{},
		&WithPublishMenuProduct{},
	)
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))
	b.MenuOrder(
		"WithPublishProduct",
		"permissionRN",
	)
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:presets:with_publish_products:*"),
		),
	)
	// @snippet_begin(PublishConfigureView)
	mb := b.Model(&WithPublishProduct{})
	mb2 := configWithPublishMenuProduct(b)
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

	ab := activity.New(db, func(ctx context.Context) *activity.User {
		return &activity.User{
			ID:   "1",
			Name: "John",
		}
	}).
		AutoMigrate()
	publisher := publish.New(db, nil).Activity(ab)
	b.Use(publisher)
	mb.Use(publisher)
	mb2.Use(publisher)
	// run the publisher job if Schedule is used
	go publish.RunPublisher(db, nil, publisher)
	// @snippet_end
	return b
}
