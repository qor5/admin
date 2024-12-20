package media

import (
	"fmt"
)

func mainPortalName(field string) string {
	return fmt.Sprintf("%s_portal", field)
}

func deleteConfirmPortalName(field string) string {
	return fmt.Sprintf("%s_deleteConfirm_portal", field)
}

func mediaBoxThumbnailsPortalName(field string) string {
	return fmt.Sprintf("%s_portal_thumbnails", field)
}

func cropperPortalName(field string) string {
	return fmt.Sprintf("%s_cropper_portal", field)
}

func dialogContentPortalName(field string) string {
	return fmt.Sprintf("%s_dialog_content", field)
}

func searchKeywordName(inMediaLibrary bool, field string) string {
	if inMediaLibrary {
		return "keyword"
	}
	return fmt.Sprintf("%s_file_chooser_search_keyword", field)
}

func currentPageName(inMediaLibrary bool, field string) string {
	if inMediaLibrary {
		return "page"
	}
	return fmt.Sprintf("%s_file_chooser_current_page", field)
}

func fileCroppingVarName(id uint) string {
	return fmt.Sprintf("fileChooser%d_cropping", id)
}

func folderGroupPortalName(id uint) string {
	return fmt.Sprintf("%v_folder_portal_name", id)
}

const (
	newFolderDialogPortalName         = "media_new_folder_dialog_portal_name"
	renameDialogPortalName            = "media_rename_dialog_portal_name"
	updateDescriptionDialogPortalName = "media_update_description_dialog_portal_name"
	moveToFolderDialogPortalName      = "media_move_to_folder_dialog_portal_name"
)
