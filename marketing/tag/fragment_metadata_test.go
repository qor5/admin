package tag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFragmentMetadata_matchExpression verifies that expression matching logic works correctly
// for different operator types to ensure fragment conditional display rules function properly
func TestFragmentMetadata_matchExpression(t *testing.T) {
	metadata := &FragmentMetadata{}

	tests := []struct {
		name      string
		value     any
		expr      map[string]any
		expected  bool
		expectErr bool
	}{
		{
			name:  "IN operator - value in list",
			value: "NE",
			expr: map[string]any{
				string(SkipOperatorIN): []string{"EQ", "NE", "LT"},
			},
			expected:  true,
			expectErr: false,
		},
		{
			name:  "IN operator - value not in list",
			value: "GT",
			expr: map[string]any{
				string(SkipOperatorIN): []string{"EQ", "NE", "LT"},
			},
			expected:  false,
			expectErr: false,
		},
		{
			name:  "EQ operator - values equal",
			value: "admin",
			expr: map[string]any{
				string(SkipOperatorEQ): "admin",
			},
			expected:  true,
			expectErr: false,
		},
		{
			name:  "EQ operator - values not equal",
			value: "user",
			expr: map[string]any{
				string(SkipOperatorEQ): "admin",
			},
			expected:  false,
			expectErr: false,
		},
		{
			name:  "EQ operator - numeric values equal",
			value: 42,
			expr: map[string]any{
				string(SkipOperatorEQ): 42,
			},
			expected:  true,
			expectErr: false,
		},
		{
			name:  "EQ operator - numeric values not equal",
			value: 42,
			expr: map[string]any{
				string(SkipOperatorEQ): 24,
			},
			expected:  false,
			expectErr: false,
		},
		{
			name:  "Unsupported operator",
			value: "admin",
			expr: map[string]any{
				"NE": "user",
			},
			expected:  false,
			expectErr: true,
		},
		{
			name:      "Empty expression",
			value:     "admin",
			expr:      map[string]any{},
			expected:  false,
			expectErr: true,
		},
		{
			name:  "IN operator with non-slice value",
			value: "admin",
			expr: map[string]any{
				string(SkipOperatorIN): "not-a-slice",
			},
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := metadata.matchExpression(tt.value, tt.expr)

			// Check error expectation
			if (err != nil) != tt.expectErr {
				t.Errorf("matchExpression() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			// If we expect an error, no need to check the boolean result
			if tt.expectErr {
				return
			}

			// Check boolean result
			if result != tt.expected {
				t.Errorf("matchExpression() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFragmentMetadata_SkipIf(t *testing.T) {
	// Create a view with conditional logic
	view := &View{
		Fragments: []Fragment{
			&SelectFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "operator",
					Required: true,
				},
				Options: []*Option{
					{Value: "EQ", Label: "Equals"},
					{Value: "NE", Label: "Not Equals"},
					{Value: "LT", Label: "Less Than"},
					{Value: "GT", Label: "Greater Than"},
					{Value: "BETWEEN", Label: "Between"},
				},
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "value",
					Required: true,
				},
				Min: 0,
				Max: 100,
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "maxValue",
					Required: true,
					// Skip validation of maxValue when operator is not BETWEEN
					SkipIf: map[string]any{
						"$operator": map[string]any{
							string(SkipOperatorIN): []string{"EQ", "NE", "LT", "GT"},
						},
					},
				},
				Min: 0,
				Max: 100,
			},
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		params        map[string]any
		expectedError string
	}{
		{
			name: "Scenario 1 - operator=EQ, no maxValue",
			params: map[string]any{
				"operator": "EQ",
				"value":    float64(50),
				// Note: maxValue is omitted
			},
			expectedError: "",
		},
		{
			name: "Scenario 2 - operator=BETWEEN, no maxValue",
			params: map[string]any{
				"operator": "BETWEEN",
				"value":    float64(30),
				// Missing required maxValue
			},
			expectedError: "required parameter missing: maxValue",
		},
		{
			name: "Scenario 3 - operator=BETWEEN, with maxValue",
			params: map[string]any{
				"operator": "BETWEEN",
				"value":    float64(30),
				"maxValue": float64(70),
			},
			expectedError: "",
		},
	}

	// Execute tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Only test the maxValue fragment directly
			fragment := view.Fragments[2].(*NumberInputFragment)
			err := fragment.FragmentMetadata.Validate(context.Background(), tc.params)

			// For skipped validations, we should get ErrorShouldSkipValidate
			if tc.name == "Scenario 1 - operator=EQ, no maxValue" {
				assert.Equal(t, ErrorShouldSkipValidate, err)
				return
			}

			// Otherwise check the expected error
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestFragmentMetadata_SkipUnless(t *testing.T) {
	// Create a view with conditional logic and nested condition
	view := &View{
		Fragments: []Fragment{
			&SelectFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "operator",
					Required: true,
				},
				Options: []*Option{
					{Value: "EQ", Label: "Equals"},
					{Value: "NE", Label: "Not Equals"},
					{Value: "LT", Label: "Less Than"},
					{Value: "GT", Label: "Greater Than"},
					{Value: "BETWEEN", Label: "Between"},
				},
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "value",
					Required: true,
				},
				Min: 0,
				Max: 100,
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "maxValue",
					Required: true,
					// Only validate maxValue when operator is BETWEEN
					SkipUnless: map[string]any{
						"$operator": map[string]any{
							string(SkipOperatorIN): []string{"BETWEEN"},
						},
					},
				},
				Min: 0,
				Max: 100,
			},
			&TextInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "advancedOptions",
					Required: true,
				},
			},
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		params        map[string]any
		expectedError string
	}{
		{
			name: "Scenario 1 - operator=EQ, no maxValue",
			params: map[string]any{
				"operator": "EQ",
				"value":    float64(50),
				// maxValue is omitted and should be skipped
				"advancedOptions": "standard",
			},
			expectedError: "",
		},
		{
			name: "Scenario 2 - operator=BETWEEN, missing maxValue",
			params: map[string]any{
				"operator": "BETWEEN",
				"value":    float64(30),
				// Missing required maxValue
				"advancedOptions": "inclusive",
			},
			expectedError: "required parameter missing: maxValue",
		},
		{
			name: "Scenario 3 - operator=BETWEEN, all required fields",
			params: map[string]any{
				"operator":        "BETWEEN",
				"value":           float64(30),
				"maxValue":        float64(70),
				"advancedOptions": "inclusive",
			},
			expectedError: "",
		},
		{
			name: "Scenario 4 - operator=GT, no maxValue",
			params: map[string]any{
				"operator":        "GT",
				"value":           float64(30),
				"advancedOptions": "exclusive",
				// maxValue can be omitted
			},
			expectedError: "",
		},
	}

	// Execute tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Only test the maxValue fragment directly
			fragment := view.Fragments[2].(*NumberInputFragment)
			err := fragment.FragmentMetadata.Validate(context.Background(), tc.params)

			// For skipped validations, we should get ErrorShouldSkipValidate
			if tc.name == "Scenario 1 - operator=EQ, no maxValue" ||
				tc.name == "Scenario 4 - operator=GT, no maxValue" {
				assert.Equal(t, ErrorShouldSkipValidate, err)
				return
			}

			// Otherwise check the expected error
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}
