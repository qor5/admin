package media

import "net/http"

// DO NOT associate media_library permissions with parent resources
// WRONG: permPolicy.On("*:post:*")
// right: permPolicy.On("*")
// right: permPolicy.On("*:media_libraries:*")
// right: permPolicy.On("*:media_libraries:1")
const (
	PermUpload      = "perm_media_library_upload"
	PermDelete      = "perm_media_library_delete"
	PermUpdateDesc  = "perm_media_library_update_desc"
	PermUpdateName  = "perm_media_library_update_name"
	PermMovieTo     = "perm_media_library_move_to"
	PermCopy        = "perm_media_library_copy"
	PermNewFolder   = "perm_media_library_new_folder"
	PermListFolders = "perm_media_library_list_folders"
)

func (mb *Builder) uploadIsAllowed(r *http.Request) error {
	return mb.permVerifier.Do(PermUpload).On("media_libraries").WithReq(r).IsAllowed()
}

func (mb *Builder) copyIsAllowed(r *http.Request) error {
	return mb.permVerifier.Do(PermCopy).On("media_libraries").WithReq(r).IsAllowed()
}

func (mb *Builder) moveToIsAllowed(r *http.Request) error {
	return mb.permVerifier.Do(PermMovieTo).On("media_libraries").WithReq(r).IsAllowed()
}

func (mb *Builder) deleteIsAllowed(r *http.Request, obj interface{}) error {
	if obj == nil {
		return mb.permVerifier.Do(PermDelete).On("media_libraries").WithReq(r).IsAllowed()
	}
	return mb.permVerifier.Do(PermDelete).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) updateDescIsAllowed(r *http.Request, obj interface{}) error {
	return mb.permVerifier.Do(PermUpdateDesc).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) updateNameIsAllowed(r *http.Request, obj interface{}) error {
	return mb.permVerifier.Do(PermUpdateName).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) newFolderIsAllowed(r *http.Request) error {
	return mb.permVerifier.Do(PermNewFolder).On("media_libraries").WithReq(r).IsAllowed()
}

func (mb *Builder) listFoldersIsAllowed(r *http.Request) error {
	return mb.permVerifier.Do(PermListFolders).On("media_libraries").WithReq(r).IsAllowed()
}
