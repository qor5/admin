package media

import (
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

const (
	openFileChooserEvent    = "mediaLibrary_OpenFileChooserEvent"
	deleteFileEvent         = "mediaLibrary_DeleteFileEvent"
	cropImageEvent          = "mediaLibrary_CropImageEvent"
	loadImageCropperEvent   = "mediaLibrary_LoadImageCropperEvent"
	imageSearchEvent        = "mediaLibrary_ImageSearchEvent"
	imageJumpPageEvent      = "mediaLibrary_ImageJumpPageEvent"
	uploadFileEvent         = "mediaLibrary_UploadFileEvent"
	chooseFileEvent         = "mediaLibrary_ChooseFileEvent"
	updateDescriptionEvent  = "mediaLibrary_UpdateDescriptionEvent"
	deleteConfirmationEvent = "mediaLibrary_DeleteConfirmationEvent"
	doDeleteEvent           = "mediaLibrary_DoDelete"
)

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB) {
	hub.RegisterEventFunc(openFileChooserEvent, fileChooser(db))
	hub.RegisterEventFunc(deleteFileEvent, deleteFileField())
	hub.RegisterEventFunc(cropImageEvent, cropImage(db))
	hub.RegisterEventFunc(loadImageCropperEvent, loadImageCropper(db))
	hub.RegisterEventFunc(imageSearchEvent, searchFile(db))
	hub.RegisterEventFunc(imageJumpPageEvent, jumpPage(db))
	hub.RegisterEventFunc(uploadFileEvent, uploadFile(db))
	hub.RegisterEventFunc(chooseFileEvent, chooseFile(db))
	hub.RegisterEventFunc(updateDescriptionEvent, updateDescription(db))
	hub.RegisterEventFunc(deleteConfirmationEvent, deleteConfirmation(db))
	hub.RegisterEventFunc(doDeleteEvent, doDelete(db))
}
