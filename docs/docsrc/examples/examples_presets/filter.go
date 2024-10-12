package examples_presets

// @snippet_begin(FilterSample)
import (
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint
	Title     string
	Body      string
	UpdatedAt time.Time
	CreatedAt time.Time
	Disabled  bool

	Status string

	CategoryID uint
}

func PresetsBasicFilter(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	err := db.AutoMigrate(&Post{})
	if err != nil {
		panic(err)
	}
	// create a ModelBuilder
	postBuilder := b.Model(&Post{})

	// get its ListingBuilder
	listing := postBuilder.Listing()

	// Call FilterDataFunc
	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		// Prepare filter options, it is a two dimension array: [][]string{"text", "value"}
		options := []*vuetifyx.SelectItem{
			{Text: "Draft", Value: publish.StatusDraft},
			{Text: "Online", Value: publish.StatusOnline},
		}

		return []*vuetifyx.FilterItem{
			{
				Key:      "status",
				Label:    "Status",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `status %s ?`,
				Options:      options,
			},
			{
				Key:                   "title",
				Label:                 "titleNoChoose",
				ItemType:              vuetifyx.ItemTypeString,
				Modifier:              vuetifyx.ModifierEquals, // default modifier is contains
				SQLCondition:          `title = ?`,
				DisableChooseModifier: true,
			},
			{
				Key:          "warpBody",
				Label:        "warpBody",
				ItemType:     vuetifyx.ItemTypeString,
				Modifier:     vuetifyx.ModifierEquals,
				SQLCondition: `body = ?`,
				Options:      options,
				WrapInput: func(val string) interface{} {
					return val + "wrap"
				},
			},
		}
	})
	return
}

// @snippet_end
