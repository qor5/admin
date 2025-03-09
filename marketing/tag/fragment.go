package tag

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FragmentType defines the type of template fragment
type FragmentType string

const (
	FragmentDatePicker  FragmentType = "DATE_PICKER"
	FragmentNumberInput FragmentType = "NUMBER_INPUT"
	FragmentTextInput   FragmentType = "TEXT_INPUT"
	FragmentSelect      FragmentType = "SELECT"
	FragmentIcon        FragmentType = "ICON"
	FragmentText        FragmentType = "TEXT"
	FragmentHidden      FragmentType = "HIDDEN"
)

// TextInputFragment represents a text input fragment
var _ Fragment = &TextInputFragment{}

type TextInputFragment struct {
	FragmentMetadata
}

// Type returns the fragment type
func (f *TextInputFragment) Type() FragmentType {
	return FragmentTextInput
}

// Validate validates the text input fragment
func (f *TextInputFragment) Validate(ctx context.Context, params map[string]any) error {
	if f.Key == "" {
		return errors.New("text input fragment must have a non-empty key")
	}

	if err := f.FragmentMetadata.Validate(ctx, params); err != nil {
		return err
	}

	// Additional validation for text input
	if value, exists := params[f.Key]; exists {
		if _, ok := value.(string); !ok {
			return errors.Errorf("value for %q must be a string", f.Key)
		}
	}
	return nil
}

// NumberInputFragment represents a number input fragment
var _ Fragment = &NumberInputFragment{}

type NumberInputFragment struct {
	FragmentMetadata
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Type returns the fragment type
func (f *NumberInputFragment) Type() FragmentType {
	return FragmentNumberInput
}

// Validate validates the number input fragment
func (f *NumberInputFragment) Validate(ctx context.Context, params map[string]any) error {
	if f.Key == "" {
		return errors.New("number input fragment must have a non-empty key")
	}

	if err := f.FragmentMetadata.Validate(ctx, params); err != nil {
		return err
	}

	// Additional validation for number input
	if value, exists := params[f.Key]; exists {
		numValue, ok := value.(float64)
		if !ok {
			return errors.Errorf("value for %q must be a number", f.Key)
		}
		// Only validate non-zero min/max to avoid false positives
		if f.Min != 0 && numValue < f.Min {
			return errors.Errorf("value for %q must be at least %v", f.Key, f.Min)
		}

		if f.Max != 0 && numValue > f.Max {
			return errors.Errorf("value for %q must be at most %v", f.Key, f.Max)
		}
	}

	return nil
}

// DatePickerFragment represents a date picker fragment
var _ Fragment = &DatePickerFragment{}

type DatePickerFragment struct {
	FragmentMetadata
	Min         time.Time `json:"min"`
	Max         time.Time `json:"max"`
	IncludeTime bool      `json:"includeTime"`
}

// Type returns the fragment type
func (f *DatePickerFragment) Type() FragmentType {
	return FragmentDatePicker
}

// Validate validates the date picker fragment
func (f *DatePickerFragment) Validate(ctx context.Context, params map[string]any) error {
	if f.Key == "" {
		return errors.New("date picker fragment must have a non-empty key")
	}

	if err := f.FragmentMetadata.Validate(ctx, params); err != nil {
		return err
	}

	var dateValue time.Time

	// Additional validation for date picker
	if value, exists := params[f.Key]; exists {
		strValue, ok := value.(string)
		if !ok {
			return errors.Errorf("value for %q must be a date string", f.Key)
		}

		// Parse the date string
		layout := time.RFC3339
		if !f.IncludeTime {
			layout = "2006-01-02"
		}

		var err error
		dateValue, err = time.Parse(layout, strValue)
		if err != nil {
			return errors.Errorf("value for %q must be a valid date", f.Key)
		}

		// Validate min date
		if !f.Min.IsZero() && dateValue.Before(f.Min) {
			return errors.Errorf("value for %q must not be before %v", f.Key, f.Min)
		}

		// Validate max date
		if !f.Max.IsZero() && dateValue.After(f.Max) {
			return errors.Errorf("value for %q must not be after %v", f.Key, f.Max)
		}
	}

	return nil
}

// Option represents a selection option
type Option struct {
	Value any    `json:"value"`
	Label string `json:"label"`
}

// SelectFragment represents a dropdown selector fragment
var _ Fragment = &SelectFragment{}

type SelectFragment struct {
	FragmentMetadata
	Options  []*Option `json:"options"`
	Multiple bool      `json:"multiple"`
}

// Type returns the fragment type
func (f *SelectFragment) Type() FragmentType {
	return FragmentSelect
}

// Validate validates the select fragment
func (f *SelectFragment) Validate(ctx context.Context, params map[string]any) error {
	if f.Key == "" {
		return errors.New("select fragment must have a non-empty key")
	}

	// Validate that options are provided
	if len(f.Options) == 0 {
		return errors.New("select fragment must have at least one option")
	}

	// Call parent Validate method
	if err := f.FragmentMetadata.Validate(ctx, params); err != nil {
		return err
	}

	// Additional validation for select
	if value, exists := params[f.Key]; exists && value != nil {
		var valueSlice []any

		if f.Multiple {
			var ok bool
			valueSlice, ok = toSlice(value)
			if !ok {
				return errors.Errorf("value for %q must be an array", f.Key)
			}
		} else {
			valueSlice = []any{value}
		}

		for _, item := range valueSlice {
			found := false
			for _, option := range f.Options {
				if reflect.DeepEqual(option.Value, item) {
					found = true
					break
				}
			}

			if !found {
				return errors.Errorf("value for %q contains an invalid option", f.Key)
			}
		}
	}

	return nil
}

// toSlice converts various input types to a slice of values
// Supports converting actual slices, arrays, and comma-separated strings
func toSlice(value any) ([]any, bool) {
	rv := reflect.ValueOf(value)

	// If value is already a slice or array
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result, true
	}

	// If value is a comma-separated string
	if str, ok := value.(string); ok {
		if str == "" {
			return []any{}, true
		}
		items := strings.Split(str, ",")
		result := make([]any, len(items))
		for i, item := range items {
			result[i] = strings.TrimSpace(item)
		}
		return result, true
	}

	// Not a slice
	return nil, false
}

// IconFragment represents an icon fragment
var _ Fragment = &IconFragment{}

type IconFragment struct {
	FragmentMetadata
}

// Type returns the fragment type
func (f *IconFragment) Type() FragmentType {
	return FragmentIcon
}

// TextFragment represents a static text fragment
var _ Fragment = &TextFragment{}

type TextFragment struct {
	FragmentMetadata
	Text string `json:"text"`
}

// Type returns the fragment type
func (f *TextFragment) Type() FragmentType {
	return FragmentText
}

// HiddenFragment represents a hidden field fragment
var _ Fragment = &HiddenFragment{}

type HiddenFragment struct {
	FragmentMetadata
}

// Type returns the fragment type
func (f *HiddenFragment) Type() FragmentType {
	return FragmentHidden
}

// RegisterStandardFragments registers all standard fragment types with the given registry
// This enables the registry to create the appropriate fragment type based on the type string
func RegisterStandardFragments(registry *Registry) error {
	fragmentFactories := map[FragmentType]func() Fragment{
		FragmentTextInput:   func() Fragment { return &TextInputFragment{} },
		FragmentNumberInput: func() Fragment { return &NumberInputFragment{} },
		FragmentDatePicker:  func() Fragment { return &DatePickerFragment{} },
		FragmentSelect:      func() Fragment { return &SelectFragment{} },
		FragmentIcon:        func() Fragment { return &IconFragment{} },
		FragmentText:        func() Fragment { return &TextFragment{} },
		FragmentHidden:      func() Fragment { return &HiddenFragment{} },
	}

	for fragmentType, factory := range fragmentFactories {
		if err := registry.RegisterFragment(fragmentType, factory); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	// Register standard fragments with the default registry
	if err := RegisterStandardFragments(DefaultRegistry); err != nil {
		panic(err)
	}
}
