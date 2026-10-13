package media_library

import (
	"strings"
	"testing"
)

func TestMediaLibraryStorageOriginalFileName(t *testing.T) {
	var s MediaLibraryStorage
	s.FileName = "HOMES.pdf"
	s.OriginalFileName = "HOME'S.pdf"

	v, err := s.Value()
	if err != nil {
		t.Fatal(err)
	}

	var got MediaLibraryStorage
	if err := got.Scan(v); err != nil {
		t.Fatal(err)
	}
	if got.OriginalFileName != "HOME'S.pdf" || got.FileName != "HOMES.pdf" {
		t.Errorf("round trip: OriginalFileName = %q, FileName = %q", got.OriginalFileName, got.FileName)
	}

	// Records whose name was not changed must not grow a new key.
	s.OriginalFileName = ""
	v, _ = s.Value()
	if strings.Contains(v.(string), "OriginalFileName") {
		t.Errorf("empty OriginalFileName was serialized: %s", v)
	}
}
