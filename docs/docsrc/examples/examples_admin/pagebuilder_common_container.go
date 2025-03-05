package examples_admin

import (
	"context"
	"net/http"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/qor5/x/v3/perm"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
)

func PageBuilderCommonContainerExample(b *presets.Builder, db *gorm.DB) http.Handler {
	b.DataOperator(gorm2op.DataOperator(db))
	storage := filesystem.New("/tmp/publish")
	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:     "1",
			Name:   "John",
			Avatar: "https://i.pravatar.cc/300",
		}, nil
	}).AutoMigrate()

	puBuilder := publish.New(db, storage)
	if b.GetPermission() == nil {
		b.Permission(
			perm.New().Policies(
				perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
			),
		)
	}
	b.Use(puBuilder)

	pb := commonContainer.New(db, b, b.GetURIPrefix()+"/page_builder", func(body h.HTMLComponent, input *pagebuilder.PageLayoutInput, ctx *web.EventContext) h.HTMLComponent {
		var freeStyleCss h.HTMLComponent
		if len(input.FreeStyleCss) > 0 {
			freeStyleCss = h.Style(strings.Join(input.FreeStyleCss, "\n"))
		}

		head := h.Components(
			input.SeoTags,
			input.CanonicalLink,
			h.Meta().Attr("http-equiv", "X-UA-Compatible").Content("IE=edge"),
			h.Meta().Content("true").Name("HandheldFriendly"),
			h.Meta().Content("yes").Name("apple-mobile-web-app-capable"),
			h.Meta().Content("black").Name("apple-mobile-web-app-status-bar-style"),
			h.Meta().Name("format-detection").Content("telephone=no"),
			h.If(len(input.EditorCss) > 0, input.EditorCss...),
			freeStyleCss,
			input.StructuredData,
			pagebuilder.ScriptWithCodes(input.FreeStyleTopJs),
		)
		ctx.Injector.HTMLLang(input.LocaleCode)
		if input.WrapHead != nil {
			head = input.WrapHead(head)
		}
		ctx.Injector.HeadHTML(h.MustString(head, nil))
		bodies := h.Components(
			h.If(input.Header != nil, input.Header),
			body,
			h.If(input.Footer != nil, input.Footer),
			pagebuilder.ScriptWithCodes(input.FreeStyleBottomJs),
		)
		if input.WrapBody != nil {
			bodies = input.WrapBody(bodies)
		}
		return h.Body(
			bodies,
		)
	}).
		AutoMigrate().
		Activity(ab).
		PreviewOpenNewTab(true).
		Publisher(puBuilder)
	err := commonContainer.AutoMigrate(db)
	if err != nil {
		panic(err)
	}
	// use demo container and media etc. plugins
	b.Use(pb)
	return TestHandler(pb, b)
}
