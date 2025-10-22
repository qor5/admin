package oss

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/oss/filesystem"
)

var (
	// URLTemplate default URL template
	URLTemplate = "/system/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	// Storage the storage used to save medias
	Storage oss.StorageInterface = filesystem.New("public")
	_       base.Media           = &OSS{}
)

// OSS common storage interface
type OSS struct {
	base.Base
}

// DefaultURLTemplateHandler used to generate URL and save into database
var DefaultURLTemplateHandler = func(oss OSS, option *base.Option) (url string) {
	if url = option.Get("URL"); url == "" {
		url = URLTemplate
	}

	url = strings.TrimSuffix(Storage.GetEndpoint(context.Background()), "/") + "/" + strings.TrimPrefix(url, "/")
	if strings.HasPrefix(url, "/") {
		return url
	}

	for _, prefix := range []string{"https://", "http://"} {
		url = strings.TrimPrefix(url, prefix)
	}

	// convert `getqor.com/hello` => `//getqor.com/hello`
	return "//" + url
}

// GetURLTemplate URL's template
func (o OSS) GetURLTemplate(option *base.Option) (url string) {
	return DefaultURLTemplateHandler(o, option)
}

// DefaultStoreHandler used to store reader with default Storage
var DefaultStoreHandler = func(oss OSS, path string, option *base.Option, reader io.Reader) error {
	_, err := Storage.Put(context.Background(), path, reader)
	return err
}

// Store save reader's content with path
func (o OSS) Store(path string, option *base.Option, reader io.Reader) error {
	return DefaultStoreHandler(o, path, option, reader)
}

// DefaultRetrieveHandler used to retrieve file
var DefaultRetrieveHandler = func(oss OSS, path string) (base.FileInterface, error) {
	result, err := Storage.GetStream(context.Background(), path)
	if f, ok := result.(base.FileInterface); ok {
		return f, err
	}
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(result)
	if err != nil {
		return nil, err
	}
	rs := ClosingReadSeeker{bytes.NewReader(buf)}
	rs.Seek(0, 0)
	return rs, nil
}

// Retrieve retrieve file content with url
func (o OSS) Retrieve(path string) (base.FileInterface, error) {
	return DefaultRetrieveHandler(o, path)
}

// URL return file's url with given style
func (o OSS) URL(styles ...string) string {
	url := o.Base.URL(styles...)

	newurl, err := Storage.GetURL(context.Background(), url)
	if err != nil || newurl == "" {
		return url
	}

	return newurl
}

func (o OSS) String() string {
	url := o.Base.URL()

	newurl, err := Storage.GetURL(context.Background(), url)
	if err != nil || newurl == "" {
		return url
	}

	return newurl
}

// ClosingReadSeeker implement Closer interface for ReadSeeker
type ClosingReadSeeker struct {
	io.ReadSeeker
}

// Close implement Closer interface for Buffer
func (ClosingReadSeeker) Close() error {
	return nil
}
