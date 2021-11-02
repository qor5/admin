package main

import (
	"fmt"
	"os"
	"time"

	"github.com/qor/qor5/exchange"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Variant struct {
	gorm.Model
	Code         string `gorm:"uniqueIndex;not null;"`
	ProductID    uint
	ProductCode  string
	Price        uint64
	SellingPrice uint64

	Properties []*VariantPropertyShip
}

type VariantPropertyShip struct {
	ID                uint `gorm:"primary_key"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	VariantID         uint
	VariantPropertyID uint
	Value             string `gorm:"size:512"`
}

type VariantProperty struct {
	gorm.Model
	Name string
}

type Product struct {
	ID   uint `gorm:"primarykey"`
	Code string
}

func main() {
	var err error
	db, err := gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	db.Logger = db.Logger.LogMode(logger.Info)
	if err = db.AutoMigrate(&Variant{}, &VariantPropertyShip{}); err != nil {
		panic(err)
	}

	doImport := false
	doExport := false
	switch os.Args[1] {
	// import
	case "1":
		doImport = true
		// export
	case "2":
		doExport = true
	default:
		panic("add import/export param")
	}

	properties := []*VariantProperty{}
	if err = db.Find(&properties).Error; err != nil {
		panic(err)
	}
	propertiesNameMap := make(map[string]*VariantProperty)
	for i, _ := range properties {
		p := properties[i]
		propertiesNameMap[p.Name] = p
	}

	products := []*Product{}
	if err = db.Find(&products).Error; err != nil {
		panic(err)
	}
	productsCodeMap := make(map[string]*Product)
	for i, _ := range products {
		p := products[i]
		productsCodeMap[p.Code] = p
	}

	associations := []string{"Properties"}
	metas := []*exchange.Meta{
		exchange.NewMeta("Code").Header("JANコード").PrimaryKey(true),
		exchange.NewMeta("ProductCode").Header("品番").Setter(func(record interface{}, value string, metaValues exchange.MetaValues) error {
			r := record.(*Variant)
			r.ProductCode = value
			if r.ProductID == 0 {
				product, ok := productsCodeMap[value]
				if ok {
					r.ProductID = product.ID
				}
			}
			return nil
		}),
		exchange.NewMeta("Price").Header("上代1"),
		exchange.NewMeta("SellingPrice").Header("上代2"),
		exchange.NewMeta("ColorCode").Header("カラー").
			Setter(propertyMetaSetter(propertiesNameMap["ColorCode"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["ColorCode"].ID)),
		exchange.NewMeta("SizeCode").Header("サイズ").
			Setter(propertyMetaSetter(propertiesNameMap["SizeCode"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["SizeCode"].ID)),
		exchange.NewMeta("SeasonName").Header("シーズン").
			Setter(propertyMetaSetter(propertiesNameMap["SeasonName"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["SeasonName"].ID)),
		exchange.NewMeta("FranceColorCode").Header("仏カラー").
			Setter(propertyMetaSetter(propertiesNameMap["FranceColorCode"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["FranceColorCode"].ID)),
		exchange.NewMeta("SaleCornerOnly").Header("SaleCornerOnly").
			Setter(propertyMetaSetter(propertiesNameMap["SaleCornerOnly"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["SaleCornerOnly"].ID)),
		exchange.NewMeta("Badge").Header("Badge").
			Setter(propertyMetaSetter(propertiesNameMap["Badge"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["Badge"].ID)),
		exchange.NewMeta("AllowShipFromStore").Header("AllowShipFromStore").
			Setter(propertyMetaSetter(propertiesNameMap["AllowShipFromStore"].ID)).
			Valuer(propertyMetaValuer(propertiesNameMap["AllowShipFromStore"].ID)),
	}

	if doImport {
		importer := exchange.NewImporter(&Variant{}).Associations(associations...).Metas(metas...)

		f, err := os.Open(os.Args[2])
		if err != nil {
			panic(err)
		}
		defer f.Close()
		r, err := exchange.NewCSVReader(f)
		if err != nil {
			panic(err)
		}
		err = importer.Exec(db, r)
		if err != nil {
			panic(err)
		}
	}
	if doExport {
		exporter := exchange.NewExporter(&Variant{}).Associations(associations...).Metas(metas...)
		f, err := os.Create(fmt.Sprintf("variant-e%s.csv", time.Now().Format("200601021504")))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w, err := exchange.NewCSVWriter(f)
		if err != nil {
			panic(err)
		}
		err = exporter.Exec(db, w)
		if err != nil {
			panic(err)
		}
	}
}

func propertyMetaSetter(propertyID uint) exchange.MetaSetter {
	return func(record interface{}, value string, metaValues exchange.MetaValues) error {
		r := record.(*Variant)
		has := false
		for _, p := range r.Properties {
			if p.VariantPropertyID == propertyID {
				has = true
				p.Value = value
				break
			}
		}
		if !has {
			r.Properties = append(r.Properties, &VariantPropertyShip{
				VariantPropertyID: propertyID,
				Value:             value,
			})
		}
		return nil
	}
}

func propertyMetaValuer(propertyID uint) exchange.MetaValuer {
	return func(record interface{}) (string, error) {
		r := record.(*Variant)
		for _, p := range r.Properties {
			if p.VariantPropertyID == propertyID {
				return p.Value, nil
			}
		}
		return "", nil
	}
}
