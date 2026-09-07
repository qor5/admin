package views

import (
	"github.com/qor5/admin/media/shorturl"
	"github.com/qor5/web"
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

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB, shortURLCfg *shorturl.Config) {
	hub.RegisterEventFunc(openFileChooserEvent, fileChooser(db, shortURLCfg))
	hub.RegisterEventFunc(deleteFileEvent, deleteFileField())
	hub.RegisterEventFunc(cropImageEvent, cropImage(db))
	hub.RegisterEventFunc(loadImageCropperEvent, loadImageCropper(db))
	hub.RegisterEventFunc(imageSearchEvent, searchFile(db, shortURLCfg))
	hub.RegisterEventFunc(imageJumpPageEvent, jumpPage(db, shortURLCfg))
	hub.RegisterEventFunc(uploadFileEvent, uploadFile(db, shortURLCfg))
	hub.RegisterEventFunc(chooseFileEvent, chooseFile(db, shortURLCfg))
	hub.RegisterEventFunc(updateDescriptionEvent, updateDescription(db, shortURLCfg))
	hub.RegisterEventFunc(deleteConfirmationEvent, deleteConfirmation(db))
	hub.RegisterEventFunc(doDeleteEvent, doDelete(db, shortURLCfg))
}
