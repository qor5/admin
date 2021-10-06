package example

import (
	"log"
	"net/http"
	"os"

	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/pagebuilder"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TextAndImage struct {
	Text  string
	Image media_library.MediaBox
}

func main() {
	db := connectDB()

	pb := pagebuilder.New(db)

	textAndImage := pb.NewContainer("text_and_image")
	textAndImage.Model(&TextAndImage{}).Editing("Text", "Image")

	log.Println("Listen on http://localhost:9600")
	log.Fatal(http.ListenAndServe(":9600", pb))
}

func connectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}
