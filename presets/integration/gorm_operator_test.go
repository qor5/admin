package integration_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/theplant/gofixtures"
)

type TestVariant struct {
	ProductCode string
	ColorCode   string
	Name        string
}

var emptyData = gofixtures.Data(gofixtures.Sql(``, []string{"test_variants"}))

func (tv *TestVariant) PrimarySlug() string {
	return fmt.Sprintf("%s_%s", tv.ProductCode, tv.ColorCode)
}

func (*TestVariant) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"product_code": segs[0],
		"color_code":   segs[1],
	}
}

func TestPrimarySlugger(t *testing.T) {
	db := TestDB
	db.AutoMigrate(&TestVariant{})
	rawDB, _ := db.DB()
	emptyData.TruncatePut(rawDB)
	op := gorm2op.DataOperator(db)
	ctx := new(web.EventContext)
	err := op.Save(&TestVariant{ProductCode: "P01", ColorCode: "C01", Name: "Product 1"}, "", ctx)
	if err != nil {
		panic(err)
	}

	err = op.Save(&TestVariant{ProductCode: "P01", ColorCode: "C01", Name: "Product 2"}, "P01_C01", ctx)
	if err != nil {
		panic(err)
	}

	tv, err := op.Fetch(&TestVariant{}, "P01_C01", ctx)
	if err != nil {
		panic(err)
	}

	if tv.(*TestVariant).Name != "Product 2" {
		t.Error("didn't update product 2", tv)
	}

	err = op.Delete(&TestVariant{}, "P01_C01", ctx)
	if err != nil {
		panic(err)
	}

	tv, err = op.Fetch(&TestVariant{}, "P01_C01", ctx)
	if err != presets.ErrRecordNotFound {
		t.Error("didn't return not found after delete", tv, err)
	}
}
