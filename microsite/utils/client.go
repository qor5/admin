package utils

import "github.com/qor5/x/v3/oss/s3"

func NewClient(client *s3.Client) S3Client {
	return S3Client{client}
}

type S3Client struct {
	*s3.Client
}

func (this S3Client) GetBucket() string {
	return this.Config.Bucket
}
