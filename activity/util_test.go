package activity_test

import (
	"reflect"
	"testing"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/publish"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestFirstUpperWord(t *testing.T) {
	assert.Equal(t, activity.FirstUpperWord(""), "")
	assert.Equal(t, activity.FirstUpperWord("xxx"), "X")
	assert.Equal(t, activity.FirstUpperWord("Yxx"), "Y")
	assert.Equal(t, activity.FirstUpperWord("你好"), "你")
	assert.Equal(t, activity.FirstUpperWord("フィールド"), "フ")
}

func TestModelName(t *testing.T) {
	type TestActivityModel struct {
		ID uint `gorm:"primaryKey"`
	}
	assert.Equal(t, "TestActivityModel", activity.ParseModelName(&TestActivityModel{}))
	assert.Equal(t, "TestActivityModel", activity.ParseModelName(TestActivityModel{}))
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

	type Version struct {
		*publish.Version
		ID string
	}

	tests := []struct {
		name     string
		input    any
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
		{
			name: "Struct with Version",
			input: struct {
				publish.Version
				Name string
			}{
				Version: publish.Version{
					Version: "ver0",
				},
				Name: "bar",
			},
			keys:     []string{"Version", "Name"},
			expected: "ver0:bar",
		},
		{
			name: "Struct with Version ptr",
			input: struct {
				*publish.Version
				Name string
			}{
				Version: &publish.Version{
					Version: "ver0",
				},
				Name: "bar",
			},
			keys:     []string{"Version", "Name"},
			expected: "ver0:bar",
		},
		{
			name: "Struct with Version multilevel",
			input: struct {
				*Version
				Name string
			}{
				Version: &Version{
					Version: &publish.Version{
						Version: "ver0",
					},
					ID: "foo",
				},
				Name: "bar",
			},
			keys:     []string{"ID", "Version", "Name"},
			expected: "foo:ver0:bar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := activity.KeysValue(test.input, test.keys, ":")
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestParsePrimaryFields(t *testing.T) {
	type TestModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	type EmbeddedModel struct {
		TestModel
	}

	type NestedModel struct {
		EmbeddedModel
		Version string `gorm:"primaryKey"`
	}

	tests := []struct {
		Name     string
		Model    any
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
			s, err := activity.ParseSchema(test.Model)
			if err != nil {
				t.Errorf("Error occurred: %v", err)
			}
			keys := lo.Map(s.PrimaryFields, func(f *schema.Field, _ int) string {
				return f.Name
			})
			if !reflect.DeepEqual(keys, test.Expected) {
				t.Errorf("Expected primary fields %v, but got %v", test.Expected, s.PrimaryFields)
			}
		})
	}
}

func TestParsePrimaryKeys(t *testing.T) {
	type TestModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	type EmbeddedModel struct {
		TestModel
	}

	type NestedModel struct {
		*EmbeddedModel
		Version string `gorm:"primaryKey"`
	}

	type Version struct {
		Version     string `gorm:"primaryKey"`
		VersionName string
	}

	type WithVersion struct {
		ID uint `gorm:"primaryKey"`
		Version
	}

	type WithVersionAndIgnore struct {
		ID      uint `gorm:"primaryKey"`
		Version `gorm:"-"`
	}

	type EmbeddedVersion struct {
		Version     string `gorm:"primaryKey"`
		VersionName string
	}

	type WithEmbeddedVersion struct {
		ID      uint   `gorm:"primaryKey"`
		Version string `gorm:"primaryKey"`
		EmbeddedVersion
	}

	type WithEmbeddedVersionAndIgnore struct {
		ID      uint   `gorm:"primaryKey"`
		Version string `gorm:"-"`
		EmbeddedVersion
	}

	tests := []struct {
		Name        string
		Model       any
		UseBindName bool
		Expected    []string
	}{
		{
			Name:        "TestModel",
			Model:       TestModel{},
			UseBindName: true,
			Expected:    []string{"ID"},
		},
		{
			Name:        "TestModel",
			Model:       TestModel{},
			UseBindName: false,
			Expected:    []string{"ID"},
		},
		{
			Name:        "EmbeddedModel",
			Model:       EmbeddedModel{},
			UseBindName: true,
			Expected:    []string{"TestModel.ID"},
		},
		{
			Name:        "EmbeddedModel",
			Model:       EmbeddedModel{},
			UseBindName: false,
			Expected:    []string{"ID"},
		},
		{
			Name:        "NestedModel",
			Model:       NestedModel{},
			UseBindName: true,
			Expected:    []string{"EmbeddedModel.TestModel.ID", "Version"},
		},
		{
			Name:        "NestedModel",
			Model:       NestedModel{},
			UseBindName: false,
			Expected:    []string{"ID", "Version"},
		},
		{
			Name:        "WithVersion",
			Model:       WithVersion{},
			UseBindName: true,
			Expected:    []string{"ID", "Version.Version"},
		},
		{
			Name:        "WithVersion",
			Model:       WithVersion{},
			UseBindName: false,
			Expected:    []string{"ID", "Version"},
		},
		{
			Name:        "WithVersionAndIgnore",
			Model:       WithVersionAndIgnore{},
			UseBindName: true,
			Expected:    []string{"ID"},
		},
		{
			Name:        "WithVersionAndIgnore",
			Model:       WithVersionAndIgnore{},
			UseBindName: false,
			Expected:    []string{"ID"},
		},
		{
			Name:        "WithEmbeddedVersion",
			Model:       WithEmbeddedVersion{},
			UseBindName: true,
			Expected:    []string{"ID", "Version"},
		},
		{
			Name:        "WithEmbeddedVersion",
			Model:       WithEmbeddedVersion{},
			UseBindName: false,
			Expected:    []string{"ID", "Version"},
		},
		{
			Name:        "WithEmbeddedVersionAndIgnore",
			Model:       WithEmbeddedVersionAndIgnore{},
			UseBindName: true,
			Expected:    []string{"ID", "EmbeddedVersion.Version"},
		},
		{
			Name:        "WithEmbeddedVersionAndIgnore",
			Model:       WithEmbeddedVersionAndIgnore{},
			UseBindName: false,
			Expected:    []string{"ID", "Version"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			{
				keys := activity.ParsePrimaryKeys(test.Model, test.UseBindName)
				if !reflect.DeepEqual(keys, test.Expected) {
					t.Errorf("Expected primary fields %v, but got %v", test.Expected, keys)
				}
			}
			{ // ptr test
				keys := activity.ParsePrimaryKeys(&(test.Model), test.UseBindName)
				if !reflect.DeepEqual(keys, test.Expected) {
					t.Errorf("Expected primary fields %v, but got %v", test.Expected, keys)
				}
			}
		})
	}
}
