package utils

type DeleteObjecter interface {
	DeleteObjects(paths []string) (err error)
}
