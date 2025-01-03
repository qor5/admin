package base

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/qor5/imaging"

	"github.com/qor5/admin/v3/utils"
)

func GetImageFormat(url string) (*imaging.Format, error) {
	formats := map[string]imaging.Format{
		".jpg":  imaging.JPEG,
		".jpeg": imaging.JPEG,
		".png":  imaging.PNG,
		".tif":  imaging.TIFF,
		".tiff": imaging.TIFF,
		".bmp":  imaging.BMP,
		".gif":  imaging.GIF,
	}

	ext := strings.ToLower(regexp.MustCompile(`(\?.*?$)`).ReplaceAllString(filepath.Ext(url), ""))
	if f, ok := formats[ext]; ok {
		return &f, nil
	}
	return nil, imaging.ErrUnsupportedFormat
}

// IsImageFormat check filename is image or not
func IsImageFormat(name string) bool {
	_, err := GetImageFormat(name)
	return err == nil
}

// IsVideoFormat check filename is video or not
func IsVideoFormat(name string) bool {
	formats := []string{".mp4", ".m4p", ".m4v", ".m4v", ".mov", ".mpeg", ".webm", ".avi", ".ogg", ".ogv"}

	ext := strings.ToLower(regexp.MustCompile(`(\?.*?$)`).ReplaceAllString(filepath.Ext(name), ""))

	for _, format := range formats {
		if format == ext {
			return true
		}
	}

	return false
}

func IsSVGFormat(name string) bool {
	formats := []string{".svg", ".svgz"}

	ext := strings.ToLower(regexp.MustCompile(`(\?.*?$)`).ReplaceAllString(filepath.Ext(name), ""))

	for _, format := range formats {
		if format == ext {
			return true
		}
	}

	return false
}

func parseTagOption(str string) *Option {
	option := Option(utils.ParseTagOption(str))
	return &option
}

func ByteCountSI(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	format := "%.1f%cB"
	suffix := "kMGTPE"[exp]
	if suffix == 'k' {
		format = "%.f%cB"
	}
	return fmt.Sprintf(format,
		float64(b)/float64(div), suffix)
}

func SaleUpDown(width, height int, size *Size) {
	if size.Height == 0 && size.Width > 0 {
		size.Height = int(float64(size.Width) / float64(width) * float64(height))
	} else if size.Height > 0 && size.Width == 0 {
		size.Width = int(float64(size.Height) / float64(height) * float64(width))
	}
}
