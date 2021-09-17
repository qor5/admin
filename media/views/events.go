package views

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/perm"
	"github.com/jinzhu/gorm"
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

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB, permVerifier *perm.Verifier) {
	hub.RegisterEventFunc(openFileChooserEvent, fileChooser(db, permVerifier))
	hub.RegisterEventFunc(deleteFileEvent, deleteFileField())
	hub.RegisterEventFunc(cropImageEvent, cropImage(db))
	hub.RegisterEventFunc(loadImageCropperEvent, loadImageCropper(db))
	hub.RegisterEventFunc(imageSearchEvent, searchFile(db, permVerifier))
	hub.RegisterEventFunc(imageJumpPageEvent, jumpPage(db, permVerifier))
	hub.RegisterEventFunc(uploadFileEvent, uploadFile(db, permVerifier))
	hub.RegisterEventFunc(chooseFileEvent, chooseFile(db))
	hub.RegisterEventFunc(updateDescriptionEvent, updateDescription(db, permVerifier))
	hub.RegisterEventFunc(deleteConfirmationEvent, deleteConfirmation(db))
	hub.RegisterEventFunc(doDeleteEvent, doDelete(db, permVerifier))
}
