package examples_admin

import (
	"context"
	"net/http"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func ActivityExample(b *presets.Builder, db *gorm.DB) http.Handler {
	// @snippet_begin(NewActivitySample)
	b.DataOperator(gorm2op.DataOperator(db))

	activityBuilder := activity.New(db)
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
	productModel := b.Model(&WithActivityProduct{}).Use(activityBuilder)

	bt := productModel.Detailing("Content", activity.Timeline).Drawer(true)
	bt.Section("Content").
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			return Div().Text("text")
		}).Editing("Title", "Code", "Price")

	activityBuilder.RegisterModel(productModel).EnableActivityInfoTab().AddKeys("Title").AddIgnoredFields("Code").SkipDelete()
	// @snippet_end

	// @snippet_begin(ActivityRecordLogSample)
	currentCtx := context.WithValue(context.Background(), activity.CreatorContextKey, "user1")

	activityBuilder.AddRecords("Publish", currentCtx, &WithActivityProduct{Title: "Product 1", Code: "P1", Price: 100})

	activityBuilder.AddRecords("Update Price", currentCtx, &WithActivityProduct{Title: "Product 1", Code: "P1", Price: 200})
	// @snippet_end
	return b
}
