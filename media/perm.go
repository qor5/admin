package media

import "net/http"

// DO NOT associate media_library permissions with parent resources
// WRONG: permPolicy.On("*:post:*")
// right: permPolicy.On("*")
// right: permPolicy.On("*:media_libraries:*")
// right: permPolicy.On("*:media_libraries:1")
const (
	PermUpload     = "perm_media_library_upload"
	PermDelete     = "perm_media_library_delete"
	PermUpdateDesc = "perm_media_library_update_desc"
)

func uploadIsAllowed(r *http.Request) error {
	return permVerifier.Do(PermUpload).On("media_libraries").WithReq(r).IsAllowed()
}

func deleteIsAllowed(r *http.Request, obj interface{}) error {
	return permVerifier.Do(PermDelete).ObjectOn(obj).WithReq(r).IsAllowed()
}

func updateDescIsAllowed(r *http.Request, obj interface{}) error {
	return permVerifier.Do(PermUpdateDesc).ObjectOn(obj).WithReq(r).IsAllowed()
}
