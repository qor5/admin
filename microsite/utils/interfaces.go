package utils

type DeleteObjectsInterface interface {
	DeleteObjects(paths []string) (err error)
}

type CopyInterface interface {
	Copy(from, to string) (err error)
}

type GetBucketInterface interface {
	GetBucket() string
}
