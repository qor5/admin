package admin

import (
	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/presets"
)

func configL10nModel(b *presets.Builder) (*presets.ModelBuilder, *presets.ModelBuilder) {
	if err := db.AutoMigrate(
		&models.L10nModel{},
		&models.L10nModelWithVersion{},
	); err != nil {
		panic(err)
	}
	l10nM := b.Model(&models.L10nModel{}).Label("L10n Models")
	l10nM.Listing("Title", "Locale")
	l10nVM := b.Model(&models.L10nModelWithVersion{}).Label("L10n Models With Versions")
	l10nVM.Listing("Title", "Locale", "Status", "Draft Count", "Online")

	return l10nM, l10nVM
}
