package tag

import (
	"context"
)

// Validator defines the interface for validating form elements
type Validator interface {
	// Validate validates the element with the given parameters
	Validate(ctx context.Context, params map[string]any) error
}

// Validation defines validation rules
type Validation struct {
	Pattern      string `json:"pattern"`
	ErrorMessage string `json:"errorMessage"`
}

// FragmentMetadata contains common metadata for all fragment types
type FragmentMetadata struct {
	Key          string         `json:"key"`
	DefaultValue any            `json:"defaultValue"`
	Validation   *Validation    `json:"validation"` // regular expression for validation
	Required     bool           `json:"required"`
	SkipIf       map[string]any `json:"skipIf"`
	SkipUnless   map[string]any `json:"skipUnless"`
}

// Fragment defines the interface for all fragment types
type Fragment interface {
	// Metadata returns the fragment's metadata
	Metadata() FragmentMetadata
	// Type returns the fragment's type
	Type() FragmentType
}

// View defines the UI rendering information
type View struct {
	Fragments []Fragment `json:"fragments"`
}

// Ensure View implements Validator interface
var _ Validator = &View{}

// Category defines a grouping for tag builders
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CategoryWithBuilders represents a category with all its associated builder metadatas
type CategoryWithBuilders struct {
	*Category
	Builders []*Metadata `json:"builders"`
}

// Metadata contains basic information about a tag builder
type Metadata struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  string `json:"categoryID"`
	View        *View  `json:"view"`
}

// Builder interface
type Builder interface {
	Metadata(ctx context.Context) *Metadata
}

// Tag represents a SQL building operation
type Tag struct {
	BuilderID string         `json:"builderID"`
	Params    map[string]any `json:"params"`
}

// Expression represents a logical expression for filtering users
// An expression can be one of:
// 1. Tag (Tag non-nil, Intersect/Union/Except all empty)
// 2. INTERSECT expression (Intersect non-empty, others empty)
// 3. UNION expression (Union non-empty, others empty)
// 4. EXCEPT expression (Except non-empty with at least 2 elements, others empty)
type Expression struct {
	Intersect []*Expression `json:"intersect,omitempty"`
	Union     []*Expression `json:"union,omitempty"`
	Except    []*Expression `json:"except,omitempty"`
	Tag       *Tag          `json:"tag,omitempty"`
}

// SQL represents a SQL query with its arguments
type SQL struct {
	Query string `json:"query"`
	Args  []any  `json:"args"`
}

// SQLBuilder interface defines methods for building SQL queries from parameters
type SQLBuilder interface {
	Builder
	BuildSQL(ctx context.Context, params map[string]any) (*SQL, error)
}

// SQLDialect defines the dialect-specific SQL generation behavior
type SQLDialect interface {
	// Intersect combines result sets using INTERSECT operation
	Intersect(queries []string) string
	// Union combines result sets using UNION operation
	Union(queries []string) string
	// Except performs set difference operation (A EXCEPT B)
	Except(queries []string) string
	// Parentheses wraps an expression in parentheses if needed
	Parentheses(query string) string
}

// SQLProcessor handles SQL generation from expressions
type SQLProcessor interface {
	// Process generates SQL from an expression
	Process(ctx context.Context, expr *Expression) (*SQL, error)
}
