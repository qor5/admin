package redirection

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss"
)

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func (r *Builder) Get(ctx context.Context, path string) (*os.File, error) {
	return r.storage.Get(ctx, path)
}

func (r *Builder) GetStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return r.storage.GetStream(ctx, path)
}

func (r *Builder) Put(ctx context.Context, path string, reader io.Reader) (obj *oss.Object, err error) {
	defer func() {
		if err == nil {
			r.createEmptyTargetRecord(path)
		}
	}()
	return r.storage.Put(ctx, path, reader)
}

func (r *Builder) Delete(ctx context.Context, path string) error {
	return r.storage.Delete(ctx, path)
}

func (r *Builder) List(ctx context.Context, path string) ([]*oss.Object, error) {
	return r.storage.List(ctx, path)
}

func (r *Builder) GetURL(ctx context.Context, path string) (string, error) {
	return r.storage.GetURL(ctx, path)
}

func (r *Builder) GetEndpoint(ctx context.Context) string {
	return r.storage.GetEndpoint(ctx)
}

func (r *Builder) createEmptyTargetRecord(path string) {
	var (
		tx  = r.db.Begin()
		err error
		m   ObjectRedirection
	)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	tx.Where(`Source = ?`, path).Order("created_at desc").First(&m)
	if m.ID != 0 && m.Target == "" {
		return
	}
	err = tx.Create(&ObjectRedirection{
		Source: path,
		Target: "",
	}).Error
}

var (
	repeatedSourceErrors = errors.New("repeated source")
)

func (r *Builder) importRecords(ctx *web.EventContext) (err error) {
	var (
		uf   uploadFiles
		file multipart.File

		data           []map[string]string
		records        []ObjectRedirection
		existedSource  = make(map[string]bool)
		repeatedSource []string
	)
	ctx.MustUnmarshalForm(&uf)
	for _, fh := range uf.NewFiles {
		if file, err = fh.Open(); err != nil {
			return
		}
		if data, err = gocsv.CSVToMaps(file); err != nil {
			return
		}
		for _, val := range data {
			for k, v := range val {
				if existedSource[k] {
					repeatedSource = append(repeatedSource, v)
					continue
				}
				existedSource[k] = true
				records = append(records, ObjectRedirection{
					Source: k,
					Target: v,
				})
			}
		}
	}
	if len(repeatedSource) > 0 {
		return repeatedSourceErrors
	}
	return
}
