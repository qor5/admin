package media

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
)

func testFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("NewFiles", filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	form, err := multipart.NewReader(&body, w.Boundary()).ReadForm(1 << 20)
	require.NoError(t, err)
	t.Cleanup(func() { _ = form.RemoveAll() })
	return form.File["NewFiles"][0]
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	return buf.Bytes()
}

const testSVGXSS = `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(document.domain)</script></svg>`

func TestCheckUploadFile(t *testing.T) {
	pngData := testPNG(t)
	mp4Data := []byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom")
	imgOnError := []byte(`<img src=x onerror=alert(document.domain)>`)
	xhtmlNS := []byte(`<x:script xmlns:x="http://www.w3.org/1999/xhtml">alert(1)</x:script>`)
	svgNS := []byte(`<s:svg xmlns:s="http://www.w3.org/2000/svg"><s:script>alert(1)</s:script></s:svg>`)
	html := []byte(`<html><script>alert(1)</script></html>`)
	procInst := []byte(`<?foo bar?><x:script xmlns:x="http://www.w3.org/1999/xhtml">alert(1)</x:script>`)

	cases := []struct {
		name     string
		builder  *Builder
		filename string
		content  []byte
		want     bool
	}{
		{"png image", New(nil), "a.png", pngData, true},
		{"uppercase ext", New(nil), "A.PNG", pngData, true},
		{"svg denied by default", New(nil), "x.svg", []byte(testSVGXSS), false},
		{"svg trailing dot", New(nil), "x.svg.", []byte(testSVGXSS), false},
		{"html denied", New(nil), "x.html", []byte("<html></html>"), false},
		{"svg disguised as png", New(nil), "x.png", []byte(testSVGXSS), false},
		{"html disguised as mp4", New(nil), "x.mp4", []byte("<html><script>1</script></html>"), false},
		{"svg disguised as txt", New(nil), "x.txt", []byte(testSVGXSS), false},
		{"plain txt", New(nil), "x.txt", []byte("hello"), true},
		{"csv", New(nil), "x.csv", []byte("a,b\n1,2\n"), true},
		{"pdf", New(nil), "x.pdf", []byte("%PDF-1.4\n"), true},
		{"docx (zip)", New(nil), "x.docx", []byte("PK\x03\x04\x14\x00\x06\x00"), true},
		{"mp4", New(nil), "x.mp4", mp4Data, true},
		{"text as png", New(nil), "x.png", []byte("not an image"), false},
		{"allowlist hit", New(nil).AllowedExtensions("PNG", ".pdf"), "a.png", pngData, true},
		{"allowlist miss", New(nil).AllowedExtensions("png"), "x.txt", []byte("hello"), false},
		{"allowlisted svg", New(nil).AllowedExtensions(".svg"), "x.svg", []byte(testSVGXSS), true},
		{"allowlist keeps image sniff", New(nil).AllowedExtensions(".png"), "x.png", []byte(testSVGXSS), false},

		// Extensions served as HTML/XML that are not in the static list.
		{"ehtml img onerror", New(nil), "a.ehtml", imgOnError, false},
		{"rdf namespaced xhtml", New(nil), "a.rdf", xhtmlNS, false},
		{"xbl namespaced xhtml", New(nil), "a.xbl", xhtmlNS, false},
		{"xbl processing instruction", New(nil), "a.xbl", procInst, false},
		// Structural markup check: any leading '<' in a plain file or an
		// extension without a known MIME type.
		{"namespaced svg as txt", New(nil), "a.txt", svgNS, false},
		{"img onerror as txt", New(nil), "a.txt", imgOnError, false},
		{"utf-16 markup as txt", New(nil), "a.txt", []byte("\xff\xfe<\x00h\x00"), false},
		{"markup with unknown ext", New(nil), "a.xyz", imgOnError, false},
		{"text with unknown ext", New(nil), "a.xyz", []byte("hello"), true},
		// Unsafe names and allowlisted extensions never let markup in.
		{"allowlisted png trailing space", New(nil).AllowedExtensions("png", "pdf"), "a.png ", html, false},
		{"allowlisted png trailing dot", New(nil).AllowedExtensions("png", "pdf"), "a.png.", html, false},
		{"allowlisted txt with html", New(nil).AllowedExtensions("pdf", "txt"), "a.txt", html, false},
		{"png trailing space", New(nil), "a.png ", html, false},
		{"control char in name", New(nil), "a\x01b.txt", []byte("hello"), false},
		// Images are governed by magic bytes; a GIF polyglot is served as
		// image/gif, which browsers do not render as a document.
		{"gif polyglot", New(nil), "a.gif", append([]byte("GIF89a"), html...), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Build with a neutral name, then set the raw one: the multipart
			// reader would refuse some of the hostile names before our check.
			fh := testFileHeader(t, "upload.bin", c.content)
			fh.Filename = c.filename
			typ := uploadSelectedType(base.FileExtension(c.filename))
			got, err := c.builder.checkUploadFile(fh, typ)
			require.NoError(t, err)
			require.Equal(t, c.want, got)

			// The file must still be readable from the start afterwards.
			f, err := fh.Open()
			require.NoError(t, err)
			defer f.Close()
			head := make([]byte, len(c.content))
			_, err = f.Read(head)
			require.NoError(t, err)
			require.Equal(t, c.content, head)
		})
	}
}

func TestUploadSelectedType(t *testing.T) {
	require.Equal(t, media_library.ALLOW_TYPE_IMAGE, uploadSelectedType(base.FileExtension("a.PNG")))
	require.Equal(t, media_library.ALLOW_TYPE_VIDEO, uploadSelectedType(base.FileExtension("a.mov")))
	require.Equal(t, media_library.ALLOW_TYPE_FILE, uploadSelectedType(base.FileExtension("a.pdf")))
	require.Equal(t, media_library.ALLOW_TYPE_FILE, uploadSelectedType(base.FileExtension("noext")))
}
