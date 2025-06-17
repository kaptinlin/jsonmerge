package jsonmerge

// Document represents the supported document types for JSON Merge Patch operations.
// This constraint allows for type-safe operations across different JSON representations.
type Document interface {
	~[]byte | ~string | map[string]any | any
}

// Result wraps the merged document with type safety.
// The generic parameter T preserves the original document type through the merge operation.
type Result[T Document] struct {
	Doc T // The merged document of the same type as the input
}

// Options contains configuration for merge operations.
type Options struct {
	Mutate bool // If true, modifies the target document in place for performance
}

// Option is a functional option type for configuring merge operations.
type Option func(*Options)

// WithMutate configures whether to modify the target document in place.
// By default, merge operations are immutable and create a new document.
// Setting mutate to true can improve performance but may affect thread safety.
//
// Example:
//
//	result, err := Merge(target, patch, WithMutate(true))
func WithMutate(mutate bool) Option {
	return func(opts *Options) {
		opts.Mutate = mutate
	}
}
