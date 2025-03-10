package tag

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/samber/lo"
)

func strQuote(v string, c string) string {
	return c + strings.ReplaceAll(v, c, c+c) + c
}

func quote(v any, c string) string {
	switch val := v.(type) {
	case string:
		return strQuote(val, c)
	case fmt.Stringer:
		return strQuote(val.String(), c)
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.String {
		return strQuote(rv.String(), c)
	}
	panic(fmt.Sprintf("quote only accepts string-like types, got %T", v))
}

func SingleQuote(v any) string {
	return quote(v, "'")
}

func BacktickQuote(v any) string {
	return quote(v, "`")
}

func QuoteJoin(fn func(v any) string) func(sep string, values any) string {
	return func(sep string, values any) string {
		if strs, ok := values.([]string); ok {
			return strings.Join(lo.Map(strs, func(item string, _ int) string {
				return fn(item)
			}), sep)
		}
		rv := reflect.ValueOf(values)
		if rv.Kind() != reflect.Slice {
			panic(fmt.Sprintf("quoteJoin only accepts slice, got %T", values))
		}
		length := rv.Len()
		if length == 0 {
			return ""
		}
		quotedValues := make([]string, length)
		for i := 0; i < length; i++ {
			elemValue := rv.Index(i).Interface()
			quotedValues[i] = fn(elemValue)
		}
		return strings.Join(quotedValues, sep)
	}
}

// SQLTemplateFuncs provides common string manipulation functions for templates
var SQLTemplateFuncs = template.FuncMap{
	"toLower":           strings.ToLower,
	"toUpper":           strings.ToUpper,
	"trim":              strings.Trim,
	"trimSuffix":        strings.TrimSuffix,
	"hasPrefix":         strings.HasPrefix,
	"hasSuffix":         strings.HasSuffix,
	"replaceAll":        strings.ReplaceAll,
	"split":             strings.Split,
	"join":              strings.Join,
	"camelCase":         lo.CamelCase,
	"snakeCase":         lo.SnakeCase,
	"pascalCase":        lo.PascalCase,
	"kebabCase":         lo.KebabCase,
	"capitalize":        lo.Capitalize,
	"plural":            inflection.Plural,
	"singular":          inflection.Singular,
	"singleQuote":       SingleQuote,
	"backtickQuote":     BacktickQuote,
	"singleQuoteJoin":   QuoteJoin(SingleQuote),
	"backtickQuoteJoin": QuoteJoin(BacktickQuote),
}

// arg returns a function that can be used to collect SQL parameters.
// It records each parameter value and returns appropriate placeholder.
func arg(collector *[]any, placeholderFn func(index int) string) func(value any) string {
	return func(value any) string {
		*collector = append(*collector, value)
		return placeholderFn(len(*collector) - 1)
	}
}

// argEach returns a function that processes array parameters for SQL IN clauses.
// For arrays, it expands each element into individual parameters with placeholders.
// For non-arrays, it behaves like the regular arg function.
// Empty arrays return "NULL" as a special case.
func argEach(collector *[]any, placeholderFn func(index int) string) func(values any) string {
	return func(values any) string {
		rv := reflect.ValueOf(values)

		// If not a slice, handle as a single value
		if rv.Kind() != reflect.Slice {
			*collector = append(*collector, values)
			return placeholderFn(len(*collector) - 1)
		}

		// Empty slice handling
		length := rv.Len()
		if length == 0 {
			return "NULL" // Special representation for empty arrays
		}

		// Collect each item and build placeholders
		placeholders := make([]string, length)
		for i := 0; i < length; i++ {
			elemValue := rv.Index(i).Interface()
			*collector = append(*collector, elemValue)
			placeholders[i] = placeholderFn(len(*collector) - 1)
		}

		return strings.Join(placeholders, ", ")
	}
}
