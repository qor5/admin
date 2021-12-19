package publish_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/publish"
	"github.com/theplant/sliceutils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Product struct {
	gorm.Model
	Name string
	Code string

	publish.Version
	publish.Schedule
	publish.Status
	publish.List
}

func (p *Product) getContent() string {
	return p.Code + p.Name + p.VersionName
}

func (p *Product) getUrl() string {
	return fmt.Sprintf("test/product/%s/index.html", p.Code)
}

func (p *Product) getListUrl() string {
	return "test/product/list/index.html"
}

func (p *Product) getListContent() string {
	return fmt.Sprintf("list page  %s", p.Code)
}

func (p *Product) GetPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.getUrl(),
		Content:  p.getContent(),
		IsDelete: false,
	})
	p.SetOnlineUrl(p.getUrl())

	var liveRecord Product
	db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline).First(&liveRecord)
	if liveRecord.ID == 0 {
		return
	}

	if liveRecord.GetOnlineUrl() != p.GetOnlineUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      liveRecord.getUrl(),
			IsDelete: true,
		})
	}

	if val, ok := ctx.Value("skip_list").(bool); ok && val {
		return
	}
	objs = append(objs, &publish.PublishAction{
		Url:      p.getListUrl(),
		Content:  p.getListContent(),
		IsDelete: false,
	})
	return
}
func (p *Product) GetUnPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.GetOnlineUrl(),
		IsDelete: true,
	})
	if val, ok := ctx.Value("skip_list").(bool); ok && val {
		return
	}
	objs = append(objs, &publish.PublishAction{
		Url:      p.getListUrl(),
		IsDelete: true,
	})
	return
}

type ProductWithoutVersion struct {
	gorm.Model
	Name string
	Code string

	publish.Status
	publish.List
}

func (p *ProductWithoutVersion) getContent() string {
	return p.Code + p.Name
}

func (p *ProductWithoutVersion) getUrl() string {
	return fmt.Sprintf("test/product_no_version/%s/index.html", p.Code)
}

func (p *ProductWithoutVersion) GetPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.getUrl(),
		Content:  p.getContent(),
		IsDelete: false,
	})

	if p.GetStatus() == publish.StatusOnline && p.GetOnlineUrl() != p.getUrl() {
		objs = append(objs, &publish.PublishAction{
			Url:      p.GetOnlineUrl(),
			IsDelete: true,
		})
	}

	p.SetOnlineUrl(p.getUrl())
	return
}
func (p *ProductWithoutVersion) GetUnPublishActions(db *gorm.DB, ctx context.Context) (objs []*publish.PublishAction) {
	objs = append(objs, &publish.PublishAction{
		Url:      p.GetOnlineUrl(),
		IsDelete: true,
	})
	return
}

func (this ProductWithoutVersion) GetListUrl(pageNumber string) string {
	return fmt.Sprintf("/product_without_version/list/%v.html", pageNumber)
}

func (this ProductWithoutVersion) GetListContent(db *gorm.DB, onePageItems *publish.OnePageItems) string {
	pageNumber := onePageItems.PageNumber
	var result string
	for _, item := range onePageItems.Items {
		record := item.(*ProductWithoutVersion)
		result = result + fmt.Sprintf("product:%v ", record.Name)
	}
	result = result + fmt.Sprintf("pageNumber:%v", pageNumber)
	return result
}

func (this ProductWithoutVersion) Sort(array []interface{}) {
	var temp []*ProductWithoutVersion
	sliceutils.Unwrap(array, &temp)
	sort.Sort(SliceProductWithoutVersion(temp))
	for k, v := range temp {
		array[k] = v
	}
	return
}

type SliceProductWithoutVersion []*ProductWithoutVersion

func (x SliceProductWithoutVersion) Len() int           { return len(x) }
func (x SliceProductWithoutVersion) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x SliceProductWithoutVersion) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type MockStorage struct {
	oss.StorageInterface
	Objects        map[string][]string
	DeletedObjects map[string]bool
}

func (m *MockStorage) Get(path string) (*os.File, error) {
	return nil, errors.New("no file")
}

func (m *MockStorage) Put(path string, r io.Reader) (*oss.Object, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	if m.Objects == nil {
		m.Objects = make(map[string][]string)
	}
	if _, ok := m.Objects[path]; !ok {
		m.Objects[path] = []string{}
	}
	m.Objects[path] = append([]string{string(b)}, m.Objects[path]...)

	return &oss.Object{}, nil
}

func (m *MockStorage) Delete(path string) error {
	fmt.Println("Calling mock s3 client - Delete: ", path)
	if m.DeletedObjects == nil {
		m.DeletedObjects = make(map[string]bool)
	}
	m.DeletedObjects[path] = true
	return nil
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

func TestPublishVersionContentToS3(t *testing.T) {
	db := ConnectDB()
	db.AutoMigrate(&Product{})
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: session.Must(session.NewSession()),
	})

	productV1 := Product{
		Model:   gorm.Model{ID: 1},
		Code:    "0001",
		Name:    "coffee",
		Status:  publish.Status{Status: publish.StatusDraft},
		Version: publish.Version{VersionName: "v1"},
	}
	productV2 := Product{
		Model:   gorm.Model{ID: 1},
		Code:    "0002",
		Name:    "coffee",
		Status:  publish.Status{Status: publish.StatusDraft},
		Version: publish.Version{VersionName: "v2"},
	}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV1)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV2)

	p := publish.New(db, storage)

	// publish v1
	if err := p.WithValue("skip_list", true).Publish(&productV1); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV1, publish.StatusOnline, productV1.getUrl()); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV1, storage); err != nil {
		t.Error(err)
	}
	if err := assertUploadListFile(&productV1, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New("skip_list failed"))
	}

	// publish v2
	if err := p.WithValue("skip_list", false).Publish(&productV2); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV2, publish.StatusOnline, productV2.getUrl()); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV2, storage); err != nil {
		t.Error(err)
	}
	if err := assertUploadListFile(&productV2, storage); err != nil {
		t.Error(err)
	}
	// if delete v1 file
	if err := assertUploadFile(&productV1, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", productV1.getUrl())))
	}
	// if update v1 status to offline
	if err := assertUpdateStatus(db, &productV1, publish.StatusOffline, productV1.getUrl()); err != nil {
		t.Error(err)
	}

	// unpublish v2
	if err := p.UnPublish(&productV2); err != nil {
		t.Error(err)
	}
	if err := assertUpdateStatus(db, &productV2, publish.StatusOffline, productV2.getUrl()); err != nil {
		t.Error(err)
	}
	if err := assertUploadFile(&productV2, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", productV2.getUrl())))
	}
	if err := assertUploadListFile(&productV2, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New("delete list file %s failed"), productV2.getListUrl())
	}
}

func TestPublishList(t *testing.T) {
	db := ConnectDB()
	db.AutoMigrate(&ProductWithoutVersion{})
	storage := &MockStorage{}

	productV1 := ProductWithoutVersion{
		Model:  gorm.Model{ID: 1},
		Code:   "1",
		Name:   "1",
		Status: publish.Status{Status: publish.StatusDraft},
	}
	productV2 := ProductWithoutVersion{
		Model:  gorm.Model{ID: 2},
		Code:   "2",
		Name:   "2",
		Status: publish.Status{Status: publish.StatusDraft},
	}

	productV3 := ProductWithoutVersion{
		Model:  gorm.Model{ID: 3},
		Code:   "3",
		Name:   "3",
		Status: publish.Status{Status: publish.StatusDraft},
	}

	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV1)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV2)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV3)

	publisher := publish.New(db, storage)
	listPublisher := publish.NewListPublishBuilder(db, storage)

	publisher.Publish(&productV1)
	publisher.Publish(&productV3)
	if err := listPublisher.Run(ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	var expected string
	expected = "product:1 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"][0] != expected {
		t.Error(errors.New(fmt.Sprintf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"][0])))
	}

	publisher.Publish(&productV2)
	if err := listPublisher.Run(ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 product:2 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"][0] != expected {
		t.Error(errors.New(fmt.Sprintf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"][0])))
	}

	publisher.UnPublish(&productV2)
	if err := listPublisher.Run(ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"][0] != expected {
		t.Error(errors.New(fmt.Sprintf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"][0])))
	}

	publisher.UnPublish(&productV3)
	if err := listPublisher.Run(ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"][0] != expected {
		t.Error(errors.New(fmt.Sprintf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"][0])))
	}
}

func TestPublishContentWithoutVersionToS3(t *testing.T) {
	db := ConnectDB()
	db.AutoMigrate(&ProductWithoutVersion{})
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: session.Must(session.NewSession()),
	})

	product1 := ProductWithoutVersion{
		Model:  gorm.Model{ID: 1},
		Code:   "0001",
		Name:   "tea",
		Status: publish.Status{Status: publish.StatusDraft},
	}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&product1)

	p := publish.New(db, storage)
	// publish product1
	if err := p.Publish(&product1); err != nil {
		t.Error(err)
	}
	if err := assertNoVersionUpdateStatus(db, &product1, publish.StatusOnline, product1.getUrl()); err != nil {
		t.Error(err)
	}
	if err := assertNoVersionUploadFile(&product1, storage); err != nil {
		t.Error(err)
	}

	product1Clone := product1
	product1Clone.Code = "0002"

	// publish product1 again
	if err := p.Publish(&product1Clone); err != nil {
		t.Error(err)
	}
	if err := assertNoVersionUpdateStatus(db, &product1Clone, publish.StatusOnline, product1Clone.getUrl()); err != nil {
		t.Error(err)
	}
	if err := assertNoVersionUploadFile(&product1Clone, storage); err != nil {
		t.Error(err)
	}
	// if delete product1 old file
	if err := assertNoVersionUploadFile(&product1, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", product1.getUrl())))
	}

	// unpublish product1
	if err := p.UnPublish(&product1Clone); err != nil {
		t.Error(err)
	}
	if err := assertNoVersionUpdateStatus(db, &product1Clone, publish.StatusOffline, product1Clone.getUrl()); err != nil {
		t.Error(err)
	}
	// if delete product1 file
	if err := assertNoVersionUploadFile(&product1Clone, storage); err != nil && strings.HasPrefix(err.Error(), "NoSuchKey: The specified key does not exist") {
	} else {
		t.Error(errors.New(fmt.Sprintf("delete file %s failed", product1Clone.getUrl())))
	}

}
func assertUpdateStatus(db *gorm.DB, p *Product, assertStatus string, asserOnlineUrl string) (err error) {
	var pindb Product
	err = db.Model(&Product{}).Where("id = ? AND version_name = ?", p.ID, p.VersionName).First(&pindb).Error
	if err != nil {
		return err
	}
	if pindb.GetStatus() != assertStatus || pindb.GetOnlineUrl() != asserOnlineUrl {
		return errors.New("update status failed")
	}
	return
}

func assertNoVersionUpdateStatus(db *gorm.DB, p *ProductWithoutVersion, assertStatus string, asserOnlineUrl string) (err error) {
	var pindb ProductWithoutVersion
	err = db.Model(&ProductWithoutVersion{}).Where("id = ?", p.ID).First(&pindb).Error
	if err != nil {
		return err
	}
	if pindb.GetStatus() != assertStatus || pindb.GetOnlineUrl() != asserOnlineUrl {
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

func assertUploadListFile(p *Product, storage oss.StorageInterface) error {
	f, err := storage.Get(p.getListUrl())
	if err != nil {
		return err
	}
	c, err := ioutil.ReadAll(f)
	if string(c) != p.getListContent() {
		return errors.New("wrong content")
	}
	return nil
}

func assertNoVersionUploadFile(p *ProductWithoutVersion, storage oss.StorageInterface) error {
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
