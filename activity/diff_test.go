package activity

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/qor5/admin/media/media_library"
)

type (
	Post struct {
		ID            uint `gorm:"primarykey"`
		CreatedAt     time.Time
		UpdatedAt     time.Time
		PublishedDate time.Time
		Image         media_library.MediaBox

		Title    string
		Content  string
		Author   Author
		Comments []Comment
		Tags     map[string]Tag
	}
	Tag struct {
		Name string
	}
	Author struct {
		Name string
		Age  int
	}
	Comment struct {
		Text string
	}
)

func TestDiff(t *testing.T) {
	testCases := []struct {
		description  string
		modelBuilder *ModelBuilder
		old          Post
		now          Post
		want         []Diff
	}{
		{
			description:  "Simple basic update",
			modelBuilder: &ModelBuilder{},
			old:          Post{Title: "test", Content: ""},
			now:          Post{Title: "test1", Content: "124"},
			want: []Diff{
				{
					Field: "Title",
					Old:   "test",
					Now:   "test1",
				},
				{
					Field: "Content",
					Old:   "",
					Now:   "124",
				}},
		},
		{
			description:  "Default type handles",
			modelBuilder: &ModelBuilder{},
			old:          Post{Image: media_library.MediaBox{ID: json.Number("1"), Url: "https://s3.com/1.jpg", Description: "test"}},
			now:          Post{Image: media_library.MediaBox{ID: json.Number("2"), Url: "https://s3.com/2.jpg", Description: "test2"}},
			want: []Diff{
				{
					Field: "Image.Url",
					Old:   "https://s3.com/1.jpg",
					Now:   "https://s3.com/2.jpg",
				},
				{
					Field: "Image.Description",
					Old:   "test",
					Now:   "test2",
				},
			},
		},
		{
			description:  "Default ingored fields",
			modelBuilder: &ModelBuilder{},
			old:          Post{ID: 1, CreatedAt: time.Unix(1257894000, 0)},
			now:          Post{ID: 2, CreatedAt: time.Unix(1457894000, 0)},
			want:         nil,
		},
		{
			description:  "Using model ingored fields",
			modelBuilder: (&ModelBuilder{}).AddIgnoredFields("Name"),
			old:          Post{Author: Author{Name: "test", Age: 10}},
			now:          Post{Author: Author{Name: "test1", Age: 12}},
			want: []Diff{
				{
					Field: "Author.Age",
					Old:   "10",
					Now:   "12",
				},
			},
		},
		{
			description: "Using model type handles",
			modelBuilder: (&ModelBuilder{}).AddTypeHanders(Author{}, func(old, now interface{}, prefixField string) (diffs []Diff) {
				oldAuthor := old.(Author)
				nowAuthor := now.(Author)
				if oldAuthor.Name != nowAuthor.Name {
					diffs = append(diffs, Diff{Field: fmt.Sprintf("%s.Name", prefixField), Old: oldAuthor.Name, Now: nowAuthor.Name})
				}
				return diffs
			}),
			old: Post{Author: Author{Name: "test", Age: 10}},
			now: Post{Author: Author{Name: "test1", Age: 12}},
			want: []Diff{
				{
					Field: "Author.Name",
					Old:   "test",
					Now:   "test1",
				},
			},
		},
		{
			description:  "Test slice data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Comments: []Comment{{Text: "1"}, {Text: "2"}}},
			now:          Post{Comments: []Comment{{Text: "1.1"}, {Text: "2.2"}}},
			want: []Diff{
				{
					Field: "Comments.0.Text",
					Old:   "1",
					Now:   "1.1",
				},
				{
					Field: "Comments.1.Text",
					Old:   "2",
					Now:   "2.2",
				},
			},
		},
		{
			description:  "Test deleting slice data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Comments: []Comment{{Text: "1"}, {Text: "2"}}},
			now:          Post{Comments: []Comment{{Text: "1.1"}}},
			want: []Diff{
				{
					Field: "Comments.0.Text",
					Old:   "1",
					Now:   "1.1",
				},
				{
					Field: "Comments.1",
					Old:   "{Text:2}",
					Now:   "",
				},
			},
		},
		{
			description:  "Test adding slice data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Comments: []Comment{{Text: "1"}}},
			now:          Post{Comments: []Comment{{Text: "1.1"}, {Text: "2"}}},
			want: []Diff{
				{
					Field: "Comments.0.Text",
					Old:   "1",
					Now:   "1.1",
				},
				{
					Field: "Comments.1",
					Old:   "",
					Now:   "{Text:2}",
				},
			},
		},
		{
			description:  "Test creating slice data",
			modelBuilder: &ModelBuilder{},
			old:          Post{},
			now:          Post{Comments: []Comment{{Text: "1.1"}, {Text: "2"}}},
			want: []Diff{
				{
					Field: "Comments",
					Old:   "",
					Now:   "[{Text:1.1} {Text:2}]",
				},
			},
		},
		{
			description:  "Test remove all slice data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Comments: []Comment{{Text: "1.1"}, {Text: "2"}}},
			now:          Post{},
			want: []Diff{
				{
					Field: "Comments",
					Old:   "[{Text:1.1} {Text:2}]",
					Now:   "",
				},
			},
		},
		{
			description:  "Test map data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}}},
			now:          Post{Tags: map[string]Tag{"tag1": {Name: "tst12"}}},
			want: []Diff{
				{
					Field: "Tags.tag1.Name",
					Old:   "tst1",
					Now:   "tst12",
				},
			},
		},
		{
			description:  "Test adding map data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}}},
			now:          Post{Tags: map[string]Tag{"tag1": {Name: "tst12"}, "tag2": {Name: "tst121"}}},
			want: []Diff{
				{
					Field: "Tags.tag1.Name",
					Old:   "tst1",
					Now:   "tst12",
				},
				{
					Field: "Tags.tag2",
					Old:   "",
					Now:   "{Name:tst121}",
				},
			},
		},
		{
			description:  "Test deleting map data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}, "tag2": {Name: "tst1"}}},
			now:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}}},
			want: []Diff{
				{
					Field: "Tags.tag2",
					Old:   "{Name:tst1}",
					Now:   "",
				},
			},
		},
		{
			description:  "Test creating map data",
			modelBuilder: &ModelBuilder{},
			old:          Post{},
			now:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}}},
			want: []Diff{
				{
					Field: "Tags",
					Old:   "",
					Now:   "map[tag1:{Name:tst1}]",
				},
			},
		},
		{
			description:  "Test remove all map data",
			modelBuilder: &ModelBuilder{},
			old:          Post{Tags: map[string]Tag{"tag1": {Name: "tst1"}}},
			now:          Post{Tags: nil},
			want: []Diff{
				{
					Field: "Tags",
					Old:   "map[tag1:{Name:tst1}]",
					Now:   "",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			diffs, err := NewDiffBuilder(test.modelBuilder).Diff(test.old, test.now)
			if err != nil {
				t.Fatalf("want: %v, but got error: %v", test.want, err)
			}
			w, _ := json.Marshal(test.want)
			d, _ := json.Marshal(diffs)
			if string(w) != string(d) {
				t.Fatalf("want: %v, but got: %v", string(w), string(d))
			}
		})
	}
}

func TestDiffTypesError(t *testing.T) {
	_, err := NewDiffBuilder(&ModelBuilder{}).Diff(Post{Title: "123"}, Author{Name: "ccc"})

	if err.Error() != "old and now type mismatch: activity.Post != activity.Author" {
		t.Fatalf("difference type error")
	}

	_, err = NewDiffBuilder(&ModelBuilder{}).Diff(Post{Title: "123"}, struct{}{})
	if err.Error() != "old and now type mismatch: activity.Post != struct {}" {
		t.Fatalf("difference type error")
	}
}

func BenchmarkSimpleDiff(b *testing.B) {
	builder := NewDiffBuilder(&ModelBuilder{})
	for i := 0; i < b.N; i++ {
		builder.Diff(Author{Name: "test1", Age: 10}, Author{Name: "test12", Age: 18})
	}
}

func BenchmarkComplexDiff(b *testing.B) {
	old := Post{
		ID:            1,
		CreatedAt:     time.Now(),
		PublishedDate: time.Now(),
		Image:         media_library.MediaBox{ID: json.Number("1"), Url: "https://s3.com/1.jpg", Description: "test"},
		Title:         "title",
		Content:       "content111",
		Author:        Author{Name: "author1", Age: 10},
		Comments:      []Comment{},
		Tags:          map[string]Tag{},
	}

	for i := 0; i < 50; i++ {
		old.Comments = append(old.Comments, Comment{Text: fmt.Sprintf("text - %d", i)})
		old.Tags[fmt.Sprintf("tag - %d", i)] = Tag{Name: fmt.Sprintf("title - %d", i)}
	}

	now := Post{
		ID:            1,
		CreatedAt:     time.Now().Add(1 * time.Hour),
		PublishedDate: time.Now().Add(3 * time.Hour),
		Image:         media_library.MediaBox{ID: json.Number("2"), Url: "https://s3.com/2.jpg", Description: "test2"},
		Title:         "title1",
		Content:       "content111",
		Author: Author{
			Name: "author2",
			Age:  19,
		},
		Comments: []Comment{},
		Tags:     map[string]Tag{},
	}

	for i := 0; i < 80; i++ {
		now.Comments = append(now.Comments, Comment{Text: fmt.Sprintf("text ---%d", i)})
		old.Tags[fmt.Sprintf("tag - %d", i)] = Tag{Name: fmt.Sprintf("title - %d", i)}
	}

	builder := NewDiffBuilder(&ModelBuilder{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Diff(old, now)
	}
}

// goos: darwin
// goarch: amd64
// pkg: github.com/qor5/admin/activity
// cpu: Intel(R) Core(TM) i5-6267U CPU @ 2.90GHz
// BenchmarkSimpleDiff-4    	  669444	      1869 ns/op
// BenchmarkComplexDiff-4   	    1381	    729444 ns/op
