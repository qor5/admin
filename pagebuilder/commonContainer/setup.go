package commonContainer

import (
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/footer"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/header"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/heroImageList"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/heroImageVertical"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/imageWithText"
	"github.com/qor5/admin/v3/tiptap"
)

const (
	HeroImageHorizontal   ContainerType = "heroImageHorizontal"
	TailWindHeroList      ContainerType = "tailWindHeroList"
	TailWindHeroVertical  ContainerType = "tailWindHeroVertical"
	TailWindExampleHeader ContainerType = "tailWindExampleHeader"
	TailWindExampleFooter ContainerType = "tailWindExampleFooter"
	ImageWithText         ContainerType = "imageWithText"
)

type (
	ContainerType     string
	containerRegister struct {
		ContainerType ContainerType
		Register      func(*pagebuilder.Builder, *gorm.DB)
		Model         interface{}
	}
)

var (
	allContainerType = []ContainerType{HeroImageHorizontal, TailWindHeroList, TailWindHeroVertical, TailWindExampleHeader, TailWindExampleFooter}

	register = []containerRegister{
		{
			ContainerType: TailWindHeroList,
			Register:      heroImageList.RegisterContainer,
			Model:         &heroImageList.TailWindHeroList{},
		},
		{
			ContainerType: TailWindHeroVertical,
			Register:      heroImageVertical.RegisterContainer,
			Model:         &heroImageVertical.TailWindHeroVertical{},
		},
		{
			ContainerType: TailWindExampleHeader,
			Register:      header.RegisterContainer,
			Model:         &header.TailWindExampleHeader{},
		},
		{
			ContainerType: TailWindExampleFooter,
			Register:      footer.RegisterContainer,
			Model:         &footer.TailWindExampleFooter{},
		},
		{
			ContainerType: ImageWithText,
			Register:      imageWithText.RegisterContainer,
			Model:         &imageWithText.ImageWithText{},
		},
	}
)

func autoMigrate(db *gorm.DB, ct ...ContainerType) error {
	var models []interface{}
	for _, containerType := range ct {
		for _, r := range register {
			if r.ContainerType == containerType {
				models = append(models, r.Model)
				break
			}
		}
	}

	return db.AutoMigrate(models...)
}

func Setup(pb *pagebuilder.Builder, db *gorm.DB, layout pagebuilder.PageLayoutFunc, ct ...ContainerType) *pagebuilder.Builder {
	if layout != nil {
		pb.PageLayout(pagebuilder.WrapDefaultPageLayout(layout))
	}
	pb.GetPresetsBuilder().ExtraAsset("/tiptap.css", "text/css", tiptap.ThemeGithubCSSComponentsPack())
	if err := autoMigrate(db, ct...); err != nil {
		panic(err)
	}
	for _, containerType := range ct {
		for _, r := range register {
			if r.ContainerType == containerType {
				r.Register(pb, db)
				break
			}
		}
	}

	return pb
}
