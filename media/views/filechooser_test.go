package views

import (
	"mime/multipart"
	"strings"
	"testing"
)

func TestSanitizeUploadFileName(t *testing.T) {
	t.Cleanup(func() { SetFileNameSanitizer(nil) })

	SetFileNameSanitizer(nil)
	fh := &multipart.FileHeader{Filename: "HOME'S.pdf"}
	if got := sanitizeUploadFileName(fh); got != "" || fh.Filename != "HOME'S.pdf" {
		t.Errorf("without sanitizer: original = %q, name = %q; want untouched", got, fh.Filename)
	}

	SetFileNameSanitizer(func(s string) string { return strings.ReplaceAll(s, "'", "") })

	fh = &multipart.FileHeader{Filename: "HOME'S.pdf"}
	if got := sanitizeUploadFileName(fh); got != "HOME'S.pdf" || fh.Filename != "HOMES.pdf" {
		t.Errorf("changed name: original = %q, name = %q; want %q, %q", got, fh.Filename, "HOME'S.pdf", "HOMES.pdf")
	}

	fh = &multipart.FileHeader{Filename: "plain.pdf"}
	if got := sanitizeUploadFileName(fh); got != "" || fh.Filename != "plain.pdf" {
		t.Errorf("unchanged name: original = %q, name = %q; want \"\", %q", got, fh.Filename, "plain.pdf")
	}
}
