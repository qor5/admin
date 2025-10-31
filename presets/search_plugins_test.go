package presets

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/qor5/web/v3"
)

func TestWrapSearchOps_InOperatorSnakeAndPascal(t *testing.T) {
	ids := []string{"u1", "u2"}
	op := func(ec *web.EventContext, _ *SearchParams) (*Filter, error) {
		// Business op returns Filter only; framework compiles SQL and merges Filter.
		return &Filter{Condition: &FieldCondition{
			Field:    "UserID", // should remain PascalCase in Filter
			Operator: FilterOperatorIn,
			Value:    ids,
		}}, nil
	}

	var captured *SearchParams
	wrapped := WrapSearchOps(op)(func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
		// capture params after wrapper work
		cp := *params
		captured = &cp
		return &SearchResult{}, nil
	})

	ec := &web.EventContext{R: &http.Request{}}
	_, err := wrapped(ec, &SearchParams{})
	if err != nil {
		t.Fatalf("wrapped search returned error: %v", err)
	}
	if captured == nil {
		t.Fatalf("captured params is nil")
	}

	// Filter should contain PascalCase field
	if captured.Filter == nil || captured.Filter.Condition == nil {
		t.Fatalf("expected filter condition to be set")
	}
	if captured.Filter.Condition.Field != "UserID" {
		t.Fatalf("expected Filter field UserID, got %s", captured.Filter.Condition.Field)
	}

	// SQLConditions should be compiled from op filter, using snake_case column
	if len(captured.SQLConditions) != 1 {
		t.Fatalf("expected 1 SQLCondition, got %d", len(captured.SQLConditions))
	}
	sc := captured.SQLConditions[0]
	if sc == nil {
		t.Fatalf("SQLCondition is nil")
	}
	if sc.Query != "user_id IN ?" {
		t.Fatalf("unexpected query: %s", sc.Query)
	}
	expectedArgs := []any{ids}
	if !reflect.DeepEqual(sc.Args, expectedArgs) {
		t.Fatalf("unexpected args: %#v", sc.Args)
	}
}

func TestWrapSearchOps_FoldPreferenceAffectsLike(t *testing.T) {
	// Base filter sets fold=false for Name
	base := &Filter{Condition: &FieldCondition{Field: "Name", Operator: FilterOperatorFold, Value: false}}
	// Op uses Contains => should compile to LIKE (not ILIKE)
	op := func(ec *web.EventContext, _ *SearchParams) (*Filter, error) {
		return &Filter{Condition: &FieldCondition{
			Field:    "Name",
			Operator: FilterOperatorContains,
			Value:    "Alpha",
		}}, nil
	}

	var captured *SearchParams
	wrapped := WrapSearchOps(op)(func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
		cp := *params
		captured = &cp
		return &SearchResult{}, nil
	})

	ec := &web.EventContext{R: &http.Request{}}
	_, err := wrapped(ec, &SearchParams{Filter: base})
	if err != nil {
		t.Fatalf("wrapped search returned error: %v", err)
	}
	if captured == nil || len(captured.SQLConditions) != 1 {
		t.Fatalf("expected 1 SQLCondition")
	}
	q := captured.SQLConditions[0].Query
	if !strings.Contains(q, " LIKE ") || strings.Contains(q, " ILIKE ") {
		t.Fatalf("expected LIKE (fold=false), got %s", q)
	}
	if !strings.HasPrefix(q, "name ") {
		t.Fatalf("expected snake_case column name, got %s", q)
	}
}

func TestWrapSearchOps_EmptyInProducesNoMatch(t *testing.T) {
	op := func(ec *web.EventContext, _ *SearchParams) (*Filter, error) {
		return &Filter{Condition: &FieldCondition{
			Field:    "UserID",
			Operator: FilterOperatorIn,
			Value:    []string{},
		}}, nil
	}

	var captured *SearchParams
	wrapped := WrapSearchOps(op)(func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
		cp := *params
		captured = &cp
		return &SearchResult{}, nil
	})

	ec := &web.EventContext{R: &http.Request{}}
	_, err := wrapped(ec, &SearchParams{})
	if err != nil {
		t.Fatalf("wrapped search returned error: %v", err)
	}
	if captured == nil || len(captured.SQLConditions) != 1 {
		t.Fatalf("expected 1 SQLCondition")
	}
	if captured.SQLConditions[0].Query != "1 = 0" {
		t.Fatalf("expected '1 = 0', got %s", captured.SQLConditions[0].Query)
	}
}

func TestWrapSearchOps_FieldNameNormalization_SnakeToPascalAndSnakeSQL(t *testing.T) {
	op := func(ec *web.EventContext, _ *SearchParams) (*Filter, error) {
		return &Filter{Condition: &FieldCondition{
			Field:    "user_id", // snake_case input
			Operator: FilterOperatorEq,
			Value:    "u1",
		}}, nil
	}

	var captured *SearchParams
	wrapped := WrapSearchOps(op)(func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
		cp := *params
		captured = &cp
		return &SearchResult{}, nil
	})

	ec := &web.EventContext{R: &http.Request{}}
	_, err := wrapped(ec, &SearchParams{})
	if err != nil {
		t.Fatalf("wrapped search returned error: %v", err)
	}
	if captured == nil || captured.Filter == nil || captured.Filter.Condition == nil {
		t.Fatalf("expected filter condition to be set")
	}
	// PascalCase normalization for Filter field
	if captured.Filter.Condition.Field != "UserId" { // ToCamel("user_id") => "UserId"
		t.Fatalf("expected Filter field UserId, got %s", captured.Filter.Condition.Field)
	}
	if len(captured.SQLConditions) != 1 || captured.SQLConditions[0] == nil {
		t.Fatalf("expected 1 SQLCondition")
	}
	// snake_case column in SQL
	if captured.SQLConditions[0].Query != "user_id = ?" {
		t.Fatalf("unexpected SQL query: %s", captured.SQLConditions[0].Query)
	}
}

func TestWrapSearchOps_FieldNameNormalization_LowerToPascalAndSnakeSQL(t *testing.T) {
	op := func(ec *web.EventContext, _ *SearchParams) (*Filter, error) {
		return &Filter{Condition: &FieldCondition{
			Field:    "name", // lower input
			Operator: FilterOperatorContains,
			Value:    "abc",
		}}, nil
	}

	var captured *SearchParams
	wrapped := WrapSearchOps(op)(func(ec *web.EventContext, params *SearchParams) (*SearchResult, error) {
		cp := *params
		captured = &cp
		return &SearchResult{}, nil
	})

	ec := &web.EventContext{R: &http.Request{}}
	_, err := wrapped(ec, &SearchParams{})
	if err != nil {
		t.Fatalf("wrapped search returned error: %v", err)
	}
	if captured == nil || captured.Filter == nil || captured.Filter.Condition == nil {
		t.Fatalf("expected filter condition to be set")
	}
	// PascalCase normalization for Filter field
	if captured.Filter.Condition.Field != "Name" {
		t.Fatalf("expected Filter field Name, got %s", captured.Filter.Condition.Field)
	}
	if len(captured.SQLConditions) != 1 || captured.SQLConditions[0] == nil {
		t.Fatalf("expected 1 SQLCondition")
	}
	q := captured.SQLConditions[0].Query
	if !strings.Contains(q, "name ILIKE ?") { // default fold=true -> ILIKE
		t.Fatalf("unexpected SQL LIKE query: %s", q)
	}
}
