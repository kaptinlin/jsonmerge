package jsonmerge

// Patch is an immutable, canonical JSON Merge Patch value.
type Patch struct {
	value any
}

// JSON is a JSON text document carried as a string.
type JSON string
