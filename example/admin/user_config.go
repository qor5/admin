package admin

import (
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/example/models"
	"gorm.io/gorm"
)

func configUser(b *presets.Builder, db *gorm.DB) {
	user := b.Model(&models.User{})

	user.Editing(
		"Name",
		"Company",
		"Email",
		"Permission",
		"Status",
	)

}
