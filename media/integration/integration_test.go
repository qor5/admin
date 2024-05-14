package integration_test

import (
	"embed"
	"testing"

	"github.com/qor5/admin/v3/media/base"
	"github.com/theplant/osenv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/web/v3/multipartestutils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed *.png
var box embed.FS

var (
	testDBParams = osenv.Get("TEST_DB_PARAMS", "test database connection string", "user=test password=test dbname=test sslmode=disable host=localhost port=6432 TimeZone=Asia/Tokyo")
	s3Bucket     = osenv.Get("S3_Bucket", "s3-bucket for media library storage", "example")
	s3Region     = osenv.Get("S3_Region", "s3-region for media library storage", "ap-northeast-1")
)

func setup() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(testDBParams), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db = db.Debug()
	// db.Logger = db.Logger.LogMode(logger.Info)

	if err = db.AutoMigrate(
		&media_library.MediaLibrary{},
	); err != nil {
		panic(err)
	}

	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  s3Bucket,
		Region:  s3Region,
		ACL:     s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Session: sess,
	})

	return
}

func TestUpload(t *testing.T) {
	db := setup()
	f, err := box.ReadFile("testfile.png")
	if err != nil {
		panic(err)
	}

	fh := multipartestutils.CreateMultipartFileHeader("test.png", f)
	m := media_library.MediaLibrary{}

	err = m.File.Scan(fh)
	if err != nil {
		t.Fatal(err)
	}

	err = base.SaveUploadAndCropImage(db, &m)
	if err != nil {
		t.Fatal(err)
	}
}
