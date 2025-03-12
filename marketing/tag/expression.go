package tag

import (
	"github.com/pkg/errors"
)

// Validate ensures the expression is in a valid format
func (e *Expression) Validate() error {
	if e == nil {
		return errors.New("expression cannot be nil")
	}

	fieldCount := 0

	if len(e.Intersect) > 0 {
		fieldCount++
	}

	if len(e.Union) > 0 {
		fieldCount++
	}

	if len(e.Except) > 0 {
		fieldCount++
		// Except must have at least 2 expressions
		if len(e.Except) < 2 {
			return errors.New("except must have at least 2 expressions")
		}
	}

	if e.Tag != nil {
		fieldCount++
	}

	if fieldCount != 1 {
		return errors.New("expression must have exactly one non-empty field")
	}

	// Validate nested expressions
	expressions := [][]*Expression{e.Intersect, e.Union, e.Except}
	for _, exprSlice := range expressions {
		for _, subExpr := range exprSlice {
			if err := subExpr.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Simplify flattens nested expressions of the same type
func (e *Expression) Simplify() *Expression {
	if e == nil {
		return nil
	}

	if e.Tag != nil {
		return e
	}

	if len(e.Intersect) > 0 {
		simplifiedIntersect := make([]*Expression, 0, len(e.Intersect))

		for _, subExpr := range e.Intersect {
			simplified := subExpr.Simplify()
			if simplified == nil {
				continue
			}

			if len(simplified.Intersect) > 0 && len(simplified.Union) == 0 &&
				len(simplified.Except) == 0 && simplified.Tag == nil {
				simplifiedIntersect = append(simplifiedIntersect, simplified.Intersect...)
			} else {
				simplifiedIntersect = append(simplifiedIntersect, simplified)
			}
		}

		if len(simplifiedIntersect) == 1 {
			return simplifiedIntersect[0]
		}

		return &Expression{Intersect: simplifiedIntersect}
	}

	if len(e.Union) > 0 {
		simplifiedUnion := make([]*Expression, 0, len(e.Union))

		for _, subExpr := range e.Union {
			simplified := subExpr.Simplify()
			if simplified == nil {
				continue
			}

			if len(simplified.Union) > 0 && len(simplified.Intersect) == 0 &&
				len(simplified.Except) == 0 && simplified.Tag == nil {
				simplifiedUnion = append(simplifiedUnion, simplified.Union...)
			} else {
				simplifiedUnion = append(simplifiedUnion, simplified)
			}
		}

		if len(simplifiedUnion) == 1 {
			return simplifiedUnion[0]
		}

		return &Expression{Union: simplifiedUnion}
	}

	if len(e.Except) > 0 {
		simplifiedExcept := make([]*Expression, 0, len(e.Except))

		for _, subExpr := range e.Except {
			simplified := subExpr.Simplify()
			if simplified == nil {
				continue
			}
			simplifiedExcept = append(simplifiedExcept, simplified)
		}

		return &Expression{Except: simplifiedExcept}
	}

	return e
}
