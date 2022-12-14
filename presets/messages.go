package presets

import (
	"strings"
)

type Messages struct {
	SuccessfullyUpdated                        string
	Search                                     string
	New                                        string
	Update                                     string
	Delete                                     string
	Edit                                       string
	FormTitle                                  string
	OK                                         string
	Cancel                                     string
	Clear                                      string
	Create                                     string
	DeleteConfirmationTextTemplate             string
	CreatingObjectTitleTemplate                string
	EditingObjectTitleTemplate                 string
	ListingObjectTitleTemplate                 string
	DetailingObjectTitleTemplate               string
	FiltersClear                               string
	FiltersAdd                                 string
	FilterApply                                string
	FilterByTemplate                           string
	FiltersDateInTheLast                       string
	FiltersDateEquals                          string
	FiltersDateBetween                         string
	FiltersDateIsAfter                         string
	FiltersDateIsAfterOrOn                     string
	FiltersDateIsBefore                        string
	FiltersDateIsBeforeOrOn                    string
	FiltersDateDays                            string
	FiltersDateMonths                          string
	FiltersDateAnd                             string
	FiltersDateTo                              string
	FiltersNumberEquals                        string
	FiltersNumberBetween                       string
	FiltersNumberGreaterThan                   string
	FiltersNumberLessThan                      string
	FiltersNumberAnd                           string
	FiltersStringEquals                        string
	FiltersStringContains                      string
	FiltersMultipleSelectIn                    string
	FiltersMultipleSelectNotIn                 string
	PaginationRowsPerPage                      string
	ListingNoRecordToShow                      string
	ListingSelectedCountNotice                 string
	ListingClearSelection                      string
	BulkActionNoAvailableRecords               string
	BulkActionSelectedIdsProcessNoticeTemplate string
	ConfirmationDialogText                     string
	Language                                   string
	Colon                                      string
}

func (msgr *Messages) DeleteConfirmationText(id string) string {
	return strings.NewReplacer("{id}", id).
		Replace(msgr.DeleteConfirmationTextTemplate)
}

func (msgr *Messages) CreatingObjectTitle(modelName string) string {
	return strings.NewReplacer("{modelName}", modelName).
		Replace(msgr.CreatingObjectTitleTemplate)
}

func (msgr *Messages) EditingObjectTitle(label string, name string) string {
	return strings.NewReplacer("{id}", name, "{modelName}", label).
		Replace(msgr.EditingObjectTitleTemplate)
}
func (msgr *Messages) ListingObjectTitle(label string) string {
	return strings.NewReplacer("{modelName}", label).
		Replace(msgr.ListingObjectTitleTemplate)
}
func (msgr *Messages) DetailingObjectTitle(label string, name string) string {
	return strings.NewReplacer("{id}", name, "{modelName}", label).
		Replace(msgr.DetailingObjectTitleTemplate)
}

func (msgr *Messages) BulkActionSelectedIdsProcessNotice(ids string) string {
	return strings.NewReplacer("{ids}", ids).
		Replace(msgr.BulkActionSelectedIdsProcessNoticeTemplate)
}

func (msgr *Messages) FilterBy(filter string) string {
	return strings.NewReplacer("{filter}", filter).
		Replace(msgr.FilterByTemplate)
}

var Messages_en_US = &Messages{
	SuccessfullyUpdated:            "Successfully Updated",
	Search:                         "Search",
	New:                            "New",
	Update:                         "Update",
	Delete:                         "Delete",
	Edit:                           "Edit",
	FormTitle:                      "Form",
	OK:                             "OK",
	Cancel:                         "Cancel",
	Clear:                          "Clear",
	Create:                         "Create",
	DeleteConfirmationTextTemplate: "Are you sure you want to delete object with id: {id}?",
	CreatingObjectTitleTemplate:    "New {modelName}",
	EditingObjectTitleTemplate:     "Editing {modelName} {id}",
	ListingObjectTitleTemplate:     "Listing {modelName}",
	DetailingObjectTitleTemplate:   "{modelName} {id}",
	FiltersClear:                   "Clear Filters",
	FiltersAdd:                     "Add Filters",
	FilterApply:                    "Apply",
	FilterByTemplate:               "Filter by {filter}",
	FiltersDateInTheLast:           "is in the last",
	FiltersDateEquals:              "is equal to",
	FiltersDateBetween:             "is between",
	FiltersDateIsAfter:             "is after",
	FiltersDateIsAfterOrOn:         "is on or after",
	FiltersDateIsBefore:            "is before",
	FiltersDateIsBeforeOrOn:        "is before or on",
	FiltersDateDays:                "days",
	FiltersDateMonths:              "months",
	FiltersDateAnd:                 "and",
	FiltersDateTo:                  "to",
	FiltersNumberEquals:            "is equal to",
	FiltersNumberBetween:           "between",
	FiltersNumberGreaterThan:       "is greater than",
	FiltersNumberLessThan:          "is less than",
	FiltersNumberAnd:               "and",
	FiltersStringEquals:            "is equal to",
	FiltersStringContains:          "contains",
	FiltersMultipleSelectIn:        "in",
	FiltersMultipleSelectNotIn:     "not in",
	PaginationRowsPerPage:          "Rows per page: ",
	ListingNoRecordToShow:          "No records to show",
	ListingSelectedCountNotice:     "{count} records are selected. ",
	ListingClearSelection:          "clear selection",
	BulkActionNoAvailableRecords:   "None of the selected records can be executed with this action.",
	BulkActionSelectedIdsProcessNoticeTemplate: "Partially selected records cannot be executed with this action: {ids}.",
	ConfirmationDialogText:                     "Are you sure?",
	Language:                                   "Language",
	Colon:                                      ":",
}

var Messages_zh_CN = &Messages{
	SuccessfullyUpdated:            "???????????????",
	Search:                         "??????",
	New:                            "??????",
	Update:                         "??????",
	Delete:                         "??????",
	Edit:                           "??????",
	FormTitle:                      "??????",
	OK:                             "??????",
	Cancel:                         "??????",
	Clear:                          "??????",
	Create:                         "??????",
	DeleteConfirmationTextTemplate: "?????????????????????????????????????????????ID: {id}?",
	CreatingObjectTitleTemplate:    "??????{modelName}",
	EditingObjectTitleTemplate:     "??????{modelName} {id}",
	ListingObjectTitleTemplate:     "{modelName}??????",
	DetailingObjectTitleTemplate:   "{modelName} {id}",
	FiltersClear:                   "???????????????",
	FiltersAdd:                     "???????????????",
	FilterApply:                    "??????",
	FilterByTemplate:               "???{filter}??????",
	FiltersDateInTheLast:           "??????",
	FiltersDateEquals:              "??????",
	FiltersDateBetween:             "??????",
	FiltersDateIsAfter:             "??????",
	FiltersDateIsAfterOrOn:         "???????????????",
	FiltersDateIsBefore:            "??????",
	FiltersDateIsBeforeOrOn:        "???????????????",
	FiltersDateDays:                "???",
	FiltersDateMonths:              "???",
	FiltersDateAnd:                 "???",
	FiltersDateTo:                  "???",
	FiltersNumberEquals:            "??????",
	FiltersNumberBetween:           "??????",
	FiltersNumberGreaterThan:       "??????",
	FiltersNumberLessThan:          "??????",
	FiltersNumberAnd:               "???",
	FiltersStringEquals:            "??????",
	FiltersStringContains:          "??????",
	FiltersMultipleSelectIn:        "??????",
	FiltersMultipleSelectNotIn:     "?????????",
	PaginationRowsPerPage:          "??????: ",
	ListingNoRecordToShow:          "????????????????????????",
	ListingSelectedCountNotice:     "{count}?????????????????????",
	ListingClearSelection:          "????????????",
	BulkActionNoAvailableRecords:   "???????????????????????????????????????????????????",
	BulkActionSelectedIdsProcessNoticeTemplate: "????????????????????????????????????????????????: {ids}???",
	ConfirmationDialogText:                     "?????????????",
	Language:                                   "??????",
	Colon:                                      "???",
}
