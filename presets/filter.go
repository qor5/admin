package presets

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/x/v3/jsonx"
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

const (
	PascalCase = "PascalCase"
	LowerCase  = "LowerCase"
)

// ----- Unmarshal options and normalization helpers -----

type unmarshalOptions struct {
	casing string // "lowerCamel" (default) or "upperCamel"
}

func defaultUnmarshalOptions() unmarshalOptions {
	return unmarshalOptions{casing: LowerCase}
}

// UnmarshalOption configures Unmarshal behavior.
type UnmarshalOption func(*unmarshalOptions)

// WithPascalCase forces both field and operator keys to UpperCamel (PascalCase).
// Default (without this option) uses lowerCamel for both.
func WithPascalCase() UnmarshalOption {
	return func(o *unmarshalOptions) { o.casing = PascalCase }
}

// applyCasingKey converts a simple identifier into the desired casing.
// PascalCase -> UpperCamel (e.g., "contains" -> "Contains");
// LowerCase  -> lowerCamel (e.g., "contains" -> "contains").
func applyCasingKey(s string, casing string) string {
	if casing == PascalCase {
		return strcase.ToCamel(s)
	}
	return strcase.ToLowerCamel(s)
}

// ternaryCasing returns casing when option exists; otherwise returns default LowerCase
func ternaryCasing(o *unmarshalOptions) string {
	if o == nil || o.casing == "" {
		return LowerCase
	}
	return o.casing
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
		field := strcase.ToCamel(segs[idx])
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

// Unmarshal populates dst with the content of Filters (plus Keyword/KeywordColumns as OR contains conditions).
// dst can be either a pointer to a request wrapper that has a field named "Filter",
// or a pointer to the filter struct itself. This function is generic and not limited to any specific request type.
// Unmarshal performs the same as Unmarshal but allows customization via options.
func (p *SearchParams) Unmarshal(dst any, opts ...UnmarshalOption) error {
	if dst == nil {
		return errors.New("dst is nil")
	}

	// Build an augmented filter tree that includes keyword conditions
	p.buildAugmentedFilters()
	// Options
	uo := defaultUnmarshalOptions()
	for _, o := range opts {
		if o != nil {
			o(&uo)
		}
	}
	// Convert filter tree to proto-ready JSON-map using options
	norm, err := filterToJSONMap(p.Filter, &uo)
	if err != nil {
		return errors.Wrap(err, "filter to json map")
	}

	// Coerce values based on destination operator field types before unmarshalling
	if err := coerceNormValuesForDst(norm, dst); err != nil {
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
		hadContent := filterHasContent(p.Filter)
		or := &Filter{}
		// collect fold preferences from existing filter tree
		prefs := map[string]*bool{}
		if p.Filter != nil {
			collectFoldPrefs(p.Filter, prefs)
		}
		// collect present fields in existing tree to optionally scope keyword columns
		present := map[string]bool{}
		if p.Filter != nil {
			collectPresentFields(p.Filter, present)
		}
		// detect whether any of the keyword columns are already present
		restrictToPresent := false
		for _, col := range p.KeywordColumns {
			if present[strcase.ToCamel(col)] {
				restrictToPresent = true
				break
			}
		}
		for _, col := range p.KeywordColumns {
			if restrictToPresent && !present[strcase.ToCamel(col)] {
				continue
			}
			or.Or = append(or.Or, &Filter{Condition: &FieldCondition{Field: col, Operator: FilterOperatorContains, Value: p.Keyword, Fold: true}})
		}
		if len(or.Or) > 0 {
			if !hadContent {
				// If no existing content, place keyword conditions directly at root OR
				p.Filter.Or = append(p.Filter.Or, or.Or...)
			} else {
				p.Filter.And = append(p.Filter.And, or)
			}
		}
	}
}

func filterHasContent(f *Filter) bool {
	if f == nil {
		return false
	}
	if f.Condition != nil && f.Condition.Field != "" {
		return true
	}
	if len(f.And) > 0 || len(f.Or) > 0 {
		return true
	}
	if f.Not != nil {
		return filterHasContent(f.Not)
	}
	return false
}

// collectFoldPrefs walks the filter tree and records Fold preferences per field.
func collectFoldPrefs(f *Filter, prefs map[string]*bool) {
	if f == nil {
		return
	}
	if f.Condition != nil && f.Condition.Field != "" {
		field := strcase.ToCamel(f.Condition.Field)
		// condition carries Fold flag directly
		if f.Condition.Fold {
			bv := true
			prefs[field] = &bv
		}
		// standalone Fold operator node: Operator == "Fold", Value is bool
		if strings.EqualFold(string(f.Condition.Operator), "Fold") {
			if b, ok := f.Condition.Value.(bool); ok {
				bv := b
				prefs[field] = &bv
			}
		}
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

// collectPresentFields walks the filter tree and records which fields have any condition.
func collectPresentFields(f *Filter, present map[string]bool) {
	if f == nil {
		return
	}
	if f.Condition != nil && f.Condition.Field != "" {
		present[strcase.ToCamel(f.Condition.Field)] = true
	}
	for _, ch := range f.And {
		collectPresentFields(ch, present)
	}
	for _, ch := range f.Or {
		collectPresentFields(ch, present)
	}
	if f.Not != nil {
		collectPresentFields(f.Not, present)
	}
}

// ----- JSON map builder (Filter -> map[string]any) -----

func filterToJSONMap(f *Filter, o *unmarshalOptions) (map[string]any, error) {
	if f == nil {
		return nil, nil
	}
	// Step 1: marshal Filter into a generic JSON map
	b, err := jsonx.Marshal(f)
	if err != nil {
		return nil, errors.Wrap(err, "marshal filter")
	}
	var raw map[string]any
	if err := jsonx.Unmarshal(b, &raw); err != nil {
		return nil, errors.Wrap(err, "unmarshal filter into raw map")
	}
	// Step 2: transform the raw map (Condition/And/Or/Not) into proto-target map
	return transformFilterRawToProtoMap(raw, o)
}

func transformFilterRawToProtoMap(raw map[string]any, o *unmarshalOptions) (map[string]any, error) {
	if raw == nil {
		return nil, nil
	}
	dst := map[string]any{}

	// And: merge children maps into current scope
	if v, ok := raw["And"]; ok {
		if arr, ok := v.([]any); ok {
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					cm, err := transformFilterRawToProtoMap(m, o)
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
	if v, ok := raw["Condition"]; ok {
		if cm, ok := v.(map[string]any); ok {
			field, _ := cm["Field"].(string)
			if field != "" {
				fieldKey := applyCasingKey(field, ternaryCasing(o))
				sub, _ := dst[fieldKey].(map[string]any)
				if sub == nil {
					sub = map[string]any{}
				}
				opStr, _ := cm["Operator"].(string)
				opKey := strings.TrimSpace(opStr)
				if opKey == "" {
					return nil, errors.New("empty operator in Condition")
				}
				opKey = applyCasingKey(opKey, ternaryCasing(o))
				if opKey != "" {
					val := cm["Value"]
					// Defer coercion to destination-type-directed phase
					sub[opKey] = val
					if fb, ok := cm["Fold"].(bool); ok && fb {
						foldKey := applyCasingKey("fold", ternaryCasing(o))
						sub[foldKey] = true
					}
					dst[fieldKey] = sub
				}
			}
		}
	}

	// Or: collect child maps into array under key (or/Or)
	if v, ok := raw["Or"]; ok {
		if arr, ok := v.([]any); ok {
			var out []any
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					cm, err := transformFilterRawToProtoMap(m, o)
					if err != nil {
						return nil, err
					}
					if len(cm) > 0 {
						out = append(out, cm)
					}
				}
			}
			if len(out) > 0 {
				orKey := applyCasingKey("or", ternaryCasing(o))
				if exist, ok := dst[orKey].([]any); ok {
					dst[orKey] = append(exist, out...)
				} else {
					dst[orKey] = out
				}
			}
		}
	}

	// Not: nested map under key (not/Not)
	if v, ok := raw["Not"]; ok {
		if m, ok := v.(map[string]any); ok {
			nm, err := transformFilterRawToProtoMap(m, o)
			if err != nil {
				return nil, err
			}
			if len(nm) > 0 {
				notKey := applyCasingKey("not", ternaryCasing(o))
				if exist, ok := dst[notKey].(map[string]any); ok {
					mergeFilterJSONMap(exist, nm)
					dst[notKey] = exist
				} else {
					dst[notKey] = nm
				}
			}
		}
	}
	return dst, nil
}

// operatorJSONKey removed: we rely on lowerCamel of the operator string (default to "eq")

func mergeFilterJSONMap(dst, src map[string]any) {
	for k, v := range src {
		if k == "or" || k == "Or" {
			// append arrays
			var arr []any
			// allow merging irrespective of key casing by normalizing destination key presence
			if exist, ok := dst[k].([]any); ok {
				arr = exist
			} else {
				// check alternate casing
				alt := "or"
				if k == "or" {
					alt = "Or"
				}
				if exist2, ok2 := dst[alt].([]any); ok2 {
					arr = exist2
					k = alt
				}
			}
			if nv, ok := v.([]any); ok {
				arr = append(arr, nv...)
				dst[k] = arr
				continue
			}
		}
		if k == "not" || k == "Not" {
			if exist, ok := dst[k].(map[string]any); ok {
				if nv, ok2 := v.(map[string]any); ok2 {
					mergeFilterJSONMap(exist, nv)
					dst[k] = exist
					continue
				}
			} else {
				// check alternate casing
				alt := "not"
				if k == "not" {
					alt = "Not"
				}
				if exist2, ok2 := dst[alt].(map[string]any); ok2 {
					if nv, ok3 := v.(map[string]any); ok3 {
						mergeFilterJSONMap(exist2, nv)
						dst[alt] = exist2
						continue
					}
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

func coerceNormValuesForDst(norm map[string]any, dst any) error {
	if norm == nil || dst == nil {
		return nil
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil
	}
	rt := rv.Type().Elem()
	return coerceAgainstType(norm, rt)
}

func coerceAgainstType(norm map[string]any, t reflect.Type) error {
	if norm == nil {
		return nil
	}
	// Deref pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	// If this is a wrapper that contains a Filter field, coerce against that field type
	if f, ok := t.FieldByName("Filter"); ok {
		ft := f.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			if err := coerceAgainstType(norm, ft); err != nil {
				return err
			}
		}
	}

	// Handle Or children (support both "Or" and "or" keys)
	orKey := "Or"
	raw, ok := norm[orKey]
	if !ok {
		if v, ok2 := norm["or"]; ok2 {
			raw = v
			ok = true
			orKey = "or"
		}
	}
	if ok {
		if f, ok2 := t.FieldByName("Or"); ok2 {
			ft := f.Type
			if ft.Kind() == reflect.Slice {
				elem := ft.Elem()
				for elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					if arr, ok3 := raw.([]any); ok3 {
						for _, child := range arr {
							if m, ok4 := child.(map[string]any); ok4 {
								if err := coerceAgainstType(m, elem); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}

	// Handle Not child (support both "Not" and "not" keys)
	notKey := "Not"
	raw, ok = norm[notKey]
	if !ok {
		if v, ok2 := norm["not"]; ok2 {
			raw = v
			ok = true
			notKey = "not"
		}
	}
	if ok {
		if f, ok2 := t.FieldByName("Not"); ok2 {
			ft := f.Type
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				if m, ok3 := raw.(map[string]any); ok3 {
					if err := coerceAgainstType(m, ft); err != nil {
						return err
					}
				}
			}
		}
	}

	// Coerce each field operator map according to destination operator types
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		// Skip control fields handled above
		if sf.Name == "Or" || sf.Name == "And" || sf.Name == "Not" || sf.Name == "Filter" {
			continue
		}

		// Support both destination struct field name (UpperCamel) and lowerCamel JSON key
		fieldKey := sf.Name
		raw, ok := norm[fieldKey]
		usedKey := fieldKey
		if !ok {
			lc := strcase.ToLowerCamel(fieldKey)
			if v, ok2 := norm[lc]; ok2 {
				raw = v
				ok = true
				usedKey = lc
			}
		}
		if !ok {
			continue
		}
		opsMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		opsT := sf.Type
		for opsT.Kind() == reflect.Ptr {
			opsT = opsT.Elem()
		}
		if opsT.Kind() != reflect.Struct {
			continue
		}

		// For each operator field in the destination ops struct, support both UpperCamel and lowerCamel keys
		for j := 0; j < opsT.NumField(); j++ {
			of := opsT.Field(j)
			opKeyUpper := of.Name
			usedOpKey := opKeyUpper
			v, ok2 := opsMap[opKeyUpper]
			if !ok2 {
				opKeyLower := strcase.ToLowerCamel(opKeyUpper)
				if vv, ok3 := opsMap[opKeyLower]; ok3 {
					v = vv
					ok2 = true
					usedOpKey = opKeyLower
				}
			}
			if ok2 {
				opsMap[usedOpKey] = coerceJSONValToType(v, of.Type)
			}
		}

		// write back in case the map was replaced (preserve original key casing)
		norm[usedKey] = opsMap
	}
	return nil
}

func coerceJSONValToType(val any, t reflect.Type) any {
	// Deref pointer
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Special-case protobuf Timestamp: convert to {seconds, nanos}
	if t == reflect.TypeOf(timestamppb.Timestamp{}) {
		switch x := val.(type) {
		case string:
			if tm, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(x)); err == nil {
				sec := tm.Unix()
				ns := tm.Nanosecond()
				return map[string]any{
					"seconds": sec,
					"nanos":   ns,
				}
			}
			return val
		case map[string]any:
			return x
		default:
			return val
		}
	}

	switch t.Kind() {
	case reflect.String:
		switch x := val.(type) {
		case string:
			return x
		case bool:
			if x {
				return "true"
			}
			return "false"
		case float64:
			return strconv.FormatFloat(x, 'f', -1, 64)
		default:
			return val
		}
	case reflect.Bool:
		switch x := val.(type) {
		case bool:
			return x
		case string:
			ls := strings.ToLower(strings.TrimSpace(x))
			return ls == "true" || ls == "1"
		case float64:
			return x != 0
		default:
			return val
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch x := val.(type) {
		case float64:
			// JSON numbers are fine for integer targets
			return x
		case string:
			if fv, err := strconv.ParseFloat(strings.TrimSpace(x), 64); err == nil {
				return fv
			}
			return val
		default:
			return val
		}
	case reflect.Float64:
		switch x := val.(type) {
		case float64:
			return x
		case string:
			if fv, err := strconv.ParseFloat(strings.TrimSpace(x), 64); err == nil {
				return fv
			}
			return val
		default:
			return val
		}
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
		case []string:
			out := make([]any, 0, len(x))
			for _, s := range x {
				out = append(out, coerceJSONValToType(s, elem))
			}
			return out
		case string:
			// Best-effort CSV split if destination expects a slice
			parts := strings.Split(x, ",")
			out := make([]any, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
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

// parseNumberOrBool removed with coerceValueForJSON
