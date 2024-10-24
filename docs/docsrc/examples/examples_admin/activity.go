package examples_admin

import (
	"context"
	"net/http"
	"time"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func ActivityExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return activityExample(b, db, func(mb *presets.ModelBuilder, ab *activity.Builder) {
		b.Use(ab)
	})
}

func activityExample(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder, ab *activity.Builder)) http.Handler {
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

	// @snippet_end

	// @snippet_begin(ActivityRegisterPresetsModelsSample)
	type WithActivityProduct struct {
		gorm.Model
		Title     string
		Code      string
		Approved  bool
		Edited    bool
		Price     float64
		StockedAt time.Time
		AppovedAt *time.Time
	}

	err := db.AutoMigrate(&WithActivityProduct{})
	if err != nil {
		panic(err)
	}

	mb := b.Model(&WithActivityProduct{})
	mb.Listing("Title", activity.ListFieldNotes, "Code", "Price", "StockedAt", "AppovedAt")
	dp := mb.Detailing("Content").Drawer(true)
	contentSection := presets.NewSectionBuilder(mb, "Content").Editing("Title", "Code", "Approved", "Edited", "Price", "StockedAt", "AppovedAt")
	dp.Section(contentSection)
	dp.SidePanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return ab.MustGetModelBuilder(mb).NewTimelineCompo(ctx, obj, "_side")
	})

	ab.RegisterModel(mb)
	// @snippet_end

	if customize != nil {
		customize(mb, ab)
	}
	return b
}
