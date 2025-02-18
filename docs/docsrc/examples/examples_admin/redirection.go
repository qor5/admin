package examples_admin

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/qor5/x/v3/oss/s3"
	"github.com/theplant/osenv"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	microsite_utils "github.com/qor5/admin/v3/microsite/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/redirection"
)

var (
	s3PublishBucket = osenv.Get("S3_Publish_Bucket", "s3-bucket for publish", "example-publish")
	s3PublishRegion = osenv.Get("S3_Publish_Region", "s3-region for publish", "ap-northeast-1")
	publishURL      = osenv.Get("PUBLISH_URL", "publish url", "https://www.xxxxx.com")
)

func RedirectionExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return redirectionExample(b, db, func(rb *redirection.Builder) {
		b.Use(rb)
	})
}

func redirectionExample(b *presets.Builder, db *gorm.DB, customize func(rb *redirection.Builder)) http.Handler {
	b.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)

	// @snippet_begin(NewRedirectionSample)
	b.DataOperator(gorm2op.DataOperator(db))

	s3Client := s3.New(&s3.Config{
		Bucket:   s3PublishBucket,
		Region:   s3PublishRegion,
		ACL:      string(types.ObjectCannedACLBucketOwnerFullControl),
		Endpoint: publishURL,
	})
	publisher := publish.New(db, microsite_utils.NewClient(s3Client))
	rb := redirection.New(s3Client, db, publisher).AutoMigrate()
	// @snippet_end
	if customize != nil {
		customize(rb)
	}
	return b
}
