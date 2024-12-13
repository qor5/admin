package vips

import (
	"bytes"
	"io"
	"path"
	"strings"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/imaging"
	"github.com/theplant/bimg"
)

var (
	EnableGenerateWebp = false
	PNGtoWebpQuality   = 90
	JPEGtoWebpQuality  = 80
	GIFtoWebpQuality   = 85
	JPEGQuality        = 80
	PNGQuality         = 90
	PNGCompression     = 9
)

type Config struct {
	EnableGenerateWebp bool
	PNGtoWebpQuality   int
	JPEGtoWebpQuality  int
	JPEGQuality        int
	PNGQuality         int
	PNGCompression     int
}

type bimgImageHandler struct{}

func (bimgImageHandler) CouldHandle(media base.Media) bool {
	return media.IsImage()
}

// Crop & Resize
func (bimgImageHandler) Handle(m base.Media, file base.FileInterface, option *base.Option) (err error) {
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, file); err != nil {
		return err
	}
	fileSizes := m.GetFileSizes()
	fileSizes["original"] = buffer.Len()

	// Auto Rotate by EXIF
	img := copyImage(buffer.Bytes())
	metaData, _ := img.Metadata()
	base.SetWeightHeight(m, metaData.Size.Width, metaData.Size.Height)
	if metaData.EXIF.Orientation > 1 {
		rotatedBuf, err := img.AutoRotate()
		if err != nil {
			return err
		}
		buffer.Reset()
		if _, err := io.Copy(buffer, bytes.NewBuffer(rotatedBuf)); err != nil {
			return err
		}
	}

	// Save Original Image
	{
		if err = m.Store(m.URL("original"), option, bytes.NewReader(buffer.Bytes())); err != nil {
			return err
		}

		img := copyImage(buffer.Bytes())
		if err = generateWebp(m, option, bimg.Options{}, img, "original"); err != nil {
			return err
		}
	}

	// TODO support crop gif
	if isGif(m.URL()) {
		if err = m.Store(m.URL(), option, file); err != nil {
			return err
		}
		img := copyImage(buffer.Bytes())
		bimgOption := bimg.Options{Palette: true, Compression: PNGCompression}
		if err = generateWebp(m, option, bimgOption, img); err != nil {
			return err
		}
		for key := range m.GetSizes() {
			if key == base.DefaultSizeKey {
				continue
			}
			img := copyImage(buffer.Bytes())
			if err = m.Store(m.URL(key), option, file); err != nil {
				return err
			}
			fileSizes[key] = buffer.Len()
			if err = generateWebp(m, option, bimgOption, img, key); err != nil {
				return err
			}
		}
		base.SetFileSizes(m, fileSizes)
		return
	}

	quality := getQualityByImageType(m.URL())

	// Handle default image
	{
		img := copyImage(buffer.Bytes())
		bimgOption := bimg.Options{Quality: quality, Palette: true, Compression: PNGCompression}
		// Crop original image if specified
		if cropOption := m.GetCropOption(base.DefaultSizeKey); cropOption != nil {
			options := bimg.Options{
				Quality:    100, // Don't compress twice
				Top:        cropOption.Min.Y,
				Left:       cropOption.Min.X,
				AreaWidth:  cropOption.Max.X - cropOption.Min.X,
				AreaHeight: cropOption.Max.Y - cropOption.Min.Y,
			}
			if options.Top == 0 && options.Left == 0 {
				options.Top = -1
			}
			if _, err := img.Process(options); err != nil {
				return err
			}
		}
		copy := copyImage(img.Image())
		if buf, err := img.Process(bimgOption); err == nil {
			if err = m.Store(m.URL(), option, bytes.NewReader(buf)); err != nil {
				return err
			}
			fileSizes[base.DefaultSizeKey] = len(buf)
		} else {
			return err
		}
		if err = generateWebp(m, option, bimgOption, copy); err != nil {
			return err
		}
	}

	// Handle size images
	for key, size := range m.GetSizes() {
		if key == base.DefaultSizeKey {
			continue
		}
		img := copyImage(buffer.Bytes())
		if cropOption := m.GetCropOption(key); cropOption != nil {
			options := bimg.Options{
				Quality:    100, // Don't compress twice
				Top:        cropOption.Min.Y,
				Left:       cropOption.Min.X,
				AreaWidth:  cropOption.Max.X - cropOption.Min.X,
				AreaHeight: cropOption.Max.Y - cropOption.Min.Y,
			}
			if options.Top == 0 && options.Left == 0 {
				options.Top = -1
			}
			if _, err := img.Process(options); err != nil {
				return err
			}
		}
		copy := copyImage(img.Image())
		bimgOption := bimg.Options{
			Width:       size.Width,
			Height:      size.Height,
			Quality:     quality,
			Compression: PNGCompression,
			Palette:     true,
			Enlarge:     true,
		}
		// Process & Save size image
		if buf, err := img.Process(bimgOption); err == nil {
			if err = m.Store(m.URL(key), option, bytes.NewReader(buf)); err != nil {
				return err
			}
			fileSizes[key] = len(buf)
		} else {
			return err
		}
		if err = generateWebp(m, option, bimgOption, copy, key); err != nil {
			return err
		}
	}
	base.SetFileSizes(m, fileSizes)
	return
}

func generateWebp(m base.Media, option *base.Option, bimgOption bimg.Options, img *bimg.Image, size ...string) (err error) {
	if !EnableGenerateWebp {
		return
	}
	bimgOption.Type = bimg.WEBP
	bimgOption.Quality = getWebpQualityByImageType(m.URL())
	if buf, err := img.Process(bimgOption); err == nil {
		url := m.URL(size...)
		ext := path.Ext(url)
		extArr := strings.Split(ext, "?")
		i := strings.LastIndex(url, ext)
		webpUrl := url[:i] + strings.Replace(url[i:], extArr[0], ".webp", 1)
		m.Store(webpUrl, option, bytes.NewReader(buf))
	} else {
		return err
	}
	return
}

func copyImage(buffer []byte) (img *bimg.Image) {
	bs := make([]byte, len(buffer))
	copy(bs, buffer)
	img = bimg.NewImage(bs)
	return
}

func getQualityByImageType(url string) int {
	imgType, err := base.GetImageFormat(url)
	if err != nil {
		return 0
	}
	switch *imgType {
	case imaging.JPEG:
		return JPEGQuality
	case imaging.PNG:
		return PNGQuality
	}
	return 0
}

func getWebpQualityByImageType(url string) int {
	imgType, err := base.GetImageFormat(url)
	if err != nil {
		return 0
	}
	switch *imgType {
	case imaging.JPEG:
		return JPEGtoWebpQuality
	case imaging.PNG:
		return PNGtoWebpQuality
	case imaging.GIF:
		return GIFtoWebpQuality
	}
	return 0
}

func isGif(url string) bool {
	imgType, err := base.GetImageFormat(url)
	return err == nil && *imgType == imaging.GIF
}

func UseVips(cfg Config) {
	if cfg.EnableGenerateWebp {
		EnableGenerateWebp = true
	}
	if cfg.JPEGtoWebpQuality > 0 && cfg.JPEGtoWebpQuality <= 100 {
		JPEGtoWebpQuality = cfg.JPEGtoWebpQuality
	}
	if cfg.PNGtoWebpQuality > 0 && cfg.PNGtoWebpQuality <= 100 {
		PNGtoWebpQuality = cfg.PNGtoWebpQuality
	}
	if cfg.JPEGQuality > 0 && cfg.JPEGQuality <= 100 {
		JPEGQuality = cfg.JPEGQuality
	}
	if cfg.PNGQuality > 0 && cfg.PNGQuality <= 100 {
		PNGQuality = cfg.PNGQuality
	}
	if cfg.PNGCompression > 0 && cfg.PNGCompression <= 9 {
		PNGCompression = cfg.PNGCompression
	}
	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)
	base.RegisterMediaHandler("image_handler", bimgImageHandler{})
}
