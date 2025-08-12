package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/qor5/admin/v3/media/base"
)

var _ base.Media = &FileSystem{}

// FileSystem defined a media library storage using file system
type FileSystem struct {
	base.Base
}

// GetFullPath return full file path from a relative file path
func (FileSystem) GetFullPath(url string, option *base.Option) (string, error) {
	basePath := "./public"
	if option != nil && option.Get("path") != "" {
		basePath = option.Get("path")
	}

	// Convert base path to absolute and join with URL in one step
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute base path: %w", err)
	}

	// Join and get absolute path directly
	absPath, err := filepath.Abs(filepath.Join(absBasePath, url))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Security check: ensure path is within base directory
	if relPath, err := filepath.Rel(absBasePath, absPath); err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	} else if filepath.IsAbs(relPath) || relPath == ".." || (len(relPath) >= 2 && relPath[:2] == "..") {
		return "", fmt.Errorf("path '%s' is outside the allowed base directory '%s'", absPath, absBasePath)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(absPath); dir != "" {
		if _, err = os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	return absPath, nil
}

// Store save reader's context with name
func (f FileSystem) Store(name string, option *base.Option, reader io.Reader) error {
	fullpath, err := f.GetFullPath(name, option)
	if err != nil {
		return err
	}

	dst, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, reader)
	return err
}

// Retrieve retrieve file content with url
func (f FileSystem) Retrieve(url string) (base.FileInterface, error) {
	fullpath, err := f.GetFullPath(url, nil)
	if err != nil {
		return nil, os.ErrNotExist
	}
	return os.Open(fullpath)
}
