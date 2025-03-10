package tag

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestViewJSON verifies that view structures correctly serialize to JSON
// to ensure client applications receive properly formatted view definitions
func TestViewJSON(t *testing.T) {
	// Parse test dates
	minDate, err := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	require.NoError(t, err)
	maxDate, err := time.Parse(time.RFC3339, "2023-12-31T23:59:59Z")
	require.NoError(t, err)

	view := &View{
		Fragments: []Fragment{
			&TextInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "username",
					Required: true,
				},
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "age",
					Required: false,
				},
				Min: 0,
				Max: 1000,
			},
			&DatePickerFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "birthday",
					Required: true,
				},
				Min:         minDate,
				Max:         maxDate,
				IncludeTime: false,
			},
			&SelectFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "gender",
					Required: true,
				},
				Options: []*Option{
					{Value: "MALE", Label: "Male"},
					{Value: "FEMALE", Label: "Female"},
				},
			},
			&SelectFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "interests",
					Required: false,
				},
				Multiple: true,
				Options: []*Option{
					{Value: "sports", Label: "Sports"},
					{Value: "music", Label: "Music"},
					{Value: "reading", Label: "Reading"},
				},
			},
			&IconFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "status",
					Required: false,
				},
			},
			&TextFragment{
				FragmentMetadata: FragmentMetadata{
					Key: "subtitle",
				},
				Text: "This is a static text",
			},
			&HiddenFragment{
				FragmentMetadata: FragmentMetadata{
					Key:          "hidden_field",
					DefaultValue: "hidden_value",
				},
			},
			&SelectFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "status",
					Required: true,
					Validation: &Validation{
						Pattern:      `^(active|inactive)$`,
						ErrorMessage: "Status must be either 'active' or 'inactive'",
					},
				},
				Options: []*Option{
					{Value: "active", Label: "Active"},
					{Value: "inactive", Label: "Inactive"},
				},
			},
		},
	}

	data, err := json.Marshal(view)
	require.NoError(t, err)

	t.Logf("unmarshaledView: %s", string(data))

	var unmarshaledView *View
	require.NoError(t, json.Unmarshal(data, &unmarshaledView))

	assert.Equal(t, view, unmarshaledView)
}

func TestViewJSON_PrettyPrint(t *testing.T) {
	view := &View{
		Fragments: []Fragment{
			&TextInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "username",
					Required: true,
				},
			},
		},
	}

	data, err := json.MarshalIndent(view, "", "  ")
	require.NoError(t, err)

	t.Logf("Pretty-printed JSON:\n%s", string(data))

	var unmarshaledView View
	require.NoError(t, json.Unmarshal(data, &unmarshaledView))

	assert.Equal(t, 1, len(unmarshaledView.Fragments))
	assert.Equal(t, FragmentTextInput, unmarshaledView.Fragments[0].Type())
	assert.Equal(t, "username", unmarshaledView.Fragments[0].Metadata().Key)
}

func TestViewValidate(t *testing.T) {
	tests := []struct {
		name           string
		view           *View
		params         map[string]any
		expectErrorMsg string
	}{
		{
			name: "valid params",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "username",
							Required: true,
						},
					},
				},
			},
			params: map[string]any{
				"username": "testuser",
			},
			expectErrorMsg: "",
		},
		{
			name: "missing required param",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "username",
							Required: true,
						},
					},
				},
			},
			params:         map[string]any{},
			expectErrorMsg: "required parameter missing: username",
		},
		{
			name: "validation pattern success",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "email",
							Required: true,
							Validation: &Validation{
								Pattern:      `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
								ErrorMessage: "Invalid email format",
							},
						},
					},
				},
			},
			params: map[string]any{
				"email": "test@example.com",
			},
		},
		{
			name: "validation pattern failure",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "email",
							Required: true,
							Validation: &Validation{
								Pattern:      `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
								ErrorMessage: "Invalid email format",
							},
						},
					},
				},
			},
			params: map[string]any{
				"email": "invalid-email",
			},
			expectErrorMsg: "Invalid email format",
		},
		{
			name: "invalid validation pattern",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "code",
							Required: true,
							Validation: &Validation{
								Pattern:      "[", // Invalid regex pattern
								ErrorMessage: "Invalid code format",
							},
						},
					},
				},
			},
			params: map[string]any{
				"code": "ABC123",
			},
			expectErrorMsg: "invalid validation pattern for \"code\"",
		},
		{
			name: "non-string value with pattern validation",
			view: &View{
				Fragments: []Fragment{
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "count",
							Required: true,
							Validation: &Validation{
								Pattern:      `^[0-9]+$`,
								ErrorMessage: "Count must be a positive integer",
							},
						},
					},
				},
			},
			params: map[string]any{
				"count": float64(42), // Number, not string
			},
		},
		{
			name: "non-string value failing pattern validation",
			view: &View{
				Fragments: []Fragment{
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "count",
							Required: true,
							Validation: &Validation{
								Pattern:      `^[a-z]+$`,
								ErrorMessage: "Count must contain only lowercase letters",
							},
						},
					},
				},
			},
			params: map[string]any{
				"count": float64(42), // Number, not string
			},
			expectErrorMsg: "Count must contain only lowercase letters",
		},
		{
			name: "string value with pattern validation",
			view: &View{
				Fragments: []Fragment{
					&TextInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "status",
							Required: true,
							Validation: &Validation{
								Pattern:      `^(active|inactive)$`,
								ErrorMessage: "Status must be either 'active' or 'inactive'",
							},
						},
					},
				},
			},
			params: map[string]any{
				"status": "active",
			},
		},
		{
			name: "text fragment without ID",
			view: &View{
				Fragments: []Fragment{
					&TextFragment{
						Text: "Static text doesn't need validation",
					},
				},
			},
			params: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.view.Validate(context.Background(), tt.params)
			if tt.expectErrorMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorMsg)
			}
		})
	}
}

func TestViewValidate_ShouldSkipValidate(t *testing.T) {
	// Create a view with a fragment that should be skipped based on conditions
	view := &View{
		Fragments: []Fragment{
			&TextInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "name",
					Required: true,
				},
			},
			&NumberInputFragment{
				FragmentMetadata: FragmentMetadata{
					Key:      "age",
					Required: true,
					// Skip validation when role is "guest"
					SkipIf: map[string]any{
						"role": "guest",
					},
				},
			},
		},
	}

	// Test case where age should be skipped
	params := map[string]any{
		"name": "John",
		"role": "guest",
		// Note: age is missing but should be skipped
	}

	// We've modified the code to return ErrorShouldSkipValidate, but View.Validate should handle it
	err := view.Validate(context.Background(), params)
	require.NoError(t, err)

	// Change params to make the skip condition not apply
	params["role"] = "admin"
	// Now validation should fail because age is required
	err = view.Validate(context.Background(), params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "required parameter missing: age")
}
