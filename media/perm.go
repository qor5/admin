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
	PermCopyURL     = "perm_media_library_copy_url"
	PermNewFolder   = "perm_media_library_new_folder"
	PermListFolders = "perm_media_library_list_folders"
)

func (mb *Builder) uploadIsAllowed(r *http.Request) error {
	return mb.mb.Info().Verifier().Do(PermUpload).WithReq(r).IsAllowed()
}

func (mb *Builder) copyURLIsAllowed(r *http.Request) error {
	return mb.mb.Info().Verifier().Do(PermCopyURL).WithReq(r).IsAllowed()
}

func (mb *Builder) moveToIsAllowed(r *http.Request) error {
	return mb.mb.Info().Verifier().Do(PermMovieTo).WithReq(r).IsAllowed()
}

func (mb *Builder) deleteIsAllowed(r *http.Request, obj interface{}) error {
	if obj == nil {
		return mb.mb.Info().Verifier().Do(PermDelete).WithReq(r).IsAllowed()
	}
	return mb.mb.Info().Verifier().Do(PermDelete).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) updateDescIsAllowed(r *http.Request, obj interface{}) error {
	return mb.mb.Info().Verifier().Do(PermUpdateDesc).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) updateNameIsAllowed(r *http.Request, obj interface{}) error {
	return mb.mb.Info().Verifier().Do(PermUpdateName).ObjectOn(obj).WithReq(r).IsAllowed()
}

func (mb *Builder) newFolderIsAllowed(r *http.Request) error {
	return mb.mb.Info().Verifier().Do(PermNewFolder).WithReq(r).IsAllowed()
}

func (mb *Builder) listFoldersIsAllowed(r *http.Request) error {
	return mb.mb.Info().Verifier().Do(PermListFolders).WithReq(r).IsAllowed()
}
