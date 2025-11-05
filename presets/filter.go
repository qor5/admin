package presets

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/x/v3/jsonx"
)

// centralized tokens to avoid hard-coded strings
const (
	groupOpKey = "__op"
	groupOpAnd = "and"
	groupOpOr  = "or"
	prefixNot  = "not"
	csvSep     = ","
)

// modifiers (case-insensitive in parsing)
const (
	modIlike = "ilike"
	modGte   = "gte"
	modLte   = "lte"
	modGt    = "gt"
	modLt    = "lt"
	modIn    = "in"
	modNotIn = "notin"
	modFold  = "fold"
)

// (operator alias table removed; JSON path no longer needs it)

// Types for query aggregation (extracted to reduce function complexity)
type filterItem struct {
	field  string
	mod    string
	values []string
	isNot  bool
}

type filterGroupAgg struct {
	items []filterItem
	// tri-state: nil=unset, true/false=explicit via field.fold
	foldMap map[string]*bool
	op      string
}

// Helper: detect group operators from query values
func detectGroupOps(values url.Values) map[string]string {
	groupOp := map[string]string{}
	for k, arr := range values {
		key := k
		if strings.HasPrefix(key, "f_") {
			key = key[2:]
		}
		segs := strings.Split(key, ".")
		if len(segs) >= 2 && segs[1] == groupOpKey {
			if len(arr) > 0 {
				groupOp[segs[0]] = strings.ToLower(arr[len(arr)-1])
			}
		}
	}
	return groupOp
}

// Helper: aggregate query into groups
func aggregateGroups(values url.Values, groupOp map[string]string) map[string]*filterGroupAgg {
	groups := map[string]*filterGroupAgg{}
	getGroup := func(gid string) *filterGroupAgg {
		g := groups[gid]
		if g == nil {
			g = &filterGroupAgg{foldMap: map[string]*bool{}, op: strings.ToLower(groupOp[gid])}
			if g.op == "" {
				g.op = groupOpAnd
			}
			groups[gid] = g
		}
		return g
	}
	for k, arr := range values {
		key := k
		if strings.HasPrefix(key, "f_") {
			key = key[2:]
		}
		segs := strings.Split(key, ".")
		if len(segs) == 0 {
			continue
		}
		gid := ""
		idx := 0
		if _, ok := groupOp[segs[0]]; ok {
			gid = segs[0]
			idx = 1
		}
		if idx >= len(segs) {
			continue
		}
		if segs[idx] == groupOpKey {
			continue
		}
		isNot := false
		if segs[idx] == prefixNot {
			isNot = true
			idx++
		}
		if idx >= len(segs) {
			continue
		}
		field := strcase.ToCamel(segs[idx])
		mod := ""
		if idx+1 < len(segs) {
			mod = strings.ToLower(segs[idx+1])
		}
		g := getGroup(gid)
		if mod == modFold {
			if len(arr) > 0 {
				v := strings.TrimSpace(arr[len(arr)-1])
				// treat presence without value as true by default
				val := true
				if v != "" {
					val = strings.EqualFold(v, "true") || v == "1"
				}
				g.foldMap[field] = &val
			}
			continue
		}
		g.items = append(g.items, filterItem{field: field, mod: mod, values: arr, isNot: isNot})
	}
	return groups
}

// Helper: build filters from groups
func buildFiltersFromGroups(groups map[string]*filterGroupAgg) []*Filter {
	var out []*Filter
	buildChild := func(field string, mod string, sv string, foldOn bool, isNot bool) *Filter {
		op := mapModToOperator(mod)
		node := &Filter{Condition: &FieldCondition{Field: field, Operator: op, Value: sv}}
		if foldOn {
			node = &Filter{And: []*Filter{
				node,
				{Condition: &FieldCondition{Field: field, Operator: FilterOperatorFold, Value: true}},
			}}
		}
		if isNot {
			node = &Filter{Not: node}
		}
		return node
	}
	for _, g := range groups {
		groupNode := &Filter{}
		usedFields := map[string]bool{}
		for _, it := range g.items {
			var foldOn bool
			if bv, ok := g.foldMap[it.field]; ok && bv != nil {
				foldOn = *bv
			} else if strings.EqualFold(it.mod, modIlike) {
				foldOn = true
			}
			usedFields[it.field] = true
			op := mapModToOperator(it.mod)
			if (op == FilterOperatorIn || op == FilterOperatorNotIn) && len(it.values) > 0 {
				vals := make([]string, 0)
				for _, raw := range it.values {
					for _, p := range strings.Split(raw, csvSep) {
						p = strings.TrimSpace(p)
						if p != "" {
							vals = append(vals, p)
						}
					}
				}
				var valAny any = vals
				node := &Filter{Condition: &FieldCondition{Field: it.field, Operator: op, Value: valAny}}
				if it.isNot {
					node = &Filter{Not: node}
				}
				if g.op == groupOpOr {
					groupNode.Or = append(groupNode.Or, node)
				} else {
					groupNode.And = append(groupNode.And, node)
				}
				continue
			}
			if len(it.values) > 1 {
				if g.op == groupOpOr {
					for _, sv := range it.values {
						groupNode.Or = append(groupNode.Or, buildChild(it.field, it.mod, sv, foldOn, it.isNot))
					}
				} else {
					orGroup := &Filter{}
					for _, sv := range it.values {
						orGroup.Or = append(orGroup.Or, buildChild(it.field, it.mod, sv, foldOn, it.isNot))
					}
					groupNode.And = append(groupNode.And, orGroup)
				}
				continue
			}
			sv := ""
			if len(it.values) > 0 {
				sv = it.values[0]
			}
			node := buildChild(it.field, it.mod, sv, foldOn, it.isNot)
			if g.op == groupOpOr {
				groupNode.Or = append(groupNode.Or, node)
			} else {
				groupNode.And = append(groupNode.And, node)
			}
		}
		// Emit standalone Fold nodes when provided without any other operator for that field
		for field, pv := range g.foldMap {
			if pv == nil {
				continue
			}
			if usedFields[field] {
				continue
			}
			node := &Filter{Condition: &FieldCondition{Field: field, Operator: FilterOperatorFold, Value: *pv}}
			// place under AND so it applies alongside other group criteria regardless of group op
			groupNode.And = append(groupNode.And, node)
		}
		if len(groupNode.And) == 0 && len(groupNode.Or) == 0 && groupNode.Not == nil && groupNode.Condition.Field == "" {
			continue
		}
		out = append(out, groupNode)
	}
	return out
}

const (
	FilterOperatorEq         FilterOperator = "Eq"
	FilterOperatorNeq        FilterOperator = "Neq"
	FilterOperatorLt         FilterOperator = "Lt"
	FilterOperatorLte        FilterOperator = "Lte"
	FilterOperatorGt         FilterOperator = "Gt"
	FilterOperatorGte        FilterOperator = "Gte"
	FilterOperatorIn         FilterOperator = "In"
	FilterOperatorNotIn      FilterOperator = "NotIn"
	FilterOperatorIsNull     FilterOperator = "IsNull"
	FilterOperatorContains   FilterOperator = "Contains"
	FilterOperatorStartsWith FilterOperator = "StartsWith"
	FilterOperatorEndsWith   FilterOperator = "EndsWith"
	FilterOperatorFold       FilterOperator = "Fold"
)

// BuildFiltersFromQuery transforms a filter query string to a slice of Filters.
// It supports:
// - Groups: <group>.<field>.<mod>=value and <group>.__op in {and,or}
// - NOT: not.<field>.<mod>=value
// - FOLD: <field>.fold=true|1
// - IN/NOT IN: value is CSV
func BuildFiltersFromQuery(qs string) []*Filter {
	if qs == "" {
		return nil
	}
	values, err := url.ParseQuery(qs)
	if err != nil {
		return nil
	}
	groupOp := detectGroupOps(values)
	groups := aggregateGroups(values, groupOp)
	return buildFiltersFromGroups(groups)
}

func mapModToOperator(mod string) FilterOperator {
	switch strings.ToLower(mod) {
	case modIlike:
		return FilterOperatorContains
	case modGte:
		return FilterOperatorGte
	case modLte:
		return FilterOperatorLte
	case modGt:
		return FilterOperatorGt
	case modLt:
		return FilterOperatorLt
	case modIn:
		return FilterOperatorIn
	case modNotIn:
		return FilterOperatorNotIn
	case modFold:
		return FilterOperatorFold
	default:
		return FilterOperatorEq
	}
}

// Unmarshal populates dst with the content of Filters (plus Keyword/KeywordColumns as OR contains conditions).
// dst can be either a pointer to a request wrapper that has a field named "Filter",
// or a pointer to the filter struct itself. This function is generic and not limited to any specific request type.
func (p *SearchParams) Unmarshal(dst any) error {
	if dst == nil {
		return errors.New("dst is nil")
	}

	// Build an augmented filter list that includes keyword conditions
	filters := p.buildAugmentedFilters()
	if len(filters) == 0 {
		// Nothing to set
		return nil
	}

	// Build a JSON-map representation and unmarshal via jsonx (protojson for proto targets)
	root := map[string]any{}
	for _, f := range filters {
		fm := filterToJSONMap(f)
		if len(fm) == 0 {
			continue
		}
		mergeFilterJSONMap(root, fm)
	}

	b, err := jsonx.Marshal(root)
	if err != nil {
		return errors.Wrap(err, "marshal filter json map")
	}
	if err := jsonx.Unmarshal(b, dst); err != nil {
		return errors.Wrap(err, "unmarshal filter json map")
	}
	return nil
}

// buildAugmentedFilters returns Filters plus synthesized keyword OR-contains conditions.
func (p *SearchParams) buildAugmentedFilters() []*Filter {
	var r []*Filter
	if p.Filter != nil {
		r = append(r, p.Filter)
	}
	if p.Keyword != "" && len(p.KeywordColumns) > 0 {
		or := &Filter{}
		// collect fold preferences from existing filter tree
		prefs := map[string]*bool{}
		if p.Filter != nil {
			collectFoldPrefs(p.Filter, prefs)
		}
		for _, col := range p.KeywordColumns {
			// If any field prefs exist, limit keyword columns to only those present in prefs
			if len(prefs) > 0 {
				if _, ok := prefs[strcase.ToCamel(col)]; !ok {
					continue
				}
			}
			// default fold is true; override if preference exists
			fv := true
			if pv, ok := prefs[strcase.ToCamel(col)]; ok && pv != nil {
				fv = *pv
			}
			or.Or = append(or.Or, &Filter{And: []*Filter{
				{Condition: &FieldCondition{Field: col, Operator: FilterOperatorContains, Value: p.Keyword}},
				{Condition: &FieldCondition{Field: col, Operator: FilterOperatorFold, Value: fv}},
			}})
		}
		r = append(r, or)
	}
	return r
}

// collectFoldPrefs walks the filter tree and records Fold preferences per field.
func collectFoldPrefs(f *Filter, prefs map[string]*bool) {
	if f == nil {
		return
	}
	if f.Condition != nil && f.Condition.Field != "" && strings.EqualFold(string(f.Condition.Operator), string(FilterOperatorFold)) {
		// normalize field name to Camel
		field := strcase.ToCamel(f.Condition.Field)
		// value may be bool or missing; default to true
		val := true
		if b, ok := f.Condition.Value.(bool); ok {
			val = b
		}
		bv := val
		prefs[field] = &bv
	}
	for _, ch := range f.And {
		collectFoldPrefs(ch, prefs)
	}
	for _, ch := range f.Or {
		collectFoldPrefs(ch, prefs)
	}
	if f.Not != nil {
		collectFoldPrefs(f.Not, prefs)
	}
}

// ----- JSON map builder (Filter -> map[string]any) -----

func filterToJSONMap(f *Filter) map[string]any {
	if f == nil {
		return nil
	}
	dst := map[string]any{}

	// AND children: merge into current scope
	for _, ch := range f.And {
		cm := filterToJSONMap(ch)
		if len(cm) == 0 {
			continue
		}
		mergeFilterJSONMap(dst, cm)
	}

	// Condition at this node
	if fc := f.Condition; fc != nil && fc.Field != "" {
		fieldKey := strcase.ToLowerCamel(fc.Field)
		sub, _ := dst[fieldKey].(map[string]any)
		if sub == nil {
			sub = map[string]any{}
		}
		opKey := operatorJSONKey(fc.Operator)
		if opKey != "" {
			// value coercion: numbers to numeric, booleans to bool, timestamps to proto JSON {seconds,nanos}
			sub[opKey] = coerceValueForJSON(fieldKey, opKey, fc.Value)
			dst[fieldKey] = sub
		}
	}

	// OR group
	if len(f.Or) > 0 {
		var arr []any
		for _, ch := range f.Or {
			cm := filterToJSONMap(ch)
			if len(cm) == 0 {
				continue
			}
			arr = append(arr, cm)
		}
		if len(arr) > 0 {
			if exist, ok := dst["or"].([]any); ok {
				dst["or"] = append(exist, arr...)
			} else {
				dst["or"] = arr
			}
		}
	}

	// NOT group
	if f.Not != nil {
		nm := filterToJSONMap(f.Not)
		if len(nm) > 0 {
			if exist, ok := dst["not"].(map[string]any); ok {
				mergeFilterJSONMap(exist, nm)
				dst["not"] = exist
			} else {
				dst["not"] = nm
			}
		}
	}

	return dst
}

func operatorJSONKey(op FilterOperator) string {
	switch op {
	case FilterOperatorEq, "":
		return "eq"
	case FilterOperatorNeq:
		return "neq"
	case FilterOperatorLt:
		return "lt"
	case FilterOperatorLte:
		return "lte"
	case FilterOperatorGt:
		return "gt"
	case FilterOperatorGte:
		return "gte"
	case FilterOperatorIn:
		return "in"
	case FilterOperatorNotIn:
		return "notIn"
	case FilterOperatorContains:
		return "contains"
	case FilterOperatorStartsWith:
		return "startsWith"
	case FilterOperatorEndsWith:
		return "endsWith"
	case FilterOperatorFold:
		return "fold"
	case FilterOperatorIsNull:
		return "isNull"
	default:
		return "eq"
	}
}

func mergeFilterJSONMap(dst, src map[string]any) {
	for k, v := range src {
		if k == "or" {
			// append arrays
			var arr []any
			if exist, ok := dst[k].([]any); ok {
				arr = exist
			}
			if nv, ok := v.([]any); ok {
				arr = append(arr, nv...)
				dst[k] = arr
				continue
			}
		}
		if k == "not" {
			if exist, ok := dst[k].(map[string]any); ok {
				if nv, ok2 := v.(map[string]any); ok2 {
					mergeFilterJSONMap(exist, nv)
					dst[k] = exist
					continue
				}
			}
		}
		if dm, ok := dst[k].(map[string]any); ok {
			if sm, ok2 := v.(map[string]any); ok2 {
				mergeFilterJSONMap(dm, sm)
				dst[k] = dm
				continue
			}
		}
		dst[k] = v
	}
}

func coerceValueForJSON(fieldKey string, opKey string, val any) any {
	// String pattern operators should keep raw string values, even if numeric-looking
	if opKey == "contains" || opKey == "startsWith" || opKey == "endsWith" {
		if s, ok := val.(string); ok {
			return s
		}
		return val
	}
	// For slice operators, coerce each element
	if opKey == "in" || opKey == "notIn" {
		switch x := val.(type) {
		case []string:
			out := make([]any, 0, len(x))
			for _, s := range x {
				out = append(out, tryParseNumberBoolOrTime(s))
			}
			return out
		case []any:
			out := make([]any, 0, len(x))
			for _, e := range x {
				if s, ok := e.(string); ok {
					out = append(out, tryParseNumberBoolOrTime(s))
				} else {
					out = append(out, e)
				}
			}
			return out
		default:
			// single value as slice of one
			if s, ok := val.(string); ok {
				// support CSV string values
				parts := strings.Split(s, ",")
				out := make([]any, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					out = append(out, tryParseNumberBoolOrTime(p))
				}
				if len(out) == 0 {
					return []any{}
				}
				return out
			}
			return []any{val}
		}
	}

	// fold and isNull are booleans
	if opKey == "fold" || opKey == "isNull" {
		if b, ok := val.(bool); ok {
			return b
		}
		if s, ok := val.(string); ok {
			ls := strings.TrimSpace(strings.ToLower(s))
			return ls == "true" || ls == "1"
		}
		return val
	}

	// other scalar: try timestamp (by heuristic), then number, then bool, else keep string/raw
	if s, ok := val.(string); ok {
		return tryParseNumberBoolOrTime(s)
	}
	return val
}

func tryParseNumberBoolOrTime(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// try to parse explicit timestamp formats only (no epoch seconds inference)
	if tm, ok := parseTimeFlexible(s); ok {
		secs := tm.Unix()
		nanos := tm.Nanosecond()
		return map[string]any{"seconds": secs, "nanos": nanos}
	}
	// numbers first
	// int
	if iv, err := strconv.ParseInt(s, 10, 64); err == nil {
		return iv
	}
	// float
	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return fv
	}
	// bool (true/false only; do not coerce 1/0 here to avoid ambiguity with numeric fields)
	ls := strings.ToLower(s)
	if ls == "true" || ls == "false" {
		return ls == "true"
	}
	// keep original string
	return s
}

func parseTimeFlexible(s string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}
