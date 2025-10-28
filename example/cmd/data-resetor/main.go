package main

import (
	"github.com/qor5/admin/v3/example/admin"
)

func main() {
	db := admin.ConnectDB()
	tbs := admin.GetNonIgnoredTableNames(db)
	admin.EmptyDB(db, tbs)
	admin.InitDB(db, tbs)
	admin.ErasePublicUsersData(db)
}
