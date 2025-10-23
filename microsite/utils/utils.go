package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/qor5/x/v3/oss"
)

func Upload(storage oss.StorageInterface, path string, reader io.Reader) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			// todo error log
			log.Println(err)
		} else {
			log.Printf("upload: %s, time_spent_ms: %f \n", path, float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond))
		}
	}()
	_, err = storage.Put(context.Background(), path, reader)
	if err != nil {
		err = fmt.Errorf("upload error: %w, path: %s", err, path)
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
			log.Printf("delete: %v, time_spent_ms: %f \n", paths, float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond))
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
				err = fmt.Errorf("delete error: %w, path: %v", err, paths[left:right])
				return
			}
		}
		return
	}

	for _, v := range paths {
		err = storage.Delete(context.Background(), v)
		if err != nil {
			err = fmt.Errorf("delete error: %w, path: %s", err, v)
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
			log.Printf("copy: from %s to %s, time_spent_ms: %f \n", from, to, float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond))
		}
	}()

	if storage, ok := storage.(GetBucketInterface); ok {
		from = path.Join(storage.GetBucket(), from)
	}

	err = storage.(CopyInterface).Copy(from, to)
	if err != nil {
		err = fmt.Errorf("copy error: %w, from: %s, to: %s", err, from, to)
	}
	return
}
