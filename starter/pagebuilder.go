package starter

import (
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/s3x"

	vx "github.com/qor5/x/v3/ui/vuetifyx"
)

var SetupPageBuilderForHandler = []any{
	CreateSEOBuilder,
	CreatePublishStorage,
	CreatePublisher,
	CreatePageBuilder,
}

// CreateSEOBuilder creates and configures the SEO builder
func CreateSEOBuilder(a *Handler, l10nBuilder *l10n.Builder) *seo.Builder {
	b := seo.New(a.DB, seo.WithLocales(l10nBuilder.GetSupportLocaleCodes()...))
	a.Use(b)
	return b
}

// CreatePublishStorage configures S3 storage for publishing
func CreatePublishStorage(a *Handler) oss.StorageInterface {
	a.S3Publish.ACL = string(types.ObjectCannedACLBucketOwnerFullControl)
	return s3x.SetupClient(&a.S3Publish, nil)
}

// CreatePublisher creates and configures the publisher
func CreatePublisher(a *Handler, publishStorage oss.StorageInterface, l10nBuilder *l10n.Builder, activityBuilder *activity.Builder) *publish.Builder {
	publisher := publish.New(a.DB, publishStorage).
		ContextValueFuncs(l10nBuilder.ContextValueProvider).
		Activity(activityBuilder)

	a.Use(publisher)
	return publisher
}

// CreatePageBuilder creates and configures the page builder
func CreatePageBuilder(a *Handler, presetsBuilder *presets.Builder, mediaBuilder *media.Builder, l10nBuilder *l10n.Builder, activityBuilder *activity.Builder, publisher *publish.Builder, seoBuilder *seo.Builder, mux *http.ServeMux) *pagebuilder.Builder {
	pageBuilder := pagebuilder.New("/page_builder", a.DB, presetsBuilder).
		Media(mediaBuilder).
		L10n(l10nBuilder).
		Activity(activityBuilder).
		Publisher(publisher).
		SEO(seoBuilder).
		PreviewContainer(false).
		WrapPageInstall(func(in presets.ModelInstallFunc) presets.ModelInstallFunc {
			return func(pb *presets.Builder, pm *presets.ModelBuilder) (err error) {
				err = in(pb, pm)
				if err != nil {
					return
				}
				pmListing := pm.Listing()
				pmListing.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
					item, err := activityBuilder.MustGetModelBuilder(pm).NewHasUnreadNotesFilterItem(ctx.R.Context(), "")
					if err != nil {
						panic(err)
					}
					return []*vx.FilterItem{item}
				})

				pmListing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
					msgr := i18n.MustGetModuleMessages(ctx.R, I18nDemoKey, Messages_en_US).(*Messages)

					tab, err := activityBuilder.MustGetModelBuilder(pm).NewHasUnreadNotesFilterTab(ctx.R.Context())
					if err != nil {
						panic(err)
					}
					return []*presets.FilterTab{
						{
							Label: msgr.FilterTabsAll,
							ID:    "all",
							Query: url.Values{"all": []string{"1"}},
						},
						tab,
					}
				})
				return nil
			}
		})

	mux.Handle("/page_builder/", pageBuilder)

	a.Use(pageBuilder)
	return pageBuilder
}
