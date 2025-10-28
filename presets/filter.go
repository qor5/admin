package presets

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// timestamp types for robust detection instead of relying on PkgPath/Name string checks
	_tsPtrType = reflect.TypeOf((*timestamppb.Timestamp)(nil))
	_tsValType = reflect.TypeOf(timestamppb.Timestamp{})
)

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
func BuildFiltersFromQuery(_ vuetifyx.FilterData, qs string) []*Filter {
	if qs == "" {
		return nil
	}
	values, err := url.ParseQuery(qs)
	if err != nil {
		return nil
	}

	// detect group ops
	groupOp := map[string]string{}
	for k, arr := range values {
		segs := strings.Split(k, ".")
		if len(segs) >= 2 && segs[1] == "__op" {
			if len(arr) > 0 {
				groupOp[segs[0]] = strings.ToLower(arr[len(arr)-1])
			}
		}
	}

	type item struct {
		field  string
		mod    string
		values []string
		isNot  bool
	}
	type groupAgg struct {
		items   []item
		foldMap map[string]bool
		op      string
	}
	groups := map[string]*groupAgg{}
	getGroup := func(gid string) *groupAgg {
		g := groups[gid]
		if g == nil {
			g = &groupAgg{foldMap: map[string]bool{}, op: strings.ToLower(groupOp[gid])}
			if g.op == "" {
				g.op = "and"
			}
			groups[gid] = g
		}
		return g
	}

	// parse
	for k, arr := range values {
		segs := strings.Split(k, ".")
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
		if segs[idx] == "__op" {
			continue
		}
		isNot := false
		if segs[idx] == "not" {
			isNot = true
			idx++
		}
		if idx >= len(segs) {
			continue
		}
		field := toCamel(segs[idx])
		mod := ""
		if idx+1 < len(segs) {
			mod = strings.ToLower(segs[idx+1])
		}
		g := getGroup(gid)
		if mod == "fold" {
			if len(arr) > 0 {
				v := arr[len(arr)-1]
				g.foldMap[field] = strings.EqualFold(v, "true") || v == "1"
			}
			continue
		}
		g.items = append(g.items, item{field: field, mod: mod, values: arr, isNot: isNot})
	}

	// build
	var out []*Filter
	buildChild := func(field string, mod string, sv string, foldOn bool, isNot bool) *Filter {
		op := mapModToOperator(mod)
		node := &Filter{Condition: FieldCondition{Field: field, Operator: op, Value: sv}}
		if foldOn {
			node = &Filter{And: []*Filter{
				node,
				{Condition: FieldCondition{Field: field, Operator: FilterOperatorFold, Value: true}},
			}}
		}
		if isNot {
			node = &Filter{Not: node}
		}
		return node
	}
	for gid, g := range groups {
		_ = gid
		groupNode := &Filter{}
		for _, it := range g.items {
			foldOn := g.foldMap[it.field]
			if (mapModToOperator(it.mod) == FilterOperatorIn || mapModToOperator(it.mod) == FilterOperatorNotIn) && len(it.values) > 0 {
				sv := it.values[len(it.values)-1]
				var valAny any
				if sv != "" {
					valAny = strings.Split(sv, ",")
				} else {
					valAny = []string{}
				}
				node := &Filter{Condition: FieldCondition{Field: it.field, Operator: mapModToOperator(it.mod), Value: valAny}}
				if it.isNot {
					node = &Filter{Not: node}
				}
				if g.op == "or" {
					groupNode.Or = append(groupNode.Or, node)
				} else {
					groupNode.And = append(groupNode.And, node)
				}
				continue
			}
			if len(it.values) > 1 {
				if g.op == "or" {
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
			if g.op == "or" {
				groupNode.Or = append(groupNode.Or, node)
			} else {
				groupNode.And = append(groupNode.And, node)
			}
		}
		if len(groupNode.And) == 0 && len(groupNode.Or) == 0 && groupNode.Not == nil && groupNode.Condition.Field == "" {
			continue
		}
		out = append(out, groupNode)
	}
	return out
}

func mapModToOperator(mod string) FilterOperator {
	switch strings.ToLower(mod) {
	case "ilike":
		return FilterOperatorContains
	case "gte":
		return FilterOperatorGte
	case "lte":
		return FilterOperatorLte
	case "gt":
		return FilterOperatorGt
	case "lt":
		return FilterOperatorLt
	case "in":
		return FilterOperatorIn
	case "notin":
		return FilterOperatorNotIn
	case "fold":
		return FilterOperatorFold
	default:
		return FilterOperatorEq
	}
}

// applyFilterRecursively maps a Filter tree into the destination filter struct using reflection.
func applyFilterRecursively(src *Filter, dst reflect.Value) error {
	if !dst.IsValid() || dst.Kind() != reflect.Struct {
		return errors.New("dst must be a struct value")
	}

	// Handle logical groups first
	if len(src.And) > 0 {
		// Flatten AND: merge all children into current node
		for _, ch := range src.And {
			if err := applyFilterRecursively(ch, dst); err != nil {
				return err
			}
		}
	}
	if len(src.Or) > 0 {
		if err := appendChildren("Or", src.Or, dst); err != nil {
			return err
		}
	}
	if src.Not != nil {
		// Field name Not
		f, ok := findExportedFieldByName(dst, "Not")
		if ok {
			// Ensure pointer to struct
			if f.Kind() == reflect.Ptr {
				if f.IsNil() {
					f.Set(reflect.New(f.Type().Elem()))
				}
				if err := applyFilterRecursively(src.Not, f.Elem()); err != nil {
					return err
				}
			} else if f.Kind() == reflect.Struct {
				if err := applyFilterRecursively(src.Not, f); err != nil {
					return err
				}
			}
		}
	}

	// Handle field condition
	if src.Condition.Field != "" {
		fieldName := toCamel(src.Condition.Field)
		fv, ok := findExportedFieldByName(dst, fieldName)
		if !ok {
			// Try alternate name without underscores/case-insensitivity
			if fv, ok = findExportedFieldByInsensitive(dst, src.Condition.Field); !ok {
				return nil // Silently ignore unknown fields
			}
		}

		// Ensure we have a struct to write operator fields into
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			fv = fv.Elem()
		}
		if fv.Kind() != reflect.Struct {
			return errors.Errorf("field %s is not a struct for operators", fieldName)
		}

		opName := string(src.Condition.Operator)
		if strings.EqualFold(opName, "Fold") {
			// Special handling: Fold is a boolean flag on the field filter struct
			// If the target struct has a bool field named Fold, set it to true (or bool value if provided)
			if foldField, ok := findExportedFieldByName(fv, "Fold"); ok {
				val := true
				if b, ok2 := src.Condition.Value.(bool); ok2 {
					val = b
				}
				if foldField.Kind() == reflect.Bool {
					foldField.SetBool(val)
				} else if foldField.Kind() == reflect.Ptr && foldField.Type().Elem().Kind() == reflect.Bool {
					if foldField.IsNil() {
						foldField.Set(reflect.New(foldField.Type().Elem()))
					}
					foldField.Elem().SetBool(val)
				}
			}
			return nil
		}

		if opName == "" {
			opName = "Eq"
		}
		of, ok := findExportedFieldByName(fv, opName)
		if !ok {
			// Some schemas may use ilike for contains; try alternative mappings
			alt := operatorAlternatives(opName)
			var found bool
			for _, an := range alt {
				if of, ok = findExportedFieldByName(fv, an); ok {
					found = true
					break
				}
			}
			if !found {
				return nil // Operator not supported by target, ignore
			}
		}

		// Set value
		if err := setValueForField(of, src.Condition.Value, opName); err != nil {
			return err
		}
	}
	return nil
}

// (And-flattening helper removed; AND is always flattened now)

func appendChildren(field string, children []*Filter, dst reflect.Value) error {
	fv, ok := findExportedFieldByName(dst, field)
	if !ok {
		return nil
	}
	// Expect slice of pointers to the same filter struct type
	if fv.Kind() != reflect.Slice {
		return nil
	}
	elemType := fv.Type().Elem()
	for _, child := range children {
		var childVal reflect.Value
		if elemType.Kind() == reflect.Ptr {
			childVal = reflect.New(elemType.Elem())
			if err := applyFilterRecursively(child, childVal.Elem()); err != nil {
				return err
			}
		} else if elemType.Kind() == reflect.Struct {
			childVal = reflect.New(elemType).Elem()
			if err := applyFilterRecursively(child, childVal); err != nil {
				return err
			}
		} else {
			continue
		}
		fv.Set(reflect.Append(fv, childVal))
	}
	return nil
}

func setValueForField(f reflect.Value, val any, opName string) error {
	// If pointer, allocate
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}
		f = f.Elem()
	}

	// For IN/NotIn, ensure slice
	if strings.EqualFold(opName, "In") || strings.EqualFold(opName, "NotIn") {
		if f.Kind() != reflect.Slice {
			return errors.Errorf("operator %s requires slice field", opName)
		}
		elems, err := coerceToSlice(f.Type().Elem(), val)
		if err != nil {
			return err
		}
		f.Set(elems)
		return nil
	}

	// Scalar
	coerced, err := coerceScalar(f.Type(), val)
	if err != nil {
		return err
	}
	f.Set(coerced)
	return nil
}

func coerceToSlice(elemType reflect.Type, v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	out := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)
	push := func(iv any) error {
		sv, err := coerceScalar(elemType, iv)
		if err != nil {
			return err
		}
		out = reflect.Append(out, sv)
		return nil
	}
	switch rv.Kind() {
	case reflect.String:
		s := rv.String()
		// split by comma if present
		parts := splitCSV(s)
		for _, p := range parts {
			if err := push(p); err != nil {
				return reflect.Value{}, err
			}
		}
		return out, nil
	case reflect.Slice, reflect.Array:
		l := rv.Len()
		for i := 0; i < l; i++ {
			if err := push(rv.Index(i).Interface()); err != nil {
				return reflect.Value{}, err
			}
		}
		return out, nil
	default:
		// single value
		if err := push(v); err != nil {
			return reflect.Value{}, err
		}
		return out, nil
	}
}

func coerceScalar(t reflect.Type, v any) (reflect.Value, error) {
	// Handle pointer target
	if t.Kind() == reflect.Ptr {
		elem := t.Elem()
		sv, err := coerceScalar(elem, v)
		if err != nil {
			return reflect.Value{}, err
		}
		pv := reflect.New(elem)
		pv.Elem().Set(sv)
		return pv, nil
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		// zero
		return reflect.Zero(t), nil
	}

	// Special-case protobuf Timestamp using concrete types (avoid fragile package/name string checks)
	if t == _tsValType || t.AssignableTo(_tsPtrType) {
		// Accept time.Time or various string/int inputs and convert via timestamppb.New
		var tm time.Time
		switch x := v.(type) {
		case time.Time:
			tm = x
		case *time.Time:
			if x != nil {
				tm = *x
			}
		case string:
			// try common formats
			layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00", "2006-01-02"}
			var parsed bool
			xs := strings.TrimSpace(x)
			for _, layout := range layouts {
				if t1, e := time.Parse(layout, xs); e == nil {
					tm = t1
					parsed = true
					break
				}
			}
			if !parsed {
				// try seconds since epoch
				if sec, err := parseInt64(xs); err == nil {
					tm = time.Unix(sec, 0)
					parsed = true
				}
			}
			if !parsed {
				return reflect.Value{}, errors.Errorf("cannot parse time: %q", x)
			}
		default:
			if s, ok := toString(v); ok {
				if sec, err := parseInt64(s); err == nil {
					tm = time.Unix(sec, 0)
				} else {
					// fallback: RFC3339
					t1, err2 := time.Parse(time.RFC3339, s)
					if err2 != nil {
						return reflect.Value{}, errors.Wrap(err2, "parse time")
					}
					tm = t1
				}
			} else {
				return reflect.Value{}, errors.Errorf("unsupported time value type: %T", v)
			}
		}
		pb := timestamppb.New(tm)
		if t.Kind() == reflect.Ptr {
			return reflect.ValueOf(pb), nil
		}
		// target is non-pointer timestamppb.Timestamp (value)
		return reflect.Indirect(reflect.ValueOf(pb)).Convert(t), nil
	}

	// Directly assignable
	if rv.Type().AssignableTo(t) {
		return rv.Convert(t), nil
	}
	// Convertible
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t), nil
	}
	// Try string based conversions
	if s, ok := toString(v); ok {
		switch t.Kind() {
		case reflect.String:
			return reflect.ValueOf(s).Convert(t), nil
		case reflect.Bool:
			b := strings.EqualFold(s, "true") || s == "1"
			return reflect.ValueOf(b).Convert(t), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if iv, err := parseInt64(s); err == nil {
				return reflect.ValueOf(iv).Convert(t), nil
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if iv, err := parseUint64(s); err == nil {
				return reflect.ValueOf(iv).Convert(t), nil
			}
		case reflect.Float32, reflect.Float64:
			if fv, err := parseFloat64(s); err == nil {
				return reflect.ValueOf(fv).Convert(t), nil
			}
		}
	}
	return reflect.Value{}, errors.Errorf("cannot coerce %T to %s", v, t.String())
}

var nonWord = regexp.MustCompile(`[^A-Za-z0-9]+`)

func toCamel(s string) string {
	if s == "" {
		return s
	}
	// If contains non-word separators, convert each segment to Title-case
	if nonWord.MatchString(s) {
		s = nonWord.ReplaceAllString(s, "_")
		parts := strings.Split(s, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
		return strings.Join(parts, "")
	}
	// Otherwise, treat as already camelCase/CamelCase: only uppercase the first rune
	return strings.ToUpper(s[:1]) + s[1:]
}

func findExportedFieldByName(v reflect.Value, name string) (reflect.Value, bool) {
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	f := v.FieldByName(name)
	if f.IsValid() {
		return f, true
	}
	return reflect.Value{}, false
}

func findExportedFieldByInsensitive(v reflect.Value, raw string) (reflect.Value, bool) {
	camel := toCamel(raw)
	return findExportedFieldByName(v, camel)
}

func operatorAlternatives(op string) []string {
	switch op {
	case "Contains":
		return []string{"Contains", "Ilike"}
	default:
		return []string{op}
	}
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	segs := strings.Split(s, ",")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func toString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case fmt.Stringer:
		return x.String(), true
	default:
		rv := reflect.ValueOf(v)
		if !rv.IsValid() {
			return "", false
		}
		return fmt.Sprint(v), true
	}
}

func parseInt64(s string) (int64, error)     { return strconvParseInt(s, 10, 64) }
func parseUint64(s string) (uint64, error)   { return strconvParseUint(s, 10, 64) }
func parseFloat64(s string) (float64, error) { return strconvParseFloat(s, 64) }

// Minimal wrappers to avoid importing strconv directly in multiple places
func strconvParseInt(s string, base int, bitSize int) (int64, error) {
	return parseInt(s, base, bitSize)
}
func strconvParseUint(s string, base int, bitSize int) (uint64, error) {
	return parseUint(s, base, bitSize)
}
func strconvParseFloat(s string, bitSize int) (float64, error) { return parseFloat(s, bitSize) }

// Thin helpers backed by strconv
func parseInt(s string, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), base, bitSize)
}
func parseUint(s string, base int, bitSize int) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(s), base, bitSize)
}
func parseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), bitSize)
}

// Unmarshal populates dst with the content of Filters (plus Keyword/KeywordColumns as OR contains conditions).
// dst can be either a pointer to a request wrapper that has a field named "Filter",
// or a pointer to the filter struct itself. This function is generic and not limited to any specific request type.
func (p *SearchParams) Unmarshal(dst any) error {
	if dst == nil {
		return errors.New("dst is nil")
	}
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("dst must be a non-nil pointer")
	}

	// dst is expected to be a pointer to a filter struct or a pointer-to-pointer to a filter struct (e.g., &req.Filter)
	target := v.Elem()
	var filterVal reflect.Value
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		if target.Elem().Kind() == reflect.Struct {
			filterVal = target.Elem()
		}
	} else if target.Kind() == reflect.Struct {
		filterVal = target
	}

	if !filterVal.IsValid() || filterVal.Kind() != reflect.Struct {
		return errors.New("cannot resolve filter target in dst")
	}

	// Build an augmented filter list that includes keyword conditions
	filters := p.buildAugmentedFilters()
	if len(filters) == 0 {
		// Nothing to set
		return nil
	}

	// Apply each top-level filter onto the same root filter struct (logical AND)
	for _, f := range filters {
		if err := applyFilterRecursively(f, filterVal); err != nil {
			return errors.Wrap(err, "apply filter")
		}
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
		for _, col := range p.KeywordColumns {
			or.Or = append(or.Or, &Filter{Condition: FieldCondition{Field: col, Operator: FilterOperatorContains, Value: p.Keyword}})
		}
		r = append(r, or)
	}
	return r
}
