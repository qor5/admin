package main

import (
	"github.com/qor5/admin/example/admin"
)

func main() {
	db := admin.ConnectDB()
	tbs := admin.GetNonIgnoredTableNames()
	admin.EmptyDB(db, tbs)
	admin.InitDB(db, tbs)
	return
}
