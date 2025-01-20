package redirection

import (
	"net/url"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/oss/s3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
)

type (
	Builder struct {
		s3Client  *s3.Client
		baseURL   *url.URL
		db        *gorm.DB
		mb        *presets.ModelBuilder
		publisher *publish.Builder
		storage   oss.StorageInterface
	}
)

func New(s3Client *s3.Client, db *gorm.DB, publisher *publish.Builder) *Builder {
	return &Builder{
		s3Client:  s3Client,
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
	if b.s3Client == nil || b.s3Client.Config.Endpoint == "" {
		return
	}
	if b.baseURL, err = url.Parse(b.s3Client.Config.Endpoint); err != nil {
		return
	}
	pb.GetI18n().
		RegisterForModule(language.English, I18nRedirectionKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nRedirectionKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nRedirectionKey, Messages_ja_JP)

	m := &Redirection{}
	b.mb = pb.Model(m)
	b.mb.RegisterEventFunc(UploadFileEvent, b.uploadFile)
	listing := b.mb.Listing("Source", "Target")
	listing.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		cell.SetAttr("@click", "")
		return cell
	})
	listing.RowMenu().Empty()
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return web.Scope(
			vx.VXBtn("UploadFile").
				Attr(":loading", "xLocals.loading").
				PrependIcon("mdi-upload").Color(v.ColorPrimary).
				Attr("@click", "$refs.uploadInput.click()"),
			h.Input("").
				Attr("ref", "uploadInput").
				Attr("accept", ".csv").
				Type("file").
				Style("display:none").
				Attr("@change",
					"form.NewFiles = [...$event.target.files];"+
						web.Plaid().
							BeforeScript("xLocals.loading=true").
							ThenScript("$refs.uploadInput.value=null;xLocals.loading=false").
							EventFunc(UploadFileEvent).
							Go()),
		).VSlot("{locals:xLocals}").Init("{loading:false}")
	})
	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			params.OrderBys = append(params.OrderBys, relay.OrderBy{
				Field: "CreatedAt",
				Desc:  true,
			})
			//		params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
			//			Query: fmt.Sprintf(`%s.created_at=(SELECT MAX(created_at)
			//FROM %s tb
			//WHERE source = %s.source)`, m.TableName(), m.TableName(), m.TableName()),
			//		})
			return in(ctx, params)
		}
	})
	b.publisher.WrapStorage(func(v oss.StorageInterface) oss.StorageInterface {
		b.storage = v
		return b
	})
	return
}
