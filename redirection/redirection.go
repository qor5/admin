package redirection

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
)

type (
	Builder struct {
		db        *gorm.DB
		mb        *presets.ModelBuilder
		publisher *publish.Builder
		storage   oss.StorageInterface
	}
)

func New(db *gorm.DB, publisher *publish.Builder) *Builder {
	return &Builder{
		db:        db,
		publisher: publisher,
	}
}

func (b *Builder) AutoMigrate() *Builder {
	if err := AutoMigrate(b.db); err != nil {
		panic(err)
	}
	return b
}

func (b *Builder) Install(pb *presets.Builder) (err error) {
	m := &ObjectRedirection{}
	b.mb = pb.Model(m)
	b.mb.RegisterEventFunc(UploadFileEvent, b.uploadFile)
	listing := b.mb.Listing("Source", "Target")
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return h.Div(
			vx.VXBtn("UploadFile").PrependIcon("mdi-upload").Color(v.ColorPrimary).
				Attr("@click", "$refs.uploadInput.click()"),
			h.Input("").
				Attr("ref", "uploadInput").
				Attr("accept", ".csv").
				Type("file").
				Style("display:none").
				Attr("@change",
					"form.NewFiles = [...$event.target.files];"+
						web.Plaid().
							EventFunc(UploadFileEvent).
							Go()),
		)

	})
	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			params.OrderBys = append(params.OrderBys, relay.OrderBy{
				Field: "CreatedAt",
				Desc:  true,
			})
			params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
				Query: fmt.Sprintf(`%s.created_at=(SELECT MAX(created_at)
    FROM %s tb
    WHERE source = %s.source)`, m.TableName(), m.TableName(), m.TableName()),
			})
			return in(ctx, params)
		}
	})
	b.publisher.WrapStorage(func(v oss.StorageInterface) oss.StorageInterface {
		b.storage = v
		return b
	})
	return
}
