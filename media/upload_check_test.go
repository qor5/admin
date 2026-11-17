package media

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"

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

	cases := []struct {
		name     string
		builder  *Builder
		filename string
		content  []byte
		typ      string
		want     bool
	}{
		{"png image", New(nil), "a.png", pngData, media_library.ALLOW_TYPE_IMAGE, true},
		{"uppercase ext", New(nil), "A.PNG", pngData, media_library.ALLOW_TYPE_IMAGE, true},
		{"svg denied by default", New(nil), "x.svg", []byte(testSVGXSS), media_library.ALLOW_TYPE_FILE, false},
		{"svg trailing dot", New(nil), "x.svg.", []byte(testSVGXSS), media_library.ALLOW_TYPE_FILE, false},
		{"html denied", New(nil), "x.html", []byte("<html></html>"), media_library.ALLOW_TYPE_FILE, false},
		{"svg disguised as png", New(nil), "x.png", []byte(testSVGXSS), media_library.ALLOW_TYPE_IMAGE, false},
		{"html disguised as mp4", New(nil), "x.mp4", []byte("<html><script>1</script></html>"), media_library.ALLOW_TYPE_VIDEO, false},
		{"svg disguised as txt", New(nil), "x.txt", []byte(testSVGXSS), media_library.ALLOW_TYPE_FILE, false},
		{"plain txt", New(nil), "x.txt", []byte("hello"), media_library.ALLOW_TYPE_FILE, true},
		{"pdf", New(nil), "x.pdf", []byte("%PDF-1.4\n"), media_library.ALLOW_TYPE_FILE, true},
		{"mp4", New(nil), "x.mp4", mp4Data, media_library.ALLOW_TYPE_VIDEO, true},
		{"text as png", New(nil), "x.png", []byte("not an image"), media_library.ALLOW_TYPE_IMAGE, false},
		{"allowlist hit", New(nil).AllowedExtensions("PNG", ".pdf"), "a.png", pngData, media_library.ALLOW_TYPE_IMAGE, true},
		{"allowlist miss", New(nil).AllowedExtensions("png"), "x.txt", []byte("hello"), media_library.ALLOW_TYPE_FILE, false},
		{"allowlisted svg", New(nil).AllowedExtensions(".svg"), "x.svg", []byte(testSVGXSS), media_library.ALLOW_TYPE_FILE, true},
		{"allowlist keeps image sniff", New(nil).AllowedExtensions(".png"), "x.png", []byte(testSVGXSS), media_library.ALLOW_TYPE_IMAGE, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fh := testFileHeader(t, c.filename, c.content)
			got, err := c.builder.checkUploadFile(fh, c.typ)
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
