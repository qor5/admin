package activity

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestFirstUpperWord(t *testing.T) {
	assert.Equal(t, firstUpperWord(""), "")
	assert.Equal(t, firstUpperWord("xxx"), "X")
	assert.Equal(t, firstUpperWord("Yxx"), "Y")
	assert.Equal(t, firstUpperWord("你好"), "你")
	assert.Equal(t, firstUpperWord("フィールド"), "フ")
}

func TestKeysValue(t *testing.T) {
	type InnerStruct struct {
		Field1 string
		Field2 int
		Field3 string
	}

	type OuterStruct struct {
		InnerStruct
		Field3 float64
		Field4 string
	}

	tests := []struct {
		name     string
		input    interface{}
		keys     []string
		expected string
	}{
		{
			name:     "Empty input",
			input:    nil,
			keys:     []string{"Field1", "Field2"},
			expected: "",
		},
		{
			name: "Struct with embedded field",
			input: OuterStruct{
				InnerStruct: InnerStruct{
					Field1: "Value1",
					Field2: 42,
					Field3: "Value3",
				},
				Field3: 3.14,
				Field4: "",
			},
			keys:     []string{"Field1", "Field3"},
			expected: "Value1:3.14",
		},
		{
			name: "Struct with empty fields",
			input: OuterStruct{
				Field3: 3.14,
				Field4: "",
			},
			keys:     []string{"Field1", "Field3", "Field4"},
			expected: ":3.14:",
		},
		{
			name: "input pointer",
			input: &OuterStruct{
				Field3: 3.14,
				Field4: "",
			},
			keys:     []string{"Field1", "Field3", "Field4"},
			expected: ":3.14:",
		},
		{
			name: "Struct without embedded field",
			input: struct {
				Field1 string
				Field2 int
			}{
				Field1: "Value1",
				Field2: 42,
			},
			keys:     []string{"Field1", "Field2"},
			expected: "Value1:42",
		},
		{
			name: "Struct with fields not exists",
			input: struct {
				Field1 string
				Field2 int
			}{
				Field1: "Value1",
				Field2: 42,
			},
			keys:     []string{"Field1", "Field2", "FieldNotExist"},
			expected: "Value1:42",
		},
		{
			name: "Struct with gorm.Model embedded",
			input: struct {
				gorm.Model
				Name string
			}{
				Name: "foo",
			},
			keys:     []string{"ID", "Name"},
			expected: "0:foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := keysValue(test.input, test.keys)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParseGormPrimaryFieldNames(t *testing.T) {
	type TestModel struct {
		ID   uint `gorm:"primary_key"`
		Name string
	}

	type EmbeddedModel struct {
		TestModel
	}

	type NestedModel struct {
		EmbeddedModel
		Version string `gorm:"primary_key"`
	}

	tests := []struct {
		Name     string
		Model    interface{}
		Expected []string
	}{
		{
			Name:     "TestModel",
			Model:    TestModel{},
			Expected: []string{"ID"},
		},
		{
			Name:     "EmbeddedModel",
			Model:    EmbeddedModel{},
			Expected: []string{"ID"},
		},
		{
			Name:     "NestedModel",
			Model:    NestedModel{},
			Expected: []string{"ID", "Version"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fields, err := ParseGormPrimaryFieldNames(test.Model)
			if err != nil {
				t.Errorf("Error occurred: %v", err)
			}
			if !reflect.DeepEqual(fields, test.Expected) {
				t.Errorf("Expected primary fields %v, but got %v", test.Expected, fields)
			}
		})
	}
}
