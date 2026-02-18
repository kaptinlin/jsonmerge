# JSON Merge Patch

RFC 7386 JSON Merge Patch implementation for Go with type-safe generics, minimal API, and high performance.

## Commands

```bash
# Testing
task test                # Run all tests with race detection
task test-verbose        # Run tests with verbose output
task test-coverage       # Generate coverage report (coverage.html)
task bench              # Run benchmarks

# Code Quality
task lint               # Run golangci-lint + mod tidy check
task fmt                # Format code
task vet                # Run go vet

# Complete Verification
task verify             # Run deps, fmt, vet, lint, test

# Dependency Management
task deps               # Download and tidy dependencies
task clean              # Clean build artifacts and caches
```

## Architecture

### Core Components

**jsonmerge.go** - Core implementation
- `Merge[T Document](target, patch T, ...Option) (*Result[T], error)` - Apply RFC 7386 merge patch
- `Generate[T Document](source, target T) (T, error)` - Create patch between documents
- `Valid[T Document](patch T) bool` - Validate merge patch
- `applyPatch(target, patch any) any` - RFC 7386 algorithm (lines 132-163)
- Type conversion system: `convertToInterface`, `convertFromInterface`

**types.go** - Type system and options
- `Document` interface - Generic constraint for supported types ([]byte, string, map[string]any, structs)
- `Result[T]` - Type-safe wrapper preserving document type
- `Options` and `Option` - Functional options pattern
- `Error` - Sentinel error type

### Key Types and Interfaces

```go
// Document constraint supports multiple JSON representations
type Document interface {
    ~[]byte | ~string | map[string]any | any
}

// Result preserves type through merge operations
type Result[T Document] struct {
    Doc T
}

// Sentinel errors for error checking with errors.Is
const (
    ErrMarshal    Error = "marshal failed"
    ErrUnmarshal  Error = "unmarshal failed"
    ErrConversion Error = "type conversion failed"
)
```

## Design Philosophy

- **RFC 7386 First** - Every implementation decision prioritizes RFC 7386 compliance. The `applyPatch()` algorithm directly implements RFC 7386 Section 2.
- **Type Safety Through Generics** - Go 1.26+ generics with Document constraint provide compile-time type safety while supporting structs, maps, JSON bytes, and JSON strings.
- **Immutable by Default** - Uses `github.com/kaptinlin/deepclone` to prevent side effects. Optional `WithMutate(true)` enables in-place modification for 3x performance improvement.
- **Minimal API Surface** - Only 3 public functions (Merge, Generate, Valid) keep the API simple and focused.
- **Zero-Copy Optimization** - `map[string]any` operations have zero conversion overhead; type conversion minimizes JSON marshal/unmarshal cycles.
- **Benchmark-Driven Performance** - All optimizations verified with benchmarks; failed optimization attempts documented to prevent regression.

## Coding Rules

### Must Follow

- Go 1.26+
- RFC 7386 compliance is non-negotiable:
  - Objects merge recursively
  - `null` values delete fields
  - Non-objects completely replace target
  - Arrays replace entirely (never element-wise merge)
  - Primitives replace corresponding target values
- Immutable operations by default (use deepclone unless WithMutate specified)
- Go idiomatic error handling:
  - Define minimal sentinel errors for `errors.Is()` checking
  - Use `%w` to wrap errors and preserve error chain
  - Keep error messages lowercase, concise, descriptive
  - Use `fmt.Errorf` directly, no helper functions like `wrapError()`
- Keep hot path code inline - function call overhead matters in performance-critical paths

### Forbidden

- No deviation from RFC 7386 algorithm specification
- No helper function extraction in hot paths (causes 50%+ performance regression)
- No unused sentinel errors (define only what's necessary)
- No generic fluff in documentation ("write clean code", "follow best practices")
- Failed optimization patterns (documented to prevent repetition):
  - Using `bytes.Equal` instead of string comparison (34% slower)
  - Extracting helpers from `convertToInterface` (53% slower)
  - Adding `a == b` fast path in `deepEqual` (panics on uncomparable types)
  - Using `reflect.DeepEqual` (panics on uncomparable types)

## Testing

All tests use `-race` flag. Test suite includes:
- RFC 7386 Appendix A compliance tests (all must pass)
- Edge cases: Unicode, deep nesting, large arrays, mixed types
- Type safety tests for structs, maps, JSON bytes, JSON strings
- Benchmark tests for performance verification

```bash
# Run specific test
go test -race -run TestMerge -v

# Run benchmarks for specific function
go test -bench=BenchmarkMerge -run=^$
```

## Dependencies

**Production:**
- `github.com/go-json-experiment/json` - JSON operations (better performance than stdlib)
- `github.com/kaptinlin/deepclone` - Deep cloning for immutable operations

**Test:**
- `github.com/stretchr/testify` - Test assertions

## Error Handling

Sentinel errors defined as `const` with custom `Error` type for immutability:

```go
const (
    ErrMarshal    Error = "marshal failed"    // JSON marshaling failed
    ErrUnmarshal  Error = "unmarshal failed"  // JSON unmarshaling failed
    ErrConversion Error = "type conversion failed" // Type conversion failed
)
```

Use `errors.Is()` for error checking:

```go
result, err := jsonmerge.Merge(target, patch)
if err != nil {
    if errors.Is(err, jsonmerge.ErrMarshal) {
        // Handle marshaling error
    }
}
```

Error wrapping pattern: Use `%w` to preserve error chain, add context when needed.

## Performance

### Optimization Principles

- Benchmark-driven: verify with `make bench` before and after changes
- Hot path optimization: keep frequently-called code inline
- Simplicity over abstraction: Go runtime optimizes common operations
- Make ONE change at a time, measure impact, revert if performance regresses

### Performance Tips

- Use `WithMutate(true)` for performance-critical scenarios (3x faster)
- `map[string]any` has zero conversion overhead (most efficient)
- JSON bytes/strings more efficient than structs for large data
- Deep nesting has minimal performance impact

### Benchmark Results (Apple M3)

```
BenchmarkMerge-8                  952150     1357 ns/op    1273 B/op   17 allocs/op
BenchmarkMergeWithMutate-8       2400202      466 ns/op     345 B/op    4 allocs/op
BenchmarkMergeStructs-8           154684     8722 ns/op    3993 B/op   78 allocs/op
BenchmarkMergeJSONStrings-8       246922     5458 ns/op    3743 B/op   77 allocs/op
BenchmarkMergeJSONBytes-8         206040     5934 ns/op    3416 B/op   74 allocs/op
```

## Linting

golangci-lint v2.9.0 (managed via `.golangci.version`)
- Timeout: 5m for Go 1.24+
- 30+ enabled linters including errcheck, govet, staticcheck, gosec, exhaustive
- Examples directory excluded from most linters

## Agent Skills

Package-local skills in `.agents/skills/`:

- **agent-md-creating** - Generate CLAUDE.md for Go projects
- **code-simplifying** - Refine recently written Go code for clarity and consistency
- **committing** - Create conventional commits following project conventions
- **dependency-selecting** - Select Go dependencies from kaptinlin/agentable ecosystem
- **go-best-practices** - Google Go coding best practices and style guide
- **linting** - Set up and run golangci-lint v2
- **modernizing** - Go 1.20-1.26 modernization guide
- **ralphy-initializing** - Initialize Ralphy AI coding loop configuration
- **ralphy-todo-creating** - Create Ralphy TODO.yaml task files
- **readme-creating** - Generate README.md for Go libraries
- **releasing** - Guide release process with semantic versioning
- **testing** - Write Go tests with testify and Go 1.25+ features
