package publish_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
	"github.com/stretchr/testify/require"
	"github.com/theplant/sliceutils"
	"github.com/theplant/testenv"
	"github.com/theplant/testingutils"
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

func (*Product) getListUrl() string {
	return "test/product/list/index.html"
}

func (p *Product) getListContent() string {
	return fmt.Sprintf("list page  %s", p.Code)
}

func (p *Product) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      p.getUrl(),
		Content:  p.getContent(),
		IsDelete: false,
	})
	p.OnlineUrl = p.getUrl()

	var liveRecord Product
	db.Where("id = ? AND status = ?", p.ID, publish.StatusOnline).First(&liveRecord)
	if liveRecord.ID == 0 {
		return
	}

	if liveRecord.OnlineUrl != p.OnlineUrl {
		actions = append(actions, &publish.PublishAction{
			Url:      liveRecord.getUrl(),
			IsDelete: true,
		})
	}

	if val, ok := ctx.Value(ctxKeySkipList{}).(bool); ok && val {
		return
	}
	actions = append(actions, &publish.PublishAction{
		Url:      p.getListUrl(),
		Content:  p.getListContent(),
		IsDelete: false,
	})
	return
}

func (p *Product) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      p.OnlineUrl,
		IsDelete: true,
	})
	if val, ok := ctx.Value(ctxKeySkipList{}).(bool); ok && val {
		return
	}
	actions = append(actions, &publish.PublishAction{
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

func (p *ProductWithoutVersion) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      p.getUrl(),
		Content:  p.getContent(),
		IsDelete: false,
	})

	if p.Status.Status == publish.StatusOnline && p.OnlineUrl != p.getUrl() {
		actions = append(actions, &publish.PublishAction{
			Url:      p.OnlineUrl,
			IsDelete: true,
		})
	}

	p.OnlineUrl = p.getUrl()
	return
}

func (p *ProductWithoutVersion) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	actions = append(actions, &publish.PublishAction{
		Url:      p.OnlineUrl,
		IsDelete: true,
	})
	return
}

func (ProductWithoutVersion) GetListUrl(pageNumber string) string {
	return fmt.Sprintf("/product_without_version/list/%v.html", pageNumber)
}

func (ProductWithoutVersion) GetListContent(db *gorm.DB, onePageItems *publish.OnePageItems) string {
	pageNumber := onePageItems.PageNumber
	var result string
	for _, item := range onePageItems.Items {
		record := item.(*ProductWithoutVersion)
		result = result + fmt.Sprintf("product:%v ", record.Name)
	}
	result = result + fmt.Sprintf("pageNumber:%v", pageNumber)
	return result
}

func (ProductWithoutVersion) Sort(array []interface{}) {
	var temp []*ProductWithoutVersion
	sliceutils.Unwrap(array, &temp)
	sort.Sort(SliceProductWithoutVersion(temp))
	for k, v := range temp {
		array[k] = v
	}
}

type SliceProductWithoutVersion []*ProductWithoutVersion

func (x SliceProductWithoutVersion) Len() int           { return len(x) }
func (x SliceProductWithoutVersion) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x SliceProductWithoutVersion) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type MockStorage struct {
	oss.StorageInterface
	Objects map[string]string
}

func (m *MockStorage) Get(ctx context.Context, path string) (f *os.File, err error) {
	content, exist := m.Objects[path]
	if !exist {
		err = fmt.Errorf("NoSuchKey: %s", path)
		return
	}

	pattern := fmt.Sprintf("s3*%d", time.Now().Unix())

	if f, err = os.CreateTemp("/tmp", pattern); err == nil {
		f.WriteString(content)
		f.Seek(0, 0)
	}
	return
}

func (m *MockStorage) Put(ctx context.Context, path string, r io.Reader) (*oss.Object, error) {
	fmt.Println("Calling mock s3 client - Put: ", path)
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	if m.Objects == nil {
		m.Objects = make(map[string]string)
	}
	m.Objects[path] = string(b)

	return &oss.Object{}, nil
}

func (m *MockStorage) Delete(ctx context.Context, path string) error {
	fmt.Println("Calling mock s3 client - Delete: ", path)
	delete(m.Objects, path)
	return nil
}

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)

	m.Run()
}

type ctxKeySkipList struct{}

type ProductWithError struct {
	Product
	publishActionError   error
	afterPublishError    error
	unpublschActionError error
	afterUnPublishError  error
}

func (*ProductWithError) TableName() string {
	return "products"
}

func (p *ProductWithError) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	if p.publishActionError != nil {
		return nil, p.publishActionError
	}
	return p.Product.GetPublishActions(ctx, db, storage)
}

func (p *ProductWithError) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	if p.unpublschActionError != nil {
		return nil, p.unpublschActionError
	}
	return p.Product.GetUnPublishActions(ctx, db, storage)
}

func (p *ProductWithError) AfterPublish(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) error {
	if p.afterPublishError != nil {
		return p.afterPublishError
	}
	return nil
}

func (p *ProductWithError) AfterUnPublish(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) error {
	if p.afterUnPublishError != nil {
		return p.afterUnPublishError
	}
	return nil
}

func TestPublishVersionContentToS3(t *testing.T) {
	db := TestDB
	db.AutoMigrate(&Product{})
	storage := &MockStorage{}

	productV1 := Product{
		Model:   gorm.Model{ID: 1},
		Code:    "0001",
		Name:    "coffee",
		Status:  publish.Status{Status: publish.StatusDraft},
		Version: publish.Version{Version: "v1"},
	}
	productV2 := Product{
		Model:   gorm.Model{ID: 1},
		Code:    "0002",
		Name:    "coffee",
		Status:  publish.Status{Status: publish.StatusDraft},
		Version: publish.Version{Version: "v2"},
	}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV1)
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV2)

	p := publish.New(db, storage)
	// publish v1
	skipListTrueContext := context.WithValue(context.Background(), ctxKeySkipList{}, true)
	skipListFalseContext := context.WithValue(context.Background(), ctxKeySkipList{}, false)
	if err := p.Publish(skipListTrueContext, &productV1); err != nil {
		t.Error(err)
	}
	assertUpdateStatus(t, db, &productV1, publish.StatusOnline, productV1.getUrl())
	assertUploadFile(t, productV1.getContent(), productV1.getUrl(), storage)
	// assertUploadFile(t, productV1.getListContent(), productV1.getListUrl(), storage)

	// publish v2
	if err := p.Publish(skipListFalseContext, &productV2); err != nil {
		t.Error(err)
	}
	assertUpdateStatus(t, db, &productV2, publish.StatusOnline, productV2.getUrl())
	assertUploadFile(t, productV2.getContent(), productV2.getUrl(), storage)
	assertUploadFile(t, productV2.getListContent(), productV2.getListUrl(), storage)
	// if delete v1 file
	assertContentDeleted(t, productV1.getUrl(), storage)
	// if update v1 status to offline
	assertUpdateStatus(t, db, &productV1, publish.StatusOffline, productV1.getUrl())

	// unpublish v2
	if err := p.UnPublish(skipListFalseContext, &productV2); err != nil {
		t.Error(err)
	}
	assertUpdateStatus(t, db, &productV2, publish.StatusOffline, productV2.getUrl())
	assertContentDeleted(t, productV2.getUrl(), storage)
	assertContentDeleted(t, productV2.getListUrl(), storage)

	err := p.Publish(skipListFalseContext, &ProductWithError{
		Product:            productV2,
		publishActionError: fmt.Errorf("publish error"),
	})
	require.ErrorContains(t, err, "publish error")
	assertUpdateStatus(t, db, &productV2, publish.StatusOffline, productV2.getUrl())

	err = p.Publish(skipListFalseContext, &ProductWithError{
		Product:           productV2,
		afterPublishError: fmt.Errorf("after publish error"),
	})
	require.ErrorContains(t, err, "after publish error")
	assertUpdateStatus(t, db, &productV2, publish.StatusOffline, productV2.getUrl())

	err = p.Publish(skipListFalseContext, &ProductWithError{
		Product: productV2,
	})
	require.NoError(t, err)
	assertUpdateStatus(t, db, &productV2, publish.StatusOnline, productV2.getUrl())

	err = p.UnPublish(skipListFalseContext, &ProductWithError{
		Product:              productV2,
		unpublschActionError: fmt.Errorf("unpublish error"),
	})
	require.ErrorContains(t, err, "unpublish error")
	assertUpdateStatus(t, db, &productV2, publish.StatusOnline, productV2.getUrl())

	err = p.UnPublish(skipListFalseContext, &ProductWithError{
		Product:             productV2,
		afterUnPublishError: fmt.Errorf("after unpublish error"),
	})
	require.ErrorContains(t, err, "after unpublish error")
	assertUpdateStatus(t, db, &productV2, publish.StatusOnline, productV2.getUrl())
}

func TestPublishList(t *testing.T) {
	db := TestDB
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

	publisher.Publish(context.Background(), &productV1)
	publisher.Publish(context.Background(), &productV3)
	if err := listPublisher.Run(context.Background(), ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	var expected string
	expected = "product:1 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"] != expected {
		t.Errorf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"])
	}

	publisher.Publish(context.Background(), &productV2)
	if err := listPublisher.Run(context.Background(), ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 product:2 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"] != expected {
		t.Errorf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"])
	}

	publisher.UnPublish(context.Background(), &productV2)
	if err := listPublisher.Run(context.Background(), ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 product:3 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"] != expected {
		t.Errorf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"])
	}

	publisher.UnPublish(context.Background(), &productV3)
	if err := listPublisher.Run(context.Background(), ProductWithoutVersion{}); err != nil {
		panic(err)
	}

	expected = "product:1 pageNumber:1"
	if storage.Objects["/product_without_version/list/1.html"] != expected {
		t.Errorf(`
want: %v
get: %v
`, expected, storage.Objects["/product_without_version/list/1.html"])
	}
}

type ProductWithSchedulePublisherDBScope struct {
	Product
}

func (ProductWithSchedulePublisherDBScope) TableName() string {
	return "products"
}

func (ProductWithSchedulePublisherDBScope) SchedulePublishDBScope(db *gorm.DB) *gorm.DB {
	return db.Where("name <> ?", "should_ignored_by_publish")
}

func (ProductWithSchedulePublisherDBScope) ScheduleUnPublishDBScope(db *gorm.DB) *gorm.DB {
	return db.Where("name <> ?", "should_ignored_by_unpublish")
}

func TestSchedulePublish(t *testing.T) {
	db := TestDB
	db.Migrator().DropTable(&Product{})
	db.AutoMigrate(&Product{})
	storage := &MockStorage{}

	productV1 := Product{
		Model:   gorm.Model{ID: 1},
		Version: publish.Version{Version: "2021-12-19-v01"},
		Code:    "1",
		Name:    "1",
		Status:  publish.Status{Status: publish.StatusDraft},
	}

	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&productV1)

	publisher := publish.New(db, storage)
	publisher.Publish(context.Background(), &productV1)

	var expected string
	expected = "11"
	if storage.Objects["test/product/1/index.html"] != expected {
		t.Errorf(`
	want: %v
	get: %v
	`, expected, storage.Objects["test/product/1/index.html"])
	}

	productV1.Name = "2"
	startAt := db.NowFunc().Add(-24 * time.Hour)
	productV1.ScheduledStartAt = &startAt
	if err := db.Save(&productV1).Error; err != nil {
		panic(err)
	}
	schedulePublisher := publish.NewSchedulePublishBuilder(publisher)
	if err := schedulePublisher.Run(context.Background(), productV1); err != nil {
		panic(err)
	}
	expected = "12"
	if storage.Objects["test/product/1/index.html"] != expected {
		t.Errorf(`
	want: %v
	get: %v
	`, expected, storage.Objects["test/product/1/index.html"])
	}

	endAt := startAt.Add(time.Second * 2)
	productV1.ScheduledEndAt = &endAt
	if err := db.Save(&productV1).Error; err != nil {
		panic(err)
	}
	if err := schedulePublisher.Run(context.Background(), productV1); err != nil {
		panic(err)
	}
	expected = ""
	if storage.Objects["test/product/1/index.html"] != expected {
		t.Errorf(`
	want: %v
	get: %v
	`, expected, storage.Objects["test/product/1/index.html"])
	}

	productV1.ScheduledEndAt = nil
	// expect ignored
	{
		productV1.Name = "should_ignored_by_publish"
		startAt := db.NowFunc().Add(-24 * time.Hour)
		productV1.ScheduledStartAt = &startAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(context.Background(), ProductWithSchedulePublisherDBScope{
			Product: productV1,
		})
		require.NoError(t, err)
		require.Equal(t, "", storage.Objects["test/product/1/index.html"])
	}
	// expect ok
	{
		productV1.Name = "3"
		startAt := db.NowFunc().Add(-24 * time.Hour)
		productV1.ScheduledStartAt = &startAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(context.Background(), ProductWithSchedulePublisherDBScope{
			Product: productV1,
		})
		require.NoError(t, err)
		require.Equal(t, "13", storage.Objects["test/product/1/index.html"])
	}
	productV1.ScheduledStartAt = nil
	// expect ignored
	{
		productV1.Name = "should_ignored_by_unpublish"
		endAt := startAt.Add(time.Second * 2)
		productV1.ScheduledEndAt = &endAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(context.Background(), ProductWithSchedulePublisherDBScope{
			Product: productV1,
		})
		require.NoError(t, err)
		require.Equal(t, "13", storage.Objects["test/product/1/index.html"])
	}
	// expect ok
	{
		productV1.Name = "3"
		endAt := startAt.Add(time.Second * 2)
		productV1.ScheduledEndAt = &endAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(context.Background(), ProductWithSchedulePublisherDBScope{
			Product: productV1,
		})
		require.NoError(t, err)
		require.Equal(t, "", storage.Objects["test/product/1/index.html"])
	}

	finderCtx := publish.WithScheduleRecordsFinder(context.Background(), func(ctx context.Context, operation publish.ScheduleOperation, b *publish.Builder, db *gorm.DB, records any) error {
		switch operation {
		case publish.ScheduleOperationPublish:
			if err := db.Where("scheduled_start_at <= ? AND name <> ?", db.NowFunc(), "should_ignored_by_finder").
				Order("scheduled_start_at").Find(records).Error; err != nil {
				return err
			}
		case publish.ScheduleOperationUnPublish:
			if err := db.Where("scheduled_end_at <= ? AND name <> ?", db.NowFunc(), "should_ignored_by_finder").
				Order("scheduled_end_at").Find(records).Error; err != nil {
				return err
			}
		}
		return nil
	})
	finderWithErrCtx := publish.WithScheduleRecordsFinder(context.Background(), func(ctx context.Context, operation publish.ScheduleOperation, b *publish.Builder, db *gorm.DB, records any) error {
		return fmt.Errorf("finder error")
	})

	productV1.ScheduledEndAt = nil
	{
		productV1.Name = "should_ignored_by_finder"
		startAt := db.NowFunc().Add(-24 * time.Hour)
		productV1.ScheduledStartAt = &startAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		// expect error
		err = schedulePublisher.Run(finderWithErrCtx, productV1)
		require.ErrorContains(t, err, "finder error")

		// expect ignored
		err = schedulePublisher.Run(finderCtx, productV1)
		require.NoError(t, err)
		require.Equal(t, "", storage.Objects["test/product/1/index.html"])

		// expect ok
		productV1.Name = "3"
		err = db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(finderCtx, productV1)
		require.NoError(t, err)
		require.Equal(t, "13", storage.Objects["test/product/1/index.html"])
	}

	productV1.ScheduledStartAt = nil
	{
		productV1.Name = "should_ignored_by_finder"
		endAt := startAt.Add(time.Second * 2)
		productV1.ScheduledEndAt = &endAt

		err := db.Save(&productV1).Error
		require.NoError(t, err)

		// expect error
		err = schedulePublisher.Run(finderWithErrCtx, productV1)
		require.ErrorContains(t, err, "finder error")

		// expect ignored
		err = schedulePublisher.Run(finderCtx, productV1)
		require.NoError(t, err)
		require.Equal(t, "13", storage.Objects["test/product/1/index.html"])

		// expect ok
		productV1.Name = "3"
		err = db.Save(&productV1).Error
		require.NoError(t, err)

		err = schedulePublisher.Run(finderCtx, productV1)
		require.NoError(t, err)
		require.Equal(t, "", storage.Objects["test/product/1/index.html"])
	}
}

func TestPublishContentWithoutVersionToS3(t *testing.T) {
	db := TestDB
	db.AutoMigrate(&ProductWithoutVersion{})
	storage := &MockStorage{}

	product1 := ProductWithoutVersion{
		Model:  gorm.Model{ID: 1},
		Code:   "0001",
		Name:   "tea",
		Status: publish.Status{Status: publish.StatusDraft},
	}
	db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&product1)
	ctx := context.Background()

	p := publish.New(db, storage)
	// publish product1
	if err := p.Publish(ctx, &product1); err != nil {
		t.Error(err)
	}
	assertNoVersionUpdateStatus(t, db, &product1, publish.StatusOnline, product1.getUrl())
	assertUploadFile(t, product1.getContent(), product1.getUrl(), storage)

	product1Clone := product1
	product1Clone.Code = "0002"

	// publish product1 again
	if err := p.Publish(ctx, &product1Clone); err != nil {
		t.Error(err)
	}
	assertNoVersionUpdateStatus(t, db, &product1Clone, publish.StatusOnline, product1Clone.getUrl())

	assertUploadFile(t, product1Clone.getContent(), product1Clone.getUrl(), storage)

	// if delete product1 old file
	assertContentDeleted(t, product1.getUrl(), storage)
	// unpublish product1
	if err := p.UnPublish(ctx, &product1Clone); err != nil {
		t.Error(err)
	}
	assertNoVersionUpdateStatus(t, db, &product1Clone, publish.StatusOffline, product1Clone.getUrl())
	// if delete product1 file
	assertContentDeleted(t, product1Clone.getUrl(), storage)

	{
		// wrap publish
		var spends time.Duration
		warpperCalled := false
		p.WrapPublish(func(in publish.PublishFunc) publish.PublishFunc {
			return func(ctx context.Context, record any) error {
				start := time.Now()
				defer func() {
					spends = time.Since(start)
					warpperCalled = true
				}()
				return in(ctx, record)
			}
		})
		require.NoError(t, p.Publish(ctx, &product1Clone))
		require.True(t, spends > 0)
		require.True(t, warpperCalled)
	}

	{
		// wrap unpublish
		var spends time.Duration
		warpperCalled := false
		p.WrapUnPublish(func(in publish.UnPublishFunc) publish.UnPublishFunc {
			return func(ctx context.Context, record any) error {
				start := time.Now()
				defer func() {
					spends = time.Since(start)
					warpperCalled = true
				}()
				return in(ctx, record)
			}
		})
		require.NoError(t, p.UnPublish(ctx, &product1Clone))
		require.True(t, spends > 0)
		require.True(t, warpperCalled)
	}
}

func assertUpdateStatus(t *testing.T, db *gorm.DB, p *Product, assertStatus, asserOnlineUrl string) {
	t.Helper()

	var pindb Product
	err := db.Model(&Product{}).Where("id = ? AND version = ?", p.ID, p.Version.Version).First(&pindb).Error
	if err != nil {
		t.Fatal(err)
	}

	diff := testingutils.PrettyJsonDiff(publish.Status{
		Status:    assertStatus,
		OnlineUrl: asserOnlineUrl,
	}, pindb.Status)
	if diff != "" {
		t.Error(diff)
	}
}

func assertContentDeleted(t *testing.T, url string, storage oss.StorageInterface) {
	t.Helper()

	t.Helper()
	_, err := storage.Get(context.Background(), url)
	if err == nil {
		t.Errorf("content for %s should be deleted", url)
	}
}

func assertNoVersionUpdateStatus(t *testing.T, db *gorm.DB, p *ProductWithoutVersion, assertStatus, asserOnlineUrl string) {
	var pindb ProductWithoutVersion
	err := db.Model(&ProductWithoutVersion{}).Where("id = ?", p.ID).First(&pindb).Error
	if err != nil {
		t.Fatal(err)
	}
	diff := testingutils.PrettyJsonDiff(publish.Status{
		Status:    assertStatus,
		OnlineUrl: asserOnlineUrl,
	}, pindb.Status)
	if diff != "" {
		t.Error(diff)
	}
}

func assertUploadFile(t *testing.T, content, url string, storage oss.StorageInterface) {
	t.Helper()
	f, err := storage.Get(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	c, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	diff := testingutils.PrettyJsonDiff(content, string(c))
	if diff != "" {
		t.Error(diff)
	}
}
