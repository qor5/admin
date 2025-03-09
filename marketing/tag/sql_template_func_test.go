package tag

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// StringEnum is used for testing string enum type handling
type StringEnum string

const (
	EnumValueOne   StringEnum = "ONE"
	EnumValueTwo   StringEnum = "TWO"
	EnumValueThree StringEnum = "THREE"
)

type stringer struct {
	value string
}

func (s stringer) String() string {
	return s.value
}

// TestSingleQuote verifies the singleQuote function properly escapes single quotes in strings
// to ensure SQL strings are correctly formatted for different database dialects
func TestSingleQuote(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  string
		shouldErr bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
		{
			name:     "normal string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with embedded single quotes",
			input:    "it's",
			expected: "'it''s'",
		},
		{
			name:     "stringer implementation",
			input:    stringer{"world"},
			expected: "'world'",
		},
		{
			name:     "string enum",
			input:    EnumValueOne,
			expected: "'ONE'",
		},
		{
			name:      "non-string type",
			input:     42,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				assert.Panics(t, func() {
					SingleQuote(tt.input)
				})
			} else {
				result := SingleQuote(tt.input)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBacktickQuote(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  string
		shouldErr bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "``",
		},
		{
			name:     "normal string",
			input:    "column_name",
			expected: "`column_name`",
		},
		{
			name:     "string with embedded backticks",
			input:    "column`name",
			expected: "`column``name`",
		},
		{
			name:     "stringer implementation",
			input:    stringer{"table_name"},
			expected: "`table_name`",
		},
		{
			name:     "string enum",
			input:    EnumValueTwo,
			expected: "`TWO`",
		},
		{
			name:      "non-string type",
			input:     42,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				assert.Panics(t, func() {
					BacktickQuote(tt.input)
				})
			} else {
				result := BacktickQuote(tt.input)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestQuoteJoin(t *testing.T) {
	singleQuoteJoin := QuoteJoin(SingleQuote)
	backtickQuoteJoin := QuoteJoin(BacktickQuote)

	tests := []struct {
		name        string
		fn          func(string, any) string
		separator   string
		input       any
		expected    string
		shouldPanic bool
	}{
		{
			name:      "single quote join with string slice",
			fn:        singleQuoteJoin,
			separator: ", ",
			input:     []string{"a", "b", "c"},
			expected:  "'a', 'b', 'c'",
		},
		{
			name:      "single quote join with string slice and embedded quotes",
			fn:        singleQuoteJoin,
			separator: ", ",
			input:     []string{"it's", "they're"},
			expected:  "'it''s', 'they''re'",
		},
		{
			name:      "backtick quote join with string slice",
			fn:        backtickQuoteJoin,
			separator: ", ",
			input:     []string{"col1", "col2", "col3"},
			expected:  "`col1`, `col2`, `col3`",
		},
		{
			name:      "single quote join with empty string slice",
			fn:        singleQuoteJoin,
			separator: ", ",
			input:     []string{},
			expected:  "",
		},
		{
			name:      "single quote join with empty interface slice",
			fn:        singleQuoteJoin,
			separator: ", ",
			input:     []interface{}{},
			expected:  "",
		},
		{
			name:      "single quote join with enum slice",
			fn:        singleQuoteJoin,
			separator: " | ",
			input:     []StringEnum{EnumValueOne, EnumValueTwo, EnumValueThree},
			expected:  "'ONE' | 'TWO' | 'THREE'",
		},
		{
			name:      "backtick quote join with enum slice",
			fn:        backtickQuoteJoin,
			separator: ", ",
			input:     []StringEnum{EnumValueOne, EnumValueTwo, EnumValueThree},
			expected:  "`ONE`, `TWO`, `THREE`",
		},
		{
			name:      "single quote join with interface slice",
			fn:        singleQuoteJoin,
			separator: " AND ",
			input:     []interface{}{"name", "age", "address"},
			expected:  "'name' AND 'age' AND 'address'",
		},
		{
			name:        "single quote join with non-slice",
			fn:          singleQuoteJoin,
			separator:   ", ",
			input:       "not a slice",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					tt.fn(tt.separator, tt.input)
				})
			} else {
				result := tt.fn(tt.separator, tt.input)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Bug test for the panic message in quoteJoin
func TestQuoteJoinPanicMessage(t *testing.T) {
	singleQuoteJoin := QuoteJoin(SingleQuote)

	defer func() {
		if r := recover(); r != nil {
			panicMsg, ok := r.(string)
			assert.True(t, ok, "Panic value should be a string")
			assert.Contains(t, panicMsg, "quoteJoin only accepts slice")
		}
	}()

	singleQuoteJoin(", ", "not a slice")
	t.Fail() // Should not reach here
}

func TestValuesTemplate(t *testing.T) {
	tpl, err := template.New("sql").Funcs(SQLTemplateFuncs).Parse(`SELECT uid FROM users WHERE gender IN ({{ .values | singleQuoteJoin ", " }})`)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]any{"values": []any{"male", "female"}})
	require.NoError(t, err)
	assert.Equal(t, "SELECT uid FROM users WHERE gender IN ('male', 'female')", buf.String())
}
