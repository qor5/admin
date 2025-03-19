package bq

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/marketing/tag"
)

// SQLDialect implements the SQLDialect interface for SQL
var _ tag.SQLDialect = &SQLDialect{}

type SQLDialect struct{}

// NewSQLDialect creates a new SQLDialect
func NewSQLDialect() *SQLDialect {
	return &SQLDialect{}
}

// Intersect uses INTERSECT DISTINCT to find the intersection of the result sets
func (d *SQLDialect) Intersect(queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	if len(queries) == 1 {
		return queries[0]
	}

	return strings.Join(queries, "\nINTERSECT DISTINCT\n")
}

// Union uses UNION DISTINCT to combine result sets
func (d *SQLDialect) Union(queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	if len(queries) == 1 {
		return queries[0]
	}

	return strings.Join(queries, "\nUNION DISTINCT\n")
}

// Except implements set difference using EXCEPT DISTINCT
func (d *SQLDialect) Except(queries []string) string {
	if len(queries) < 2 {
		return "" // Should not happen as Validate already checks this
	}

	// Join all queries with EXCEPT DISTINCT
	return strings.Join(queries, "\nEXCEPT DISTINCT\n")
}

// Parentheses wraps an expression in parentheses for BigQuery
func (d *SQLDialect) Parentheses(query string) string {
	return fmt.Sprintf("(%s)", query)
}
