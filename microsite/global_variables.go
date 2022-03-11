package microsite

import (
	"errors"
	"os"
)

var INVALID_ARCHIVER_ERROR = errors.New("unarr: No valid RAR, ZIP, 7Z or TAR archive")
var TOO_MANY_FILE_ERROR = errors.New("Too many uploaded files, please contact the administrator")

var PackageAndPreviewPrepath = "microsite"
var LocalTempFileDir = "./temp"
var MaxNumberOfFilesUploadedAtTheSameTime = 10

// MaximumNumberOfFilesInArchive can't be lager than 1000, otherwise, s3's DeleteObjects will return error.
// The more files, the longer the upload time.
var MaximumNumberOfFilesInArchive = 200

func GetTempFileDir() (string, error) {
	_, err := os.Stat(LocalTempFileDir)
	if err != nil {
		err = os.Mkdir(LocalTempFileDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return LocalTempFileDir, nil
}
