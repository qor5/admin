package main

import (
	"fmt"
	"os"

	"github.com/qor5/admin/example/admin"
	"github.com/qor5/admin/media/media_library"
	"gorm.io/gorm"
)

func main() {
	db := admin.ConnectDB()

	ignoredTableNames := map[string]struct{}{
		"users":          {},
		"roles":          {},
		"user_role_join": {},
		"login_sessions": {},
	}

	var rawTableNames []string
	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public';").Scan(&rawTableNames).
		Error; err != nil {
		panic(err)
	}

	var tableNames []string
	for _, n := range rawTableNames {
		if _, ok := ignoredTableNames[n]; !ok {
			tableNames = append(tableNames, n)
		}
	}

	emptyDB(db, tableNames)
	initDB(db, tableNames)

	return
}

func emptyDB(db *gorm.DB, tables []string) {
	for _, name := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", name)).Error; err != nil {
			panic(err)
		}
	}
}

func initDB(db *gorm.DB, tables []string) {
	var err error
	// Users
	admin.GenInitialUser()
	// Page Builder
	if err = db.Exec(initPageBuilderSQL).Error; err != nil {
		panic(err)
	}
	// Orders
	if err = db.Exec(initOrdersSQL).Error; err != nil {
		panic(err)
	}
	// Workers
	if err = db.Exec(initWorkersSQL).Error; err != nil {
		panic(err)
	}
	// Categories
	if err = db.Exec(initCategoriesSQL).Error; err != nil {
		panic(err)
	}
	// InputDemos
	if err = db.Exec(initInputDemosSQL).Error; err != nil {
		panic(err)
	}
	// Posts
	if err = db.Exec(initPostsSQL).Error; err != nil {
		panic(err)
	}
	// NestedFieldDemos
	if err = db.Exec(initNestedFieldDemosSQL).Error; err != nil {
		panic(err)
	}
	// ListModels
	if err = db.Exec(initListModelsSQL).Error; err != nil {
		panic(err)
	}
	// MicrositeModels
	if err = db.Exec(initMicrositeModelsSQL).Error; err != nil {
		panic(err)
	}
	// Products
	if err = db.Exec(initProductsSQL).Error; err != nil {
		panic(err)
	}
	// Media Library
	if err = db.Model(&media_library.MediaLibrary{}).Create(&[]map[string]interface{}{
		{"id": 1, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"aigle.png","Url":"%s","Width":320,"Height":84,"FileSizes":{"@qor_preview":17065,"default":3159,"original":3159},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/1/file.png"))},
		{"id": 2, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"asics.png","Url":"%s","Width":254,"Height":84,"FileSizes":{"@qor_preview":15571,"default":3060,"original":3060},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/2/file.png"))},
		{"id": 3, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20210903061739.png","Url":"%s","Width":1722,"Height":196,"FileSizes":{"@qor_preview":627,"default":6887,"original":6887},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/3/file.png"))},
		{"id": 4, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211006224452.jpg","Url":"%s","Width":2880,"Height":720,"FileSizes":{"@qor_preview":19981,"default":257343,"original":257343},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/4/file.jpg"))},
		{"id": 5, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007041906.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":79999,"default":234306,"original":234306},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/5/file.png"))},
		{"id": 6, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007042027.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":65623,"default":203098,"original":203098},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/6/file.png"))},
		{"id": 7, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007042131.png","Url":"%s","Width":481,"Height":741,"FileSizes":{"@qor_preview":64838,"default":189979,"original":189979},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/7/file.png"))},
		{"id": 8, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.20211007051449.png","Url":"%s","Width":2880,"Height":1097,"FileSizes":{"@qor_preview":75734,"default":2236473,"original":2236473},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/8/file.png"))},
		{"id": 9, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"file.png","Url":"%s","Width":1252,"Height":658,"FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/9/file.png"))},
		{"id": 10, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"lacoste.png","Url":"%s","Width":470,"Height":84,"FileSizes":{"@qor_preview":11839,"default":4714,"original":4714},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/10/file.png"))},
		{"id": 11, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"mob-mv.mov","Url":"%s","Video":"","SelectedType":"","Description":""}`, composeS3Path("/11/file.mov"))},
		{"id": 12, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"mob.jpg","Url":"%s","Width":1536,"Height":2876,"FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/12/file.jpg"))},
		{"id": 13, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"nhk.png","Url":"%s","Width":202,"Height":84,"FileSizes":{"@qor_preview":14500,"default":2066,"original":2066},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/13/file.png"))},
		{"id": 14, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"pc-mv.mov","Url":"%s","Video":"","SelectedType":"","Description":""}`, composeS3Path("/14/file.mov"))},
		{"id": 15, "selected_type": "image", "file": fmt.Sprintf(`{"FileName":"pc.jpg","Url":"%s","Width":2560,"Height":1440,"FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Video":"","SelectedType":"","Description":""}`, composeS3Path("/15/file.jpg"))},
	}).Error; err != nil {
		panic(err)
	}
	// Seq
	for _, name := range tables {
		if err := db.Exec(fmt.Sprintf("SELECT setval('%s_id_seq', (SELECT MAX(id) FROM %s));", name, name)).Error; err != nil {
			panic(err)
		}
	}
}

// composeS3Path to generate file path as https://cdn.qor5.theplant-dev.com/system/media_libraries/236/file.jpeg
func composeS3Path(filePath string) string {
	endPoint := os.Getenv("S3_Endpoint")
	return fmt.Sprintf("%s/system/media_libraries%s", endPoint, filePath)
}
