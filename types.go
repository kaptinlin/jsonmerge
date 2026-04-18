package jsonmerge

// Document is the set of types supported by Merge, Generate, and Valid.
type Document interface {
	~[]byte | ~string | map[string]any | any
}

// Result holds a merged document.
type Result[T Document] struct {
	// Doc is the merged document.
	Doc T
}

// Options configures merge behavior.
type Options struct {
	// Mutate allows in-place updates.
	Mutate bool
}

// Option configures Options.
type Option func(*Options)

// WithMutate allows Merge to update map targets in place.
// The default is false.
func WithMutate(mutate bool) Option {
	return func(opts *Options) {
		opts.Mutate = mutate
	}
}
