package emailbuilder

import (
	"os"

	"github.com/theplant/osenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbParamsString = osenv.Get("DB_PARAMS", "email builder example database connection string", "")

var (
	EmailEditorField = "EmailEditor"
	EmailDetailField = "EmailDetail"
)

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

func LoadToEmailAddress() string {
	to := os.Getenv("TO_ADDRESS")
	return to
}

func LoadSenderConfig() (config SESDriverConfig) {
	from := os.Getenv("FROM_ADDRESS")
	return SESDriverConfig{
		FromEmailAddress:               from,
		FromName:                       "",
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
