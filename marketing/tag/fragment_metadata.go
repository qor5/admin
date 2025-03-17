package tag

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// SkipOperatorType defines the type of condition operators
type SkipOperatorType string

const (
	SkipOperatorIN SkipOperatorType = "IN" // IN operator checks if a value is in a list
	SkipOperatorEQ SkipOperatorType = "EQ" // EQ operator checks if a value equals another value
)

// ErrorShouldSkipValidate is returned when validation is skipped due to conditional logic
var ErrorShouldSkipValidate = errors.New("validation should be skipped")

// Metadata returns the fragment's metadata
func (f FragmentMetadata) Metadata() FragmentMetadata {
	return f
}

// ShouldSkip determines if a field should be skipped based on conditions
// Returns (skip, error) where skip indicates if validation should be skipped
// and error reports any issues during condition evaluation
func (f *FragmentMetadata) ShouldSkip(params map[string]any) (bool, error) {
	if f.SkipIf != nil && len(f.SkipIf) > 0 {
		matches, err := f.skipIf(params, f.SkipIf)
		if err != nil {
			return false, errors.Wrap(err, "evaluating skipIf condition")
		}
		if matches {
			return true, nil
		}
	}

	if f.SkipUnless != nil && len(f.SkipUnless) > 0 {
		matches, err := f.skipIf(params, f.SkipUnless)
		if err != nil {
			return false, errors.Wrap(err, "evaluating skipUnless condition")
		}
		if !matches {
			return true, nil
		}
	}

	return false, nil
}

// skipIf checks if the provided values match all conditions in the condition map
func (f *FragmentMetadata) skipIf(values map[string]any, condition map[string]any) (bool, error) {
	if condition == nil || len(condition) == 0 {
		return false, nil
	}

	for key, conditionValue := range condition {
		var fieldName string
		var expr map[string]any
		if strings.HasPrefix(key, "$") {
			fieldName = key[1:] // Remove $ prefix

			exprValue, ok := conditionValue.(map[string]any)
			if !ok {
				return false, errors.Errorf("expression for field %q must be a map type", fieldName)
			}
			expr = exprValue
		} else {
			fieldName = key // Use key directly as field name
			expr = map[string]any{string(SkipOperatorEQ): conditionValue}
		}
		fieldValue, exists := values[fieldName]
		if !exists {
			return false, errors.Errorf("field %q referenced in condition does not exist", fieldName)
		}
		matches, err := f.matchExpression(fieldValue, expr)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// matchExpression evaluates expression conditions
func (f *FragmentMetadata) matchExpression(fieldValue any, expr map[string]any) (bool, error) {
	// Return error when no operators exist
	if len(expr) == 0 {
		return false, errors.New("no valid operators found in expression")
	}

	for op, operands := range expr {
		// Ensure all operators must evaluate to true (AND logic)
		// Short-circuit return on first false condition for efficiency
		switch SkipOperatorType(strings.ToUpper(op)) {
		case SkipOperatorIN:
			operandsValue := reflect.ValueOf(operands)
			if operandsValue.Kind() != reflect.Slice && operandsValue.Kind() != reflect.Array {
				return false, errors.Errorf("IN operator requires a slice or array, got %T", operands)
			}

			found := false
			for i := 0; i < operandsValue.Len(); i++ {
				item := operandsValue.Index(i).Interface()
				if reflect.DeepEqual(fieldValue, item) {
					found = true
					break
				}
			}
			// Short-circuit on first false condition
			if !found {
				return false, nil
			}

		case SkipOperatorEQ:
			// Short-circuit on first false condition
			if !reflect.DeepEqual(fieldValue, operands) {
				return false, nil
			}

		default:
			return false, errors.Errorf("unsupported operator %q", op)
		}
	}

	// Only return true when all conditions are satisfied
	return true, nil
}

// Validate validates this metadata against the provided parameters
func (f *FragmentMetadata) Validate(ctx context.Context, params map[string]any) error {
	skip, err := f.ShouldSkip(params)
	if err != nil {
		return errors.Wrap(err, "checking if validation should be skipped")
	}
	if skip {
		return ErrorShouldSkipValidate
	}

	if f.Key == "" {
		return nil // Allow empty key, but don't perform further validation
	}

	value, exists := params[f.Key]
	if f.Required && (!exists || value == nil) {
		return errors.Errorf("required parameter missing: %s", f.Key)
	}

	if exists && f.Validation != nil && f.Validation.Pattern != "" {
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case fmt.Stringer:
			strValue = v.String()
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		re, err := regexp.Compile(f.Validation.Pattern)
		if err != nil {
			return errors.Wrapf(err, "invalid validation pattern for %q", f.Key)
		}

		if !re.MatchString(strValue) {
			errMsg := f.Validation.ErrorMessage
			if errMsg == "" {
				errMsg = fmt.Sprintf("value for %q does not match required pattern", f.Key)
			}
			return errors.New(errMsg)
		}
	}

	return nil
}
