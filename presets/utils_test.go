package presets

import (
	"fmt"
	"testing"

	"gorm.io/gorm"
)

type foo struct {
	gorm.Model
	Version string
}

func (v *foo) PrimarySlug() string {
	return fmt.Sprintf("%d_%s", v.ID, v.Version)
}

func TestObjectID(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name: "",
			input: struct {
				gorm.Model
				Name string
			}{
				Name: "foo",
			},
			expected: "",
		},
		{
			name: "",
			input: struct {
				gorm.Model
				Name string
			}{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "foo",
			},
			expected: "1",
		},
		{
			name: "",
			input: struct {
				ID   string
				Name string
			}{
				ID:   "",
				Name: "foo",
			},
			expected: "",
		},
		{
			name: "",
			input: struct {
				ID   string
				Name string
			}{
				ID:   "xxx",
				Name: "foo",
			},
			expected: "xxx",
		},
		{
			name: "",
			input: &foo{
				Model: gorm.Model{
					ID: 1,
				},
				Version: "v1",
			},
			expected: "1_v1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ObjectID(test.input); got != test.expected {
				t.Errorf("ObjectID() = %v, expected %v", got, test.expected)
			}
		})
	}
}
