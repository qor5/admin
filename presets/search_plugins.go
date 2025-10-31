package presets

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
)

// SearchOp produces a Filter (using FilterOperator constants) based on the current request context
// and current SearchParams. The returned Filter will be AND-merged into params.Filter automatically.
// Implementations SHOULD NOT mutate params directly; they should only construct and return a Filter.
type SearchOp func(ec *web.EventContext, params *SearchParams) (*Filter, error)

// WrapSearchOps wraps a SearchFunc to:
// 1) Auto-build a Filter from ListingCompo.FilterQuery when params.Filter is nil
// 2) Run user-provided ops to obtain additional Filters
// 3) AND-merge all op Filters into params.Filter
// 4) Compile ONLY the op Filters into SQLConditions (to avoid duplicating listing-provided conditions)
//   - Field naming: Filter uses PascalCase (ToCamel)
//   - SQL columns: snake_case (ToSnake)
func WrapSearchOps(ops ...SearchOp) func(in SearchFunc) SearchFunc {
	return func(in SearchFunc) SearchFunc {
		return func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
			// Step 0: Auto build QS -> Filter when not present
			if params.Filter == nil {
				if compo := ListingCompoFromEventContext(ec); compo != nil && compo.FilterQuery != "" {
					if fs := BuildFiltersFromQuery(compo.FilterQuery); len(fs) == 1 && fs[0] != nil {
						// Ensure PascalCase on fields
						normalizeFilterFieldsCamel(fs[0])
						params.Filter = fs[0]
					}
				}
			}

			// Step 1: Collect filters from ops
			var opFilters []*Filter
			for _, op := range ops {
				f, err := op(ec, params)
				if err != nil {
					return nil, errors.Wrap(err, "search op failed")
				}
				if f == nil {
					continue
				}
				normalizeFilterFieldsCamel(f)
				opFilters = append(opFilters, f)
			}

			// Step 2: Merge op filters into params.Filter with AND semantics
			opsMerged := andAll(opFilters...)
			if opsMerged != nil {
				params.Filter = andMerge(params.Filter, opsMerged)
			}

			// Step 3: Compile ONLY op filters into SQLConditions (avoid duplicating listing-generated conditions)
			if opsMerged != nil {
				// collect fold preferences from the full tree (QS + ops) so ops respect existing fold hints
				prefs := map[string]*bool{}
				if params.Filter != nil {
					collectFoldPrefs(params.Filter, prefs)
				}
				if q, args, ok := compileFilterSQL(opsMerged, prefs); ok && q != "" {
					params.SQLConditions = append(params.SQLConditions, &SQLCondition{Query: q, Args: args})
				}
			}

			return in(ec, params)
		}
	}
}

// andMerge merges base AND f.
func andMerge(base *Filter, f *Filter) *Filter {
	if base == nil {
		return f
	}
	if f == nil {
		return base
	}
	return &Filter{And: []*Filter{base, f}}
}

// andAll builds an AND group of all non-nil filters.
func andAll(fs ...*Filter) *Filter {
	var xs []*Filter
	for _, f := range fs {
		if f != nil {
			xs = append(xs, f)
		}
	}
	if len(xs) == 0 {
		return nil
	}
	if len(xs) == 1 {
		return xs[0]
	}
	return &Filter{And: xs}
}

// normalizeFilterFieldsCamel enforces PascalCase on Filter.Condition.Field (recursively).
func normalizeFilterFieldsCamel(f *Filter) {
	if f == nil {
		return
	}
	if f.Condition != nil && f.Condition.Field != "" {
		f.Condition.Field = toFilterFieldName(f.Condition.Field)
	}
	for _, ch := range f.And {
		normalizeFilterFieldsCamel(ch)
	}
	for _, ch := range f.Or {
		normalizeFilterFieldsCamel(ch)
	}
	if f.Not != nil {
		normalizeFilterFieldsCamel(f.Not)
	}
}

// toFilterFieldName normalizes field name for Filter nodes.
// - If contains underscore, convert to PascalCase via ToCamel
// - If starts with lowercase, convert to PascalCase
// - Otherwise, keep as-is (preserve existing acronyms like UserID)
func toFilterFieldName(field string) string {
	if field == "" {
		return field
	}
	if strings.Contains(field, "_") {
		return strcase.ToCamel(field)
	}
	r, _ := utf8.DecodeRuneInString(field)
	if unicode.IsLower(r) {
		return strcase.ToCamel(field)
	}
	return field
}

// compileFilterSQL compiles a Filter tree into a single SQL condition string and args.
// Pattern operators choose ILIKE when fold preference is true (default), else LIKE.
// Column names are generated via strcase.ToSnake.
func compileFilterSQL(f *Filter, foldPrefs map[string]*bool) (string, []any, bool) {
	if f == nil {
		return "", nil, false
	}

	// Build SQL recursively
	var build func(*Filter) (string, []any, bool)
	build = func(n *Filter) (string, []any, bool) {
		if n == nil {
			return "", nil, false
		}

		// Leaf condition
		if n.Condition != nil && n.Condition.Field != "" {
			field := n.Condition.Field
			col := strcase.ToSnake(field)
			op := n.Condition.Operator
			val := n.Condition.Value

			switch op {
			case FilterOperatorEq, "":
				return col + " = ?", []any{val}, true
			case FilterOperatorNeq:
				return col + " <> ?", []any{val}, true
			case FilterOperatorGt:
				return col + " > ?", []any{val}, true
			case FilterOperatorGte:
				return col + " >= ?", []any{val}, true
			case FilterOperatorLt:
				return col + " < ?", []any{val}, true
			case FilterOperatorLte:
				return col + " <= ?", []any{val}, true
			case FilterOperatorIsNull:
				b := false
				if v, ok := val.(bool); ok {
					b = v
				}
				if b {
					return col + " IS NULL", nil, true
				}
				return col + " IS NOT NULL", nil, true
			case FilterOperatorIn:
				// Empty IN should yield no matches
				if isEmptySlice(val) {
					return "1 = 0", nil, true
				}
				return col + " IN ?", []any{val}, true
			case FilterOperatorNotIn:
				// Empty NOT IN => no-op
				if isEmptySlice(val) {
					return "", nil, false
				}
				return col + " NOT IN ?", []any{val}, true
			case FilterOperatorContains:
				like := likeOpFor(field, foldPrefs)
				return col + like + "?", []any{wrapLike(val, "%", "%")}, true
			case FilterOperatorStartsWith:
				like := likeOpFor(field, foldPrefs)
				return col + like + "?", []any{wrapLike(val, "", "%")}, true
			case FilterOperatorEndsWith:
				like := likeOpFor(field, foldPrefs)
				return col + like + "?", []any{wrapLike(val, "%", "")}, true
			case FilterOperatorFold:
				// Fold is a preference hint only
				return "", nil, false
			default:
				return col + " = ?", []any{val}, true
			}
		}

		// AND group
		var andParts []string
		var andArgs []any
		for _, ch := range n.And {
			if q, a, ok := build(ch); ok && q != "" {
				andParts = append(andParts, wrapParen(q))
				andArgs = append(andArgs, a...)
			}
		}
		// OR group
		var orParts []string
		var orArgs []any
		for _, ch := range n.Or {
			if q, a, ok := build(ch); ok && q != "" {
				orParts = append(orParts, wrapParen(q))
				orArgs = append(orArgs, a...)
			}
		}
		// NOT group
		if n.Not != nil {
			if q, a, ok := build(n.Not); ok && q != "" {
				andParts = append(andParts, "NOT "+wrapParen(q))
				andArgs = append(andArgs, a...)
			}
		}

		// Combine precedence: ANDs and ORs at same level compose as:
		// - if both present: (AND...) AND ((OR...)) to preserve semantics
		var parts []string
		var args []any
		if len(andParts) > 0 {
			parts = append(parts, strings.Join(andParts, " AND "))
			args = append(args, andArgs...)
		}
		if len(orParts) > 0 {
			orJoined := strings.Join(orParts, " OR ")
			if len(parts) > 0 {
				parts = append(parts, wrapParen(orJoined))
			} else {
				parts = append(parts, orJoined)
			}
			args = append(args, orArgs...)
		}
		if len(parts) == 0 {
			return "", nil, false
		}
		if len(parts) == 1 {
			return parts[0], args, true
		}
		return strings.Join(parts, " AND "), args, true
	}

	return build(f)
}

func likeOpFor(field string, prefs map[string]*bool) string {
	// default case-insensitive
	useILike := true
	if pv, ok := prefs[strcase.ToCamel(field)]; ok && pv != nil {
		useILike = *pv
	}
	if useILike {
		return " ILIKE "
	}
	return " LIKE "
}

func wrapLike(val any, pre, suf string) string {
	s := fmt.Sprint(val)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return pre + s + suf
}

func wrapParen(q string) string {
	if q == "" {
		return q
	}
	return "(" + q + ")"
}

func isEmptySlice(v any) bool {
	switch x := v.(type) {
	case []string:
		return len(x) == 0
	case []any:
		return len(x) == 0
	default:
		return false
	}
}
