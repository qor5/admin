package commonContainer

import (
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/footer"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/header"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/heroImageHorizontal"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/heroImageList"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/heroImageVertical"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&heroImageHorizontal.Hero{},
		&heroImageList.TailWindHeroList{},
		&heroImageVertical.TailWindHeroVertical{},
		&header.TailWindExampleHeader{},
		&footer.TailWindExampleFooter{})
}
func New(db *gorm.DB, b *presets.Builder, prefix string, layout pagebuilder.PageLayoutFunc) *pagebuilder.Builder {
	pb := pagebuilder.New(prefix, db, b)
	if layout != nil {
		pb.PageLayout(pagebuilder.WrapDefaultPageLayout(layout))
	}
	pb.GetPresetsBuilder().ExtraAsset("/tiptap.css", "text/css", tiptap.ThemeGithubCSSComponentsPack())
	footer.RegisterContainer(pb, db)
	header.RegisterContainer(pb, db)
	heroImageHorizontal.RegisterContainer(pb, db)
	heroImageVertical.RegisterContainer(pb, db)
	heroImageList.RegisterContainer(pb, db)
	return pb
}
