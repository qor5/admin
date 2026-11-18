package base

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

const svgXSS = `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(document.domain)</script></svg>`

func TestNormalizeExtension(t *testing.T) {
	for in, want := range map[string]string{
		"svg": ".svg", ".SVG": ".svg", " Png ": ".png", "": "", ".": "",
	} {
		if got := NormalizeExtension(in); got != want {
			t.Errorf("NormalizeExtension(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFileExtension(t *testing.T) {
	for in, want := range map[string]string{
		"a.png":     ".png",
		"A.SVG":     ".svg",
		"a.svg.":    ".svg",
		"a.svg . ":  ".svg",
		"a.png.svg": ".svg",
		"noext":     "",
	} {
		if got := FileExtension(in); got != want {
			t.Errorf("FileExtension(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsDeniedExtension(t *testing.T) {
	for _, ext := range []string{
		".svg", "SVG", ".svgz", ".html", ".htm", ".xhtml", ".xml", ".js", ".php", ".swf",
		".ehtml", ".rdf", ".xbl", ".xsd", ".rss", ".atom", ".xul", ".mml", ".mht", ".mhtml", ".wsdl",
	} {
		if !IsDeniedExtension(ext) {
			t.Errorf("%q should be denied", ext)
		}
	}
	for _, ext := range []string{".png", ".jpg", ".pdf", ".mp4", ".txt", ".csv", ".docx", ".xlsx", ".zip", ""} {
		if IsDeniedExtension(ext) {
			t.Errorf("%q should not be denied", ext)
		}
	}
}

func TestReadHead(t *testing.T) {
	head, err := ReadHead(strings.NewReader("short"))
	if err != nil || string(head) != "short" {
		t.Fatalf("got %q, %v", head, err)
	}
	head, err = ReadHead(bytes.NewReader(make([]byte, 2000)))
	if err != nil || len(head) != SniffLen {
		t.Fatalf("got len %d, %v", len(head), err)
	}
	head, err = ReadHead(strings.NewReader(""))
	if err != nil || len(head) != 0 {
		t.Fatalf("got %q, %v", head, err)
	}
}

func TestIsMarkupContent(t *testing.T) {
	markup := []string{
		svgXSS,
		"  \n" + svgXSS,
		"\xef\xbb\xbf" + svgXSS,
		`<?xml version="1.0"?><svg/>`,
		`<!DOCTYPE html><html></html>`,
		`<html><script>alert(1)</script></html>`,
		`<SCRIPT>alert(1)</SCRIPT>`,
	}
	for _, s := range markup {
		if !IsMarkupContent([]byte(s)) {
			t.Errorf("expected markup: %q", s)
		}
	}
	for _, b := range [][]byte{pngBytes(t), []byte("hello, plain text"), []byte("%PDF-1.4\n"), {}} {
		if IsMarkupContent(b) {
			t.Errorf("unexpected markup: %q", b)
		}
	}
}

func TestIsImageContent(t *testing.T) {
	ok := [][]byte{
		pngBytes(t),
		[]byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00"),
		[]byte("GIF89a\x01\x00\x01\x00"),
		[]byte("BM\x00\x00\x00\x00"),
		[]byte("II*\x00\x08\x00\x00\x00"),
		[]byte("MM\x00*\x00\x00\x00\x08"),
	}
	for _, b := range ok {
		if !IsImageContent(b) {
			t.Errorf("expected image: %q", b)
		}
	}
	for _, s := range []string{svgXSS, "<html></html>", "plain text", "\x00\x01\x02\x03"} {
		if IsImageContent([]byte(s)) {
			t.Errorf("unexpected image: %q", s)
		}
	}
}

func TestIsVideoContent(t *testing.T) {
	ok := [][]byte{
		[]byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom"), // mp4
		[]byte("\x00\x00\x00\x14ftypqt  \x00\x00\x00\x00qt  "),     // QuickTime: octet-stream
		[]byte("\x1a\x45\xdf\xa3\x9f\x42\x86\x81\x01\x42\xf7\x81\x01\x42\xf2\x81\x04\x42\xf3\x81\x08\x42\x82\x84webm"),
		[]byte("OggS\x00\x02"),
	}
	for _, b := range ok {
		if !IsVideoContent(b) {
			t.Errorf("expected video: %q", b)
		}
	}
	for _, s := range []string{svgXSS, "<html></html>", "plain text"} {
		if IsVideoContent([]byte(s)) {
			t.Errorf("unexpected video: %q", s)
		}
	}
}

func TestIsSafeUploadFileName(t *testing.T) {
	for _, name := range []string{"a.png", "report #3.pdf", "写真.jpg", "noext"} {
		if !IsSafeUploadFileName(name) {
			t.Errorf("%q should be safe", name)
		}
	}
	for _, name := range []string{"", "a.png.", "a.png ", "a.png\t", "a.png\u00a0", "a\x00.png", "a\x1b.txt", "a\u0085.txt"} {
		if IsSafeUploadFileName(name) {
			t.Errorf("%q should be unsafe", name)
		}
	}
}

func TestStartsWithMarkup(t *testing.T) {
	for _, s := range []string{
		"<a>", "  \n<x:script>", "\xef\xbb\xbf<svg>", "\xff\xfe<\x00s\x00", "\xfe\xff\x00<\x00s",
	} {
		if !StartsWithMarkup([]byte(s)) {
			t.Errorf("expected markup start: %q", s)
		}
	}
	for _, s := range []string{"", "   ", "a,b\n<1>", "hello <b>", "%PDF-1.4"} {
		if StartsWithMarkup([]byte(s)) {
			t.Errorf("unexpected markup start: %q", s)
		}
	}
}

func TestHasKnownMIMEType(t *testing.T) {
	if !HasKnownMIMEType(".png") || !HasKnownMIMEType("txt") {
		t.Error("png and txt have known MIME types")
	}
	if HasKnownMIMEType(".xyz-unknown") || HasKnownMIMEType("") {
		t.Error("unknown and empty extensions have no MIME type")
	}
}

func TestIsScriptCapableMIMEType(t *testing.T) {
	for _, ct := range []string{
		"text/html; charset=utf-8", "application/xhtml+xml", "text/xml", "application/xml",
		"application/rdf+xml", "image/svg+xml", "text/javascript", "application/ecmascript",
		"application/xml-dtd",
	} {
		if !isScriptCapableMIMEType(ct) {
			t.Errorf("%q should be script-capable", ct)
		}
	}
	for _, ct := range []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain; charset=utf-8", "application/pdf", "image/png", "application/json",
	} {
		if isScriptCapableMIMEType(ct) {
			t.Errorf("%q should not be script-capable", ct)
		}
	}
}
