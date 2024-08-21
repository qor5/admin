package presets

const (
	PermModule          = "presets"
	PermList            = "presets:list"
	PermGet             = "presets:get"
	PermCreate          = "presets:create"
	PermUpdate          = "presets:update"
	PermDelete          = "presets:delete"
	PermActions         = "presets:actions:*"
	PermDoListingAction = "presets:do_listing_action:*"
	PermBulkActions     = "presets:bulk_actions:*"

	permActions         = "actions"
	permDoListingAction = "do_listing_action"
	permBulkActions     = "bulk_actions"
)

var PermRead = []string{PermList, PermGet}

// params
const (
	ParamID                       = "id"
	ParamAction                   = "action"
	ParamOverlay                  = "overlay"
	ParamOverlayAfterUpdateScript = "overlay_after_update_script"
	ParamOverlayUpdateID          = "overlay_update_id"
	ParamAfterDeleteEvent         = "presets_after_delete_event"
	ParamPortalName               = "portal_name"

	VarsPresetsDataChanged = "presetsDataChanged"

	// list editor
	ParamAddRowFormKey      = "listEditor_AddRowFormKey"
	ParamRemoveRowFormKey   = "listEditor_RemoveRowFormKey"
	ParamIsStartSort        = "listEditor_IsStartSort"
	ParamSortSectionFormKey = "listEditor_SortSectionFormKey"
	ParamSortResultFormKey  = "listEditor_SortResultFormKey"
)
