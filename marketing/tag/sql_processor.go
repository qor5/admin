package tag

import (
	"context"

	"github.com/pkg/errors"
)

// SetOperation represents set operations in SQL expressions
type SetOperation int

const (
	// SetOperationIntersect represents the INTERSECT operation
	SetOperationIntersect SetOperation = iota
	// SetOperationUnion represents the UNION operation
	SetOperationUnion
	// SetOperationExcept represents the EXCEPT operation
	SetOperationExcept
)

// String returns the string representation of the set operation
func (op SetOperation) String() string {
	switch op {
	case SetOperationIntersect:
		return "INTERSECT"
	case SetOperationUnion:
		return "UNION"
	case SetOperationExcept:
		return "EXCEPT"
	default:
		return "UNKNOWN"
	}
}

var _ SQLProcessor = &BaseSQLProcessor{}

// BaseSQLProcessor processes expressions into SQL using dialect-specific syntax
type BaseSQLProcessor struct {
	registry *Registry
	dialect  SQLDialect
}

// NewSQLProcessor creates a new SQLProcessor with the given registry and dialect
func NewSQLProcessor(registry *Registry, dialect SQLDialect) *BaseSQLProcessor {
	if registry == nil {
		panic("registry cannot be nil")
	}
	if dialect == nil {
		panic("dialect cannot be nil")
	}
	return &BaseSQLProcessor{
		registry: registry,
		dialect:  dialect,
	}
}

// Process generates SQL from an expression
func (p *BaseSQLProcessor) Process(ctx context.Context, expr *Expression) (*SQL, error) {
	// Ensure expression format is correct
	if err := expr.Validate(); err != nil {
		return nil, err
	}

	if len(expr.Intersect) > 0 {
		return p.processSetOperation(ctx, expr.Intersect, SetOperationIntersect)
	}

	if len(expr.Union) > 0 {
		return p.processSetOperation(ctx, expr.Union, SetOperationUnion)
	}

	if len(expr.Except) > 0 {
		return p.processSetOperation(ctx, expr.Except, SetOperationExcept)
	}

	if expr.Tag != nil {
		return p.processTag(ctx, expr.Tag)
	}

	return nil, errors.New("empty expression")
}

// processSetOperation processes a group of expressions with a set operation
func (p *BaseSQLProcessor) processSetOperation(ctx context.Context, expressions []*Expression, operation SetOperation) (*SQL, error) {
	if len(expressions) == 0 {
		return nil, errors.New("empty expression group")
	}

	queries := make([]string, 0, len(expressions))
	var allArgs []any

	for _, expr := range expressions {
		sql, err := p.Process(ctx, expr)
		if err != nil {
			return nil, err
		}

		// Wrap each subquery in parentheses for better readability
		wrappedQuery := p.dialect.Parentheses(sql.Query)
		queries = append(queries, wrappedQuery)
		allArgs = append(allArgs, sql.Args...)
	}

	var finalQuery string
	switch operation {
	case SetOperationIntersect:
		finalQuery = p.dialect.Intersect(queries)
	case SetOperationUnion:
		finalQuery = p.dialect.Union(queries)
	case SetOperationExcept:
		finalQuery = p.dialect.Except(queries)
	default:
		return nil, errors.Errorf("unsupported set operation: %s", operation)
	}

	return &SQL{
		Query: finalQuery,
		Args:  allArgs,
	}, nil
}

// processTag processes a single tag
func (p *BaseSQLProcessor) processTag(ctx context.Context, tag *Tag) (*SQL, error) {
	if tag == nil {
		return nil, errors.New("tag cannot be nil")
	}

	builder, found := p.registry.GetBuilder(tag.BuilderID)
	if !found {
		return nil, errors.Errorf("builder not found: %s", tag.BuilderID)
	}

	sqlBuilder, ok := builder.(SQLBuilder)
	if !ok {
		return nil, errors.Errorf("builder %s does not implement SQLBuilder", tag.BuilderID)
	}

	sql, err := sqlBuilder.BuildSQL(ctx, tag.Params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build SQL for tag %s", tag.BuilderID)
	}

	if sql == nil {
		return nil, errors.Errorf("nil SQL returned from builder %s", tag.BuilderID)
	}

	return sql, nil
}
