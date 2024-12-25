package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/qor5/admin/v3/email"
	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	mux := http.NewServeMux()
	db := ConnectDB()
	eb := email.ConfigEmailBuilder(db)
	mux.Handle("/email_template/", http.StripPrefix("/email_template", eb))
	fmt.Println("Listen on http://localhost:9800")
	err := http.ListenAndServe(":9800", mux)
	if err != nil {
		panic(err)
	}
}

var dbParamsString = osenv.Get("DB_PARAMS", "email builder example database connection string", "")

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

func LoadSenderConfig() (config email.SESDriverConfig) {
	from := os.Getenv("FROM_ADDRESS")
	if from == "" {
		panic("please set FROM_ADDRESS env")
	}
	return email.SESDriverConfig{
		FromEmailAddress:               from,
		FromName:                       "ciam",
		SubjectCharset:                 "UTF-8",
		HTMLBodyCharset:                "UTF-8",
		TextBodyCharset:                "UTF-8",
		ConfigurationSetName:           "",
		FeedbackForwardingEmailAddress: "",
		FeedbackForwardingEmailAddressIdentityArn: "",
		FromEmailAddressIdentityArn:               "",
		ContactListName:                           "",
		TopicName:                                 "",
	}
}

func LoadReceiverConfig() string {
	to := os.Getenv("TO_ADDRESS")
	if to == "" {
		panic("please set TO_ADDRESS env")
	}
	return to
}
