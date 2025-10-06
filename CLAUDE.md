# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **RFC 7386 JSON Merge Patch** implementation for Go. The library provides type-safe, generic-based JSON merge patch operations with compliance to RFC 7386 specification.

**Core Purpose**: Implement RFC 7386 JSON Merge Patch algorithm with type safety, minimal API, and high performance.

## Development Commands

### Testing
```bash
# Run all tests with race detection
make test

# Run tests with verbose output
make test-verbose

# Generate coverage report (creates coverage.html)
make test-coverage

# Run benchmarks
make bench
```

### Linting and Code Quality
```bash
# Run all linters (golangci-lint + mod tidy check)
make lint

# Run golangci-lint only
make golangci-lint

# Format code
make fmt

# Run go vet
make vet
```

### Complete Verification
```bash
# Run full verification pipeline: deps, fmt, vet, lint, test
make verify
```

### Dependency Management
```bash
# Download and tidy dependencies
make deps

# Clean build artifacts and caches
make clean
```

### Running Single Tests
```bash
# Run specific test
go test -run TestMerge -v

# Run specific test with race detection
go test -race -run TestMergeBasic -v

# Run benchmarks for specific function
go test -bench=BenchmarkMerge -run=^$
```

## Architecture Overview

### Core Implementation Philosophy

**RFC 7386 First**: Every implementation decision prioritizes RFC 7386 compliance. The algorithm in `mergePatch()` (jsonmerge.go:116-143) directly implements RFC 7386 Section 2.

**Type Safety Through Generics**: Uses Go 1.25+ generics with `Document` interface constraint to provide compile-time type safety while supporting multiple document types (structs, maps, JSON bytes, JSON strings).

**Immutable by Default**: Uses `github.com/kaptinlin/deepclone` for deep cloning to prevent side effects. Optional `WithMutate(true)` enables in-place modification for performance-critical scenarios (3x faster).

### Key Architectural Components

#### 1. Document Type System (types.go)

```go
type Document interface {
    ~[]byte | ~string | map[string]any | any
}
```

**Design Purpose**: Type constraint that allows generic functions to accept multiple document formats while maintaining type safety. The `any` case handles struct types through JSON marshaling/unmarshaling.

**Supported Types**:
- **Structs**: Full support with JSON tags (`json:"field,omitempty"`, `json:"-"`)
- **map[string]any**: Native format, most efficient
- **[]byte**: JSON bytes with automatic parsing
- **string**: JSON strings or raw string values
- **Primitives**: Direct support for bool, int variants, float variants

#### 2. Core Algorithm (jsonmerge.go:116-143)

```go
func mergePatch(target, patch interface{}) interface{}
```

**RFC 7386 Implementation**:
1. If patch is not an object → return patch (complete replacement)
2. If target is not an object → create empty object
3. For each field in patch:
   - If value is `null` → delete field from target
   - Otherwise → recursively merge

**Critical Behaviors**:
- Objects merge recursively
- Arrays replace entirely (no element-wise merging)
- `null` values delete fields
- Non-objects replace target values completely

#### 3. Type Conversion System (jsonmerge.go:212-296)

**convertToInterface[T Document]**: Converts any Document type to `interface{}` for processing
- Direct passthrough for `map[string]any`
- JSON unmarshal for `[]byte` and valid JSON strings
- Marshal→Unmarshal cycle for structs

**convertFromInterface[T Document]**: Converts `interface{}` back to original type
- Preserves type through generic parameter
- Efficient for maps (direct cast)
- JSON round-trip for structs and byte/string types

**Performance Optimization**: Minimizes JSON marshal/unmarshal cycles. Map operations have zero conversion overhead.

#### 4. Functional Options Pattern (types.go:20-34)

```go
type Option func(*Options)
func WithMutate(mutate bool) Option
```

**Design Rationale**: Extensible configuration pattern allowing future options without breaking API. Currently only `Mutate` option exists, but pattern supports adding more configuration options.

**Usage Impact**:
- Default (immutable): Safe for concurrent use, predictable behavior
- `WithMutate(true)`: 3x performance improvement, use only when thread safety guaranteed

### API Design

**Minimal Surface**: Only 3 public functions
- `Merge[T Document](target, patch T, ...Option) (*Result[T], error)` - Apply merge patch
- `Generate[T Document](source, target T) (T, error)` - Create patch between documents
- `Valid[T Document](patch T) bool` - Validate merge patch

**Generic Result Type**: `Result[T]` preserves type through operations
```go
result, err := jsonmerge.Merge(userStruct, patchStruct)
// result.Doc is automatically of the same struct type
```

### Error Handling Architecture

**Go Idiomatic Error Handling** - Follows Go best practices:

**Sentinel Errors** (jsonmerge.go:30-44):
- `ErrInvalidJSON` - Invalid JSON input
- `ErrMarshal` - JSON marshaling failed
- `ErrUnmarshal` - JSON unmarshaling failed
- `ErrConversion` - Type conversion between document types failed

**Error Wrapping Pattern**:
```go
// Use %w to preserve error chain for errors.Is() checking
fmt.Errorf("%w: %w", SentinelError, originalError)
fmt.Errorf("convert target: %w", err)  // Add context when needed
```

**Best Practices**:
- ✅ Define minimal sentinel errors for error type checking with `errors.Is()`
- ✅ Use `%w` verb to wrap errors and preserve error chain
- ✅ Error messages are lowercase, concise, and descriptive
- ✅ No redundant error wrapping - keep it simple and direct
- ❌ Don't create helper functions like `wrapError()` - violates Go idioms
- ❌ Don't define unused sentinel errors - keep only what's necessary

## Development Workflow

### Code Quality Standards

**golangci-lint Configuration**:
- Version: 2.4.0 (managed via `.golangci.version`)
- Timeout: 5m for Go 1.24+
- Enabled linters: errcheck, govet, staticcheck, gosec, exhaustive, and 20+ more
- Examples directory excluded from most linters

**Testing Requirements**:
- All tests must pass with `-race` flag
- RFC 7386 Appendix A test cases must all pass
- Edge cases: Unicode, deep nesting, large arrays, mixed types
- Benchmark tests verify performance characteristics

**Documentation Standards**:
- All public types and functions require complete GoDoc comments
- Comments reference RFC 7386 sections where applicable
- Examples demonstrate type-safe usage patterns

### Important Implementation Rules

**From .cursor/rules.md**:

1. **RFC 7386 Compliance is Paramount**: Never deviate from RFC 7386 algorithm specification
2. **Type Safety First**: Leverage Go 1.25+ generics comprehensively
3. **Immutable by Default**: Use deep cloning unless `WithMutate(true)` specified
4. **Minimal Dependencies**: Only `github.com/kaptinlin/deepclone` and `github.com/go-json-experiment/json`
5. **Go Idiomatic Error Handling**:
   - Define minimal sentinel errors for `errors.Is()` checking
   - Use `%w` to wrap errors and preserve error chain
   - Keep error messages lowercase, concise, descriptive
   - Avoid helper functions like `wrapError()` - use `fmt.Errorf` directly
   - No unused sentinel errors - define only what's necessary

**Critical RFC 7386 Rules**:
1. Objects are merged recursively
2. `null` values delete fields
3. Non-objects completely replace target
4. Arrays replace entirely (never element-wise merge)
5. Primitives replace corresponding target values

## Performance Optimization Guidelines

**Core Principles**:
- ✅ Benchmark-driven optimization - verify with `make bench` before and after changes
- ✅ Hot path optimization - keep frequently-called code inline, avoid extra function calls
- ✅ Simplicity over abstraction - Go runtime already optimizes common operations
- ❌ Don't extract helper functions in hot paths - causes 50%+ performance regression
- ❌ Don't blindly trust "optimizations" - measure actual performance impact

**Proven Performance Tips**:
- Use `WithMutate(true)` for performance-critical scenarios (3x faster)
- `map[string]any` has zero conversion overhead (most efficient)
- JSON bytes/strings more efficient than structs for large data
- Keep critical functions inline - function call overhead matters in hot paths

**Failed Optimization Attempts** (DO NOT REPEAT):
- ❌ Using `bytes.Equal` instead of `string` comparison → 34% slower (Go runtime optimizes string comparison)
- ❌ Extracting helper functions from `convertToInterface` → 53% slower (function call overhead in hot path)
- ❌ Adding `a == b` fast path in `deepEqual` → panic on uncomparable types (slices, maps)
- ❌ Using `reflect.DeepEqual` → panic on uncomparable types

**Optimization Workflow**:
1. Run baseline: `make bench > baseline.txt`
2. Make ONE change at a time
3. Run benchmarks: `make bench > optimized.txt`
4. Compare results - if performance regresses, IMMEDIATELY revert
5. Only keep changes that maintain or improve performance

## Dependencies

**Required External Dependencies**:
- `github.com/go-json-experiment/json` - JSON operations (better performance than stdlib)
- `github.com/kaptinlin/deepclone` - Deep cloning for immutable operations

**Test Dependencies**:
- `github.com/stretchr/testify` - Test assertions

**Go Version**: 1.25+ (uses generics, type inference, modern stdlib)

## Examples Directory

The `examples/` directory contains comprehensive usage examples:
- `struct-merge/` - Type-safe struct operations with JSON tags
- `map-merge/` - Dynamic map[string]any merging
- `json-bytes-merge/` - Raw JSON byte processing
- `json-string-merge/` - JSON string handling

Each example demonstrates specific use cases and best practices.

## Key Files and Their Purpose

- **jsonmerge.go**: Core implementation (algorithm, type conversion, public API)
- **types.go**: Type definitions (Document constraint, Result wrapper, Options)
- **jsonmerge_test.go**: Comprehensive tests including all RFC 7386 test cases
- **benchmark_test.go**: Performance benchmarks for all core operations
- **Makefile**: Build automation and quality checks
- **.golangci.yml**: Linter configuration with 30+ enabled linters
- **.cursor/rules.md**: Comprehensive development rules for code generation

## Critical Design Decisions

**Why go-json-experiment instead of stdlib**: Better performance and modern features while maintaining compatibility.

**Why deepclone instead of custom implementation**: Reliable, tested deep cloning that handles edge cases (circular references, complex types).

**Why functional options pattern**: Future-proof extensibility without breaking API changes.

**Why generic constraints**: Compile-time type safety while supporting multiple document types, preventing runtime type errors.

**Why immutable by default**: Predictable behavior, thread safety, easier debugging. Performance mode available when needed.

**Why simple error handling**: Following Go idioms - minimal sentinel errors, direct `fmt.Errorf` wrapping, no helper functions. Simplicity and clarity over abstraction.
