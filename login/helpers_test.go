package login

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMustSetQuery(t *testing.T) {
	for _, c := range []struct {
		name   string
		u      string
		kvs    []string
		expect string
	}{
		{
			name: "no query",
			u:    "/a/b",
			kvs: []string{
				"k1", "v1",
				"k2", "v2",
			},
			expect: "/a/b?k1=v1&k2=v2",
		},
		{
			name: "has query",
			u:    "/a/b?a=1",
			kvs: []string{
				"k1", "v1",
				"k2", "v2",
			},
			expect: "/a/b?a=1&k1=v1&k2=v2",
		},
		{
			name: "has same query",
			u:    "/a/b?a=1&k2=v22",
			kvs: []string{
				"k1", "v1",
				"k2", "v2",
			},
			expect: "/a/b?a=1&k1=v1&k2=v2",
		},
		{
			name: "full url",
			u:    "https://example.com/a/b?a=1&k2=v22",
			kvs: []string{
				"k1", "v1",
				"k2", "v2",
			},
			expect: "https://example.com/a/b?a=1&k1=v1&k2=v2",
		},
	} {
		if diff := cmp.Diff(c.expect, mustSetQuery(c.u, c.kvs...)); diff != "" {
			t.Fatalf("%s: %s\n", c.name, diff)
		}
	}
}
