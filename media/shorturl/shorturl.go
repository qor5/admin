package shorturl

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/qor/oss"
	"github.com/qor5/admin/media/media_library"
)

const (
	base62Chars  = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ShortCodeLen = 6
	IndexFile    = "index.html"
)

// Config holds configuration for the short URL feature.
type Config struct {
	// Storage is the object storage used to upload and delete redirect files.
	Storage oss.StorageInterface

	// ShortURLHost is prepended to the short path when building the full URL.
	// Example: "https://example.com"
	ShortURLHost string

	// PathPrefix is an optional S3 path segment inserted before the short code.
	// Example: "s" produces "s/oJZH0n/index.html". Empty means no prefix.
	PathPrefix string

	// ShortURLUpload overrides the default HTML redirect upload behavior.
	// If nil, a meta-refresh HTML file is uploaded via Storage.Put.
	// shortPath example: "s/aB3k9Z/index.html"
	ShortURLUpload func(storage oss.StorageInterface, shortPath, actualURL string) error

	// CodeGenerator returns a short code for the given media file.
	// If nil, the default RandomCode(ShortCodeLen) is used.
	CodeGenerator func(m *media_library.MediaLibrary) (string, error)
}

// FullURL builds the complete short URL from a stored short path.
// The trailing "/index.html" is stripped so the copied URL stays clean.
// S3 with static website hosting serves index.html automatically for directory paths.
func (c *Config) FullURL(shortPath string) string {
	displayPath := strings.TrimSuffix(shortPath, IndexFile)
	return strings.TrimSuffix(c.ShortURLHost, "/") + "/" + displayPath
}

// ShortPath builds the S3 object path from an optional prefix and short code.
// Example: prefix="s", code="oJZH0n" → "s/oJZH0n/index.html"
//
//	prefix="",  code="oJZH0n" → "oJZH0n/index.html"
func ShortPath(prefix, code string) string {
	if prefix == "" {
		return code + "/" + IndexFile
	}
	return strings.TrimSuffix(prefix, "/") + "/" + code + "/" + IndexFile
}

// Upload uploads the redirect file to storage using the configured uploader,
// falling back to the default meta-refresh HTML if none is set.
func Upload(cfg *Config, shortPath, actualURL string) error {
	if cfg.ShortURLUpload != nil {
		return cfg.ShortURLUpload(cfg.Storage, shortPath, actualURL)
	}
	if cfg.Storage == nil {
		return fmt.Errorf("shorturl: Storage is nil and no ShortURLUpload is set")
	}
	html := defaultHTML(actualURL)
	_, err := cfg.Storage.Put(shortPath, strings.NewReader(html))
	return err
}

// Delete removes the redirect file from storage.
// It is a no-op when cfg is nil, Storage is nil, or shortPath is empty.
func Delete(cfg *Config, shortPath string) error {
	if cfg == nil || cfg.Storage == nil || shortPath == "" {
		return nil
	}
	return cfg.Storage.Delete(shortPath)
}

// RandomCode generates a cryptographically random Base62 string of the given length.
// The character set is 0-9a-zA-Z (62 chars). The modulo bias is negligible for
// short URL use cases (256 is not divisible by 62, max bias ~0.4%).
func RandomCode(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	for i := range b {
		b[i] = base62Chars[int(b[i])%62]
	}
	return string(b), nil
}

// defaultHTML returns a minimal redirect page that works across all browsers.
func defaultHTML(actualURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
<script>window.location.replace(%q);</script>
</body>
</html>`, actualURL, actualURL)
}
