package media

import (
	"github.com/qor5/web/v3"
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
	CreateDirectoryEvent    = "mediaLibrary_CreateDirectoryEvent"
	NewFolderDialogEvent    = "mediaLibrary_NewFolderDialogEvent"
)

func registerEventFuncs(hub web.EventFuncHub, mb *Builder) {
	hub.RegisterEventFunc(openFileChooserEvent, fileChooser(mb))
	hub.RegisterEventFunc(deleteFileEvent, deleteFileField())
	hub.RegisterEventFunc(cropImageEvent, cropImage(mb))
	hub.RegisterEventFunc(loadImageCropperEvent, loadImageCropper(mb))
	hub.RegisterEventFunc(imageSearchEvent, searchFile(mb))
	hub.RegisterEventFunc(imageJumpPageEvent, jumpPage(mb))
	hub.RegisterEventFunc(uploadFileEvent, uploadFile(mb))
	hub.RegisterEventFunc(chooseFileEvent, chooseFile(mb))
	hub.RegisterEventFunc(updateDescriptionEvent, updateDescription(mb))
	hub.RegisterEventFunc(deleteConfirmationEvent, deleteConfirmation(mb))
	hub.RegisterEventFunc(doDeleteEvent, doDelete(mb))
	hub.RegisterEventFunc(CreateDirectoryEvent, createDirectory(mb))
	hub.RegisterEventFunc(NewFolderDialogEvent, newFolderDialog(mb))
}
