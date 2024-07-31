package examples_admin

import (
	"context"
	"net/http"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func ActivityExample(b *presets.Builder, db *gorm.DB) http.Handler {
	b.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)

	// @snippet_begin(NewActivitySample)
	b.DataOperator(gorm2op.DataOperator(db))

	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:     "1",
			Name:   "John",
			Avatar: "https://i.pravatar.cc/300",
		}, nil
	}).
		// TablePrefix("cms_"). // multitentant if needed
		AutoMigrate()
	b.Use(ab)

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

	mb := b.Model(&WithActivityProduct{})
	defer func() { ab.RegisterModel(mb) }()
	mb.Listing("Title", activity.ListFieldNotes, "Code", "Price")
	dp := mb.Detailing("Content").Drawer(true)
	dp.Section("Content").Editing("Title", "Code", "Price")
	dp.SidePanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return ab.MustGetModelBuilder(mb).NewTimelineCompo(ctx, obj, "_side")
	})
	// @snippet_end

	return b
}
