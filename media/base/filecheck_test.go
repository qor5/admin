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
	for _, ext := range []string{".svg", "SVG", ".svgz", ".html", ".htm", ".xhtml", ".xml", ".js", ".php", ".swf"} {
		if !IsDeniedExtension(ext) {
			t.Errorf("%q should be denied", ext)
		}
	}
	for _, ext := range []string{".png", ".jpg", ".pdf", ".mp4", ".txt", ""} {
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
