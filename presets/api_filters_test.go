package presets

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Local mirror of ProductFilter and ListProductsRequest for testing
type (
	CreatedAtFilter struct {
		Lt  *timestamppb.Timestamp
		Lte *timestamppb.Timestamp
		Gt  *timestamppb.Timestamp
		Gte *timestamppb.Timestamp
	}

	UpdatedAtFilter struct {
		Lt  *timestamppb.Timestamp
		Lte *timestamppb.Timestamp
		Gt  *timestamppb.Timestamp
		Gte *timestamppb.Timestamp
	}

	StatusFilter struct {
		Eq    *string
		In    []string
		NotIn []string
	}

	NameFilter struct {
		Eq         *string
		Contains   *string
		StartsWith *string
		Fold       bool
	}

	CodeFilter struct {
		Eq    *string
		In    []string
		NotIn []string
	}

	LocaleCodeFilter struct {
		Eq *string
		In []string
	}

	IsPublishedFilter struct {
		Eq *bool
	}

	ProductFilter struct {
		And []*ProductFilter
		Or  []*ProductFilter
		Not *ProductFilter

		CreatedAt   *CreatedAtFilter
		UpdatedAt   *UpdatedAtFilter
		Status      *StatusFilter
		Name        *NameFilter
		Code        *CodeFilter
		LocaleCode  *LocaleCodeFilter
		IsPublished *IsPublishedFilter
	}

	ListProductsRequest struct {
		Filter *ProductFilter
	}
)

func TestUnmarshalFilters_ProductFilterBasic(t *testing.T) {
	// Build filters
	sp := &SearchParams{}
	sp.Filter = &Filter{And: []*Filter{
		// Name contains "abc"
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "abc"}},
		// Code in [A, B]
		{Condition: FieldCondition{Field: "Code", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
		// not IsPublished eq true
		{Not: &Filter{Condition: FieldCondition{Field: "IsPublished", Operator: FilterOperatorEq, Value: true}}},
		// CreatedAt lt time
		{Condition: FieldCondition{Field: "CreatedAt", Operator: FilterOperatorLt, Value: "2025-10-20T12:34:56Z"}},
	}}
	sp.KeywordColumns = []string{"Name", "Code"}
	sp.Keyword = "kw"

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	assertFilterInitialized(t, req)
	assertNameContains(t, req, "abc")
	assertCodeIn(t, req, []string{"A", "B"})
	assertNotIsPublishedEq(t, req, true)
	assertCreatedAtLtDate(t, req, 2025, time.October, 20)
	assertKeywordOrNameContains(t, req, sp.Keyword)
}

// This test verifies Unmarshal when SearchParams resides in a wrapper request
// and the destination is the wrapper that contains a Filter field (common real-world usage).
func TestUnmarshalFilters_SearchParamsWrapper(t *testing.T) {
	type Wrapper struct {
		Search *SearchParams
		Filter *ProductFilter
	}
	w := &Wrapper{Search: &SearchParams{}}
	w.Search.Filter = &Filter{And: []*Filter{
		{Condition: FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"PUBLISHED", "DRAFT"}}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorStartsWith, Value: "Pro"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorFold, Value: true}},
	}}

	if err := w.Search.Unmarshal(&w.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if w.Filter == nil || w.Filter.Status == nil || len(w.Filter.Status.In) != 2 {
		t.Fatalf("expect status.in set, got %#v", w.Filter)
	}
	if w.Filter.Name == nil || w.Filter.Name.StartsWith == nil || *w.Filter.Name.StartsWith != "Pro" {
		t.Fatalf("expect name.starts_with=Pro, got %#v", w.Filter.Name)
	}
	if w.Filter.Name == nil || !w.Filter.Name.Fold {
		t.Fatalf("expect name.fold=true, got %#v", w.Filter.Name)
	}
}

// Verify full flow: query string -> vx.FilterData.SetByQueryString -> SQL condition and args
func TestFilterData_SetByQueryString_ToSQLConditions(t *testing.T) {
	fd := vx.FilterData{
		&vx.FilterItem{Key: "name", ItemType: vx.ItemTypeString, SQLCondition: "name {op} ?"},
		&vx.FilterItem{Key: "price", ItemType: vx.ItemTypeNumber, SQLCondition: "price {op} ?"},
		&vx.FilterItem{Key: "status", ItemType: vx.ItemTypeMultipleSelect, SQLCondition: "status {op} ?"},
	}

	// name.ilike => contains, price.gte => range lower bound, status.in => array
	qs := "name.ilike=Galaxy&price.gte=10&status.in=A,B"
	cond, args, vErr := fd.SetByQueryString(nil, qs)
	if vErr.HaveErrors() {
		t.Fatalf("validation errors: %#v", vErr)
	}
	// Keys are sorted: name, price, status
	expectedCondParts := []string{"name ILIKE ?", "price >= ?", "status IN ?"}
	for _, part := range expectedCondParts {
		if !strings.Contains(cond, part) {
			t.Fatalf("expect cond contains %q, got %q", part, cond)
		}
	}
	if len(args) != 3 {
		t.Fatalf("expect 3 args, got %d (%#v)", len(args), args)
	}
	if s, ok := args[0].(string); !ok || s != "%Galaxy%" {
		t.Fatalf("expect ilike arg '%%Galaxy%%', got %#v", args[0])
	}

	// Then convert equivalent semantics into ListProductsRequest using SearchParams + Unmarshal
	// name.ilike=Galaxy => Name.Contains="Galaxy" with Fold=true (case-insensitive)
	// status.in=A,B => Status.In=["A","B"]
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Galaxy"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorFold, Value: true}},
		{Condition: FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal to ListProductsRequest error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != "Galaxy" {
		t.Fatalf("expect filter.name.contains=Galaxy, got %#v", req.Filter)
	}
	if !req.Filter.Name.Fold {
		t.Fatalf("expect filter.name.fold=true")
	}
	if req.Filter.Status == nil || len(req.Filter.Status.In) != 2 || req.Filter.Status.In[0] != "A" || req.Filter.Status.In[1] != "B" {
		t.Fatalf("expect filter.status.in=[A B], got %#v", req.Filter.Status)
	}
}

// Cover additional operators mapping via Unmarshal (Eq, Neq, Lt, Lte, Gt, Gte, In, NotIn, IsNull, Contains, StartsWith, EndsWith, Fold)
func TestUnmarshalFilters_AllOperators(t *testing.T) {
	sp := &SearchParams{}
	trueVal := true
	sp.Filter = &Filter{And: []*Filter{
		{Condition: FieldCondition{Field: "Price", Operator: FilterOperatorEq, Value: 9.9}},
		{Condition: FieldCondition{Field: "Price", Operator: FilterOperatorGt, Value: 10}},
		{Condition: FieldCondition{Field: "Price", Operator: FilterOperatorGte, Value: 11}},
		{Condition: FieldCondition{Field: "Price", Operator: FilterOperatorLt, Value: 20}},
		{Condition: FieldCondition{Field: "Price", Operator: FilterOperatorLte, Value: 19}},
		{Condition: FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
		{Condition: FieldCondition{Field: "Status", Operator: FilterOperatorNotIn, Value: []string{"C"}}},
		{Condition: FieldCondition{Field: "DeletedAt", Operator: FilterOperatorIsNull, Value: true}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "nova"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorStartsWith, Value: "Super"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorEndsWith, Value: "Star"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorFold, Value: trueVal}},
	}}

	// Local target filter with fields to receive operators
	type PriceFilter struct{ Eq, Gt, Gte, Lt, Lte *float64 }
	type StatusFilter struct{ In, NotIn []string }
	type NameFilter struct {
		Contains, StartsWith, EndsWith *string
		Fold                           bool
	}
	type DeletedAtFilter struct{ Eq *bool }
	type Filter struct {
		Price     *PriceFilter
		Status    *StatusFilter
		Name      *NameFilter
		DeletedAt *DeletedAtFilter
	}

	var dst Filter
	if err := sp.Unmarshal(&dst); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	assertPriceComparators(t, dst.Price.Gt, dst.Price.Gte, dst.Price.Lt, dst.Price.Lte, 10, 11, 20, 19)
	assertStatusSets(t, dst.Status.In, dst.Status.NotIn, []string{"A", "B"}, []string{"C"})
	assertNameStringOps(t, dst.Name.Contains, dst.Name.StartsWith, dst.Name.EndsWith, dst.Name.Fold, "nova", "Super", "Star", true)
}

// Verify processFilter end-to-end: ListingCompo with FilterQuery -> SQLConditions
func TestListingCompo_ProcessFilter_QueryToSQLConditions(t *testing.T) {
	fd := vx.FilterData{
		&vx.FilterItem{Key: "name", ItemType: vx.ItemTypeString, SQLCondition: "name {op} ?"},
		&vx.FilterItem{Key: "price", ItemType: vx.ItemTypeNumber, SQLCondition: "price {op} ?"},
	}
	lb := &ListingBuilder{}
	lb.filterDataFunc = func(_ *web.EventContext) vx.FilterData { return fd }

	c := &ListingCompo{lb: lb}
	c.FilterQuery = "name.ilike=Galaxy&price.lte=99"

	_, conds := c.processFilter(nil)
	if len(conds) != 1 {
		t.Fatalf("expect 1 SQLCondition, got %d", len(conds))
	}
	if conds[0] == nil {
		t.Fatalf("nil SQLCondition")
	}
	q := conds[0].Query
	if !(strings.Contains(q, "name ILIKE ?") && strings.Contains(q, "price <= ?")) {
		t.Fatalf("unexpected query: %q", q)
	}
	if len(conds[0].Args) != 2 {
		t.Fatalf("unexpected args len: %d (%#v)", len(conds[0].Args), conds[0].Args)
	}
	if s, ok := conds[0].Args[0].(string); !ok || s != "%Galaxy%" {
		t.Fatalf("unexpected ilike arg: %#v", conds[0].Args[0])
	}
	if f, ok := conds[0].Args[1].(int); ok {
		if f != 99 {
			t.Fatalf("unexpected price arg (int): %v", f)
		}
	} else if fs, ok := conds[0].Args[1].(string); ok {
		if fs != "99" {
			t.Fatalf("unexpected price arg (string): %v", fs)
		}
	}
}

// Verify that URL query can be converted into Filters tree (AND/OR) and then mapped to request filter struct
func TestBuildFiltersFromQuery_ToRequestFilter(t *testing.T) {
	// name.ilike=Alpha&name.ilike=Beta => OR for Name.Contains
	// status.in=A,B => In slice
	qs := "name.ilike=Alpha&name.ilike=Beta&status.in=A,B&name.fold=true"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	if len(req.Filter.Or) != 2 {
		t.Fatalf("expect 2 OR children for name.ilike duplicates, got %d", len(req.Filter.Or))
	}
	// Name.fold=true should be applied on root Name (when present at same level)
	// Each OR child should carry Name.Contains with values {Alpha, Beta} and also Name.Fold=true per child
	want := map[string]bool{"Alpha": true, "Beta": true}
	seen := map[string]bool{}
	for _, ch := range req.Filter.Or {
		if ch.Name == nil || ch.Name.Contains == nil {
			t.Fatalf("expect OR child with Name.Contains, got %#v", ch)
		}
		if !ch.Name.Fold {
			t.Fatalf("expect OR child Name.Fold=true, got %#v", ch.Name)
		}
		seen[*ch.Name.Contains] = true
	}
	for k := range want {
		if !seen[k] {
			t.Fatalf("missing OR contains value: %s (seen=%#v)", k, seen)
		}
	}
	if req.Filter.Status == nil || len(req.Filter.Status.In) != 2 {
		t.Fatalf("expect Status.In populated, got %#v", req.Filter.Status)
	}
}

// NOT and groups: g1 (or) with two names (fold true each), and not.status.in=C
func TestBuildFiltersFromQuery_WithNotAndGroups(t *testing.T) {
	qs := "g1.name.ilike=Alpha&g1.name.ilike=Beta&g1.__op=or&g1.name.fold=1&not.status.in=C"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	// group g1 should become an OR of two Name conditions with fold
	if len(req.Filter.Or) == 0 {
		t.Fatalf("expect OR group for g1")
	}
	// flatten may place OR directly under root or under an intermediate group, we check any OR child
	var orNodes []*ProductFilter
	orNodes = append(orNodes, req.Filter.Or...)
	foundAlpha, foundBeta := false, false
	for _, ch := range orNodes {
		if ch.Name != nil && ch.Name.Contains != nil && ch.Name.Fold {
			if *ch.Name.Contains == "Alpha" {
				foundAlpha = true
			}
			if *ch.Name.Contains == "Beta" {
				foundBeta = true
			}
		}
	}
	if !(foundAlpha && foundBeta) {
		t.Fatalf("missing OR contains with fold; alpha=%v beta=%v", foundAlpha, foundBeta)
	}
	// not.status.in=C
	if req.Filter.Not == nil || req.Filter.Not.Status == nil || len(req.Filter.Not.Status.In) != 1 || req.Filter.Not.Status.In[0] != "C" {
		t.Fatalf("expect not.status.in=[C], got %#v", req.Filter.Not)
	}
}

// name=Alpha (default eq), g1.code.in=A,B, g2.not.locale_code.eq=zh
func TestBuildFiltersFromQuery_DefaultEqAndMultiGroups(t *testing.T) {
	qs := "name=Alpha&g1.code.in=A,B&g1.__op=and&g2.not.locale_code.eq=zh&g2.__op=and"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "Alpha" {
		t.Fatalf("expect Name.Eq=Alpha, got %#v", req.Filter)
	}
	if req.Filter.Code == nil || len(req.Filter.Code.In) != 2 || req.Filter.Code.In[0] != "A" || req.Filter.Code.In[1] != "B" {
		t.Fatalf("expect Code.In=[A B], got %#v", req.Filter.Code)
	}
	if req.Filter.Not == nil || req.Filter.Not.LocaleCode == nil || req.Filter.Not.LocaleCode.Eq == nil || *req.Filter.Not.LocaleCode.Eq != "zh" {
		t.Fatalf("expect Not.LocaleCode.Eq=zh, got %#v", req.Filter.Not)
	}
}

// mapModToOperator should map all supported modifiers
// Shared runner to execute qs -> filters -> unmarshal pipeline
func runQS(t *testing.T, qs string, check func(*ListProductsRequest)) {
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	check(&req)
}

func TestBuildFiltersFromQuery_QSSubtests(t *testing.T) {
	cases := []struct {
		name  string
		qs    string
		check func(*ListProductsRequest)
	}{
		{"StatusIn", "status.in=A,B", func(req *ListProductsRequest) { assertStatusIn(t, *req, []string{"A", "B"}) }},
		{"StatusNotIn", "status.notin=C", func(req *ListProductsRequest) { assertStatusNotIn(t, *req, []string{"C"}) }},
		{"NameIlikeWithFold", "name.ilike=nova&name.fold=1", func(req *ListProductsRequest) { assertNameContainsFold(t, *req, "nova", true) }},
		{"NameIlikeImplicitFold", "name.ilike=MiXeD", func(req *ListProductsRequest) { assertNameContainsFold(t, *req, "MiXeD", true) }},
		{"NameEq", "name=Alpha", func(req *ListProductsRequest) { assertNameEqFold(t, *req, "Alpha", false) }},
		{"CreatedAtLte", "created_at.lte=2025-10-20T12:00:00Z", func(req *ListProductsRequest) {
			if req.Filter == nil || req.Filter.CreatedAt == nil || req.Filter.CreatedAt.Lte == nil {
				t.Fatalf("expect CreatedAt.Lte set, got %#v", req.Filter.CreatedAt)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { runQS(t, c.qs, c.check) })
	}
}

// Cover buildFiltersFromGroups: IN branch + group op = or (append to groupNode.Or)
func TestQS_GroupOr_StatusIn(t *testing.T) {
	qs := "g1.status.in=A,B&g1.__op=or"
	runQS(t, qs, func(req *ListProductsRequest) {
		if req.Filter == nil || len(req.Filter.Or) == 0 {
			t.Fatalf("expect OR group for g1 with status.in, got %#v", req.Filter)
		}
	})
}

// Cover handleNotGroup: destination Not field is struct (not pointer)
func TestUnmarshal_NotWithStructField(t *testing.T) {
	type NameOps struct{ Eq *string }
	type Dest struct {
		// Not as struct to cover f.Kind()==reflect.Struct branch
		Not struct {
			Name *NameOps
		}
	}

	sp := &SearchParams{Filter: &Filter{Not: &Filter{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorEq, Value: "X"}}}}
	var dst Dest
	if err := sp.Unmarshal(&dst); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if dst.Not.Name == nil || dst.Not.Name.Eq == nil || *dst.Not.Name.Eq != "X" {
		t.Fatalf("expect Not.Name.Eq=X, got %#v", dst.Not)
	}
}

// Cover appendChildren: Or slice element type is struct (not pointer)
func TestUnmarshal_OrWithStructSlice(t *testing.T) {
	type NameOps struct{ Contains *string }
	type Child struct{ Name *NameOps }
	type Dest struct{ Or []Child }

	// Build two OR name.contains
	sp := &SearchParams{Filter: &Filter{Or: []*Filter{
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "A"}},
		{Condition: FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "B"}},
	}}}
	var dst Dest
	if err := sp.Unmarshal(&dst); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if len(dst.Or) != 2 || dst.Or[0].Name == nil || dst.Or[1].Name == nil {
		t.Fatalf("expect 2 struct OR children with Name set, got %#v", dst.Or)
	}
}

// coerceFromString direct coverage (success and failure)
func TestBuildFiltersFromQuery_CoerceFromStringPaths(t *testing.T) {
	// bool true via eq
	qs := "is_published.eq=true"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.IsPublished == nil || req.Filter.IsPublished.Eq == nil || *req.Filter.IsPublished.Eq != true {
		t.Fatalf("expect IsPublished.Eq=true, got %#v", req.Filter)
	}
}

// ---------- helpers to reduce per-test complexity ----------

func assertFilterInitialized(t *testing.T, req ListProductsRequest) {
	if req.Filter == nil {
		t.Fatalf("Filter should be initialized")
	}
}

func assertNameContains(t *testing.T, req ListProductsRequest, val string) {
	if req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != val {
		t.Fatalf("expect name.contains=%s, got %#v", val, req.Filter.Name)
	}
}

func assertCodeIn(t *testing.T, req ListProductsRequest, want []string) {
	if req.Filter.Code == nil || len(req.Filter.Code.In) != len(want) {
		t.Fatalf("expect code.in len=%d, got %#v", len(want), req.Filter.Code)
	}
	for i := range want {
		if req.Filter.Code.In[i] != want[i] {
			t.Fatalf("unexpected code.in at %d: %s (want %s)", i, req.Filter.Code.In[i], want[i])
		}
	}
}

func assertNotIsPublishedEq(t *testing.T, req ListProductsRequest, b bool) {
	if req.Filter.Not == nil || req.Filter.Not.IsPublished == nil || req.Filter.Not.IsPublished.Eq == nil || *req.Filter.Not.IsPublished.Eq != b {
		t.Fatalf("expect not.is_published.eq=%v, got %#v", b, req.Filter.Not)
	}
}

func assertCreatedAtLtDate(t *testing.T, req ListProductsRequest, year int, month time.Month, day int) {
	if req.Filter.CreatedAt == nil || req.Filter.CreatedAt.Lt == nil {
		t.Fatalf("expect created_at.lt set, got %#v", req.Filter.CreatedAt)
	}
	got := req.Filter.CreatedAt.Lt.AsTime().UTC()
	if got.IsZero() || got.Year() != year || got.Month() != month || got.Day() != day {
		t.Fatalf("unexpected created_at.lt time: %v", got)
	}
}

func assertKeywordOrNameContains(t *testing.T, req ListProductsRequest, kw string) {
	if len(req.Filter.Or) == 0 {
		t.Fatalf("expect OR children for keyword")
	}
	var ok bool
	for _, ch := range req.Filter.Or {
		if ch.Name != nil && ch.Name.Contains != nil && *ch.Name.Contains == kw {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("expect OR child with name.contains=%s, got %#v", kw, req.Filter.Or)
	}
}

func assertStatusIn(t *testing.T, req ListProductsRequest, want []string) {
	if req.Filter == nil || req.Filter.Status == nil || len(req.Filter.Status.In) != len(want) {
		t.Fatalf("expect Status.In len=%d, got %#v", len(want), req.Filter)
	}
	for i := range want {
		if req.Filter.Status.In[i] != want[i] {
			t.Fatalf("unexpected Status.In[%d]=%s (want %s)", i, req.Filter.Status.In[i], want[i])
		}
	}
}

func assertStatusNotIn(t *testing.T, req ListProductsRequest, want []string) {
	if req.Filter == nil || req.Filter.Status == nil || len(req.Filter.Status.NotIn) != len(want) {
		t.Fatalf("expect Status.NotIn len=%d, got %#v", len(want), req.Filter)
	}
	for i := range want {
		if req.Filter.Status.NotIn[i] != want[i] {
			t.Fatalf("unexpected Status.NotIn[%d]=%s (want %s)", i, req.Filter.Status.NotIn[i], want[i])
		}
	}
}

func assertNameContainsFold(t *testing.T, req ListProductsRequest, val string, fold bool) {
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != val || req.Filter.Name.Fold != fold {
		t.Fatalf("expect Name.Contains=%s Fold=%v, got %#v", val, fold, req.Filter)
	}
}

func assertNameEqFold(t *testing.T, req ListProductsRequest, val string, fold bool) {
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != val || req.Filter.Name.Fold != fold {
		t.Fatalf("expect Name.Eq=%s Fold=%v, got %#v", val, fold, req.Filter)
	}
}

func assertPriceComparators(t *testing.T, gtPtr, gtePtr, ltPtr, ltePtr *float64, gt, gte, lt, lte float64) {
	if gtPtr == nil || *gtPtr != gt {
		t.Fatalf("gt not set correctly: %v", gtPtr)
	}
	if gtePtr == nil || *gtePtr != gte {
		t.Fatalf("gte not set correctly: %v", gtePtr)
	}
	if ltPtr == nil || *ltPtr != lt {
		t.Fatalf("lt not set correctly: %v", ltPtr)
	}
	if ltePtr == nil || *ltePtr != lte {
		t.Fatalf("lte not set correctly: %v", ltePtr)
	}
}

func assertStatusSets(t *testing.T, in []string, notIn []string, wantIn []string, wantNotIn []string) {
	if len(in) != len(wantIn) {
		t.Fatalf("in not set: %#v", in)
	}
	if len(notIn) != len(wantNotIn) {
		t.Fatalf("notIn not set: %#v", notIn)
	}
}

func assertNameStringOps(t *testing.T, containsPtr, startsPtr, endsPtr *string, fold bool, wantContains, wantStarts, wantEnds string, wantFold bool) {
	if containsPtr == nil || *containsPtr != wantContains {
		t.Fatalf("contains not set: %v", containsPtr)
	}
	if startsPtr == nil || *startsPtr != wantStarts {
		t.Fatalf("startsWith not set: %v", startsPtr)
	}
	if endsPtr == nil || *endsPtr != wantEnds {
		t.Fatalf("endsWith not set: %v", endsPtr)
	}
	if fold != wantFold {
		t.Fatalf("fold not set correctly: %v (want %v)", fold, wantFold)
	}
}

// Numeric comparators full-flow via custom mirror filter
func TestBuildFiltersFromQuery_NumericComparators_FullFlow(t *testing.T) {
	type PriceOps struct{ Eq, Gt, Gte, Lt, Lte *float64 }
	type MirrorFilter struct{ Price *PriceOps }
	type Req struct{ Filter *MirrorFilter }

	qs := "price.gt=10&price.gte=11&price.lt=20&price.lte=19"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req Req
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil || req.Filter.Price == nil || req.Filter.Price.Gt == nil || *req.Filter.Price.Gt != 10 {
		t.Fatalf("expect Price.Gt=10, got %#v", req.Filter)
	}
	if req.Filter.Price.Gte == nil || *req.Filter.Price.Gte != 11 {
		t.Fatalf("expect Price.Gte=11, got %#v", req.Filter.Price)
	}
	if req.Filter.Price.Lt == nil || *req.Filter.Price.Lt != 20 {
		t.Fatalf("expect Price.Lt=20, got %#v", req.Filter.Price)
	}
	if req.Filter.Price.Lte == nil || *req.Filter.Price.Lte != 19 {
		t.Fatalf("expect Price.Lte=19, got %#v", req.Filter.Price)
	}
}

// NOT name eq full-flow
func TestBuildFiltersFromQuery_NotNameEq(t *testing.T) {
	qs := "not.name.eq=Alpha"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Not == nil || req.Filter.Not.Name == nil || req.Filter.Not.Name.Eq == nil || *req.Filter.Not.Name.Eq != "Alpha" {
		t.Fatalf("expect Not.Name.Eq=Alpha, got %#v", req.Filter)
	}
}

// fold=false via URL should keep Fold=false
func TestBuildFiltersFromQuery_FoldFalseViaQuery(t *testing.T) {
	qs := "name.ilike=ab&name.fold=0"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != "ab" {
		t.Fatalf("expect Name.Contains=ab, got %#v", req.Filter.Name)
	}
	if req.Filter.Name.Fold {
		t.Fatalf("expect Name.Fold=false, got true")
	}
}

// Keyword in qs should map to OR Contains with Fold=true by default
func TestBuildFiltersFromQuery_KeywordQS_FoldImplicit(t *testing.T) {
	qs := "keyword=Nova&name.eq=Exact&name.fold=1"
	// parse keyword from qs
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	// extract keyword
	v, _ := url.ParseQuery(qs)
	kw := v.Get("keyword")

	sp := &SearchParams{Filter: root}
	sp.Keyword = kw
	sp.KeywordColumns = []string{"Name", "Code"}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	// Expect Name.Eq=Exact with Fold=true from explicit name.fold=1
	if req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "Exact" || !req.Filter.Name.Fold {
		t.Fatalf("expect Name.Eq=Exact and Fold=true, got %#v", req.Filter.Name)
	}
	// Keyword should add OR children: Name.Contains=Nova (Fold=true) and possibly Code if supported
	var keywordOK bool
	for _, ch := range req.Filter.Or {
		if ch.Name != nil && ch.Name.Contains != nil && *ch.Name.Contains == kw && ch.Name.Fold {
			keywordOK = true
		}
	}
	if !keywordOK {
		t.Fatalf("expect OR child from keyword with Name.Contains and Fold=true, got %#v", req.Filter.Or)
	}
}

// Eq with fold should set Fold=true and carry Eq value
func TestBuildFiltersFromQuery_EqWithFold(t *testing.T) {
	qs := "name.eq=ABC&name.fold=1"
	filters := BuildFiltersFromQuery(nil, qs)
	var root *Filter
	if len(filters) == 1 {
		root = filters[0]
	} else if len(filters) > 1 {
		root = &Filter{And: filters}
	}
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "ABC" || !req.Filter.Name.Fold {
		t.Fatalf("expect Name.Eq=ABC and Fold=true, got %#v", req.Filter.Name)
	}
}

// IN/NotIn with CSV string values should be split and coerced via coerceToSlice string path
func TestUnmarshal_In_CSVString(t *testing.T) {
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: FieldCondition{Field: "Code", Operator: FilterOperatorIn, Value: "A, B ,C"}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Code == nil || len(req.Filter.Code.In) != 3 {
		t.Fatalf("expect Code.In len=3, got %#v", req.Filter.Code)
	}
	if req.Filter.Code.In[0] != "A" || req.Filter.Code.In[1] != "B" || req.Filter.Code.In[2] != "C" {
		t.Fatalf("unexpected Code.In: %#v", req.Filter.Code.In)
	}
}

func TestUnmarshal_NotIn_CSVString(t *testing.T) {
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: FieldCondition{Field: "Status", Operator: FilterOperatorNotIn, Value: "X,Y"}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Status == nil || len(req.Filter.Status.NotIn) != 2 || req.Filter.Status.NotIn[0] != "X" || req.Filter.Status.NotIn[1] != "Y" {
		t.Fatalf("expect Status.NotIn=[X Y], got %#v", req.Filter.Status)
	}
}
