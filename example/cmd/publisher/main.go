package main

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/example/admin"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/publish"
)

func main() {
	db := admin.ConnectDB()
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: session.Must(session.NewSession()),
	})

	listP := publish.NewListBuilder(db, storage)
	go func() {
		t := time.Tick(time.Second * 1)
		for range t {
			listP.PublishList(models.ListModel{}, []*models.ListModel{})
		}
	}()
	select {}
}
