// types.go defines the core types and functional options for JSON Merge Patch operations.
package jsonmerge

// Document represents the supported document types for JSON Merge Patch operations.
// This constraint allows type-safe operations across different JSON representations.
//
// Supported types and their behavior:
//   - []byte: must contain valid JSON; returns ErrUnmarshal if invalid
//   - string: attempts JSON parsing; if invalid JSON, treated as raw string value
//   - map[string]any: native format with zero conversion overhead (most efficient)
//   - struct types (via any): converted through JSON marshal/unmarshal cycle;
//     respects json struct tags (json:"name,omitempty", json:"-")
//   - primitive types (bool, int*, uint*, float*): passed through directly
type Document interface {
	~[]byte | ~string | map[string]any | any
}

// Result wraps the merged document with type safety.
// The generic parameter T preserves the original document type through the merge operation.
type Result[T Document] struct {
	// Doc is the merged document, preserving the same type as the input.
	Doc T
}

// Options contains configuration for merge operations.
type Options struct {
	// Mutate controls whether to modify the target document in place for performance.
	Mutate bool
}

// Option is a functional option type for configuring merge operations.
type Option func(*Options)

// WithMutate configures whether to modify the target document in place.
// By default, merge operations are immutable and create a new document.
// Setting mutate to true improves performance but affects thread safety.
//
// Default: false
//
// Example:
//
//	result, err := Merge(target, patch, WithMutate(true))
func WithMutate(mutate bool) Option {
	return func(opts *Options) {
		opts.Mutate = mutate
	}
}
