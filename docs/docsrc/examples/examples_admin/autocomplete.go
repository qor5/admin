package examples_admin

import (
	"net/http"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/autocomplete"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

type AutoCompletePost struct {
	gorm.Model
	Title  string
	Body   string
	Status string
}

func AutoCompleteBasicFilterExample(b *presets.Builder, ab *autocomplete.Builder, db *gorm.DB) http.Handler {
	// Setup the project name, ORM and Homepage
	ab.DB(db)
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
	abm1 := ab.Model(&AutoCompletePost{}).SQLCondition("title ilike ? ").
		Columns("id", "title").Paging(true)
	abm2 := ab.Model(&AutoCompletePost{}).SQLCondition("body ilike ? ").
		Columns("id", "body").UriName("auto-complete-posts-body").OrderBy("id desc")

	// Call FilterDataFunc
	listing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		// Prepare filter options, it is a two dimension array: [][]string{"text", "value"}

		titleConfig := autocomplete.NewDefaultAutocompleteDataSource(abm1.JsonHref())
		bodyConfig := autocomplete.NewDefaultAutocompleteDataSource(abm2.JsonHref())
		bodyConfig.IsPaging = false
		bodyConfig.ItemTitle = "body"
		return []*vuetifyx.FilterItem{
			{
				Key:      "select",
				Label:    "ID",
				ItemType: vuetifyx.ItemTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition: `id %s ?`,
				Options: []*vuetifyx.SelectItem{
					{Text: "test001", Value: "1"},
					{Text: "test002", Value: "2"},
				},
			},
			{
				Key:      "title",
				Label:    "Title",
				ItemType: vuetifyx.AutoCompleteTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition:           `title ilike ?`,
				AutocompleteDataSource: titleConfig,
				WrapInput: func(val string) interface{} {
					return strings.Split(val, titleConfig.Separator)[0]
				},
			},
			{
				Key:      "body",
				Label:    "Body",
				ItemType: vuetifyx.AutoCompleteTypeSelect,
				// %s is the condition. e.g. >, >=, =, <, <=, like，
				// ? is the value of selected option
				SQLCondition:           `body ilike ?`,
				AutocompleteDataSource: bodyConfig,
				WrapInput: func(val string) interface{} {
					return strings.Split(val, bodyConfig.Separator)[0]
				},
			},
		}
	})
	return b
}
