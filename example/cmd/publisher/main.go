package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/example/admin"
	"github.com/qor5/admin/publish"
)

func main() {
	db := admin.ConnectDB()
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Publish_Bucket"),
		Region:  os.Getenv("S3_Publish_Region"),
		ACL:     s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Session: session.Must(session.NewSession()),
	})
	admin.NewConfig()

	publish.RunPublisher(db, storage)

	select {}
}
