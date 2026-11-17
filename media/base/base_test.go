package base

import (
	"path"
	"regexp"
	"testing"

	"gorm.io/gorm"
)

func TestGetURLUUIDPlaceholder(t *testing.T) {
	db := &gorm.DB{Statement: &gorm.Statement{}}
	option := Option{"URL": "/system/{{uuid}}/file/{{uuid}}/{{filename}}"}
	re := regexp.MustCompile(`^/system/([0-9a-f-]{36})/file/([0-9a-f-]{36})/a\.png$`)

	b := &Base{FileName: "a.png"}
	u1 := b.GetURL(&option, db, nil, b)
	m := re.FindStringSubmatch(u1)
	if m == nil {
		t.Fatalf("unexpected url %q", u1)
	}
	if m[1] != m[2] {
		t.Fatalf("{{uuid}} must render the same value within one URL, got %q", u1)
	}

	u2 := b.GetURL(&option, db, nil, b)
	if m2 := re.FindStringSubmatch(u2); m2 == nil || m2[1] == m[1] {
		t.Fatalf("each rendering must get a fresh uuid: %q vs %q", u1, u2)
	}

	// Size variants derive from the stored URL, so they share its directory.
	b.Url = u1
	for _, style := range []string{OriginalSizeKey, "thumb"} {
		if got := path.Dir(b.URL(style)); got != path.Dir(u1) {
			t.Errorf("variant %q dir = %q, want %q", style, got, path.Dir(u1))
		}
	}
}
