package pagebuilder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFillCategoryIndentLevels(t *testing.T) {
	for _, c := range []struct {
		name   string
		cats   []*Category
		expect []*Category
	}{
		{
			name: "",
			cats: []*Category{
				{
					Path: "/",
				},
				{
					Path: "/a",
				},
				{
					Path: "/a/b",
				},
				{
					Path: "/a/b/c",
				},
				{
					Path: "/a/bb",
				},
				{
					Path: "/a/c",
				},
			},
			expect: []*Category{
				{
					Path:        "/",
					IndentLevel: 0,
				},
				{
					Path:        "/a",
					IndentLevel: 0,
				},
				{
					Path:        "/a/b",
					IndentLevel: 1,
				},
				{
					Path:        "/a/b/c",
					IndentLevel: 2,
				},
				{
					Path:        "/a/bb",
					IndentLevel: 1,
				},
				{
					Path:        "/a/c",
					IndentLevel: 1,
				},
			},
		},
	} {
		fillCategoryIndentLevels(c.cats)
		if diff := cmp.Diff(c.expect, c.cats); diff != "" {
			t.Fatalf("%s: %s\n", c.name, diff)
		}
	}
}
