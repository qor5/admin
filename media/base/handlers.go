package base

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"io"
	"math"

	"github.com/qor5/imaging"
)

var (
	mediaHandlers   = make(map[string]MediaHandler)
	DefaultSizeKey  = "default"
	OriginalSizeKey = "original"
)

// MediaHandler media library handler interface, defined which files could be handled, and the handler
type MediaHandler interface {
	CouldHandle(media Media) bool
	Handle(media Media, file FileInterface, option *Option) error
}

// RegisterMediaHandler register Media library handler
func RegisterMediaHandler(name string, handler MediaHandler) {
	mediaHandlers[name] = handler
}

// imageHandler default image handler
type imageHandler struct{}

func (imageHandler) CouldHandle(media Media) bool {
	return media.IsImage()
}

func resizeImageTo(img image.Image, size *Size, format imaging.Format) image.Image {
	imgSize := img.Bounds().Size()
	SaleUpDown(imgSize.X, imgSize.Y, size)
	switch {
	case size.Padding:
		var (
			backgroundColor = image.NewUniform(color.Transparent)
			ratioX          = float64(size.Width) / float64(imgSize.X)
			ratioY          = float64(size.Height) / float64(imgSize.Y)
			// 100x200 -> 200x300  ==>  ratioX = 2,   ratioY = 1.5  ==> resize to (x1.5) = 150x300
			// 100x200 -> 20x50    ==>  ratioX = 0.2, ratioY = 0.4  ==> resize to (x0.2) = 20x40
			// 100x200 -> 50x0     ==>  ratioX = 0.5, ratioY = 0    ==> resize to (x0.5) = 50x100
			minRatio = math.Min(ratioX, ratioY)
		)

		if format == imaging.JPEG {
			backgroundColor = image.NewUniform(color.White)
		}

		background := imaging.New(size.Width, size.Height, backgroundColor)
		fixFloat := func(x float64, y int) int {
			if math.Abs(x-float64(y)) < 1 {
				return y
			}
			return int(x)
		}

		if minRatio == 0 {
			minRatio = math.Max(ratioX, ratioY)

			if size.Width == 0 && size.Height != 0 {
				// size 50x0, source 100x200 => crop to 50x100
				newWidth := int(float64(imgSize.X) / float64(imgSize.Y) * float64(size.Height))
				background = imaging.New(newWidth, size.Height, backgroundColor)
			} else if size.Height == 0 && size.Width != 0 {
				// size 0x50, source 100x200 => crop to 25x50
				newHeight := int(float64(imgSize.Y) / float64(imgSize.X) * float64(size.Width))
				background = imaging.New(size.Width, newHeight, backgroundColor)
			} else if size.Height == 0 && size.Width == 0 {
				minRatio = 1
				background = imaging.New(imgSize.X, imgSize.Y, backgroundColor)
			}
		}

		backgroundSize := background.Bounds().Size()
		img = imaging.Resize(img, fixFloat(float64(imgSize.X)*minRatio, backgroundSize.X), fixFloat(float64(imgSize.Y)*minRatio, backgroundSize.Y), imaging.CatmullRom)
		return imaging.PasteCenter(background, img)
	default:
		width, height := size.Width, size.Height
		if width == 0 && height != 0 {
			// size 50x0, source 100x200 => crop to 50x100
			width = int(float64(imgSize.X) / float64(imgSize.Y) * float64(size.Height))
		} else if height == 0 && width != 0 {
			// size 0x50, source 100x200 => crop to 25x50
			height = int(float64(imgSize.Y) / float64(imgSize.X) * float64(size.Width))
		} else if height == 0 && width == 0 {
			width, height = imgSize.X, imgSize.Y
		}
		return imaging.Thumbnail(img, width, height, imaging.Lanczos)
	}
}

func (imageHandler) Handle(media Media, file FileInterface, option *Option) (err error) {
	var fileBytes []byte
	fileBytes, err = io.ReadAll(file)
	if err != nil {
		return
	}
	fileSizes := media.GetFileSizes()
	originalFileSize := len(fileBytes)
	fileSizes[OriginalSizeKey] = originalFileSize
	if _, err = file.Seek(0, 0); err != nil {
		return
	}
	err = media.Store(media.URL(OriginalSizeKey), option, file)
	if err != nil {
		return
	}
	if _, err = file.Seek(0, 0); err != nil {
		return
	}
	var format *imaging.Format
	format, err = GetImageFormat(media.URL())
	if err != nil {
		return
	}

	if *format == imaging.GIF {
		err = handleGIF(media, file, option, fileSizes, format)
		if err != nil {
			return
		}
		SetFileSizes(media, fileSizes)
		return
	}
	var img image.Image
	img, _, err = image.Decode(file)
	if err != nil {
		return
	}

	SetWeightHeight(media, img.Bounds().Dx(), img.Bounds().Dy())
	// Save cropped default image
	if cropOption := media.GetCropOption(DefaultSizeKey); cropOption != nil {
		var buffer bytes.Buffer
		if err = imaging.Encode(&buffer, imaging.Crop(img, *cropOption), *format); err != nil {
			return
		}
		fileSizes[DefaultSizeKey] = buffer.Len()
		if err = media.Store(media.URL(), option, &buffer); err != nil {
			return
		}
	} else {
		if _, err = file.Seek(0, 0); err != nil {
			return
		}
		// Save default image
		fileSizes[DefaultSizeKey] = originalFileSize
		if err = media.Store(media.URL(), option, file); err != nil {
			return
		}
	}

	// save sizes image
	for key, size := range media.GetSizes() {
		if key == DefaultSizeKey {
			continue
		}

		newImage := img
		if cropOption := media.GetCropOption(key); cropOption != nil {
			newImage = imaging.Crop(newImage, *cropOption)
		}
		var buffer bytes.Buffer
		if err = imaging.Encode(&buffer, resizeImageTo(newImage, size, *format), *format); err != nil {
			return
		}
		fileSizes[key] = buffer.Len()
		if err = media.Store(media.URL(key), option, &buffer); err != nil {
			return
		}
	}
	SetFileSizes(media, fileSizes)

	return
}

func handleGIF(media Media, file FileInterface, option *Option, fileSizes map[string]int, format *imaging.Format) error {
	var buffer bytes.Buffer
	g, err := gif.DecodeAll(file)
	if err != nil {
		return err
	}

	SetWeightHeight(media, g.Config.Width, g.Config.Height)
	if cropOption := media.GetCropOption(DefaultSizeKey); cropOption != nil {
		for i := range g.Image {
			img := imaging.Crop(g.Image[i], *cropOption)
			g.Image[i] = image.NewPaletted(img.Rect, g.Image[i].Palette)
			draw.Draw(g.Image[i], img.Rect, img, image.Pt(0, 0), draw.Src)
			if i == 0 {
				g.Config.Width = img.Rect.Dx()
				g.Config.Height = img.Rect.Dy()
			}
		}
	}

	gif.EncodeAll(&buffer, g)
	fileSizes[DefaultSizeKey] = buffer.Len()
	media.Store(media.URL(), option, &buffer)

	// save sizes image
	for key, size := range media.GetSizes() {
		if key == DefaultSizeKey {
			continue
		}

		file.Seek(0, 0)
		g, err := gif.DecodeAll(file)
		if err != nil {
			return err
		}

		for i := range g.Image {
			var img image.Image = g.Image[i]
			if cropOption := media.GetCropOption(key); cropOption != nil {
				img = imaging.Crop(g.Image[i], *cropOption)
			}
			img = resizeImageTo(img, size, *format)
			g.Image[i] = image.NewPaletted(image.Rect(0, 0, size.Width, size.Height), g.Image[i].Palette)
			draw.Draw(g.Image[i], image.Rect(0, 0, size.Width, size.Height), img, image.Pt(0, 0), draw.Src)
		}

		var buffer bytes.Buffer
		g.Config.Width = size.Width
		g.Config.Height = size.Height
		gif.EncodeAll(&buffer, g)
		fileSizes[key] = buffer.Len()
		media.Store(media.URL(key), option, &buffer)
	}
	return nil
}

func SetWeightHeight(media Media, width, height int) {
	result, _ := json.Marshal(map[string]int{"Width": width, "Height": height})
	media.Scan(string(result))
}

func SetFileSizes(media Media, fileSizes map[string]int) {
	result, _ := json.Marshal(map[string]map[string]int{"FileSizes": fileSizes})
	media.Scan(string(result))
}

func init() {
	RegisterMediaHandler("image_handler", imageHandler{})
}
