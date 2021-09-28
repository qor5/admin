package publish_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/publish"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Page struct {
	ID        uint
	Title     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	publish.Status
}

type Product struct {
	gorm.Model
	Name string
	Code string

	publish.Version
	publish.Schedule
	publish.Status
}

func (p *Product) getContent() string {
	return p.Code + p.Name
}

func (p *Product) getUrl() string {
	return fmt.Sprintf("test/%s/index.html", p.Code)
}

func (p *Product) GetPublishActions(db *gorm.DB) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.getUrl(),
		Content:  p.getContent(),
		IsDelete: false,
	})
	p.SetOnlineID(p.Code)

	var liveRecord Product
	db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline).First(&liveRecord)
	if liveRecord.ID == 0 {
		return
	}

	if liveRecord.GetOnlineID() != p.GetOnlineID() {
		objs = append(objs, &publish.PublishAction{
			Url:      liveRecord.getUrl(),
			IsDelete: true,
		})
	}

	return
}
func (p *Product) GetUnPublishActions(db *gorm.DB) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.getUrl(),
		IsDelete: true,
	})
	return
}

type ProductWithoutVersion struct {
	ID        uint
	Title     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	publish.Status
}

func ConnectDB() (db *gorm.DB) {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Info)
	return
}

func TestPublishContentToS3(t *testing.T) {
	db := ConnectDB()
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: session.Must(session.NewSession()),
	})
	db.AutoMigrate(&Product{}, &ProductWithoutVersion{})

	productV1 := Product{
		Model: gorm.Model{
			ID: 1,
		},
		Code: "0001",
		Name: "coffee",
		Status: publish.Status{
			Status: publish.StatusDraft,
		},
		Version: publish.Version{
			VersionName: "v1",
		},
	}
	productV2 := Product{
		Model: gorm.Model{
			ID: 1,
		},
		Code: "0002",
		Name: "coffee",
		Status: publish.Status{
			Status: publish.StatusDraft,
		},
		Version: publish.Version{
			VersionName: "v2",
		},
	}

	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&productV1)
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&productV2)

	p := publish.New(db, storage)

	// publish v1
	if err := p.Publish(&productV1); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV1, publish.StatusOnline); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV1, storage); err != nil {
		t.Error(err)
	}

	// publish v2
	if err := p.Publish(&productV2); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV2, publish.StatusOnline); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV2, storage); err != nil {
		t.Error(err)
	}
	// if delete v1 file
	if err := assertUploadFile(&productV1, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", productV1.getUrl())))
	}
	// if update v1 status to offline
	if err := assertUpdateStatus(db, &productV1, publish.StatusOffline); err != nil {
		t.Error(err)
	}

	// unpublish v2
	if err := p.UnPublish(&productV2); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV2, publish.StatusOffline); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV2, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", productV2.getUrl())))
	}

}
func assertUpdateStatus(db *gorm.DB, p *Product, assertStatus string) (err error) {
	var pindb Product
	err = db.Model(&Product{}).Where("id = ? AND version_name = ?", p.ID, p.VersionName).First(&pindb).Error
	if err != nil {
		return err
	}
	if pindb.GeStatus() != assertStatus {
		return errors.New("update status failed")
	}
	return
}
func assertUploadFile(p *Product, storage oss.StorageInterface) error {
	f, err := storage.Get(p.getUrl())
	if err != nil {
		return err
	}
	c, err := ioutil.ReadAll(f)
	if string(c) != p.getContent() {
		return errors.New("wrong content")
	}
	return nil
}
