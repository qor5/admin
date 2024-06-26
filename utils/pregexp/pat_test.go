package pregexp

import "testing"

func TestApplyPathValues(t *testing.T) {
	tests := []struct {
		name             string
		pattern          string
		values           map[string]string
		onceIfDuplicated bool
		expected         string
	}{
		{
			"single value substitution",
			"/users/{id}",
			map[string]string{"id": "123"},
			false,
			"/users/123",
		},
		{
			"multiple value substitutions",
			"/users/{id}/posts/{postID}",
			map[string]string{"id": "123", "postID": "456"},
			false,
			"/users/123/posts/456",
		},
		{
			"value substitution with duplicated keys",
			"/users/{id}/{id}",
			map[string]string{"id": "123"},
			true,
			"/users/123/{id}",
		},
		{
			"value substitution with missing key",
			"/users/{id}/posts/{postID}",
			map[string]string{"id": "123"},
			false,
			"/users/123/posts/{postID}",
		},
		{
			"empty pattern with empty values",
			"",
			map[string]string{},
			false,
			"",
		},
		{
			"empty pattern with non-empty values",
			"",
			map[string]string{"id": "123"},
			false,
			"",
		},
		{
			"value substitution with keys containing leading and trailing spaces",
			"/users/{ id}/posts/{postID }",
			map[string]string{"id": "123", "postID": "456"},
			false,
			"/users/123/posts/456",
		},
		{
			"value substitution with keys containing leading and trailing spaces (duplicated keys)",
			"/users/{ id}/{ id }",
			map[string]string{"id": "123"},
			true,
			"/users/123/{ id }",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ApplyPathValues(test.pattern, test.values, test.onceIfDuplicated)
			if result != test.expected {
				t.Errorf("ApplyPathValues(%s, %v, %t) = %s, expected %s", test.pattern, test.values, test.onceIfDuplicated, result, test.expected)
			}
		})
	}
}
