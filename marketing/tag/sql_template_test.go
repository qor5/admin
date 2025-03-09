package tag

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestGenderTemplate builds a template for gender-based filtering
// to be used in unit tests validating SQL generation behavior
func createTestGenderTemplate(categoryID string) *SQLTemplate {
	if categoryID == "" {
		categoryID = "demographic"
	}
	return NewSQLTemplate(
		&Metadata{
			ID:          "gender",
			Name:        "Gender",
			Description: "User gender filter",
			CategoryID:  categoryID,
			View: &View{
				Fragments: []Fragment{
					&TextFragment{
						Text: "Gender is ",
					},
					&SelectFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "values",
							Required: true,
							Validation: &Validation{
								Pattern:      ".+",
								ErrorMessage: "Please select at least one gender",
							},
						},
						Multiple: true,
						Options: []*Option{
							{Value: "MALE", Label: "Male"},
							{Value: "FEMALE", Label: "Female"},
						},
					},
				},
			},
		},
		"SELECT uid FROM users WHERE gender IN ({{argEach .values}})",
	)
}

func createTestAgeTemplate(categoryID string) *SQLTemplate {
	if categoryID == "" {
		categoryID = "demographic"
	}
	return NewSQLTemplate(
		&Metadata{
			ID:          "age",
			Name:        "Age",
			Description: "User age range",
			CategoryID:  categoryID,
			View: &View{
				Fragments: []Fragment{
					&TextFragment{
						Text: "Age between ",
					},
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "min",
							Required:     true,
							DefaultValue: float64(0),
							Validation: &Validation{
								Pattern:      "^\\d+$",
								ErrorMessage: "Please enter a valid age",
							},
						},
						Min: 0,
						Max: 1000,
					},
					&TextFragment{
						Text: " and ",
					},
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "max",
							Required:     true,
							DefaultValue: float64(100),
							Validation: &Validation{
								Pattern:      "^\\d+$",
								ErrorMessage: "Please enter a valid age",
							},
						},
						Min: 0,
						Max: 1000,
					},
				},
			},
		},
		"SELECT uid FROM users WHERE age BETWEEN {{arg .min}} AND {{arg .max}}",
	)
}

func createTestPurchaseTemplate(categoryID string) *SQLTemplate {
	if categoryID == "" {
		categoryID = "behavior"
	}
	return NewSQLTemplate(
		&Metadata{
			ID:          "purchase_amount",
			Name:        "Purchase Amount",
			Description: "Filter by total purchase amount in a date range",
			CategoryID:  categoryID,
			View: &View{
				Fragments: []Fragment{
					&IconFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "moneyIcon",
							DefaultValue: "money",
						},
					},
					&TextFragment{
						Text: "Purchased between ",
					},
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "minAmount",
							Required:     true,
							DefaultValue: float64(0),
							Validation: &Validation{
								Pattern:      "^\\d+$",
								ErrorMessage: "Please enter a valid amount",
							},
						},
						Min: 0,
						Max: 1000,
					},
					&TextFragment{
						Text: " and ",
					},
					&NumberInputFragment{
						FragmentMetadata: FragmentMetadata{
							Key:          "maxAmount",
							Required:     true,
							DefaultValue: float64(1000000),
							Validation: &Validation{
								Pattern:      "^\\d+$",
								ErrorMessage: "Please enter a valid amount",
							},
						},
						Min: 0,
						Max: 1000,
					},
					&TextFragment{
						Text: " from ",
					},
					&DatePickerFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "startDate",
							Required: true,
							Validation: &Validation{
								Pattern:      "^\\d{4}-\\d{2}-\\d{2}$",
								ErrorMessage: "Please enter a valid date (YYYY-MM-DD)",
							},
						},
					},
					&TextFragment{
						Text: " to ",
					},
					&DatePickerFragment{
						FragmentMetadata: FragmentMetadata{
							Key:      "endDate",
							Required: true,
							Validation: &Validation{
								Pattern:      "^\\d{4}-\\d{2}-\\d{2}$",
								ErrorMessage: "Please enter a valid date (YYYY-MM-DD)",
							},
						},
					},
				},
			},
		},
		`SELECT uid FROM events WHERE event_type = {{arg "purchase"}} AND event_date BETWEEN {{arg .startDate}} AND {{arg .endDate}} GROUP BY uid HAVING SUM(amount) BETWEEN {{arg .minAmount}} AND {{arg .maxAmount}}`,
	)
}

// Tests creation of SQL template
func TestSQLTemplate_Create(t *testing.T) {
	template := createTestGenderTemplate("")
	ctx := context.Background()

	// Validate metadata
	assert.Equal(t, "gender", template.Metadata(ctx).ID)
	assert.Equal(t, "Gender", template.Metadata(ctx).Name)
	assert.Equal(t, "User gender filter", template.Metadata(ctx).Description)

	// Validate view
	assert.Len(t, template.Metadata(ctx).View.Fragments, 2)
}

// Tests building SQL from template
func TestSQLTemplate_BuildSQL(t *testing.T) {
	template := createTestGenderTemplate("")

	sql, err := template.BuildSQL(context.Background(), map[string]any{
		"values": []any{"MALE", "FEMALE"},
	})

	require.NoError(t, err)
	require.NotNil(t, sql)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE gender IN (?, ?)"), CompactSQLQuery(sql.Query))
	assert.Equal(t, []any{"MALE", "FEMALE"}, sql.Args)
}

// Tests building SQL with missing required parameters
func TestSQLTemplate_BuildSQL_MissingParams(t *testing.T) {
	template := createTestGenderTemplate("")

	// No parameters passed
	_, err := template.BuildSQL(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "values")

	// Empty parameter map
	_, err = template.BuildSQL(context.Background(), map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "values")
}

// Tests building SQL with numeric parameters
func TestSQLTemplate_BuildSQL_NumberParams(t *testing.T) {
	template := createTestAgeTemplate("")

	sql, err := template.BuildSQL(context.Background(), map[string]any{
		"min": float64(18),
		"max": float64(65),
	})

	require.NoError(t, err)
	require.NotNil(t, sql)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE age BETWEEN ? AND ?"), CompactSQLQuery(sql.Query))
	assert.Equal(t, []any{float64(18), float64(65)}, sql.Args)
}

// Tests building SQL from a complex template
func TestSQLTemplate_BuildSQL_ComplexTemplate(t *testing.T) {
	template := createTestPurchaseTemplate("")

	sql, err := template.BuildSQL(context.Background(), map[string]any{
		"minAmount": float64(100),
		"maxAmount": float64(1000),
		"startDate": "2023-01-01",
		"endDate":   "2023-12-31",
	})

	require.NoError(t, err)
	require.NotNil(t, sql)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM events WHERE event_type = ? AND event_date BETWEEN ? AND ? GROUP BY uid HAVING SUM(amount) BETWEEN ? AND ?"), CompactSQLQuery(sql.Query))
	assert.Equal(t, []any{"purchase", "2023-01-01", "2023-12-31", float64(100), float64(1000)}, sql.Args)
}

// TestSQLTemplate_BuildSQL_WithParameters tests SQL generation with different parameter placeholder styles
func TestSQLTemplate_BuildSQL_WithParameters(t *testing.T) {
	createTestParameterizedTemplate := func(categoryID string) *SQLTemplate {
		if categoryID == "" {
			categoryID = "behavior"
		}
		return NewSQLTemplate(
			&Metadata{
				ID:          "purchase_param",
				Name:        "Purchase Parameter",
				Description: "Purchase with parameterized values",
				CategoryID:  categoryID,
				View: &View{
					Fragments: []Fragment{
						&TextFragment{
							Text: "Purchased amount between ",
						},
						&NumberInputFragment{
							FragmentMetadata: FragmentMetadata{
								Key:      "minAmount",
								Required: true,
							},
						},
						&TextFragment{
							Text: " and ",
						},
						&NumberInputFragment{
							FragmentMetadata: FragmentMetadata{
								Key:      "maxAmount",
								Required: true,
							},
						},
					},
				},
			},
			`SELECT uid FROM events WHERE event_type = {{arg "purchase"}} AND amount BETWEEN {{arg .minAmount}} AND {{arg .maxAmount}}`,
		)
	}

	params := map[string]any{
		"minAmount": float64(100),
		"maxAmount": float64(1000),
	}

	// Default placeholder style (?)
	template := createTestParameterizedTemplate("")
	sql, err := template.BuildSQL(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, sql)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM events WHERE event_type = ? AND amount BETWEEN ? AND ?"), CompactSQLQuery(sql.Query))
	assert.Equal(t, []any{"purchase", float64(100), float64(1000)}, sql.Args)

	// PostgreSQL style placeholders ($n)
	pgTemplate := NewSQLTemplate(
		&Metadata{
			ID: "pg_params",
		},
		`SELECT uid FROM events WHERE event_type = {{ "purchase" | toUpper | arg }} AND amount BETWEEN {{arg .minAmount}} AND {{arg .maxAmount}} AND location {{ .op }} {{arg .location}}`,
	).WithArgPlaceholder(func(index int) string {
		return fmt.Sprintf("$%d", index+1)
	})

	pgSQL, err := pgTemplate.BuildSQL(context.Background(), map[string]any{
		"minAmount": float64(100),
		"maxAmount": float64(1000),
		"location":  "New York",
		"op":        "=",
	})
	require.NoError(t, err)
	require.NotNil(t, pgSQL)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM events WHERE event_type = $1 AND amount BETWEEN $2 AND $3 AND location = $4"), CompactSQLQuery(pgSQL.Query))
	assert.Equal(t, []any{"PURCHASE", float64(100), float64(1000), "New York"}, pgSQL.Args)

	// Named parameters style (:name)
	namedTemplate := createTestParameterizedTemplate("")
	namedTemplate.WithArgPlaceholder(func(index int) string {
		switch index {
		case 0:
			return ":event_type"
		case 1:
			return ":min_amount"
		case 2:
			return ":max_amount"
		default:
			return ":param" + strconv.Itoa(index)
		}
	})
	namedSQL, err := namedTemplate.BuildSQL(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, namedSQL)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM events WHERE event_type = :event_type AND amount BETWEEN :min_amount AND :max_amount"), CompactSQLQuery(namedSQL.Query))
	assert.Equal(t, []any{"purchase", float64(100), float64(1000)}, namedSQL.Args)

	// Test with array parameters
	arrayTemplate := NewSQLTemplate(
		&Metadata{
			ID: "array_params",
		},
		"SELECT uid FROM users WHERE id IN ({{arg .userIDs}})",
	)

	arraySQL, err := arrayTemplate.BuildSQL(context.Background(), map[string]any{
		"userIDs": []string{"user1", "user2", "user3"},
	})
	require.NoError(t, err)
	require.NotNil(t, arraySQL)
	// 当前实现将整个数组作为单个参数处理
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE id IN (?)"), CompactSQLQuery(arraySQL.Query))
	assert.Equal(t, []any{[]string{"user1", "user2", "user3"}}, arraySQL.Args)

	// Test with argEach function for array expansion
	arrayEachTemplate := NewSQLTemplate(
		&Metadata{
			ID: "array_each_params",
		},
		"SELECT uid FROM users WHERE id IN ({{argEach .userIDs}})",
	)

	arrayEachSQL, err := arrayEachTemplate.BuildSQL(context.Background(), map[string]any{
		"userIDs": []string{"user1", "user2", "user3"},
	})
	require.NoError(t, err)
	require.NotNil(t, arrayEachSQL)
	// argEach expands array into multiple parameters, with one placeholder per array element
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE id IN (?, ?, ?)"), CompactSQLQuery(arrayEachSQL.Query))
	assert.Equal(t, []any{"user1", "user2", "user3"}, arrayEachSQL.Args)

	// Test with argEach for empty array
	emptyArraySQL, err := arrayEachTemplate.BuildSQL(context.Background(), map[string]any{
		"userIDs": []string{},
	})
	require.NoError(t, err)
	require.NotNil(t, emptyArraySQL)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE id IN (NULL)"), CompactSQLQuery(emptyArraySQL.Query))
	assert.Empty(t, emptyArraySQL.Args)

	// Test with different bracket style
	bracketTemplate := NewSQLTemplate(
		&Metadata{
			ID: "bracket_style_params",
		},
		"SELECT uid FROM users WHERE id IN [{{argEach .userIDs}}]",
	)

	bracketSQL, err := bracketTemplate.BuildSQL(context.Background(), map[string]any{
		"userIDs": []string{"user1", "user2", "user3"},
	})
	require.NoError(t, err)
	require.NotNil(t, bracketSQL)
	assert.Equal(t, CompactSQLQuery("SELECT uid FROM users WHERE id IN [?, ?, ?]"), CompactSQLQuery(bracketSQL.Query))
	assert.Equal(t, []any{"user1", "user2", "user3"}, bracketSQL.Args)
}
