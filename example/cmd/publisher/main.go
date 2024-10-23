package main

import (
	"context"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/publish"
)

func main() {
	db := admin.ConnectDB()
	config := admin.NewConfig(db, false)
	storage := admin.PublishStorage
	publish.RunPublisher(context.Background(), db, storage, config.Publisher)
	select {}
}
