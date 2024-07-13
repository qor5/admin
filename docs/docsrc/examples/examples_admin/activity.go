package examples_admin

import (
	"context"
	"net/http"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"gorm.io/gorm"
)

func ActivityExample(b *presets.Builder, db *gorm.DB) http.Handler {
	// @snippet_begin(NewActivitySample)
	b.DataOperator(gorm2op.DataOperator(db))

	activityBuilder := activity.New(db, func(ctx context.Context) *activity.User {
		return &activity.User{
			ID:     "1",
			Name:   "John",
			Avatar: "https://i.pravatar.cc/300",
		}
	}).AutoMigrate()
	b.Use(activityBuilder)

	// @snippet_end

	// @snippet_begin(ActivityRegisterPresetsModelsSample)
	type WithActivityProduct struct {
		gorm.Model
		Title string
		Code  string
		Price float64
	}

	err := db.AutoMigrate(&WithActivityProduct{})
	if err != nil {
		panic(err)
	}

	productModel := b.Model(&WithActivityProduct{})

	bt := productModel.Detailing("Content", activity.DetailFieldTimeline).Drawer(true)
	bt.Section("Content").Editing("Title", "Code", "Price")

	productModel.Use(activityBuilder)
	activityBuilder.MustGetModelBuilder(productModel).SkipDelete()

	// @snippet_end

	// @snippet_begin(ActivityRecordLogSample)

	ctx := context.Background()
	activityBuilder.Log(ctx, "Publish", &WithActivityProduct{Title: "Product 1", Code: "P1", Price: 100}, nil)
	activityBuilder.Log(ctx, "Update Price", &WithActivityProduct{Title: "Product 1", Code: "P1", Price: 200}, nil)

	// @snippet_end
	return b
}
