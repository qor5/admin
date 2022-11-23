package presets

const (
	PermModule = "presets"
	PermList   = "presets:list"
	PermGet    = "presets:get"
	PermCreate = "presets:create"
	PermUpdate = "presets:update"
	PermDelete = "presets:delete"

	PermListingActions = "listing_actions"
	PermActions        = "actions"
	PermBulkActions    = "bulk_actions"
)

var (
	PermRead = []string{PermList, PermGet}
)

// params
const (
	ParamID                       = "id"
	ParamAction                   = "action"
	ParamOverlay                  = "overlay"
	ParamOverlayAfterUpdateScript = "overlay_after_update_script"
	ParamOverlayUpdateID          = "overlay_update_id"
	ParamBulkActionName           = "bulk_action"
	ParamListingActionName        = "listing_action"
	ParamSelectedIds              = "selected_ids"
	ParamInDialog                 = "presets_in_dialog"
	ParamListingQueries           = "presets_listing_queries"
	ParamAfterDeleteEvent         = "presets_after_delete_event"
)

const (
	ColorPrimary   = "primary"
	ColorSecondary = "secondary"
)
