package main

import (
	"fmt"

	"github.com/qor5/admin/example/admin"
	"github.com/qor5/admin/media/media_library"
	"gorm.io/gorm"
)

func main() {
	db := admin.ConnectDB()
	emptyDB(db)
	initDB(db)

	return
}

func emptyDB(db *gorm.DB) {
	var tableNames []string

	if err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public';").Scan(&tableNames).
		Error; err != nil {
		panic(err)
	}

	for _, name := range tableNames {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", name)).Error; err != nil {
			panic(err)
		}
	}
}

func initDB(db *gorm.DB) {
	admin.GenInitialPasswordUser()

	var err error
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
	// Roles
	if err = db.Exec(initRolesSQL).Error; err != nil {
		panic(err)
	}
	// Media Library
	if err = db.Model(&media_library.MediaLibrary{}).Create(&[]map[string]interface{}{
		{"selected_type": "image", "file": `{"FileName":"mob.jpg","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/100/file.jpg","Width":1536,"Height":2876,"FileSizes":{"@qor_preview":33140,"default":465542,"original":465542},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"pc.jpg","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/101/file.jpg","Width":2560,"Height":1440,"FileSizes":{"@qor_preview":33019,"default":646542,"original":646542},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"image.png","Url":"//the-plant.com/system/media_libraries/283/file.20211006224452.jpg","Width":2880,"Height":720,"FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"nhk.png","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/105/file.png","Width":202,"Height":84,"FileSizes":{"@qor_preview":14500,"default":2066,"original":2066},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"aigle.png","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/106/file.png","Width":320,"Height":84,"FileSizes":{"@qor_preview":17065,"default":3159,"original":3159},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"lacoste.png","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/107/file.png","Width":470,"Height":84,"FileSizes":{"@qor_preview":11839,"default":4714,"original":4714},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"asics.png","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/108/file.png","Width":254,"Height":84,"FileSizes":{"@qor_preview":15571,"default":3060,"original":3060},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "image", "file": `{"FileName":"image.png","Url":"//qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/109/file.png","Width":1252,"Height":658,"FileSizes":{"@qor_preview":41622,"default":227103,"original":227103},"Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "video", "file": `{"FileName":"theplant-mob.mov","Url":"","Video":"","SelectedType":"","Description":""}`},
		{"selected_type": "video", "file": `{"FileName":"theplant.mov","Url":"","Video":"","SelectedType":"","Description":""}`},
	}).Error; err != nil {
		panic(err)
	}
}
