package utils

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/qor/oss"
)

func Upload(storage oss.StorageInterface, path string, reader io.Reader) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			//todo error log
			log.Println(err)
		} else {
			fmt.Printf("upload: %s, time_spent_ms: %s \n", path, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
		}
	}()
	_, err = storage.Put(path, reader)
	if err != nil {
		return
	}
	return
}

func DeleteObjects(storage oss.StorageInterface, paths []string) (err error) {
	timeBegin := time.Now()
	defer func() {
		timeFinish := time.Now()
		if err != nil {
			//todo error log
			log.Println(err)
		} else {
			fmt.Printf("delete: %s, time_spent_ms: %s \n", paths, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
		}
	}()

	if storage, ok := storage.(DeleteObjecter); ok {
		var length = len(paths)
		var i = 0
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
				return
			}
		}
		return
	}

	for _, v := range paths {
		err = storage.Delete(v)
		if err != nil {
			return
		}
	}

	return
}
