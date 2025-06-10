package presets

import "fmt"

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
	ParamOperateID                = "operate_id"

	VarsPresetsDataChanged = "presetsDataChanged"

	// list editor
	ParamAddRowFormKey      = "listEditor_AddRowFormKey"
	ParamRemoveRowFormKey   = "listEditor_RemoveRowFormKey"
	ParamIsStartSort        = "listEditor_IsStartSort"
	ParamSortSectionFormKey = "listEditor_SortSectionFormKey"
	ParamSortResultFormKey  = "listEditor_SortResultFormKey"
)

var PhraseHasPresetsDataChanged = fmt.Sprintf("Object.values(vars.%s || {} ).some(v => v)", VarsPresetsDataChanged)

const (
	setFieldErrorsScript = `	
let keys = Object.keys(dash.errorMessages);
if (dash.__currentValidateKeys) {
    for (const key of dash.__currentValidateKeys) {
        dash.errorMessages[key] = payload.field_errors ? payload.field_errors[key] : null;
    }
} else {
    for (const key in payload.field_errors) {
        dash.errorMessages[key] = payload.field_errors ? payload.field_errors[key] : null;
    }
    keys.forEach(key => {
        dash.errorMessages[key] = payload.field_errors ? payload.field_errors[key] : null;
    })
}
dash.__currentValidateKeys = [];
`
	setValidateKeysScript = `
dash.__currentValidateKeys = dash.__currentValidateKeys ?? [];
for (let key in form) {
    if (form[key] !== oldForm[key]) {
        dash.__currentValidateKeys.push(key)
        typeof dash.__findLinkageFields === "function" && dash.__findLinkageFields(key);
    }
}
`
	checkFormChangeScript = `if(JSON.stringify(form)==JSON.stringify(oldForm)){return}`
)
