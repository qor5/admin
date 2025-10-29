package presets

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	fieldOr   = "Or"
	fieldNot  = "Not"
	fieldFold = "Fold"
)

// operatorAliases defines canonical operator field name -> acceptable alias field names on target structs.
// This centralizes all hardcoded alternatives.
var operatorAliases = map[string][]string{
	"Contains":   {"Contains", "Ilike"},
	"StartsWith": {"StartsWith"},
	"EndsWith":   {"EndsWith"},
	"Eq":         {"Eq"},
	"Neq":        {"Neq"},
	"Lt":         {"Lt"},
	"Lte":        {"Lte"},
	"Gt":         {"Gt"},
	"Gte":        {"Gte"},
	"In":         {"In"},
	"NotIn":      {"NotIn"},
	"IsNull":     {"IsNull"},
}

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
		segs := strings.Split(k, ".")
		if len(segs) >= 2 && segs[1] == "__op" {
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
				g.op = "and"
			}
			groups[gid] = g
		}
		return g
	}
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
		field := strcase.ToCamel(segs[idx])
		mod := ""
		if idx+1 < len(segs) {
			mod = strings.ToLower(segs[idx+1])
		}
		g := getGroup(gid)
		if mod == "fold" {
			if len(arr) > 0 {
				v := strings.TrimSpace(arr[len(arr)-1])
				val := strings.EqualFold(v, "true") || v == "1"
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
		node := &Filter{Condition: FieldCondition{Field: field, Operator: op, Value: sv}}
		if foldOn {
			node = &Filter{And: []*Filter{
				node,
				{Condition: FieldCondition{Field: field, Operator: mapModToOperator("fold"), Value: true}},
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
			var foldOn bool
			if bv, ok := g.foldMap[it.field]; ok && bv != nil {
				foldOn = *bv
			} else if strings.EqualFold(it.mod, "ilike") {
				foldOn = true
			}
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

// findOperatorField tries to locate the struct field for the given operator name using the alias registry.
func findOperatorField(target reflect.Value, opName string) (reflect.Value, bool) {
	if !target.IsValid() {
		return reflect.Value{}, false
	}
	// try canonical directly first
	if f, ok := findExportedFieldByName(target, opName); ok {
		return f, true
	}
	// try aliases
	aliases, ok := operatorAliases[opName]
	if !ok {
		// fall back: unknown op, try name itself
		return findExportedFieldByName(target, opName)
	}
	for _, alt := range aliases {
		if f, ok := findExportedFieldByName(target, alt); ok {
			return f, true
		}
	}
	return reflect.Value{}, false
}

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
	groupOp := detectGroupOps(values)
	groups := aggregateGroups(values, groupOp)
	return buildFiltersFromGroups(groups)
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
	if err := handleAndGroup(src, dst); err != nil {
		return err
	}
	if err := handleOrGroup(src, dst); err != nil {
		return err
	}
	if err := handleNotGroup(src, dst); err != nil {
		return err
	}
	return handleFieldCondition(src, dst)
}

func handleAndGroup(src *Filter, dst reflect.Value) error {
	if len(src.And) == 0 {
		return nil
	}
	for _, ch := range src.And {
		if err := applyFilterRecursively(ch, dst); err != nil {
			return err
		}
	}
	return nil
}

func handleOrGroup(src *Filter, dst reflect.Value) error {
	if len(src.Or) == 0 {
		return nil
	}
	return appendChildren(fieldOr, src.Or, dst)
}

func handleNotGroup(src *Filter, dst reflect.Value) error {
	if src.Not == nil {
		return nil
	}
	f, ok := findExportedFieldByName(dst, fieldNot)
	if !ok {
		return nil
	}
	if f.Kind() == reflect.Ptr {
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}
		return applyFilterRecursively(src.Not, f.Elem())
	}
	if f.Kind() == reflect.Struct {
		return applyFilterRecursively(src.Not, f)
	}
	return nil
}

func handleFieldCondition(src *Filter, dst reflect.Value) error {
	if src.Condition.Field == "" {
		return nil
	}
	fieldName := strcase.ToCamel(src.Condition.Field)
	fv, ok := findExportedFieldByName(dst, fieldName)
	if !ok {
		if fv, ok = findExportedFieldByInsensitive(dst, src.Condition.Field); !ok {
			return nil
		}
	}
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
	if strings.EqualFold(opName, fieldFold) {
		return setFoldFlag(fv, src)
	}
	if opName == "" {
		opName = "Eq"
	}
	of, ok := findOperatorField(fv, opName)
	if !ok {
		return nil
	}
	return setValueForField(of, src.Condition.Value, opName)
}

func setFoldFlag(target reflect.Value, src *Filter) error {
	if foldField, ok := findExportedFieldByName(target, fieldFold); ok {
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
		parts := strings.Split(s, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
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
		return reflect.Zero(t), nil
	}

	if val, ok := coerceTimestampValue(t, v); ok {
		return val, nil
	}

	if val, ok := coerceDirectOrConvertible(t, rv); ok {
		return val, nil
	}

	if val, ok := coerceFromString(t, fmt.Sprint(v)); ok {
		return val, nil
	}

	return reflect.Value{}, errors.Errorf("cannot coerce %T to %s", v, t.String())
}

func coerceTimestampValue(t reflect.Type, v any) (reflect.Value, bool) {
	if !(t == _tsValType || t.AssignableTo(_tsPtrType)) {
		return reflect.Value{}, false
	}
	var tm time.Time
	switch x := v.(type) {
	case time.Time:
		tm = x
	case *time.Time:
		if x != nil {
			tm = *x
		}
	case string:
		layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00", "2006-01-02"}
		xs := strings.TrimSpace(x)
		for _, layout := range layouts {
			if t1, e := time.Parse(layout, xs); e == nil {
				tm = t1
				break
			}
		}
		if tm.IsZero() {
			if sec, err := strconv.ParseInt(xs, 10, 64); err == nil {
				tm = time.Unix(sec, 0)
			}
		}
		if tm.IsZero() {
			return reflect.Value{}, true
		}
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
			tm = time.Unix(sec, 0)
		} else if t1, err2 := time.Parse(time.RFC3339, s); err2 == nil {
			tm = t1
		} else {
			return reflect.Value{}, true
		}
	}
	pb := timestamppb.New(tm)
	if t.Kind() == reflect.Ptr {
		return reflect.ValueOf(pb), true
	}
	return reflect.Indirect(reflect.ValueOf(pb)).Convert(t), true
}

func coerceDirectOrConvertible(t reflect.Type, rv reflect.Value) (reflect.Value, bool) {
	if rv.Type().AssignableTo(t) {
		return rv.Convert(t), true
	}
	if rv.Type().ConvertibleTo(t) {
		return rv.Convert(t), true
	}
	return reflect.Value{}, false
}

func coerceFromString(t reflect.Type, s string) (reflect.Value, bool) {
	s = strings.TrimSpace(s)
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(s).Convert(t), true
	case reflect.Bool:
		b := strings.EqualFold(s, "true") || s == "1"
		return reflect.ValueOf(b).Convert(t), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if iv, err := strconv.ParseInt(s, 10, 64); err == nil {
			return reflect.ValueOf(iv).Convert(t), true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if iv, err := strconv.ParseUint(s, 10, 64); err == nil {
			return reflect.ValueOf(iv).Convert(t), true
		}
	case reflect.Float32, reflect.Float64:
		if fv, err := strconv.ParseFloat(s, 64); err == nil {
			return reflect.ValueOf(fv).Convert(t), true
		}
	}
	return reflect.Value{}, false
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
	camel := strcase.ToCamel(raw)
	return findExportedFieldByName(v, camel)
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
			// Keyword is case-insensitive by default: Contains + Fold=true
			or.Or = append(or.Or, &Filter{And: []*Filter{
				{Condition: FieldCondition{Field: col, Operator: FilterOperatorContains, Value: p.Keyword}},
				{Condition: FieldCondition{Field: col, Operator: FilterOperatorFold, Value: true}},
			}})
		}
		r = append(r, or)
	}
	return r
}
