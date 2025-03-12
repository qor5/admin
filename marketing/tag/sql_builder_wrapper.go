package tag

import (
	"context"
)

// BuildSQLFunc is a function type that matches the signature of SQLBuilder.BuildSQL
type BuildSQLFunc func(ctx context.Context, params map[string]any) (*SQL, error)

// MetadataFunc is a function type that matches the signature of Builder.Metadata
type MetadataFunc func(ctx context.Context) *Metadata

// WrapSQLBuilder creates a new SQLBuilder that wraps the original SQLBuilder
// and applies hooks to its BuildSQL method
func WrapSQLBuilder(original SQLBuilder, hooks ...Hook[BuildSQLFunc]) *SQLBuilderWrapper {
	return &SQLBuilderWrapper{
		original:     original,
		buildSQLHook: ChainHook(hooks...),
	}
}

// SQLBuilderWrapper implements the SQLBuilder interface by wrapping an original SQLBuilder
// and providing methods to add hooks to its methods
type SQLBuilderWrapper struct {
	original     SQLBuilder
	buildSQLHook Hook[BuildSQLFunc]
	metadataHook Hook[MetadataFunc]
}

// WithBuildSQLHook adds hooks to the BuildSQL method chain
// Returns the wrapper for method chaining
func (w *SQLBuilderWrapper) WithBuildSQLHook(hooks ...Hook[BuildSQLFunc]) *SQLBuilderWrapper {
	w.buildSQLHook = ChainHookWith(w.buildSQLHook, hooks...)
	return w
}

// WithMetadataHook adds hooks to the Metadata method chain
// Returns the wrapper for method chaining
func (w *SQLBuilderWrapper) WithMetadataHook(hooks ...Hook[MetadataFunc]) *SQLBuilderWrapper {
	w.metadataHook = ChainHookWith(w.metadataHook, hooks...)
	return w
}

// BuildSQL implements the SQLBuilder interface
func (w *SQLBuilderWrapper) BuildSQL(ctx context.Context, params map[string]any) (*SQL, error) {
	buildFn := w.original.BuildSQL
	if w.buildSQLHook != nil {
		buildFn = w.buildSQLHook(buildFn)
	}
	return buildFn(ctx, params)
}

// Metadata implements the Builder interface by delegating to the original Builder
func (w *SQLBuilderWrapper) Metadata(ctx context.Context) *Metadata {
	metadataFn := w.original.Metadata
	if w.metadataHook != nil {
		metadataFn = w.metadataHook(metadataFn)
	}
	return metadataFn(ctx)
}
