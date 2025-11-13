package presets

import (
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/x/v3/hook"
	"github.com/qor5/x/v3/jsonx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"
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

type unmarshalOptions struct {
	hook FilterUnmarshalHook
}

// keep API compatibility: variadic options type (unused)
type UnmarshalOption func(*unmarshalOptions)

// FilterUnmarshalInput is passed to path resolvers to determine the target field key at current scope.
type FilterUnmarshalInput struct {
	FilterMap map[string]any
	Field     string
}
type FilterUnmarshalOutput struct{}

type FilterUnmarshalFunc func(in *FilterUnmarshalInput) (*FilterUnmarshalOutput, error)

// FilterUnmarshalHook composes mutators relay-style: h(next)(in).
type FilterUnmarshalHook = hook.Hook[FilterUnmarshalFunc]

// WithFilterUnmarshalHook registers a relay-style middleware hook for field path resolving.
// Hooks are composed in registration order: hN(...h2(h1(default))).
func WithFilterUnmarshalHook(h FilterUnmarshalHook) UnmarshalOption {
	return func(o *unmarshalOptions) {
		if h == nil {
			return
		}
		o.hook = hook.Prepend(o.hook, h)
	}
}

// Helper: detect group operators from query values
func detectGroupOps(values url.Values) map[string]string {
	groupOp := map[string]string{}
	for k, arr := range values {
		key := k
		if strings.HasPrefix(k, "f_") {
			key = strings.TrimPrefix(k, "f_")
		}
		if strings.EqualFold(key, "keyword") {
			continue
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

// Helper: unescape all values within url.Values using url.QueryUnescape
func unescapeAllValues(values url.Values) url.Values {
	if values == nil {
		return nil
	}
	out := url.Values{}
	for k, arr := range values {
		for _, v := range arr {
			if v == "" {
				out.Add(k, v)
				continue
			}
			if uv, err := url.QueryUnescape(v); err == nil {
				out.Add(k, uv)
			} else {
				out.Add(k, v)
			}
		}
	}
	return out
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
		if strings.HasPrefix(k, "f_") {
			key = strings.TrimPrefix(k, "f_")
		}
		if strings.EqualFold(key, "keyword") {
			continue
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
		// Normalize field:
		// - Convert to CamelCase
		// - Treat "*_range" suffix as an alias to the base field, e.g., "created_at_range" -> "CreatedAt"
		rawField := segs[idx]
		lowerRaw := strings.ToLower(rawField)
		if strings.HasSuffix(lowerRaw, "_range") {
			rawField = strings.TrimSuffix(rawField, "_range")
		}
		field := lo.CamelCase(rawField)
		mod := ""
		if idx+1 < len(segs) {
			mod = strings.ToLower(segs[idx+1])
		}
		g := getGroup(gid)
		// support fold modifier as preference flag for the field
		if strings.EqualFold(mod, "fold") {
			if len(arr) > 0 {
				v := strings.TrimSpace(arr[len(arr)-1])
				val := true
				if v != "" {
					lv := strings.ToLower(v)
					val = lv == "true" || v == "1"
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
	buildChild := func(field string, mod string, sv string, isNot bool) *Filter {
		op := mapModToOperator(mod)
		fcFold := strings.EqualFold(mod, modIlike)
		node := &Filter{Condition: &FieldCondition{Field: field, Operator: op, Value: sv, Fold: fcFold}}
		if isNot {
			node = &Filter{Not: node}
		}
		return node
	}
	for _, g := range groups {
		groupNode := &Filter{}
		usedFields := map[string]bool{}
		for _, it := range g.items {
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
						groupNode.Or = append(groupNode.Or, buildChild(it.field, it.mod, sv, it.isNot))
					}
				} else {
					orGroup := &Filter{}
					for _, sv := range it.values {
						orGroup.Or = append(orGroup.Or, buildChild(it.field, it.mod, sv, it.isNot))
					}
					groupNode.And = append(groupNode.And, orGroup)
				}
				continue
			}
			sv := ""
			if len(it.values) > 0 {
				sv = it.values[0]
			}
			node := buildChild(it.field, it.mod, sv, it.isNot)
			if g.op == groupOpOr {
				groupNode.Or = append(groupNode.Or, node)
			} else {
				groupNode.And = append(groupNode.And, node)
			}
		}
		// Do not emit standalone fold preferences as operator nodes.
		// Fold preferences are consumed elsewhere (e.g., keyword injection) via foldMap.
		if len(groupNode.And) == 0 && len(groupNode.Or) == 0 && groupNode.Not == nil && (groupNode.Condition == nil || groupNode.Condition.Field == "") {
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
)

// BuildFiltersFromQuery transforms a filter query string to a slice of Filters.
// It supports:
// - Groups: <group>.<field>.<mod>=value and <group>.__op in {and,or}
// - NOT: not.<field>.<mod>=value
// - FOLD: <field>.fold=true|1
// - IN/NOT IN: value is CSV
func BuildFiltersFromQuery(qs string) *Filter {
	if qs == "" {
		return nil
	}
	values, err := url.ParseQuery(qs)
	if err != nil {
		return nil
	}
	// Ensure all values are unescaped
	values = unescapeAllValues(values)
	groupOp := detectGroupOps(values)
	groups := aggregateGroups(values, groupOp)
	filters := buildFiltersFromGroups(groups)
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return &Filter{And: filters}
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
	default:
		return FilterOperatorEq
	}
}

func (p *SearchParams) Unmarshal(dst any, opts ...UnmarshalOption) error {
	if dst == nil {
		return errors.New("dst is nil")
	}

	// Build an augmented filter tree that includes keyword conditions
	p.buildAugmentedFilters()
	// Options
	uo := unmarshalOptions{}
	for _, o := range opts {
		if o != nil {
			o(&uo)
		}
	}
	// Convert filter tree to proto-ready JSON-map using options
	norm, err := filterToJSONMap(p.Filter)
	if err != nil {
		return errors.Wrap(err, "filter to json map")
	}

	// Coerce values based on destination operator field types before unmarshalling
	if err := coerceNormValuesForDst(norm, dst, &uo); err != nil {
		return errors.Wrap(err, "coerce values for destination")
	}

	b, err := jsonx.Marshal(norm)
	if err != nil {
		return errors.Wrap(err, "marshal filter json map")
	}
	if err := jsonx.Unmarshal(b, dst); err != nil {
		return errors.Wrap(err, "unmarshal filter json map")
	}
	return nil
}

// buildAugmentedFilters returns a single Filter tree that includes the original filter
// and a synthesized OR group for keyword conditions if present.
func (p *SearchParams) buildAugmentedFilters() {
	if p.Filter == nil {
		p.Filter = &Filter{
			And: []*Filter{},
			Or:  []*Filter{},
			Not: &Filter{},
		}
	}
	if p.Keyword != "" && len(p.KeywordColumns) > 0 {
		for _, col := range p.KeywordColumns {
			p.Filter.Or = append(p.Filter.Or, &Filter{Condition: &FieldCondition{Field: col, Operator: FilterOperatorContains, Value: p.Keyword, Fold: true}})
		}
	}
}

// ----- JSON map builder (Filter -> map[string]any) -----

func filterToJSONMap(f *Filter) (map[string]any, error) {
	if f == nil {
		return nil, nil
	}
	raw, err := jsonx.ToMap(f)
	if err != nil {
		return nil, errors.Wrap(err, "to map filter")
	}
	// Step 2: transform the raw map (Condition/And/Or/Not) into proto-target map
	return transformFilterRawToProtoMap(raw)
}

func transformFilterRawToProtoMap(raw map[string]any) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	dst := map[string]any{}

	// And: merge children maps into current scope
	if v, ok := raw["and"]; ok {
		if arr, ok := v.([]any); ok {
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					cm, err := transformFilterRawToProtoMap(m)
					if err != nil {
						return nil, err
					}
					if len(cm) > 0 {
						mergeFilterJSONMap(dst, cm)
					}
				}
			}
		}
	}

	// Condition: emit field operator map
	if v, ok := raw["condition"]; ok {
		if cm, ok := v.(map[string]any); ok {
			field, _ := cm["field"].(string)
			if field != "" {
				opStr, _ := cm["operator"].(string)
				opStr = strings.TrimSpace(opStr)
				if opStr == "" {
					return nil, errors.New("empty operator in Condition")
				}

				// lowerCamel field key and operator key
				fieldKey := lo.CamelCase(field)
				sub, _ := dst[fieldKey].(map[string]any)
				if sub == nil {
					sub = map[string]any{}
				}
				opKey := lo.CamelCase(opStr)
				val := cm["value"]
				sub[opKey] = val
				if fb, ok := cm["fold"].(bool); ok && fb {
					sub["fold"] = true
				}
				dst[fieldKey] = sub
			}
		}
	}

	// Or: collect child maps into array under key (or/Or)
	if v, ok := raw["or"]; ok {
		if arr, ok := v.([]any); ok {
			var out []any
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					cm, err := transformFilterRawToProtoMap(m)
					if err != nil {
						return nil, err
					}
					if len(cm) > 0 {
						out = append(out, cm)
					}
				}
			}
			if len(out) > 0 {
				if exist, ok := dst["or"].([]any); ok {
					dst["or"] = append(exist, out...)
				} else {
					dst["or"] = out
				}
			}
		}
	}

	// Not: nested map under key (not/Not)
	if v, ok := raw["not"]; ok {
		if m, ok := v.(map[string]any); ok {
			nm, err := transformFilterRawToProtoMap(m)
			if err != nil {
				return nil, err
			}
			if len(nm) > 0 {
				if exist, ok := dst["not"].(map[string]any); ok {
					mergeFilterJSONMap(exist, nm)
					dst["not"] = exist
				} else {
					dst["not"] = nm
				}
			}
		}
	}
	return dst, nil
}

// operatorJSONKey removed: we rely on lowerCamel of the operator string (default to "eq")

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

// ----- Coercion (norm -> dst) with hooks -----
func coerceNormValuesForDst(norm map[string]any, dst any, o *unmarshalOptions) error {
	if norm == nil || dst == nil {
		return nil
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil
	}
	rt := rv.Type().Elem()
	if err := coerceAgainstType(norm, rt, nil, o); err != nil {
		return err
	}
	// prune after coercion
	pruneEmpty(norm)
	return nil
}

func coerceAgainstType(norm map[string]any, t reflect.Type, path []string, o *unmarshalOptions) error {
	if norm == nil {
		return nil
	}
	t = derefType(t)
	if t.Kind() != reflect.Struct {
		return nil
	}
	// Traverse norm keys, map to destination field type, and recurse or coerce values
	for key := range norm {
		// Pre-field hook to allow key rename before field lookup
		if o != nil && o.hook != nil {
			in := &FilterUnmarshalInput{
				FilterMap: norm,
				Field:     key,
			}
			if fn := o.hook(func(_ *FilterUnmarshalInput) (*FilterUnmarshalOutput, error) { return nil, nil }); fn != nil {
				if _, err := fn(in); err != nil {
					return err
				}
			}
		}
		sf, ok := findFieldBySnakeCase(t, key)
		if !ok {
			continue
		}
		ft := derefType(sf.Type)
		val, ok := norm[key]
		if !ok {
			continue
		}
		switch ft.Kind() {
		case reflect.Slice:
			elem := derefType(ft.Elem())
			if elem.Kind() == reflect.Struct {
				if arr, ok := val.([]any); ok {
					for idx, child := range arr {
						if m, ok2 := child.(map[string]any); ok2 {
							if err := coerceAgainstType(m, elem, append(path, key), o); err != nil {
								return err
							}
							arr[idx] = m
						}
					}
				}
			}
		case reflect.Struct:
			if m, ok := val.(map[string]any); ok {
				// apply to present operator keys (lowerCamel)
				for j := 0; j < ft.NumField(); j++ {
					of := ft.Field(j)
					opKey := lo.CamelCase(of.Name)
					if _, ok2 := m[opKey]; ok2 {
						v := m[opKey]
						m[opKey] = coerceJSONValToType(v, of.Type)
					}
				}
			}
		default:
			return nil
		}
	}
	return nil
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func findFieldBySnakeCase(t reflect.Type, key string) (reflect.StructField, bool) {
	t = derefType(t)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if lo.SnakeCase(sf.Name) == lo.SnakeCase(key) {
			return sf, true
		}
	}
	return reflect.StructField{}, false
}

// buildCoercer removed: operator coercion is inlined in makeFieldCoercer

// pruneUnsupported removed: value conversion only; structural pruning is handled by pruneEmpty()

func pruneEmpty(m map[string]any) {
	if m == nil {
		return
	}
	for k, v := range m {
		switch x := v.(type) {
		case nil:
			delete(m, k)
		case []any:
			// prune children
			var kept []any
			for _, it := range x {
				switch child := it.(type) {
				case map[string]any:
					pruneEmpty(child)
					if len(child) > 0 {
						kept = append(kept, child)
					}
				default:
					if it != nil {
						kept = append(kept, it)
					}
				}
			}
			if len(kept) == 0 {
				delete(m, k)
			} else {
				m[k] = kept
			}
		case map[string]any:
			pruneEmpty(x)
			if len(x) == 0 {
				delete(m, k)
			}
		}
	}
}

func coerceJSONValToType(val any, t reflect.Type) any {
	// Deref pointer
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return cast.ToString(val)
	case reflect.Bool:
		return cast.ToBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cast.ToInt64(val)
	case reflect.Float64, reflect.Float32:
		return cast.ToFloat64(val)
	case reflect.Struct:
		// google.protobuf.Timestamp expects RFC3339 string in protojson
		// Normalize common "YYYY-MM-DD HH:MM:SS" to "YYYY-MM-DDTHH:MM:SSZ" if timezone info is missing.
		if isProtoTimestampType(t) {
			return cast.ToTime(val).Format(time.RFC3339)
		}
		return val
	case reflect.Slice:
		elem := t.Elem()
		// Normalize to []any for JSON marshalling
		switch x := val.(type) {
		case []any:
			out := make([]any, 0, len(x))
			for _, e := range x {
				out = append(out, coerceJSONValToType(e, elem))
			}
			return out
		case string: // Best-effort CSV split if destination expects a slice
			parts := strings.Split(x, ",")
			out := make([]any, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				out = append(out, coerceJSONValToType(p, elem))
			}
			return out
		default:
			return val
		}
	default:
		return val
	}
}

func isProtoTimestampType(t reflect.Type) bool {
	t = derefType(t)
	return t == reflect.TypeOf(timestamppb.Timestamp{})
}
