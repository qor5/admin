package redirection

import (
	"net/url"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
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

const redirection_notify_error_msg = "redirection_notify_error_msg"

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
	b.mb = pb.Model(m).MenuIcon("mdi-link")
	b.mb.RegisterEventFunc(UploadFileEvent, b.uploadFile)
	listing := b.mb.Listing("Source", "Target")
	listing.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		cell.SetAttr("@click", "")
		return cell
	})
	listing.RowMenu().Empty()
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nRedirectionKey, Messages_en_US).(*Messages)

		return web.Scope(
			web.Listen(redirection_notify_error_msg, `xLocals.text = payload;xLocals.dialog=true;`),
			vx.VXDialog(
				h.P(h.Text("{{xLocals.text}}")).Style("white-space: pre-line;"),
			).Title(msgr.ErrorTips).HideFooter(true).Type(vx.DialogError).Attr("v-model", "xLocals.dialog"),
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
		).VSlot("{locals:xLocals}").Init(`{loading:false,dialog:false,text:""}`)
	})
	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			params.OrderBy = append(params.OrderBy, relay.Order{
				Field:     "CreatedAt",
				Direction: relay.OrderDirectionDesc,
			})
			return in(ctx, params)
		}
	})
	if b.publisher == nil {
		return
	}
	b.publisher.WrapStorage(func(v oss.StorageInterface) oss.StorageInterface {
		b.storage = v
		return b
	})
	return
}
