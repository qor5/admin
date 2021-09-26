package publish_test

import (
	"testing"
	"time"

	"github.com/qor/qor5/publish"
)

type Page struct {
	ID        uint
	Title     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	publish.Version
	publish.Schedule
	publish.Status
}

type Product struct {
	ID        uint
	Name      string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	publish.Version
	publish.Schedule
	publish.Status
}

func TestPublishContentToS3(t *testing.T) {
	p := publish.New() //.S3()

	err := p.Sync(&Product{}, &Page{})
	if err != nil {
		t.Error(err)
	}
}
