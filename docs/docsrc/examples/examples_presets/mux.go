package examples_presets

import (
	"fmt"

	docsexamples "github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/examples"
	"gorm.io/gorm"
)

func SamplesHandler(mux examples.Muxer, prefix string) {
	db := docsexamples.ExampleDB()
	addExample(mux, db, PresetsHelloWorld)
	addExampleX(mux, db, PresetsHelloWorldX)
	addExample(mux, db, PresetsKeywordSearchOff)
	addExample(mux, db, PresetsListingCustomizationFields)
	addExample(mux, db, PresetsListingCustomizationFilters)
	addExample(mux, db, PresetsListingCustomizationTabs)
	addExample(mux, db, PresetsListingCustomizationBulkActions)
	addExample(mux, db, PresetsEditingCustomizationDescription)
	addExample(mux, db, PresetsEditingCustomizationFileType)
	addExample(mux, db, PresetsEditingCustomizationValidation)
	addExample(mux, db, PresetsDetailPageTopNotes)
	addExample(mux, db, PresetsDetailPageDetails)
	addExample(mux, db, PresetsDetailPageCards)
	addExample(mux, db, PresetsPermissions)
	addExample(mux, db, PresetsModelBuilderExtensions)
	addExample(mux, db, PresetsBasicFilter)
	addExample(mux, db, PresetsNotificationCenterSample)
	addExample(mux, db, PresetsLinkageSelectFilterItem)
	addExample(mux, db, PresetsBrandTitle)
	addExample(mux, db, PresetsBrandFunc)
	addExample(mux, db, PresetsProfile)
	addExample(mux, db, PresetsOrderMenu)
	addExample(mux, db, PresetsGroupMenu)
	addExample(mux, db, PresetsConfirmDialog)
	addExample(mux, db, PresetsEditingCustomizationTabs)
	addExample(mux, db, PresetsListingCustomizationSearcher)
	addExample(mux, db, PresetsDetailInlineEditDetails)
	addExample(mux, db, PresetsDetailInlineEditInspectTables)
	addExample(mux, db, PresetsDetailInlineEditDetailsInspectShowFields)
	addExample(mux, db, PresetsDetailInlineEditFieldSections)
	addExample(mux, db, PresetsDetailSimple)
	return
}

type exampleFunc func(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
)

func addExample(mux examples.Muxer, db *gorm.DB, f exampleFunc) {
	path := examples.URLPathByFunc(f)
	p := presets.New().URIPrefix(path)
	f(p, db)
	fmt.Println("Example mounting at: ", path)
	mux.Handle(
		path,
		p,
	)
}

type exampleFuncX func(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
)

func addExampleX(mux examples.Muxer, db *gorm.DB, f exampleFuncX) {
	path := examples.URLPathByFunc(f)
	p := presets.New().URIPrefix(path)
	f(p, db)
	fmt.Println("Example mounting at: ", path)
	mux.Handle(
		path,
		p,
	)
}
