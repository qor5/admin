package base

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
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
	".ehtml": {},
	".rdf":   {},
	".xbl":   {},
	".xsd":   {},
	".rss":   {},
	".atom":  {},
	".xul":   {},
	".mml":   {},
	".mht":   {},
	".mhtml": {},
	".wsdl":  {},
}

// isScriptCapableMIMEType reports whether a browser may render a response of
// MIME type ct as a document or run it as script: HTML, XML (text/xml,
// application/xml, any "+xml" subtype), SVG, or JavaScript. XML is matched on
// the subtype structure so that e.g. Office "openxmlformats" types stay
// allowed.
func isScriptCapableMIMEType(ct string) bool {
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	}
	_, subtype, _ := strings.Cut(mediaType, "/")
	switch {
	case strings.Contains(subtype, "html"),
		strings.Contains(subtype, "svg"),
		strings.Contains(subtype, "javascript"),
		strings.Contains(subtype, "ecmascript"),
		subtype == "xml",
		strings.HasPrefix(subtype, "xml-"),
		strings.HasSuffix(subtype, "+xml"):
		return true
	}
	return false
}

// DeniedExtensions returns the static list of extensions rejected on upload by
// default. IsDeniedExtension additionally denies any extension whose MIME type
// (mime.TypeByExtension) is HTML, XML, SVG or JavaScript.
func DeniedExtensions() []string {
	exts := make([]string, 0, len(deniedExtensions))
	for ext := range deniedExtensions {
		exts = append(exts, ext)
	}
	return exts
}

// IsDeniedExtension reports whether files with ext are rejected on upload by
// default: ext is in the static denylist, or the MIME type a server would
// serve it with is HTML, XML, SVG or JavaScript.
func IsDeniedExtension(ext string) bool {
	ext = NormalizeExtension(ext)
	if _, ok := deniedExtensions[ext]; ok {
		return true
	}
	if ext == "" {
		return false
	}
	ct := mime.TypeByExtension(ext)
	return ct != "" && isScriptCapableMIMEType(ct)
}

// IsSafeUploadFileName reports whether name is acceptable as an upload file
// name: non-empty, without control characters, and not ending in a dot or
// whitespace — trailing characters that some storages and servers strip, which
// would make the served extension differ from the checked one.
func IsSafeUploadFileName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	last, _ := utf8.DecodeLastRuneInString(name)
	return last != '.' && !unicode.IsSpace(last)
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
// slipping past an extension check (uploads with such names are rejected
// anyway, see IsSafeUploadFileName).
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

// StartsWithMarkup reports whether the first significant byte of head — after
// any byte order mark, whitespace and NUL padding (UTF-16) — is '<'. That is
// the structural signature of HTML/XML/SVG, whatever the elements are named or
// namespaced, and it is what browsers key on when they sniff an untyped
// response.
func StartsWithMarkup(head []byte) bool {
	for _, bom := range [][]byte{{0xef, 0xbb, 0xbf}, {0xfe, 0xff}, {0xff, 0xfe}} {
		if bytes.HasPrefix(head, bom) {
			head = head[len(bom):]
			break
		}
	}
	for _, c := range head {
		switch c {
		case ' ', '\t', '\r', '\n', '\f', 0:
			continue
		}
		return c == '<'
	}
	return false
}

// HasKnownMIMEType reports whether a server would serve ext with an explicit
// MIME type (mime.TypeByExtension), rather than sniffing the content.
func HasKnownMIMEType(ext string) bool {
	ext = NormalizeExtension(ext)
	return ext != "" && mime.TypeByExtension(ext) != ""
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
