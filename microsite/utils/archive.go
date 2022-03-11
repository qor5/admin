package utils

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/qor/oss"
)

func RemoveUselessArchiveFiles(list []string) (result []string) {
	for _, v := range list {
		// Compressing in mac os may create useless files and dirs whose name contains "__MACOSX" or "DS_Store".
		// Some compressing may cause dirs to be files, so we need to remove these files(v[len(v)-1] == '/') from our list.
		if strings.Contains(v, "__MACOSX") || strings.Contains(v, "DS_Store") || v[len(v)-1] == '/' {
			continue
		}
		result = append(result, v)
	}
	return
}

func Upload(storage oss.StorageInterface, path string, reader io.Reader) (err error) {
	timeBegin := time.Now()
	_, err = storage.Put(path, reader)
	if err != nil {
		log.Println(err)
	}
	timeFinish := time.Now()
	fmt.Printf("upload: %s, time_spent_ms: %s \n", path, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
	return
}

func DeleteObjects(storage oss.StorageInterface, paths []string) (err error) {
	timeBegin := time.Now()
	for _, v := range paths {
		err = storage.Delete(v)
		if err != nil {
			log.Println(err)
		}
	}
	timeFinish := time.Now()
	fmt.Printf("delete: %s, time_spent_ms: %s \n", paths, fmt.Sprintf("%f", float64(timeFinish.Sub(timeBegin))/float64(time.Millisecond)))
	return
}
