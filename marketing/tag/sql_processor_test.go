package tag

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSQLDialect struct{}

var _ SQLDialect = &mockSQLDialect{}

// Intersect uses INTERSECT DISTINCT to find the intersection of the result sets
func (d *mockSQLDialect) Intersect(queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	if len(queries) == 1 {
		return queries[0]
	}

	return strings.Join(queries, "\nINTERSECT DISTINCT\n")
}

// Union uses UNION DISTINCT to combine result sets
func (d *mockSQLDialect) Union(queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	if len(queries) == 1 {
		return queries[0]
	}

	return strings.Join(queries, "\nUNION DISTINCT\n")
}

// Except implements set difference using EXCEPT DISTINCT
func (d *mockSQLDialect) Except(queries []string) string {
	if len(queries) < 2 {
		return "" // Should not happen as Validate already checks this
	}

	// Join all queries with EXCEPT DISTINCT
	return strings.Join(queries, "\nEXCEPT DISTINCT\n")
}

// Parentheses wraps an expression in parentheses for BigQuery
func (d *mockSQLDialect) Parentheses(query string) string {
	return fmt.Sprintf("(%s)", query)
}

// TestSQLProcessorWithBigQueryDialect verifies that complex nested expressions
func TestSQLProcessorWithBigQueryDialect(t *testing.T) {
	registry := NewRegistry()

	// Register category
	registry.MustRegisterCategory(&Category{
		ID:          "test",
		Name:        "Test",
		Description: "Test category",
	})

	registry.MustRegisterBuilder(NewSQLTemplate(
		&Metadata{
			ID:          "test_condition",
			Name:        "Test Condition",
			Description: "A simple test condition",
			CategoryID:  "test",
		},
		"SELECT id FROM users WHERE column = '{{.value}}'",
	))

	expr := &Expression{
		Intersect: []*Expression{
			{
				Tag: &Tag{
					BuilderID: "test_condition",
					Params: map[string]any{
						"value": "A",
					},
				},
			},
			{
				Union: []*Expression{
					{
						Tag: &Tag{
							BuilderID: "test_condition",
							Params: map[string]any{
								"value": "B",
							},
						},
					},
					{
						Except: []*Expression{
							{
								Tag: &Tag{
									BuilderID: "test_condition",
									Params: map[string]any{
										"value": "C",
									},
								},
							},
							{
								Tag: &Tag{
									BuilderID: "test_condition",
									Params: map[string]any{
										"value": "D",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	processor := NewSQLProcessor(registry, &mockSQLDialect{})
	sql, err := processor.Process(context.Background(), expr)
	require.NoError(t, err)
	require.NotNil(t, sql)

	expectedSQL := "(SELECT id FROM users WHERE column = 'A') INTERSECT DISTINCT ((SELECT id FROM users WHERE column = 'B') UNION DISTINCT ((SELECT id FROM users WHERE column = 'C') EXCEPT DISTINCT (SELECT id FROM users WHERE column = 'D')))"
	assert.Equal(t, CompactSQLQuery(expectedSQL), CompactSQLQuery(sql.Query))
}

func TestSQLProcessorWithMultipleTags(t *testing.T) {
	registry := NewRegistry()

	// Register categories
	registry.MustRegisterCategory(&Category{
		ID:          "demographic",
		Name:        "Demographics",
		Description: "Demographic filters",
	})

	registry.MustRegisterCategory(&Category{
		ID:          "behavior",
		Name:        "Behavior",
		Description: "Behavioral filters",
	})

	registry.MustRegisterBuilder(createTestAgeTemplate("demographic"))
	registry.MustRegisterBuilder(createTestPurchaseTemplate("behavior"))

	expr := &Expression{
		Intersect: []*Expression{
			{
				Tag: &Tag{
					BuilderID: "age",
					Params: map[string]any{
						"min": float64(30),
						"max": float64(60),
					},
				},
			},
			{
				Tag: &Tag{
					BuilderID: "purchase_amount",
					Params: map[string]any{
						"minAmount": float64(100),
						"maxAmount": float64(1000),
						"startDate": "2023-01-01",
						"endDate":   "2023-12-31",
					},
				},
			},
		},
	}

	processor := NewSQLProcessor(registry, &mockSQLDialect{})

	sql, err := processor.Process(context.Background(), expr)
	require.NoError(t, err)
	t.Logf("Generated SQL: %s", sql.Query)

	expectedSQL := "(SELECT uid FROM users WHERE age BETWEEN ? AND ?) INTERSECT DISTINCT (SELECT uid FROM events WHERE event_type = ? AND event_date BETWEEN ? AND ? GROUP BY uid HAVING SUM(amount) BETWEEN ? AND ?)"
	assert.Equal(t, CompactSQLQuery(expectedSQL), CompactSQLQuery(sql.Query), "Generated SQL should match expected SQL exactly")

	expectedArgs := []any{float64(30), float64(60), "purchase", "2023-01-01", "2023-12-31", float64(100), float64(1000)}
	assert.Equal(t, expectedArgs, sql.Args, "Generated SQL arguments should match expected arguments")
}
