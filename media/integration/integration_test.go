package integration_test

import (
	"embed"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	"github.com/qor5/admin/media/oss"
	"github.com/qor5/web/multipartestutils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed *.png
var box embed.FS

func setup() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
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
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
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

	err = media.SaveUploadAndCropImage(db, &m)
	if err != nil {
		t.Fatal(err)
	}

}
