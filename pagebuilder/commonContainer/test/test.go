package test

import (
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer"
	"gorm.io/gorm"
)

func TestCommonContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	// 使用 commonContainer 包中导出的函数
	commonContainer.HeroImageHorizontal.Register(pb, db)
	commonContainer.HeroImageList.Register(pb, db)
	commonContainer.HeroImageVertical.Register(pb, db)
	commonContainer.Header.Register(pb, db)
	commonContainer.Footer.Register(pb, db)

	// 使用 commonContainer 包中导出的类型
	_ = &commonContainer.HeroImageHorizontal.Hero
	_ = &commonContainer.HeroImageList.TailWindHeroList
	_ = &commonContainer.HeroImageVertical.TailWindHeroVertical
	_ = &commonContainer.Header.TailWindExampleHeader
	_ = &commonContainer.Footer.TailWindExampleFooter
}
