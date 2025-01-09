package integration_test

import (
	"embed"
	"os"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/stretchr/testify/require"
	"github.com/theplant/testenv"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/media/oss"
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

func checkFileExisted(t *testing.T, filename string) {
	var (
		file os.FileInfo
		err  error
	)
	if file, err = os.Stat("/tmp/media_test" + filename); err != nil {
		t.Fatalf("open file %s error %v", filename, err)
		return
	}
	if file.Size() == 0 {
		t.Fatalf("crop file %s, error %v", filename, file.Size())
		return
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
	moption.Sizes["small"] = &base.Size{
		Width:  100,
		Height: 100,
	}
	moption.Sizes["og"] = &base.Size{
		Width: 200,
	}
	moption.Sizes["large"] = &base.Size{
		Width: 400,
	}
	baseCropOption := base.CropOption{
		X:      6,
		Y:      20,
		Width:  40,
		Height: 40,
	}
	moption.CropOptions["default"] = &baseCropOption
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
	for _, name := range []string{"", "small", "og", "large"} {
		filename := m1.File.URL(name)
		if name == "" {
			filename = m1.File.URL()
		}
		checkFileExisted(t, filename)
	}
}

func TestURL(t *testing.T) {
	b := media_library.MediaBox{
		Url: "test.jpg",
	}
	if !strings.Contains(b.URLNoCached(), "test.jpg?") {
		t.Fatalf("set uncached url error %v", b.URLNoCached())
		return
	}
	b.Url = ""
	if b.URLNoCached() != "" {
		t.Fatalf("set uncached url empty error %v", b.URLNoCached())
		return
	}
	m := media_library.MediaLibrary{}
	m.File.Url = "test2.jpg"
	if !strings.Contains(m.File.URLNoCached(), "test2.jpg?") {
		t.Fatalf("set uncached url error %v", m.File.URLNoCached())
		return
	}
	m.File.Url = ""
	if m.File.URLNoCached() != "" {
		t.Fatalf("set uncached url empty error %v", m.File.URLNoCached())
		return
	}
	{
		b := oss.OSS{}
		b.Url = "test.jpg"
		require.Equal(t, "test.jpg", b.URL())
		require.Equal(t, "test.jpg", b.String())
		b.Url = ""
		require.Equal(t, "", b.URL())
		require.Equal(t, "", b.String())
	}
}
