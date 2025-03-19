package bq

import (
	"context"
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/marketing/tag"
	"github.com/samber/lo"
)

// StringOperator defines the type for string operator constants
type StringOperator string

// StringOperator constants
const (
	StringOperatorEQ         StringOperator = "EQ"
	StringOperatorNE         StringOperator = "NE"
	StringOperatorContains   StringOperator = "CONTAINS"
	StringOperatorStartsWith StringOperator = "STARTS_WITH"
	StringOperatorEndsWith   StringOperator = "ENDS_WITH"
	StringOperatorIn         StringOperator = "IN"
	StringOperatorNotIn      StringOperator = "NOT_IN"
)

// NumberOperator defines the type for number comparison operators
type NumberOperator string

// NumberOperator constants
const (
	NumberOperatorEQ      NumberOperator = "EQ"
	NumberOperatorNE      NumberOperator = "NE"
	NumberOperatorLT      NumberOperator = "LT"
	NumberOperatorLTE     NumberOperator = "LTE"
	NumberOperatorGT      NumberOperator = "GT"
	NumberOperatorGTE     NumberOperator = "GTE"
	NumberOperatorBetween NumberOperator = "BETWEEN"
)

// Accumulation defines the type for accumulation constants
type Accumulation string

// Accumulation constants
const (
	AccumulationCount Accumulation = "COUNT"
	AccumulationDays  Accumulation = "DAYS"
)

// TimeRange defines the type for time range constants
type TimeRange string

// Time range constants
const (
	TimeRange7Days  TimeRange = "7D"
	TimeRange10Days TimeRange = "10D"
	TimeRange30Days TimeRange = "30D"
	TimeRange90Days TimeRange = "90D"
)

// NumberTagBuilder creates a tag builder for number fields
func NumberTagBuilder(id, name, description, categoryID, fieldKey string, min, max float64) tag.Builder {
	md := &tag.Metadata{
		ID:          id,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		View: &tag.View{
			Fragments: []tag.Fragment{
				&tag.SelectFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "operator",
						Required: true,
					},
					Options: []*tag.Option{
						{Value: string(NumberOperatorEQ), Label: "Equals"},
						{Value: string(NumberOperatorNE), Label: "Not Equals"},
						{Value: string(NumberOperatorLT), Label: "Less Than"},
						{Value: string(NumberOperatorLTE), Label: "Less Than or Equals"},
						{Value: string(NumberOperatorGT), Label: "Greater Than"},
						{Value: string(NumberOperatorGTE), Label: "Greater Than or Equals"},
						{Value: string(NumberOperatorBetween), Label: "Between"},
					},
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "value",
						Required: true,
						SkipIf: map[string]any{
							"operator": string(NumberOperatorBetween),
						},
					},
					Min: min,
					Max: max,
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "min",
						Required: true,
						SkipUnless: map[string]any{
							"operator": string(NumberOperatorBetween),
						},
					},
					Min: min,
					Max: max,
				},
				&tag.TextFragment{
					FragmentMetadata: tag.FragmentMetadata{
						SkipUnless: map[string]any{
							"operator": string(NumberOperatorBetween),
						},
					},
					Text: "and",
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "max",
						Required: true,
						SkipUnless: map[string]any{
							"operator": string(NumberOperatorBetween),
						},
					},
					Min: min,
					Max: max,
				},
			},
		},
	}

	sqlTemplate := fmt.Sprintf(`SELECT user_id FROM users WHERE %s
{{- if eq .operator "EQ"}} = {{arg .value}}
{{- else if eq .operator "NE"}} != {{arg .value}}
{{- else if eq .operator "LT"}} < {{arg .value}}
{{- else if eq .operator "LTE"}} <= {{arg .value}}
{{- else if eq .operator "GT"}} > {{arg .value}}
{{- else if eq .operator "GTE"}} >= {{arg .value}}
{{- else if eq .operator "BETWEEN"}} BETWEEN {{arg .min}} AND {{arg .max}}
{{- else }} = {{arg .value}}
{{- end }}`, fieldKey)
	return tag.NewSQLTemplate(md, sqlTemplate)
}

// StringTagBuilder creates a tag builder for string input with selectable operators
func StringTagBuilder(id, name, description, categoryID, fieldKey string, options []*tag.Option) tag.Builder {
	var fragments []tag.Fragment

	mdForValue := tag.FragmentMetadata{
		Key:      "value",
		Required: true,
		SkipIf: map[string]any{
			"$operator": map[string]any{
				string(tag.SkipOperatorIN): []string{string(StringOperatorIn), string(StringOperatorNotIn)},
			},
		},
	}
	mdForValues := tag.FragmentMetadata{
		Key:      "values",
		Required: true,
		SkipUnless: map[string]any{
			"$operator": map[string]any{
				string(tag.SkipOperatorIN): []string{string(StringOperatorIn), string(StringOperatorNotIn)},
			},
		},
	}

	if len(options) <= 0 {
		fragments = []tag.Fragment{
			&tag.SelectFragment{
				FragmentMetadata: tag.FragmentMetadata{
					Key:          "operator",
					Required:     true,
					DefaultValue: string(StringOperatorEQ),
				},
				Options: []*tag.Option{
					{Value: string(StringOperatorEQ), Label: "equals"},
					{Value: string(StringOperatorNE), Label: "not equals"},
					{Value: string(StringOperatorContains), Label: "contains"},
					{Value: string(StringOperatorStartsWith), Label: "starts with"},
					{Value: string(StringOperatorEndsWith), Label: "ends with"},
					{Value: string(StringOperatorIn), Label: "in"},
					{Value: string(StringOperatorNotIn), Label: "not in"},
				},
			},
			&tag.TextInputFragment{
				FragmentMetadata: mdForValue,
			},
			&tag.TextInputFragment{
				FragmentMetadata: mdForValues,
			},
		}
	} else {
		fragments = []tag.Fragment{
			&tag.SelectFragment{
				FragmentMetadata: tag.FragmentMetadata{
					Key:          "operator",
					Required:     true,
					DefaultValue: string(StringOperatorEQ),
				},
				Options: []*tag.Option{
					{Value: string(StringOperatorEQ), Label: "equals"},
					{Value: string(StringOperatorNE), Label: "not equals"},
					{Value: string(StringOperatorIn), Label: "in"},
					{Value: string(StringOperatorNotIn), Label: "not in"},
				},
			},
			&tag.SelectFragment{
				FragmentMetadata: mdForValue,
				Multiple:         false,
				Options:          options,
			},
			&tag.SelectFragment{
				FragmentMetadata: mdForValues,
				Multiple:         true,
				Options:          options,
			},
		}
	}

	md := &tag.Metadata{
		ID:          id,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		View:        &tag.View{Fragments: fragments},
	}

	sqlTemplate := fmt.Sprintf(`SELECT user_id FROM users WHERE
{{- if eq .operator "EQ" }} %[1]s = {{arg .value}}
{{- else if eq .operator "NE" }} %[1]s != {{arg .value}}
{{- else if eq .operator "CONTAINS" }} %[1]s LIKE CONCAT('%%', {{arg .value}}, '%%')
{{- else if eq .operator "STARTS_WITH" }} %[1]s LIKE CONCAT({{arg .value}}, '%%')
{{- else if eq .operator "ENDS_WITH" }} %[1]s LIKE CONCAT('%%', {{arg .value}})
{{- else if eq .operator "IN" }} %[1]s IN ({{argEach .values}})
{{- else if eq .operator "NOT_IN" }} %[1]s NOT IN ({{argEach .values}})
{{- else }} %[1]s = {{arg .value}}
{{- end }}`, fieldKey)

	return tag.WrapSQLBuilder(
		tag.NewSQLTemplate(md, sqlTemplate),
		func(next tag.BuildSQLFunc) tag.BuildSQLFunc {
			return func(ctx context.Context, params map[string]any) (*tag.SQL, error) {
				if len(options) <= 0 {
					values, ok := params["values"].(string)
					if ok {
						params["values"] = lo.Map(strings.Split(values, ","), func(item string, index int) string {
							return strings.TrimSpace(item)
						})
					}
				}
				return next(ctx, params)
			}
		},
	)
}

// DateRangeTagBuilder creates a tag builder for date fields
func DateRangeTagBuilder(id, name, description, categoryID, fieldKey string, includeTime bool) tag.Builder {
	md := &tag.Metadata{
		ID:          id,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		View: &tag.View{
			Fragments: []tag.Fragment{
				&tag.TextFragment{
					Text: "Between",
				},
				&tag.DatePickerFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "start",
						Required: true,
					},
					IncludeTime: includeTime,
				},
				&tag.TextFragment{
					Text: "and",
				},
				&tag.DatePickerFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "end",
						Required: true,
					},
					IncludeTime: includeTime,
				},
			},
		},
	}

	sqlTemplate := fmt.Sprintf("SELECT user_id FROM users WHERE %s BETWEEN {{arg .start}} AND {{arg .end}}", fieldKey)
	return tag.NewSQLTemplate(md, sqlTemplate)
}

// EventTagBuilder creates a tag builder for a specific event type with customizable aggregation
func EventTagBuilder(eventName string, displayLabel string, categoryID string) tag.Builder {
	if categoryID == "" {
		categoryID = "activities"
	}
	id := fmt.Sprintf("event_%s", strings.ToLower(string(eventName)))
	name := fmt.Sprintf("%s Events", displayLabel)
	description := fmt.Sprintf("Filter users by %s events in a time period", strings.ToLower(displayLabel))

	md := &tag.Metadata{
		ID:          id,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		View: &tag.View{
			Fragments: []tag.Fragment{
				&tag.TextFragment{
					Text: "with",
				},
				&tag.SelectFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:          "accumulation",
						Required:     true,
						DefaultValue: string(Accumulation(AccumulationCount)),
					},
					Options: []*tag.Option{
						{Value: string(Accumulation(AccumulationCount)), Label: "total occurrences"},
						{Value: string(Accumulation(AccumulationDays)), Label: "unique days"},
					},
				},
				&tag.SelectFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:          "countOperator",
						Required:     true,
						DefaultValue: string(NumberOperatorGTE),
					},
					Options: []*tag.Option{
						{Value: string(NumberOperatorEQ), Label: "exactly"},
						{Value: string(NumberOperatorNE), Label: "not exactly"},
						{Value: string(NumberOperatorLT), Label: "less than"},
						{Value: string(NumberOperatorLTE), Label: "at most"},
						{Value: string(NumberOperatorGT), Label: "more than"},
						{Value: string(NumberOperatorGTE), Label: "at least"},
						{Value: string(NumberOperatorBetween), Label: "between"},
					},
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:          "countValue",
						Required:     true,
						DefaultValue: float64(1),
						SkipIf: map[string]any{
							"countOperator": string(NumberOperatorBetween),
						},
					},
					Min: 1,
					Max: 1000,
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "countMin",
						Required: true,
						SkipUnless: map[string]any{
							"countOperator": string(NumberOperatorBetween),
						},
					},
					Min: 1,
					Max: 1000,
				},
				&tag.TextFragment{
					FragmentMetadata: tag.FragmentMetadata{
						SkipUnless: map[string]any{
							"countOperator": string(NumberOperatorBetween),
						},
					},
					Text: "and",
				},
				&tag.NumberInputFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:      "countMax",
						Required: true,
						SkipUnless: map[string]any{
							"countOperator": string(NumberOperatorBetween),
						},
					},
					Min: 1,
					Max: 1000,
				},
				&tag.TextFragment{
					Text: "times in",
				},
				&tag.SelectFragment{
					FragmentMetadata: tag.FragmentMetadata{
						Key:          "timeRange",
						Required:     true,
						DefaultValue: string(TimeRange30Days),
					},
					Options: []*tag.Option{
						{Value: string(TimeRange7Days), Label: "last 7 days"},
						{Value: string(TimeRange10Days), Label: "last 10 days"},
						{Value: string(TimeRange30Days), Label: "last 30 days"},
						{Value: string(TimeRange90Days), Label: "last 90 days"},
					},
				},
			},
		},
	}

	// SQL to find users based on the selected accumulation type
	sqlTemplate := fmt.Sprintf(`SELECT user_id FROM events
WHERE event_name = %s
AND created_at >= 
{{- if eq .timeRange "7D" }} TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
{{- else if eq .timeRange "10D" }} TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 10 DAY)
{{- else if eq .timeRange "30D" }} TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
{{- else if eq .timeRange "90D" }} TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 90 DAY)
{{- else }} TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) 
{{- end }}
GROUP BY user_id
{{- if eq .accumulation "COUNT" }}
HAVING COUNT(1) 
{{- else if eq .accumulation "DAYS" }}
HAVING COUNT(DISTINCT DATE(created_at)) 
{{- end }}
{{- if eq .countOperator "EQ" }} = {{arg .countValue}}
{{- else if eq .countOperator "NE" }} != {{arg .countValue}}
{{- else if eq .countOperator "LT" }} < {{arg .countValue}}
{{- else if eq .countOperator "LTE" }} <= {{arg .countValue}}
{{- else if eq .countOperator "GT" }} > {{arg .countValue}}
{{- else if eq .countOperator "GTE" }} >= {{arg .countValue}}
{{- else if eq .countOperator "BETWEEN" }} BETWEEN {{arg .countMin}} AND {{arg .countMax}}
{{- else }} = {{arg .countValue}}
{{- end }}`, tag.SingleQuote(string(eventName)))

	return tag.NewSQLTemplate(md, sqlTemplate)
}
