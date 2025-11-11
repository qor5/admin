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
	LocaleStatus string

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
		EndsWith   *string
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

	LocaleStatusFilter struct {
		Eq    *LocaleStatus
		In    []LocaleStatus
		NotIn []LocaleStatus
	}

	IsPublishedFilter struct {
		Eq *bool
	}

	ProductFilter struct {
		And []*ProductFilter
		Or  []*ProductFilter
		Not *ProductFilter

		CreatedAt    *CreatedAtFilter
		UpdatedAt    *UpdatedAtFilter
		Status       *StatusFilter
		Name         *NameFilter
		Code         *CodeFilter
		LocaleCode   *LocaleCodeFilter
		IsPublished  *IsPublishedFilter
		LocaleStatus *LocaleStatusFilter
	}

	ListProductsRequest struct {
		Filter *ProductFilter
	}
)

// Lower camel JSON-tagged request and filters for user entity
type (
	ProductSatus int32
	UserNameOps  struct {
		Eq       *string `json:"eq"`
		Contains *string `json:"contains"`
		Fold     bool    `json:"fold"`
	}
	UserStatusOps struct {
		In []string `json:"in"`
	}
	ProductStatusOps struct {
		Eq    *ProductSatus
		In    []ProductSatus
		NotIn []ProductSatus
	}
	CreatedAtOps struct {
		Lt  *timestamppb.Timestamp `json:"lt"`
		Lte *timestamppb.Timestamp `json:"lte"`
		Gt  *timestamppb.Timestamp `json:"gt"`
		Gte *timestamppb.Timestamp `json:"gte"`
	}
	UpdatedAtOps struct {
		Lt  *timestamppb.Timestamp `json:"lt"`
		Lte *timestamppb.Timestamp `json:"lte"`
		Gt  *timestamppb.Timestamp `json:"gt"`
		Gte *timestamppb.Timestamp `json:"gte"`
	}
	UserFilter struct {
		Or            []*UserFilter     `json:"or"`
		Not           *UserFilter       `json:"not"`
		Name          *UserNameOps      `json:"name"`
		Status        *UserStatusOps    `json:"status"`
		UserName      *UserNameOps      `json:"userName"`
		ProductStatus *ProductStatusOps `json:"productStatus"`
		CreatedAt     *CreatedAtOps     `json:"createdAt"`
		UpdatedAt     *UpdatedAtOps     `json:"updatedAt"`
	}
	ListUserRequest struct {
		Filter *UserFilter `json:"filter"`
	}
)

// More custom numeric-like types to verify Unmarshal adaptability
type (
	CategoryLevel int16
	OrderState    int64
	Rank          uint32
	Score         float64

	CategoryLevelOps struct {
		Eq    *CategoryLevel
		In    []CategoryLevel
		NotIn []CategoryLevel
	}
	OrderStateOps struct {
		Eq    *OrderState
		In    []OrderState
		NotIn []OrderState
	}
	RankOps struct {
		Eq    *Rank
		In    []Rank
		NotIn []Rank
	}
	ScoreOps struct {
		Eq    *Score
		In    []Score
		NotIn []Score
	}
	CNameOps struct {
		Eq       *string `json:"eq"`
		Contains *string `json:"contains"`
		Fold     bool    `json:"fold"`
	}
	UserIDOps struct {
		Contains *string  `json:"contains"`
		Eq       *string  `json:"eq"`
		In       []string `json:"in"`
		NotIn    []string `json:"notIn"`
	}
	CustomFilter struct {
		Or            []*CustomFilter   `json:"or"`
		Not           *CustomFilter     `json:"not"`
		Name          *CNameOps         `json:"name"`
		CategoryLevel *CategoryLevelOps `json:"categoryLevel"`
		OrderState    *OrderStateOps    `json:"orderState"`
		Rank          *RankOps          `json:"rank"`
		Score         *ScoreOps         `json:"score"`
		UserId        *UserIDOps        `json:"userId"`
	}
	ListCustomRequest struct {
		Filter *CustomFilter `json:"filter"`
	}
)

const (
	LocaleStatusPublished LocaleStatus = "PUBLISHED"
	LocaleStatusDraft     LocaleStatus = "DRAFT"

	ProductStatusPublished ProductSatus = 1
	ProductStatusDraft     ProductSatus = 2
	ProductStatusArchived  ProductSatus = 3
	ProductStatusDeleted   ProductSatus = 4
	ProductStatusInactive  ProductSatus = 5
	ProductStatusActive    ProductSatus = 6
	ProductStatusPending   ProductSatus = 7
	ProductStatusCompleted ProductSatus = 8
	ProductStatusCancelled ProductSatus = 9
	// Custom numeric type constants
	LevelBasic  CategoryLevel = 1
	LevelPro    CategoryLevel = 2
	LevelElite  CategoryLevel = 3
	StateNew    OrderState    = 1
	StateClosed OrderState    = 5
	RankGold    Rank          = 7
)

func TestUnmarshalFilters_ProductFilterBasic(t *testing.T) {
	// Build filters
	sp := &SearchParams{}
	sp.Filter = &Filter{And: []*Filter{
		// Name contains "abc"
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "abc"}},
		// Code in [A, B]
		{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
		// not IsPublished eq true
		{Not: &Filter{Condition: &FieldCondition{Field: "IsPublished", Operator: FilterOperatorEq, Value: true}}},
		// CreatedAt lt time
		{Condition: &FieldCondition{Field: "CreatedAt", Operator: FilterOperatorLt, Value: "2025-10-20T12:34:56Z"}},
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
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"PUBLISHED", "DRAFT"}}},
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorStartsWith, Value: "Pro", Fold: true}},
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
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Galaxy", Fold: true}},
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
		{Condition: &FieldCondition{Field: "Price", Operator: FilterOperatorEq, Value: 9.9}},
		{Condition: &FieldCondition{Field: "Price", Operator: FilterOperatorGt, Value: 10}},
		{Condition: &FieldCondition{Field: "Price", Operator: FilterOperatorGte, Value: 11}},
		{Condition: &FieldCondition{Field: "Price", Operator: FilterOperatorLt, Value: 20}},
		{Condition: &FieldCondition{Field: "Price", Operator: FilterOperatorLte, Value: 19}},
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorNotIn, Value: []string{"C"}}},
		{Condition: &FieldCondition{Field: "DeletedAt", Operator: FilterOperatorIsNull, Value: true}},
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "nova", Fold: trueVal}},
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorStartsWith, Value: "Super"}},
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorEndsWith, Value: "Star"}},
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
	if err := sp.Unmarshal(&dst, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	assertPriceComparators(t, dst.Price.Gt, dst.Price.Gte, dst.Price.Lt, dst.Price.Lte, 10, 11, 20, 19)
	assertStatusSets(t, dst.Status.In, dst.Status.NotIn, []string{"A", "B"}, []string{"C"})
	assertNameStringOps(t, dst.Name.Contains, dst.Name.StartsWith, dst.Name.EndsWith, dst.Name.Fold, "nova", "Super", "Star", true)
}

// Direct Filter construction: multi-level nesting and multiple conditions on the same field
func TestUnmarshalFilters_DirectNestedAndDuplicateFields(t *testing.T) {
	// Build nested Filter tree directly
	root := &Filter{And: []*Filter{
		// Top-level Name contains Alpha (will be overridden later by Gamma under AND)
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Alpha"}},
		// OR group with two children
		{Or: []*Filter{
			// Child 1: AND with Name startsWith A and endsWith Z
			{And: []*Filter{
				{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorStartsWith, Value: "A"}},
				{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorEndsWith, Value: "Z"}},
			}},
			// Child 2: NOT Name Eq Beta
			{Not: &Filter{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorEq, Value: "Beta"}}},
		}},
		// AND group with Code in and Name contains Gamma (same field as top-level contains)
		{And: []*Filter{
			{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorIn, Value: []string{"X", "Y"}}},
			{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Gamma"}},
		}},
	}}

	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil {
		t.Fatalf("filter nil")
	}

	// Name: contains should be the later one (Gamma), and also include starts/ends from OR child
	if req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != "Gamma" {
		t.Fatalf("expect Name.Contains=Gamma, got %#v", req.Filter.Name)
	}

	// Ensure previous Alpha is not lingering anywhere
	alphaFound := false
	if req.Filter.Name != nil && req.Filter.Name.Contains != nil && *req.Filter.Name.Contains == "Alpha" {
		alphaFound = true
	}
	for _, ch := range req.Filter.Or {
		if ch != nil && ch.Name != nil && ch.Name.Contains != nil && *ch.Name.Contains == "Alpha" {
			alphaFound = true
		}
	}
	if alphaFound {
		t.Fatalf("unexpected lingering Name.Contains=Alpha in final filter: %#v", req.Filter)
	}

	// Code.In from AND group
	if req.Filter.Code == nil || len(req.Filter.Code.In) != 2 || req.Filter.Code.In[0] != "X" || req.Filter.Code.In[1] != "Y" {
		t.Fatalf("expect Code.In=[X Y], got %#v", req.Filter.Code)
	}

	// OR group validation: one child has Name.StartsWith=A and Name.EndsWith=Z; another child has Not.Name.Eq=Beta
	if len(req.Filter.Or) != 2 {
		t.Fatalf("expect 2 OR children, got %d", len(req.Filter.Or))
	}
	var foundStartEnd, foundNotBeta bool
	for _, ch := range req.Filter.Or {
		if ch == nil {
			continue
		}
		if ch.Name != nil && ch.Name.StartsWith != nil && *ch.Name.StartsWith == "A" && ch.Name.EndsWith != nil && *ch.Name.EndsWith == "Z" {
			foundStartEnd = true
		}
		if ch.Not != nil && ch.Not.Name != nil && ch.Not.Name.Eq != nil && *ch.Not.Name.Eq == "Beta" {
			foundNotBeta = true
		}
	}
	if !foundStartEnd || !foundNotBeta {
		t.Fatalf("expect OR children for starts/ends and not beta, got %#v", req.Filter.Or)
	}
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

	_, conds, _ := c.processFilter(nil)
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
// merged into TestBuildFiltersFromQuery_QSSubtests
func TestBuildFiltersFromQuery_ToRequestFilter(t *testing.T) {
	// name.ilike=Alpha&name.ilike=Beta => OR for Name.Contains
	// status.in=A,B => In slice
	qs := "f_name.ilike=Alpha&f_name.ilike=Beta&f_status.in=A,B"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
// merged into TestBuildFiltersFromQuery_QSSubtests
func TestBuildFiltersFromQuery_WithNotAndGroups(t *testing.T) {
	qs := "f_g1.name.ilike=Alpha&f_g1.name.ilike=Beta&f_g1.__op=or&f_not.status.in=C"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
// merged into TestBuildFiltersFromQuery_QSSubtests
func TestBuildFiltersFromQuery_DefaultEqAndMultiGroups(t *testing.T) {
	qs := "f_name=Alpha&f_g1.code.in=A,B&f_g1.__op=and&f_g2.not.locale_code.eq=zh&f_g2.__op=and"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	check(&req)
}

// Split from TestBuildFiltersFromQuery_QSSubtests
func TestBuildFiltersFromQuery_StatusIn(t *testing.T) {
	runQS(t, "status.in=A,B", func(req *ListProductsRequest) {
		if req.Filter == nil || req.Filter.Status == nil {
			t.Fatalf("expect Status present, got %#v", req.Filter)
		}
		got := req.Filter.Status.In
		if len(got) != 2 || got[0] != "A" || got[1] != "B" {
			t.Fatalf("expect Status.In=[A B], got %#v", got)
		}
	})
}

func TestBuildFiltersFromQuery_StatusNotIn(t *testing.T) {
	runQS(t, "status.notin=C", func(req *ListProductsRequest) {
		if req.Filter == nil || req.Filter.Status == nil {
			t.Fatalf("expect Status present, got %#v", req.Filter)
		}
		got := req.Filter.Status.NotIn
		if len(got) != 1 || got[0] != "C" {
			t.Fatalf("expect Status.NotIn=[C], got %#v", got)
		}
	})
}

func TestBuildFiltersFromQuery_NameIlikeWithFold(t *testing.T) {
	runQS(t, "name.ilike=nova", func(req *ListProductsRequest) {
		nf := findNameFilter(req.Filter)
		if nf == nil || nf.Contains == nil || *nf.Contains != "nova" || nf.Fold != true {
			t.Fatalf("expect Name.Contains=nova Fold=true, got %#v", req.Filter)
		}
	})
}

func TestBuildFiltersFromQuery_NameIlikeImplicitFold(t *testing.T) {
	runQS(t, "name.ilike=MiXeD", func(req *ListProductsRequest) {
		nf := findNameFilter(req.Filter)
		if nf == nil || nf.Contains == nil || *nf.Contains != "MiXeD" || nf.Fold != true {
			t.Fatalf("expect Name.Contains=MiXeD Fold=true, got %#v", req.Filter)
		}
	})
}

func TestBuildFiltersFromQuery_NameEq(t *testing.T) {
	runQS(t, "name=Alpha", func(req *ListProductsRequest) {
		nf := findNameFilter(req.Filter)
		if nf == nil || nf.Eq == nil || *nf.Eq != "Alpha" || nf.Fold != false {
			t.Fatalf("expect Name.Eq=Alpha Fold=false, got %#v", req.Filter)
		}
	})
}

func TestBuildFiltersFromQuery_CreatedAtLte(t *testing.T) {
	runQS(t, "created_at.lte=2025-10-20T12:00:00Z", func(req *ListProductsRequest) {
		if req.Filter == nil || req.Filter.CreatedAt == nil || req.Filter.CreatedAt.Lte == nil {
			t.Fatalf("expect CreatedAt.Lte set, got %#v", req.Filter.CreatedAt)
		}
	})
}

func TestBuildFiltersFromQuery_NameIlikeTwoVals_StatusIn(t *testing.T) {
	runQS(t, "name.ilike=Alpha&name.ilike=Beta&status.in=A,B", func(req *ListProductsRequest) {
		if req.Filter == nil || len(req.Filter.Or) != 2 {
			t.Fatalf("expect 2 OR children, got %#v", req.Filter)
		}
		if req.Filter.Status == nil {
			t.Fatalf("expect Status present, got %#v", req.Filter)
		}
		got := req.Filter.Status.In
		if len(got) != 2 || got[0] != "A" || got[1] != "B" {
			t.Fatalf("expect Status.In=[A B], got %#v", got)
		}
	})
}

func TestBuildFiltersFromQuery_WithNotAndGroups_Sub(t *testing.T) {
	runQS(t, "g1.name.ilike=Alpha&g1.name.ilike=Beta&g1.__op=or&not.status.in=C", func(req *ListProductsRequest) {
		if req.Filter == nil || len(req.Filter.Or) == 0 {
			t.Fatalf("expect OR group for g1")
		}
		if req.Filter.Not == nil || req.Filter.Not.Status == nil || len(req.Filter.Not.Status.In) != 1 || req.Filter.Not.Status.In[0] != "C" {
			t.Fatalf("expect not.status.in=[C], got %#v", req.Filter.Not)
		}
	})
}

func TestBuildFiltersFromQuery_DefaultEqAndMultiGroups_Sub(t *testing.T) {
	runQS(t, "name=Alpha&g1.code.in=A,B&g1.__op=and&g2.not.locale_code.eq=zh&g2.__op=and", func(req *ListProductsRequest) {
		nf := findNameFilter(req.Filter)
		if nf == nil || nf.Eq == nil || *nf.Eq != "Alpha" || nf.Fold != false {
			t.Fatalf("expect Name.Eq=Alpha Fold=false, got %#v", req.Filter)
		}
		if req.Filter.Code == nil || len(req.Filter.Code.In) != 2 || req.Filter.Code.In[0] != "A" || req.Filter.Code.In[1] != "B" {
			t.Fatalf("expect Code.In=[A B], got %#v", req.Filter.Code)
		}
		if req.Filter.Not == nil || req.Filter.Not.LocaleCode == nil || req.Filter.Not.LocaleCode.Eq == nil || *req.Filter.Not.LocaleCode.Eq != "zh" {
			t.Fatalf("expect Not.LocaleCode.Eq=zh, got %#v", req.Filter.Not)
		}
	})
}

func TestBuildFiltersFromQuery_IsPublishedEqTrue(t *testing.T) {
	runQS(t, "is_published.eq=true", func(req *ListProductsRequest) {
		if req.Filter == nil || req.Filter.IsPublished == nil || req.Filter.IsPublished.Eq == nil || *req.Filter.IsPublished.Eq != true {
			t.Fatalf("expect IsPublished.Eq=true, got %#v", req.Filter)
		}
	})
}

func TestBuildFiltersFromQuery_EqWithFold(t *testing.T) {
	runQS(t, "name.eq=ABC", func(req *ListProductsRequest) {
		nf := findNameFilter(req.Filter)
		if nf == nil || nf.Eq == nil || *nf.Eq != "ABC" || nf.Fold != false {
			t.Fatalf("expect Name.Eq=ABC Fold=false, got %#v", req.Filter)
		}
	})
}

func TestBuildFiltersFromQuery_KeywordQS_FoldImplicit(t *testing.T) {
	qs := "keyword=Nova&f_name.eq=Exact"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"Name", "Code"}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "Exact" || req.Filter.Name.Fold {
		t.Fatalf("expect Name.Eq=Exact and Fold=false, got %#v", req.Filter.Name)
	}

	if len(req.Filter.Or) != 1 {
		t.Fatalf("expect only Name OR child for keyword, got %#v", req.Filter.Or)
	}
	if !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Nova" && req.Filter.Or[0].Name.Fold) {
		t.Fatalf("expect OR Name.Contains=Nova Fold=true, got %#v", req.Filter.Or[0])
	}
}

// Unsupported operators should be pruned for all nodes and fields (direct SearchParams construction)
func TestUnmarshalFilters_UnsupportedOperator_Direct(t *testing.T) {
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		// Code does not support Contains; this condition should be pruned
		{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorContains, Value: "X"}},
		// Name supports Contains; this should remain
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Y", Fold: true}},
	}}}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	// Code must be removed entirely because it has no supported operators
	if req.Filter.Code != nil {
		t.Fatalf("expect Code pruned due to unsupported operator, got %#v", req.Filter.Code)
	}
	// Name.Contains should remain
	nf := findNameFilter(req.Filter)
	if nf == nil || nf.Contains == nil || *nf.Contains != "Y" {
		t.Fatalf("expect Name.Contains=Y, got %#v", nf)
	}
}

// Unsupported operators should be pruned for all nodes and fields (via query string + keyword)
func TestBuildFiltersFromQuery_KeywordQS_UnsupportedFields(t *testing.T) {
	qs := "keyword=Nova"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"Name", "Code", "Status"}}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	// Only Name supports Contains. Expect exactly one OR child with Name.Contains
	if len(req.Filter.Or) != 1 || !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Nova" && req.Filter.Or[0].Name.Fold) {
		t.Fatalf("expect single OR Name.Contains=Nova Fold=true, got %#v", req.Filter.Or)
	}
	// Ensure no OR child carries Code/Status
	for _, ch := range req.Filter.Or {
		if ch.Code != nil || ch.Status != nil {
			t.Fatalf("unexpected Code/Status in OR child: %#v", ch)
		}
	}
}

// NOT/OR nodes that become empty after pruning unsupported operators should be dropped
func TestUnmarshalFilters_NotOr_UnsupportedPruned(t *testing.T) {
	sp := &SearchParams{Filter: &Filter{
		Or: []*Filter{
			// Unsupported: Code.Contains -> should be removed
			{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorContains, Value: "A"}},
			// Supported: Name.Contains remains
			{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "B", Fold: true}},
		},
		// NOT becomes empty and should be removed
		Not: &Filter{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorContains, Value: "X"}},
	}}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	if len(req.Filter.Or) != 1 || !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "B") {
		t.Fatalf("expect single OR Name.Contains=B, got %#v", req.Filter.Or)
	}
	if req.Filter.Not != nil {
		t.Fatalf("expect Not pruned, got %#v", req.Filter.Not)
	}
}

func TestBuildFiltersFromQuery_KeywordQS_FoldImplicit_CamelCase(t *testing.T) {
	qs := "keyword=Nova&f_name.eq=Exact"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name", "code"}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "Exact" || req.Filter.Name.Fold {
		t.Fatalf("expect Name.Eq=Exact and Fold=false, got %#v", req.Filter.Name)
	}

	if len(req.Filter.Or) != 1 {
		t.Fatalf("expect only Name OR child for keyword, got %#v", req.Filter.Or)
	}
	if !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Nova" && req.Filter.Or[0].Name.Fold) {
		t.Fatalf("expect OR Name.Contains=Nova Fold=true, got %#v", req.Filter.Or[0])
	}
}

// Cover buildFiltersFromGroups: IN branch + group op = or (append to groupNode.Or)
// merged into TestBuildFiltersFromQuery_QSSubtests

// Cover handleNotGroup: destination Not field is struct (not pointer)
func TestUnmarshal_NotWithStructField(t *testing.T) {
	type NameOps struct{ Eq *string }
	type Dest struct {
		// Not as struct to cover f.Kind()==reflect.Struct branch
		Not struct {
			Name *NameOps
		}
	}

	sp := &SearchParams{Filter: &Filter{Not: &Filter{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorEq, Value: "X"}}}}
	var dst Dest
	if err := sp.Unmarshal(&dst, WithPascalCase()); err != nil {
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
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "A"}},
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "B"}},
	}}}
	var dst Dest
	if err := sp.Unmarshal(&dst, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if len(dst.Or) != 2 || dst.Or[0].Name == nil || dst.Or[1].Name == nil {
		t.Fatalf("expect 2 struct OR children with Name set, got %#v", dst.Or)
	}
}

// Cover coerceFromString branches: string, bool, int, uint, float via full QS pipeline
func TestBuildFiltersFromQuery_CoerceFromString_AllKinds(t *testing.T) {
	type StrOps struct{ Eq *string }
	type BoolOps struct{ Eq *bool }
	type IntOps struct{ Eq *int64 }
	type UintOps struct{ Eq *uint64 }
	type FloatOps struct{ Eq *float64 }
	type Dest struct {
		Nstr   *StrOps
		Nbool  *BoolOps
		Nint   *IntOps
		Nuint  *UintOps
		Nfloat *FloatOps
	}
	type Req struct{ Filter *Dest }

	qs := "f_nstr=hello&f_nbool=true&f_nint=123&f_nuint=456&f_nfloat=1.25"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}
	var req Req
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil || req.Filter.Nstr == nil || req.Filter.Nstr.Eq == nil || *req.Filter.Nstr.Eq != "hello" {
		t.Fatalf("expect nstr.eq=hello, got %#v", req.Filter)
	}
	if req.Filter.Nbool == nil || req.Filter.Nbool.Eq == nil || *req.Filter.Nbool.Eq != true {
		t.Fatalf("expect nbool.eq=true, got %#v", req.Filter)
	}
	if req.Filter.Nint == nil || req.Filter.Nint.Eq == nil || *req.Filter.Nint.Eq != 123 {
		t.Fatalf("expect nint.eq=123, got %#v", req.Filter)
	}
	if req.Filter.Nuint == nil || req.Filter.Nuint.Eq == nil || *req.Filter.Nuint.Eq != 456 {
		t.Fatalf("expect nuint.eq=456, got %#v", req.Filter)
	}
	if req.Filter.Nfloat == nil || req.Filter.Nfloat.Eq == nil || *req.Filter.Nfloat.Eq != 1.25 {
		t.Fatalf("expect nfloat.eq=1.25, got %#v", req.Filter)
	}
}

// coerceFromString direct coverage (success and failure)
// merged into TestBuildFiltersFromQuery_QSSubtests
func TestBuildFiltersFromQuery_CoerceFromStringPaths(t *testing.T) {}

// ---------- helpers to reduce per-test complexity ----------

func assertFilterInitialized(t *testing.T, req ListProductsRequest) {
	if req.Filter == nil {
		t.Fatalf("Filter should be initialized")
	}
}

func assertNameContains(t *testing.T, req ListProductsRequest, val string) {
	nf := findNameFilter(req.Filter)
	if nf == nil || nf.Contains == nil || *nf.Contains != val {
		t.Fatalf("expect name.contains=%s, got %#v", val, req.Filter)
	}
}

func assertCodeIn(t *testing.T, req ListProductsRequest, want []string) {
	cf := findCodeFilter(req.Filter)
	if cf == nil || len(cf.In) != len(want) {
		t.Fatalf("expect code.in len=%d, got %#v", len(want), req.Filter)
	}
	for i := range want {
		if cf.In[i] != want[i] {
			t.Fatalf("unexpected code.in at %d: %s (want %s)", i, cf.In[i], want[i])
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
	var ok bool
	var visit func(*ProductFilter)
	visit = func(p *ProductFilter) {
		if p == nil {
			return
		}
		if p.Name != nil && p.Name.Contains != nil && *p.Name.Contains == kw {
			ok = true
		}
		for _, ch := range p.Or {
			visit(ch)
		}
		for _, ch := range p.And {
			visit(ch)
		}
		if p.Not != nil {
			visit(p.Not)
		}
	}
	visit(req.Filter)
	if !ok {
		t.Fatalf("expect OR child with name.contains=%s, got %#v", kw, req.Filter)
	}
}

// assertStatusIn/assertStatusNotIn removed; use explicit inline assertions in tests

// find helpers to adapt to preserved AND/OR structure
func findNameFilter(p *ProductFilter) *NameFilter {
	if p == nil {
		return nil
	}
	if p.Name != nil {
		return p.Name
	}
	for _, ch := range p.And {
		if nf := findNameFilter(ch); nf != nil {
			return nf
		}
	}
	for _, ch := range p.Or {
		if nf := findNameFilter(ch); nf != nil {
			return nf
		}
	}
	if p.Not != nil {
		if nf := findNameFilter(p.Not); nf != nil {
			return nf
		}
	}
	return nil
}

func findStatusFilter(p *ProductFilter) *StatusFilter {
	if p == nil {
		return nil
	}
	if p.Status != nil {
		return p.Status
	}
	for _, ch := range p.And {
		if sf := findStatusFilter(ch); sf != nil {
			return sf
		}
	}
	for _, ch := range p.Or {
		if sf := findStatusFilter(ch); sf != nil {
			return sf
		}
	}
	if p.Not != nil {
		if sf := findStatusFilter(p.Not); sf != nil {
			return sf
		}
	}
	return nil
}

func findCodeFilter(p *ProductFilter) *CodeFilter {
	if p == nil {
		return nil
	}
	if p.Code != nil {
		return p.Code
	}
	for _, ch := range p.And {
		if cf := findCodeFilter(ch); cf != nil {
			return cf
		}
	}
	for _, ch := range p.Or {
		if cf := findCodeFilter(ch); cf != nil {
			return cf
		}
	}
	if p.Not != nil {
		if cf := findCodeFilter(p.Not); cf != nil {
			return cf
		}
	}
	return nil
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

	qs := "f_price.gt=10&f_price.gte=11&f_price.lt=20&f_price.lte=19"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}
	var req Req
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
	qs := "f_not.name.eq=Alpha"
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Not == nil || req.Filter.Not.Name == nil || req.Filter.Not.Name.Eq == nil || *req.Filter.Not.Name.Eq != "Alpha" {
		t.Fatalf("expect Not.Name.Eq=Alpha, got %#v", req.Filter)
	}
}

// IN/NotIn with CSV string values should be split and coerced via coerceToSlice string path
func TestUnmarshal_In_CSVString(t *testing.T) {
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: &FieldCondition{Field: "Code", Operator: FilterOperatorIn, Value: "A, B ,C"}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
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
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorNotIn, Value: "X,Y"}},
	}}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Status == nil || len(req.Filter.Status.NotIn) != 2 || req.Filter.Status.NotIn[0] != "X" || req.Filter.Status.NotIn[1] != "Y" {
		t.Fatalf("expect Status.NotIn=[X Y], got %#v", req.Filter.Status)
	}
}

// LocaleStatus filter coverage: Eq and In via both direct Filter tree and query string
func TestUnmarshalFilters_LocaleStatus(t *testing.T) {
	// Direct Filter tree
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: &FieldCondition{Field: "LocaleStatus", Operator: FilterOperatorEq, Value: string(LocaleStatusPublished)}},
		{Condition: &FieldCondition{Field: "LocaleStatus", Operator: FilterOperatorIn, Value: []string{string(LocaleStatusDraft), string(LocaleStatusPublished)}}},
	}}}

	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.LocaleStatus == nil {
		t.Fatalf("expect LocaleStatus initialized, got %#v", req.Filter)
	}
	if req.Filter.LocaleStatus.Eq == nil || *req.Filter.LocaleStatus.Eq != LocaleStatusPublished {
		t.Fatalf("expect LocaleStatus.Eq=PUBLISHED, got %#v", req.Filter.LocaleStatus)
	}
	if len(req.Filter.LocaleStatus.In) != 2 {
		t.Fatalf("expect LocaleStatus.In len=2, got %#v", req.Filter.LocaleStatus)
	}

	// Via query string
	runQS(t, "locale_status.in=PUBLISHED,DRAFT&locale_status.eq=DRAFT", func(r *ListProductsRequest) {
		if r.Filter == nil || r.Filter.LocaleStatus == nil {
			t.Fatalf("expect LocaleStatus initialized, got %#v", r.Filter)
		}
		if r.Filter.LocaleStatus.Eq == nil || *r.Filter.LocaleStatus.Eq != LocaleStatusDraft {
			t.Fatalf("expect LocaleStatus.Eq=DRAFT, got %#v", r.Filter.LocaleStatus)
		}
		if len(r.Filter.LocaleStatus.In) != 2 || r.Filter.LocaleStatus.In[0] != LocaleStatusPublished || r.Filter.LocaleStatus.In[1] != LocaleStatusDraft {
			t.Fatalf("expect LocaleStatus.In=[PUBLISHED DRAFT], got %#v", r.Filter.LocaleStatus)
		}
	})
}

// Ensure Unmarshal adapts to structs that use lowerCamel JSON tags
func TestUnmarshalFilters_JSONTagLowerCamel_Direct(t *testing.T) {
	// Direct Filter tree
	sp := &SearchParams{Filter: &Filter{And: []*Filter{
		{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: "Jack", Fold: true}},
		{Condition: &FieldCondition{Field: "Status", Operator: FilterOperatorIn, Value: []string{"A", "B"}}},
	}}}
	var req ListUserRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.Name == nil || req.Filter.Name.Contains == nil || *req.Filter.Name.Contains != "Jack" || !req.Filter.Name.Fold {
		t.Fatalf("expect name.contains=Jack fold=true, got %#v", req.Filter)
	}
	if req.Filter.Status == nil || len(req.Filter.Status.In) != 2 || req.Filter.Status.In[0] != "A" || req.Filter.Status.In[1] != "B" {
		t.Fatalf("expect status.in=[A B], got %#v", req.Filter.Status)
	}
}

func TestUnmarshalFilters_JSONTagLowerCamel_ViaQueryString(t *testing.T) {
	// reuse BuildFiltersFromQuery -> Unmarshal pipeline against ListUserRequest
	root := BuildFiltersFromQuery("f_name.ilike=Mike&f_status.in=X,Y")
	sp2 := &SearchParams{Filter: root}
	var req2 ListUserRequest
	if err := sp2.Unmarshal(&req2.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req2.Filter == nil || req2.Filter.Name == nil || req2.Filter.Name.Contains == nil || *req2.Filter.Name.Contains != "Mike" || !req2.Filter.Name.Fold {
		t.Fatalf("expect name.contains=Mike fold=true, got %#v", req2.Filter)
	}
	if req2.Filter.Status == nil || len(req2.Filter.Status.In) != 2 || req2.Filter.Status.In[0] != "X" || req2.Filter.Status.In[1] != "Y" {
		t.Fatalf("expect status.in=[X Y], got %#v", req2.Filter.Status)
	}
}

// Standalone keyword test for lowerCamel JSON tag target
func TestUnmarshalFilters_JSONTagLowerCamel_Keyword(t *testing.T) {
	// Only target name by providing KeywordColumns explicitly
	qsK := "keyword=Neo"
	v, _ := url.ParseQuery(qsK)
	root := BuildFiltersFromQuery(qsK)
	spK := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}
	var reqK ListUserRequest
	if err := spK.Unmarshal(&reqK.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if reqK.Filter == nil || len(reqK.Filter.Or) != 1 {
		t.Fatalf("expect 1 OR child for keyword, got %#v", reqK.Filter)
	}
	if !(reqK.Filter.Or[0].Name != nil && reqK.Filter.Or[0].Name.Contains != nil && *reqK.Filter.Or[0].Name.Contains == "Neo") {
		t.Fatalf("expect OR name.contains=Neo, got %#v", reqK.Filter.Or[0])
	}
}

// Nested OR/NOT with extra QS fields (ignored by target) for lowerCamel JSON tag target
func TestUnmarshalFilters_JSONTagLowerCamel_NestedAndQueryExtras(t *testing.T) {
	// QS includes:
	// - g1 OR group: two name.ilike
	// - not.status.in=C
	// - name.eq=Exact (root)
	// - code.in=A,B (field not present in ListUserRequest -> should be ignored)
	// - user_name.ilike=Omega (maps to UserName field)
	qs := "f_g1.name.ilike=Alpha&f_g1.name.ilike=Beta&f_g1.__op=or&f_not.status.in=C&f_name.eq=Exact&f_code.in=A,B&f_user_name.ilike=Omega&keyword=Neo"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}
	var req ListUserRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	// Root name.eq
	if req.Filter.Name == nil || req.Filter.Name.Eq == nil || *req.Filter.Name.Eq != "Exact" {
		t.Fatalf("expect name.eq=Exact, got %#v", req.Filter.Name)
	}
	// userName.contains=Omega
	if req.Filter.UserName == nil || req.Filter.UserName.Contains == nil || *req.Filter.UserName.Contains != "Omega" {
		t.Fatalf("expect userName.contains=Omega, got %#v", req.Filter.UserName)
	}
	// NOT status.in=C
	if req.Filter.Not == nil || req.Filter.Not.Status == nil || len(req.Filter.Not.Status.In) != 1 || req.Filter.Not.Status.In[0] != "C" {
		t.Fatalf("expect not.status.in=[C], got %#v", req.Filter.Not)
	}
	// OR group contains Name.Contains with values Alpha, Beta plus one keyword OR child Neo
	if len(req.Filter.Or) == 0 {
		t.Fatalf("expect OR children for group g1")
	}
	seenAlpha, seenBeta, seenNeo := false, false, false
	for _, ch := range req.Filter.Or {
		if ch.Name != nil && ch.Name.Contains != nil {
			switch *ch.Name.Contains {
			case "Alpha":
				seenAlpha = true
			case "Beta":
				seenBeta = true
			case "Neo":
				seenNeo = true
			}
		}
	}
	if !(seenAlpha && seenBeta && seenNeo) {
		t.Fatalf("expect OR contains Alpha, Beta and keyword Neo; got alpha=%v beta=%v neo=%v (ors=%#v)", seenAlpha, seenBeta, seenNeo, req.Filter.Or)
	}
}

// ProductStatus (int32) with QS + keyword combined for lowerCamel target
func TestUnmarshalFilters_JSONTagLowerCamel_ProductStatusWithKeyword(t *testing.T) {
	// QS includes product_status comparators and keyword; limit keyword to name via KeywordColumns
	// Using numeric values matching constants
	qs := "f_product_status.in=1,6&f_product_status.notin=9&keyword=Neo"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}

	var req ListUserRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req.Filter == nil || req.Filter.ProductStatus == nil {
		t.Fatalf("expect productStatus present, got %#v", req.Filter)
	}
	// Check IN values parsed as int32 and mapped correctly
	if len(req.Filter.ProductStatus.In) != 2 || req.Filter.ProductStatus.In[0] != ProductStatusPublished || req.Filter.ProductStatus.In[1] != ProductStatusActive {
		t.Fatalf("expect productStatus.in=[%v %v], got %#v", ProductStatusPublished, ProductStatusActive, req.Filter.ProductStatus.In)
	}
	if len(req.Filter.ProductStatus.NotIn) != 1 || req.Filter.ProductStatus.NotIn[0] != ProductStatusCancelled {
		t.Fatalf("expect productStatus.notIn=[%v], got %#v", ProductStatusCancelled, req.Filter.ProductStatus.NotIn)
	}

	// Keyword OR child should be generated for name
	if len(req.Filter.Or) != 1 {
		t.Fatalf("expect 1 OR child for keyword, got %#v", req.Filter.Or)
	}
	if !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Neo") {
		t.Fatalf("expect OR name.contains=Neo, got %#v", req.Filter.Or[0])
	}
}

// ProductStatus eq with QS + keyword combined for lowerCamel target
func TestUnmarshalFilters_JSONTagLowerCamel_ProductStatusEqWithKeyword(t *testing.T) {
	qs := "f_product_status.eq=6&keyword=Neo"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}

	var req ListUserRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.ProductStatus == nil || req.Filter.ProductStatus.Eq == nil || *req.Filter.ProductStatus.Eq != ProductStatusActive {
		t.Fatalf("expect productStatus.eq=%v, got %#v", ProductStatusActive, req.Filter.ProductStatus)
	}
	if len(req.Filter.Or) != 1 || !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Neo") {
		t.Fatalf("expect keyword OR name.contains=Neo, got %#v", req.Filter.Or)
	}
}

// Custom numeric-like types with QS + keyword combined
func TestUnmarshalFilters_JSONTagLowerCamel_CustomNumericTypesWithKeyword(t *testing.T) {
	// QS mixes various custom numeric aliases and keyword; limit keyword to name via KeywordColumns
	qs := "f_category_level.in=2,3&f_order_state.notin=5&f_rank.in=7&f_score.in=1.5,2.75&f_user_id.in=U123,U456&f_user_id.ilike=U12&keyword=Zed"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}

	var req ListCustomRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	if req.Filter.CategoryLevel == nil || len(req.Filter.CategoryLevel.In) != 2 || req.Filter.CategoryLevel.In[0] != LevelPro || req.Filter.CategoryLevel.In[1] != LevelElite {
		t.Fatalf("expect categoryLevel.in=[%v %v], got %#v", LevelPro, LevelElite, req.Filter.CategoryLevel)
	}
	if req.Filter.OrderState == nil || len(req.Filter.OrderState.NotIn) != 1 || req.Filter.OrderState.NotIn[0] != StateClosed {
		t.Fatalf("expect orderState.notIn=[%v], got %#v", StateClosed, req.Filter.OrderState)
	}
	if req.Filter.Rank == nil || len(req.Filter.Rank.In) != 1 || req.Filter.Rank.In[0] != RankGold {
		t.Fatalf("expect rank.in=[%v], got %#v", RankGold, req.Filter.Rank)
	}
	if req.Filter.Score == nil || len(req.Filter.Score.In) != 2 || req.Filter.Score.In[0] != Score(1.5) || req.Filter.Score.In[1] != Score(2.75) {
		t.Fatalf("expect score.in=[1.5 2.75], got %#v", req.Filter.Score)
	}
	if req.Filter.UserId == nil || len(req.Filter.UserId.In) != 2 || req.Filter.UserId.In[0] != "U123" || req.Filter.UserId.In[1] != "U456" {
		t.Fatalf("expect userId.in=[U123 U456], got %#v", req.Filter.UserId)
	}
	if req.Filter.UserId.Contains == nil || *req.Filter.UserId.Contains != "U12" {
		t.Fatalf("expect userId.contains=U12, got %#v", req.Filter.UserId)
	}
	if len(req.Filter.Or) != 1 || !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Zed") {
		t.Fatalf("expect keyword OR name.contains=Zed, got %#v", req.Filter.Or)
	}
}

// Time comparators (created_at/updated_at) with lowerCamel target and keyword
func TestUnmarshalFilters_JSONTagLowerCamel_TimeWithKeyword(t *testing.T) {
	qs := "f_created_at.gte=2025-10-20T12:34:56Z&f_updated_at.lte=2025-12-01T00:00:00Z&keyword=Omega"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"name"}}

	var req ListUserRequest
	if err := sp.Unmarshal(&req.Filter); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || req.Filter.CreatedAt == nil || req.Filter.CreatedAt.Gte == nil {
		t.Fatalf("expect createdAt.gte set, got %#v", req.Filter.CreatedAt)
	}
	if req.Filter == nil || req.Filter.UpdatedAt == nil || req.Filter.UpdatedAt.Lte == nil {
		t.Fatalf("expect updatedAt.lte set, got %#v", req.Filter.UpdatedAt)
	}
	gte := req.Filter.CreatedAt.Gte.AsTime().UTC()
	if gte.IsZero() || gte.Year() != 2025 || gte.Month() != time.October || gte.Day() != 20 {
		t.Fatalf("unexpected createdAt.gte time: %v", gte)
	}
	lte := req.Filter.UpdatedAt.Lte.AsTime().UTC()
	if lte.IsZero() || lte.Year() != 2025 || lte.Month() != time.December || lte.Day() != 1 {
		t.Fatalf("unexpected updatedAt.lte time: %v", lte)
	}
	if len(req.Filter.Or) != 1 || !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "Omega") {
		t.Fatalf("expect keyword OR name.contains=Omega, got %#v", req.Filter.Or)
	}
}

// Numeric keyword should still map to Name.contains as string with fold=true
func TestUnmarshalFilters_KeywordNumeric_Name(t *testing.T) {
	qs := "keyword=123"
	v, _ := url.ParseQuery(qs)
	root := BuildFiltersFromQuery(qs)
	sp := &SearchParams{Filter: root, Keyword: v.Get("keyword"), KeywordColumns: []string{"Name"}}
	var req ListProductsRequest
	if err := sp.Unmarshal(&req.Filter, WithPascalCase()); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil || len(req.Filter.Or) != 1 {
		t.Fatalf("expect 1 OR child for keyword, got %#v", req.Filter)
	}
	if !(req.Filter.Or[0].Name != nil && req.Filter.Or[0].Name.Contains != nil && *req.Filter.Or[0].Name.Contains == "123" && req.Filter.Or[0].Name.Fold) {
		t.Fatalf("expect OR name.contains=\"123\" fold=true, got %#v", req.Filter.Or[0])
	}
}

func TestUnmarshal_ScopeTypeOperatorRename(t *testing.T) {
	type AliasNameFilter struct {
		Contains *string `json:"contains"`
		Fold     bool    `json:"fold"`
	}
	type NameFilter struct {
		Contains *string `json:"contains"`
		Fold     bool    `json:"fold"`
	}
	type RootAliasFilter struct {
		// Having Name at root ensures ScopeType(NameOps) can be matched.
		AliasName *AliasNameFilter `json:"aliasName"`
		Name      *NameFilter      `json:"name"`
	}
	type Req struct{ Filter *RootAliasFilter }

	val := "Acme"
	root := &Filter{
		And: []*Filter{
			{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorContains, Value: val, Fold: true}},
		},
	}
	sp := &SearchParams{Filter: root}
	var req Req
	err := sp.Unmarshal(&req.Filter,
		WithFilterUnmarshalHook(func(next FilterUnmarshalFunc) FilterUnmarshalFunc {
			return func(in *FilterUnmarshalInput) error {
				if in.Field == "Name" {
					in.FilterMap["aliasName"] = in.FilterMap[in.Field]
					delete(in.FilterMap, in.Field)
				}
				return next(in)
			}
		}),
	)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if req.Filter == nil {
		t.Fatalf("filter nil")
	}
	if req.Filter.AliasName == nil || req.Filter.AliasName.Contains == nil || *req.Filter.AliasName.Contains != "Acme" || !req.Filter.AliasName.Fold {
		t.Fatalf("expect AliasName.Contains=Acme and fold=true, got %#v", req.Filter.AliasName)
	}
	if req.Filter.Name != nil {
		t.Fatalf("expect Name is nil  got %#v", req.Filter.Name)
	}
}
