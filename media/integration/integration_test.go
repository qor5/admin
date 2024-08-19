package integration_test

import (
	"embed"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/web/v3"
	"os"
	"strings"
	"testing"

	"github.com/qor/oss/filesystem"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

//go:embed *.png
var box embed.FS

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	m.Run()
}

func setup() (db *gorm.DB) {
	var err error
	db = TestDB

	db = db.Debug()
	// db.Logger = db.Logger.LogMode(logger.Info)

	if err = db.AutoMigrate(
		&media_library.MediaLibrary{},
	); err != nil {
		panic(err)
	}

	oss.Storage = filesystem.New("/tmp/media_test")

	return
}

func TestUpload(t *testing.T) {
	db := setup()
	f, err := box.ReadFile("testfile.png")
	if err != nil {
		panic(err)
	}

	fh := multipartestutils.CreateMultipartFileHeader("test.png", f)
	m := media_library.MediaLibrary{}

	err = m.File.Scan(fh)
	if err != nil {
		t.Fatal(err)
	}

	err = base.SaveUploadAndCropImage(db, &m, "", &web.EventContext{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCrop(t *testing.T) {
	db := setup()
	f, err := box.ReadFile("testfile.png")
	if err != nil {
		panic(err)
	}

	fh := multipartestutils.CreateMultipartFileHeader("test.png", f)
	m := media_library.MediaLibrary{}
	m1 := media_library.MediaLibrary{}

	err = m.File.Scan(fh)
	if err != nil {
		t.Fatal(err)
	}
	err = base.SaveUploadAndCropImage(db, &m, "", &web.EventContext{})
	if err != nil {
		t.Fatal(err)
	}
	TestDB.Order("id desc").First(&m1)

	moption := m1.GetMediaOption()
	if moption.CropOptions == nil {
		moption.CropOptions = make(map[string]*base.CropOption)
	}
	moption.CropOptions["default"] = &base.CropOption{
		X:      6,
		Y:      20,
		Width:  40,
		Height: 40,
	}
	moption.Crop = true
	err = m1.ScanMediaOptions(moption)
	if err != nil {
		t.Fatal(err)
	}

	err = base.SaveUploadAndCropImage(db, &m1, "", &web.EventContext{})
	if err != nil {
		t.Fatal(err)
		return
	}
	var file os.FileInfo
	if file, err = os.Stat("/tmp/media_test" + m1.File.URL("original")); err != nil {
		t.Fatalf("open file error %v", err)
		return
	}
	if file.Size() < 2<<10 {
		t.Fatalf("crop file error %v", file.Size())
		return
	}
	if file, err = os.Stat("/tmp/media_test" + m1.File.URL()); err != nil {
		t.Fatalf("open file error %v", err)
		return
	}
	if file.Size() == 0 {
		t.Fatalf("crop file error %v", file.Size())
		return
	}
}
func TestCopy(t *testing.T) {
	db := setup()
	f, err := box.ReadFile("testfile.png")
	if err != nil {
		panic(err)
	}
	mb := media.New(db)
	fh := multipartestutils.CreateMultipartFileHeader("test.png", f)
	m := media_library.MediaLibrary{}

	err = m.File.Scan(fh)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = base.SaveUploadAndCropImage(db, &m, "", &web.EventContext{})
	if err != nil {
		t.Fatal(err)
		return
	}
	oldID := m.ID
	oldCreatedTime := m.CreatedAt
	if m, err = media.CopyMediaLiMediaLibrary(mb, db, int(oldID), &web.EventContext{}); err != nil {
		t.Fatalf("copy error :%v", err)
		return
	}
	if oldID == m.ID {
		t.Fatalf("copy failed")
		return
	}
	var file os.FileInfo
	if file, err = os.Stat("/tmp/media_test" + m.File.URL()); err != nil {
		t.Fatalf("open file error %v", err)
		return
	}
	if file.Size() == 0 {
		t.Fatalf("crop file error %v", file.Size())
		return
	}
	if m.CreatedAt == oldCreatedTime {
		t.Fatalf("crop file time error  %v :%v", m.CreatedAt, oldCreatedTime)
		return
	}

}
func TestUnCachedURL(t *testing.T) {
	b := media_library.MediaBox{
		Url: "test.jpg",
	}
	if !strings.Contains(b.URLNoCached(), "test.jpg?") {
		t.Fatalf("set uncached url error %v", b.URLNoCached())
		return
	}
	m := media_library.MediaLibrary{}
	m.File.Url = "test2.jpg"
	if !strings.Contains(m.File.URLNoCached(), "test2.jpg?") {
		t.Fatalf("set uncached url error %v", m.File.URLNoCached())
		return
	}
}
