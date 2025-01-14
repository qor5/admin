package redirection

import (
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
		db      *gorm.DB
		mb      *presets.ModelBuilder
		pb      *publish.Builder
		storage oss.StorageInterface
	}
)

func New(db *gorm.DB, pb *publish.Builder) *Builder {
	return &Builder{
		db: db,
		pb: pb,
	}
}

func (r *Builder) AutoMigrate() *Builder {
	if err := AutoMigrate(r.db); err != nil {
		panic(err)
	}
	return r
}

func (r *Builder) Install(b *presets.Builder) (err error) {
	r.mb = b.Model(&ObjectRedirection{})
	r.mb.RegisterEventFunc(UploadFileEvent, r.uploadFile)
	listing := r.mb.Listing("Source", "Target")
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
			return in(ctx, params)
		}
	})
	r.pb.WrapStorage(func(v oss.StorageInterface) oss.StorageInterface {
		r.storage = v
		return r
	})
	return
}
