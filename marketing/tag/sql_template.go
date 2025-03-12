package tag

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/pkg/errors"
)

// SQLTemplate is a template-based SQL builder
var (
	_ Builder    = &SQLTemplate{}
	_ SQLBuilder = &SQLTemplate{}

	// whitespaceRegex is used to replace consecutive whitespace characters (including spaces, tabs, newlines, etc.) with a single space
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// SQLArgPlaceholder defines the function type for formatting SQL parameter placeholders
type SQLArgPlaceholder func(index int) string

// SQLTemplate struct
type SQLTemplate struct {
	md             *Metadata
	sqlTpl         string
	parseOnce      sync.Once
	parsedTpl      *template.Template
	parseErr       error
	argPlaceholder SQLArgPlaceholder // Function to format arg placeholders
}

// Metadata returns the builder's metadata
func (p *SQLTemplate) Metadata(ctx context.Context) *Metadata {
	return p.md
}

// validateParams validates the input parameters against view schema
func (p *SQLTemplate) validateParams(ctx context.Context, params map[string]any) error {
	// Skip validation if no View
	if p.md.View == nil {
		return nil
	}

	return p.md.View.Validate(ctx, params)
}

// CompactSQLQuery normalizes SQL query by replacing consecutive whitespace with a single space
func CompactSQLQuery(query string) string {
	query = whitespaceRegex.ReplaceAllString(query, " ")
	return strings.TrimSpace(query)
}

// renderTemplate renders the template and collects arguments
func (p *SQLTemplate) renderTemplate(params map[string]any) (*SQL, error) {
	// Ensure template is parsed
	p.parseOnce.Do(func() {
		p.parsedTpl, p.parseErr = template.New("sql").Funcs(SQLTemplateFuncs).Funcs(template.FuncMap{
			"arg": func(value any) string {
				panic("this is a dummy function")
			},
			"argEach": func(values any) string {
				panic("this is a dummy function")
			},
		}).Parse(p.sqlTpl)
		if p.parseErr != nil {
			p.parseErr = errors.Wrap(p.parseErr, "failed to parse SQL template")
		}
	})

	if p.parseErr != nil {
		return nil, p.parseErr
	}

	// Create a new template for execution with arg functions
	// This ensures we get a fresh argCollector for each execution
	var argCollector []any

	// Clone the original template to avoid modifying it
	execTpl, err := p.parsedTpl.Clone()
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone template")
	}

	// Add arg and argEach functions to the execution template
	execTpl.Funcs(template.FuncMap{
		"arg":     arg(&argCollector, p.argPlaceholder),
		"argEach": argEach(&argCollector, p.argPlaceholder),
	})

	var buf bytes.Buffer
	err = execTpl.Execute(&buf, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute SQL template")
	}

	return &SQL{
		Query: buf.String(),
		Args:  argCollector,
	}, nil
}

// BuildSQL implements SQLBuilder.BuildSQL
func (p *SQLTemplate) BuildSQL(ctx context.Context, params map[string]any) (*SQL, error) {
	if err := p.validateParams(ctx, params); err != nil {
		return nil, err
	}
	return p.renderTemplate(params)
}

// WithArgPlaceholder sets the parameter placeholder formatter function
func (p *SQLTemplate) WithArgPlaceholder(formatter SQLArgPlaceholder) *SQLTemplate {
	if formatter == nil {
		// Use default formatter if nil provided
		formatter = func(index int) string {
			return "?"
		}
	}
	p.argPlaceholder = formatter
	return p
}

// NewSQLTemplate creates a new template-based SQL builder
func NewSQLTemplate(md *Metadata, sqlTpl string) *SQLTemplate {
	if md == nil {
		panic("metadata is required")
	}

	if sqlTpl == "" {
		panic("sqlTpl is required")
	}

	// Default to question mark placeholders
	defaultPlaceholder := func(index int) string {
		return "?"
	}

	return &SQLTemplate{
		md:             md,
		sqlTpl:         sqlTpl,
		parseOnce:      sync.Once{},
		parsedTpl:      nil, // Will be initialized on first use
		parseErr:       nil,
		argPlaceholder: defaultPlaceholder,
	}
}
