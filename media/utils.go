package media

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/qor/qor/utils"
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

func ByteCountIEC(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
