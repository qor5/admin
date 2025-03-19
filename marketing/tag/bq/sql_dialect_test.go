package bq_test

import (
	"testing"

	"github.com/qor5/admin/v3/marketing/tag"
	"github.com/qor5/admin/v3/marketing/tag/bq"
	"github.com/stretchr/testify/assert"
)

// TestBigQuerySQLDialect verifies the BigQuery SQL dialect implementation
// to ensure proper query construction across different set operations
func TestBigQuerySQLDialect(t *testing.T) {
	dialect := bq.NewSQLDialect()

	t.Run("Intersect", func(t *testing.T) {
		t.Run("SingleQuery", func(t *testing.T) {
			queries := []string{"(SELECT id FROM users WHERE age > 30)"}
			result := dialect.Intersect(queries)
			expected := "(SELECT id FROM users WHERE age > 30)"
			assert.Equal(t, expected, result)
		})

		t.Run("MultipleQueries", func(t *testing.T) {
			queries := []string{
				"(SELECT id FROM users WHERE age > 30)",
				"(SELECT id FROM purchases WHERE amount > 100)",
				"(SELECT id FROM logins WHERE date > '2023-01-01')",
			}
			result := dialect.Intersect(queries)
			expected := "(SELECT id FROM users WHERE age > 30) INTERSECT DISTINCT (SELECT id FROM purchases WHERE amount > 100) INTERSECT DISTINCT (SELECT id FROM logins WHERE date > '2023-01-01')"
			assert.Equal(t, tag.CompactSQLQuery(expected), tag.CompactSQLQuery(result))
		})

		t.Run("EmptyQueries", func(t *testing.T) {
			queries := []string{}
			result := dialect.Intersect(queries)
			expected := ""
			assert.Equal(t, expected, result)
		})
	})

	t.Run("Union", func(t *testing.T) {
		t.Run("SingleQuery", func(t *testing.T) {
			queries := []string{"(SELECT id FROM users WHERE age > 30)"}
			result := dialect.Union(queries)
			expected := "(SELECT id FROM users WHERE age > 30)"
			assert.Equal(t, expected, result)
		})

		t.Run("MultipleQueries", func(t *testing.T) {
			queries := []string{
				"(SELECT id FROM users WHERE age > 30)",
				"(SELECT id FROM purchases WHERE amount > 100)",
				"(SELECT id FROM logins WHERE date > '2023-01-01')",
			}
			result := dialect.Union(queries)
			expected := "(SELECT id FROM users WHERE age > 30) UNION DISTINCT (SELECT id FROM purchases WHERE amount > 100) UNION DISTINCT (SELECT id FROM logins WHERE date > '2023-01-01')"
			assert.Equal(t, tag.CompactSQLQuery(expected), tag.CompactSQLQuery(result))
		})

		t.Run("EmptyQueries", func(t *testing.T) {
			queries := []string{}
			result := dialect.Union(queries)
			expected := ""
			assert.Equal(t, expected, result)
		})
	})

	t.Run("Except", func(t *testing.T) {
		t.Run("TwoQueries", func(t *testing.T) {
			queries := []string{
				"SELECT id FROM users WHERE age > 30",
				"SELECT id FROM users WHERE status = 'inactive'",
			}
			result := dialect.Except(queries)
			expected := "SELECT id FROM users WHERE age > 30 EXCEPT DISTINCT SELECT id FROM users WHERE status = 'inactive'"
			assert.Equal(t, tag.CompactSQLQuery(expected), tag.CompactSQLQuery(result))
		})

		t.Run("MultipleQueries", func(t *testing.T) {
			queries := []string{
				"SELECT id FROM all_users",
				"SELECT id FROM inactive_users",
				"SELECT id FROM blocked_users",
			}
			result := dialect.Except(queries)
			expected := "SELECT id FROM all_users EXCEPT DISTINCT SELECT id FROM inactive_users EXCEPT DISTINCT SELECT id FROM blocked_users"
			assert.Equal(t, tag.CompactSQLQuery(expected), tag.CompactSQLQuery(result))
		})

		t.Run("InvalidSingleQuery", func(t *testing.T) {
			queries := []string{"SELECT id FROM users"}
			result := dialect.Except(queries)
			expected := ""
			assert.Equal(t, expected, result)
		})
	})

	t.Run("Parentheses", func(t *testing.T) {
		query := "SELECT id FROM users WHERE age > 30"
		result := dialect.Parentheses(query)
		expected := "(SELECT id FROM users WHERE age > 30)"
		assert.Equal(t, expected, result)
	})
}
