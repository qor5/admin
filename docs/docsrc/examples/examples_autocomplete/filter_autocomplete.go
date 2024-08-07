package examples_autocompelte

// @snippet_begin(PresetsAutoCompleteBasicFilter)
import (
	"github.com/qor5/admin/v3/autocomplete"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"gorm.io/gorm"
)

type AutoCompletePost struct {
	gorm.Model
	Title  string
	Body   string
	Status string
}

func PresetsAutoCompleteBasicFilter(b *presets.Builder, ab *autocomplete.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	err := db.AutoMigrate(&AutoCompletePost{})
	if err != nil {
		panic(err)
	}
	// create a ModelBuilder
	postBuilder := b.Model(&AutoCompletePost{})
	// get its ListingBuilder
	listing := postBuilder.Listing()
	// new autocomplete builder
	abm := ab.Model(&AutoCompletePost{}).SQLCondition("title ilike ? ").
		Columns("id", "title", "body").Paging(true)
	remoteUrl := "http://localhost:7800" + abm.JsonHref()

	// Call FilterDataFunc
	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		// Prepare filter options, it is a two dimension array: [][]string{"text", "value"}

		pagingConfig := autocomplete.NewDefaultAutocompleteDataSource(remoteUrl)
		noPagingConfig := autocomplete.NewDefaultAutocompleteDataSource(remoteUrl)
		noPagingConfig.IsPaging = false
		return []*vuetifyx.FilterItem{
			{
				Key:      "title",
				Label:    "Title",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition:           `title ilike ?`,
				AutocompleteDataSource: pagingConfig,
			},
			{
				Key:      "title1",
				Label:    "TitleNOPaging",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition:           `title ilike ?`,
				AutocompleteDataSource: noPagingConfig,
			},
		}
	})
	return
}

// @snippet_end
