package main

import (
	"github.com/qor5/admin/example/admin"
	"github.com/qor5/admin/publish"
)

func main() {
	db := admin.ConnectDB()
	config := admin.NewConfig()
	storage := admin.PublishStorage
	publish.RunPublisher(db, storage, config.Publisher)
	select {}
}
