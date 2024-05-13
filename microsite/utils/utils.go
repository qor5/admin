package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/qor/oss"
)

func Upload(storage oss.StorageInterface, path string, reader io.Reader) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			// todo error log
			log.Println(err)
		} else {
			log.Printf("upload: %s, time_spent_ms: %s \n", path, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
		}
	}()
	_, err = storage.Put(path, reader)
	if err != nil {
		err = errors.New(fmt.Sprintf("upload error: %v, path: %v", err, path))
		return
	}
	return
}

func DeleteObjects(storage oss.StorageInterface, paths []string) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			// todo error log
			log.Println(err)
		} else {
			log.Printf("delete: %s, time_spent_ms: %s \n", paths, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
		}
	}()

	if storage, ok := storage.(DeleteObjectsInterface); ok {
		length := len(paths)
		i := 0
		for i < length {
			var left, right int
			left = i
			if i+1000 < length {
				right = i + 1000
			} else {
				right = length
			}
			i = right
			err = storage.DeleteObjects(paths[left:right])
			if err != nil {
				err = errors.New(fmt.Sprintf("delete error: %v, path: %v", err, paths[left:right]))
				return
			}
		}
		return
	}

	for _, v := range paths {
		err = storage.Delete(v)
		if err != nil {
			err = errors.New(fmt.Sprintf("delete error: %v, path: %v", err, v))
			return
		}
	}

	return
}

func Copy(storage oss.StorageInterface, from, to string) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			// todo error log
			log.Println(err)
		} else {
			log.Printf("copy: from %s to %s, time_spent_ms: %s \n", from, to, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
		}
	}()

	if storage, ok := storage.(GetBucketInterface); ok {
		from = path.Join(storage.GetBucket(), from)
	}

	err = storage.(CopyInterface).Copy(from, to)
	if err != nil {
		err = errors.New(fmt.Sprintf("copy error: %v, from: %v, to: %v", err, from, to))
	}
	return
}
