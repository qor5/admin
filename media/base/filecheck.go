package base

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// SniffLen is the number of leading bytes ReadHead returns, the most
// http.DetectContentType considers.
const SniffLen = 512

// deniedExtensions are file extensions whose content a browser may render as
// an active document (script-capable markup) or execute when it is served from
// the media storage. Uploads with these extensions are rejected unless the
// extension is explicitly allowlisted on the media builder.
var deniedExtensions = map[string]struct{}{
	".svg":   {},
	".svgz":  {},
	".html":  {},
	".htm":   {},
	".xhtml": {},
	".xht":   {},
	".xml":   {},
	".xsl":   {},
	".xslt":  {},
	".js":    {},
	".mjs":   {},
	".php":   {},
	".phtml": {},
	".shtml": {},
	".swf":   {},
}

// DeniedExtensions returns the extensions rejected on upload by default.
func DeniedExtensions() []string {
	exts := make([]string, 0, len(deniedExtensions))
	for ext := range deniedExtensions {
		exts = append(exts, ext)
	}
	return exts
}

// IsDeniedExtension reports whether ext (normalized with NormalizeExtension)
// is in the default upload denylist.
func IsDeniedExtension(ext string) bool {
	_, ok := deniedExtensions[NormalizeExtension(ext)]
	return ok
}

// NormalizeExtension lowercases ext and makes sure it has a leading dot, so
// "SVG", "svg" and ".svg" all compare equal. Blank input yields "".
func NormalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" || ext == "." {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

// FileExtension returns the normalized extension of a file name. Trailing dots
// and spaces are ignored, so "a.svg." and "a.svg " both yield ".svg" instead of
// slipping past an extension check.
func FileExtension(name string) string {
	name = strings.TrimRight(strings.TrimSpace(name), ". ")
	return NormalizeExtension(filepath.Ext(name))
}

// ReadHead reads up to SniffLen leading bytes from r for content sniffing.
func ReadHead(r io.Reader) ([]byte, error) {
	buf := make([]byte, SniffLen)
	n, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return buf[:n], nil
}

var markupPrefixes = [][]byte{
	[]byte("<svg"),
	[]byte("<?xml"),
	[]byte("<!doctype"),
	[]byte("<html"),
	[]byte("<head"),
	[]byte("<body"),
	[]byte("<script"),
	[]byte("<iframe"),
	[]byte("<!--"),
}

// IsMarkupContent reports whether head (the leading bytes of a file) looks like
// HTML, XML or SVG — content a browser may render as a script-capable document.
func IsMarkupContent(head []byte) bool {
	ct := strings.ToLower(http.DetectContentType(head))
	if strings.HasPrefix(ct, "text/html") || strings.Contains(ct, "xml") || strings.Contains(ct, "svg") {
		return true
	}
	trimmed := bytes.TrimPrefix(head, []byte("\xef\xbb\xbf"))
	trimmed = bytes.ToLower(bytes.TrimLeft(trimmed, " \t\r\n\f"))
	for _, p := range markupPrefixes {
		if bytes.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

// IsImageContent reports whether head is the start of a raster image of a
// format the media library treats as an image (see IsImageFormat).
func IsImageContent(head []byte) bool {
	// http.DetectContentType does not know TIFF.
	if bytes.HasPrefix(head, []byte("II*\x00")) || bytes.HasPrefix(head, []byte("MM\x00*")) {
		return true
	}
	ct := http.DetectContentType(head)
	return strings.HasPrefix(ct, "image/") && !strings.Contains(ct, "svg")
}

// IsVideoContent reports whether head is plausibly the start of a video file.
// http.DetectContentType recognizes only a few containers (mp4, webm, avi,
// ogg, mpeg); others such as QuickTime .mov come back as
// application/octet-stream, which is accepted so real videos are not
// rejected. Text content — including markup — is never accepted.
func IsVideoContent(head []byte) bool {
	ct := http.DetectContentType(head)
	switch {
	case strings.HasPrefix(ct, "video/"),
		strings.HasPrefix(ct, "audio/"),
		strings.HasPrefix(ct, "application/ogg"),
		strings.HasPrefix(ct, "application/octet-stream"):
		return true
	}
	return false
}
