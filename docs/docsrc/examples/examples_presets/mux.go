package examples_presets

import (
	"fmt"

	"github.com/qor5/web/v3/examples"
	"gorm.io/gorm"

	docsexamples "github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/presets"
)

func SamplesHandler(mux examples.Muxer) {
	db := docsexamples.ExampleDB()
	addExample(mux, db, PresetsHelloWorld)
	addExample(mux, db, PresetsKeywordSearchOff)
	addExample(mux, db, PresetsRowMenuAction)
	addExample(mux, db, PresetsListingCustomizationFields)
	addExample(mux, db, PresetsListingCustomizationFilters)
	addExample(mux, db, PresetsListingCustomizationTabs)
	addExample(mux, db, PresetsListingCustomizationBulkActions)
	addExample(mux, db, PresetsEditingCustomizationDescription)
	addExample(mux, db, PresetsEditingTiptap)
	addExample(mux, db, PresetsEditingCustomizationFileType)
	addExample(mux, db, PresetsEditingTabController)
	addExample(mux, db, PresetsEditingCustomizationValidation)
	addExample(mux, db, PresetsDetailPageTopNotes)
	addExample(mux, db, PresetsDetailPageDetails)
	addExample(mux, db, PresetsDetailPageCards)
	addExample(mux, db, PresetsDetailTabsField)
	addExample(mux, db, PresetsDetailAfterTitle)
	addExample(mux, db, PresetsDetailListSectionStatusxFieldViolations)
	addExample(mux, db, PresetsPermissions)
	addExample(mux, db, PresetsModelBuilderExtensions)
	addExample(mux, db, PresetsBasicFilter)
	addExample(mux, db, PresetsNotificationCenterSample)
	addExample(mux, db, PresetsLinkageSelectFilterItem)
	addExample(mux, db, PresetsBrandTitle)
	addExample(mux, db, PresetsBrandFunc)
	addExample(mux, db, PresetsProfile)
	addExample(mux, db, PresetsOrderMenu)
	addExample(mux, db, PresetsCustomizeMenu)
	addExample(mux, db, PresetsGroupMenu)
	addExample(mux, db, PresetsGroupMenuWithPermission)
	addExample(mux, db, PresetsMenuComponent)
	addExample(mux, db, PresetsConfirmDialog)
	addExample(mux, db, PresetsEditingCustomizationTabs)
	addExample(mux, db, PresetsEditingValidate)
	addExample(mux, db, PresetsEditingSetter)
	addExample(mux, db, PresetsEditingSection)
	addExample(mux, db, PresetsEditingSaverValidation)
	addExample(mux, db, PresetsListingCustomizationSearcher)
	addExample(mux, db, PresetsListingDatatableFunc)
	addExample(mux, db, PresetsListingFilterNotificationFunc)
	addExample(mux, db, PresetsDetailInlineEditDetails)
	addExample(mux, db, PresetsDetailSectionView)
	addExample(mux, db, PresetsDetailTabsSection)
	addExample(mux, db, PresetsDetailTabsSectionOrder)
	addExample(mux, db, PresetsDetailInlineEditInspectTables)
	addExample(mux, db, PresetsDetailSectionLabel)
	addExample(mux, db, PresetsDetailNestedMany)
	addExample(mux, db, PresetsDetailInlineEditFieldSections)
	addExample(mux, db, PresetsDetailInlineEditValidate)
	addExample(mux, db, PresetsDetailSimple)
	addExample(mux, db, PresetsDetailListSection)
	addExample(mux, db, PresetsDetailSidePanel)
	addExample(mux, db, PresetsUtilsDialog)
	addExample(mux, db, PresetsCustomPage)
	addExample(mux, db, PresetsPlainNestedField)
	addExample(mux, db, PresetsDetailDisableSave)
	addExample(mux, db, PresetsDetailSaverValidation)
	addExample(mux, db, PresetsDataOperatorWithGRPC)
	addExample(mux, db, PresetsEditingSingletonNested)
	addExample(mux, db, PresetsSectionSingleton)
	addExample(mux, db, PresetsSectionDetailingNormal)
	addExample(mux, db, PresetsSectionEditingClone)
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
	mux.Handle(path, p)
}
