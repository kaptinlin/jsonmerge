# JSON Merge Patch Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue.svg)](https://golang.org/)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/jsonmerge.svg)](https://pkg.go.dev/github.com/kaptinlin/jsonmerge)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A **type-safe**, **RFC 7386 compliant** JSON Merge Patch implementation for Go with **generic support**.

## ✨ Features

- 🔥 **RFC 7386 Compliant** - Passes all standard test cases from [RFC 7386 Appendix A](https://datatracker.ietf.org/doc/html/rfc7386#appendix-A)
- 🚀 **Type-Safe Generics** - Compile-time type safety with Go 1.18+ generics
- 🎯 **Minimal API** - Only 3 core functions: `Merge`, `Generate`, `Valid`
- ⚡ **Performance Optimized** - Zero-copy optimization with optional in-place mutation
- 🛡️ **Production Ready** - Tested, immutable by default, thread-safe
- 📦 **Multiple Document Types** - Supports structs, maps, JSON bytes, and JSON strings
- 🌐 **Unicode Support** - Full Unicode and international character support

## 🚀 Quick Start

### Installation

```bash
go get github.com/kaptinlin/jsonmerge
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/kaptinlin/jsonmerge"
)

func main() {
    // Original document
    target := map[string]any{
        "title": "Hello World",
        "author": map[string]any{
            "name":  "John Doe",
            "email": "john@example.com",
        },
        "tags": []string{"example", "demo"},
    }

    // JSON Merge Patch (RFC 7386)
    patch := map[string]any{
        "title": "Hello Go",           // Replace
        "author": map[string]any{
            "email": nil,               // Delete email field
        },
        "tags": []string{"go", "json"}, // Replace entire array
        "publishDate": "2024-01-01",    // Add new field
    }

    // Apply merge patch
    result, err := jsonmerge.Merge(target, patch)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%+v\n", result.Doc)
    // Output: map[title:Hello Go author:map[name:John Doe] tags:[go json] publishDate:2024-01-01]
}
```

## 📋 API Reference

### Core Functions

```go
// Merge applies a JSON Merge Patch to a target document
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error)

// Generate creates a merge patch between two documents  
func Generate[T Document](source, target T) (T, error)

// Valid checks if a patch is a valid JSON Merge Patch
func Valid[T Document](patch T) bool
```

### Supported Document Types

```go
type Document interface {
    ~[]byte | ~string | map[string]any | any
}
```

### Options

```go
// WithMutate enables in-place modification for performance
func WithMutate(mutate bool) Option
```

## 🎯 Use Cases & Examples

### 1. Type-Safe Struct Merging

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email,omitempty"`
    Age   int    `json:"age"`
}

user := User{Name: "John", Email: "john@example.com", Age: 30}
patch := User{Name: "Jane"} // Only update name

result, err := jsonmerge.Merge(user, patch)
// result.Doc is automatically of type User
// Name: "Jane", Email: "john@example.com", Age: 0 (set by patch)
```

### 2. Advanced Struct Features

```go
type Config struct {
    Name        string  `json:"name"`
    Port        *int    `json:"port,omitempty"`       // Pointer fields
    Debug       *bool   `json:"debug,omitempty"`      // Nil = omitted
    Description string  `json:"desc,omitempty"`       // omitempty support
    Internal    string  `json:"-"`                    // Ignored fields
    Price       float64 `json:"price,string"`         // Custom JSON tags
}

// Embedded structs are also supported
type Person struct {
    Name string `json:"name"`
    Address struct {
        Street string `json:"street"`
        City   string `json:"city"`
    } `json:"address"`
    Age int `json:"age"`
}
```

### 3. Dynamic Map Merging

```go
config := map[string]any{
    "database": map[string]any{
        "host": "localhost",
        "port": 5432,
    },
    "debug": true,
}

update := map[string]any{
    "database": map[string]any{
        "host": "prod-server",  // Update
        "ssl":  true,           // Add
    },
    "debug": nil,               // Delete
}

result, _ := jsonmerge.Merge(config, update)
```

### 4. JSON String/Bytes Processing

```go
// JSON strings
target := `{"name":"John","age":30}`
patch := `{"name":"Jane","email":"jane@example.com"}`
result, _ := jsonmerge.Merge(target, patch)

// JSON bytes  
targetBytes := []byte(`{"name":"John","age":30}`)
patchBytes := []byte(`{"name":"Jane","email":"jane@example.com"}`)
result, _ := jsonmerge.Merge(targetBytes, patchBytes)
```

### 5. Unicode and International Support

```go
target := `{"name":"José","city":"São Paulo"}`
patch := `{"name":"María","country":"España"}`
result, _ := jsonmerge.Merge(target, patch)
// Full Unicode support for international characters
```

### 6. Generate Patches

```go
original := map[string]any{"name": "John", "age": 30, "city": "NYC"}
updated := map[string]any{"name": "Jane", "age": 30, "country": "USA"}

patch, _ := jsonmerge.Generate(original, updated)
// patch: map[name:Jane country:USA city:<nil>]
```

### 7. Performance Optimization

```go
// Default: Immutable (safe for concurrent use)
result, _ := jsonmerge.Merge(doc, patch)

// High-performance: In-place mutation (use with caution)
result, _ := jsonmerge.Merge(doc, patch, jsonmerge.WithMutate(true))
```

## 🔄 RFC 7386 vs JSON Patch (RFC 6902)

| Feature | JSON Merge Patch (RFC 7386) | JSON Patch (RFC 6902) |
|---------|------------------------------|------------------------|
| **Complexity** | Simple and intuitive | Operation-based commands |
| **Learning Curve** | Low | Higher |
| **Array Operations** | Complete replacement | Precise element operations |
| **Delete Operations** | `null` values | Explicit `remove` operations |
| **Use Cases** | Form updates, config changes | Precise changes, detailed operations |

### Example Comparison

Updating: `{"name": "John", "age": 30}` → `{"name": "Jane", "age": 30, "email": "jane@example.com"}`

**JSON Merge Patch (This Library):**
```json
{
  "name": "Jane",
  "email": "jane@example.com"
}
```

**JSON Patch:**
```json
[
  {"op": "replace", "path": "/name", "value": "Jane"},
  {"op": "add", "path": "/email", "value": "jane@example.com"}
]
```

**Conclusion**: JSON Merge Patch is simpler for common scenarios.

**Related Tools**: For JSON Patch (RFC 6902) operations in Go, see [jsonpatch](https://github.com/kaptinlin/jsonpatch).

## 📊 Performance

### Benchmarks

```
goos: darwin
goarch: arm64
pkg: github.com/kaptinlin/jsonmerge
cpu: Apple M3

BenchmarkMerge-8                 1290456      920 ns/op    1273 B/op   17 allocs/op
BenchmarkMergeWithMutate-8       3509066      355 ns/op     345 B/op    4 allocs/op  
BenchmarkGenerate-8              1436236      835 ns/op     784 B/op   16 allocs/op
BenchmarkMergeStructs-8           216622     5730 ns/op    4377 B/op   81 allocs/op
BenchmarkMergeJSONStrings-8       293143     4225 ns/op    3872 B/op   78 allocs/op
BenchmarkMergeJSONBytes-8         304899     3910 ns/op    3543 B/op   75 allocs/op
BenchmarkMergeDeepNesting-8      1254697      956 ns/op    2025 B/op   14 allocs/op
BenchmarkMergeLargeArrays-8         2745   433348 ns/op  672770 B/op 10026 allocs/op
```

### Performance Tips

- Use `WithMutate(true)` for performance-critical scenarios (up to 3x faster)
- JSON bytes/strings are more efficient than struct marshaling for large data
- Immutable mode is safe for concurrent use
- Deep nesting has minimal performance impact

## 🧪 Testing & Quality

- **RFC 7386 Compliant** - Passes all [RFC 7386 Appendix A](https://datatracker.ietf.org/doc/html/rfc7386#appendix-A) test cases
- **Edge Cases Covered** - Unicode, large numbers, deep nesting, mixed types
- **Go Type System Support** - Pointer fields, embedded structs, custom JSON tags

### Run Tests

```bash
# Run all tests
go test -v

# Run benchmarks
go test -bench=. -run=^$
```

## 📚 Examples

Explore comprehensive examples in the [`examples/`](examples/) directory:

- **[struct-merge](examples/struct-merge/)** - Type-safe struct operations
- **[map-merge](examples/map-merge/)** - Dynamic data with maps  
- **[json-bytes-merge](examples/json-bytes-merge/)** - Raw JSON byte processing
- **[json-string-merge](examples/json-string-merge/)** - JSON string handling

## 🔧 Advanced Usage

### Custom Types with Generics

```go
type Config struct {
    Database DatabaseConfig `json:"database"`
    Server   ServerConfig   `json:"server"`
}

config := Config{...}
patch := Config{...}

// Type-safe merge with custom struct
result, err := jsonmerge.Merge(config, patch)
// result.Doc is of type Config
```

### Error Handling

```go
result, err := jsonmerge.Merge(target, patch)
if err != nil {
    switch {
    case errors.Is(err, jsonmerge.ErrInvalidJSON):
        // Handle invalid JSON
    case errors.Is(err, jsonmerge.ErrUnsupportedType):
        // Handle unsupported type
    case errors.Is(err, jsonmerge.ErrConversionFailed):
        // Handle type conversion errors
    default:
        // Handle other errors
    }
}
```

### Validation

```go
patches := []any{
    map[string]any{"name": "test"},  // Valid
    `{"key": "value"}`,              // Valid JSON string
    []byte(`{"data": true}`),        // Valid JSON bytes
    "simple string",                 // Valid (treated as raw string)
    42,                              // Valid (primitive value)
    nil,                             // Valid (null patch)
}

for _, patch := range patches {
    if jsonmerge.Valid(patch) {
        fmt.Println("Valid patch")
    }
}
```

### Concurrent Usage

```go
// Immutable operations are thread-safe
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        patch := map[string]any{"id": id, "processed": true}
        result, _ := jsonmerge.Merge(originalDoc, patch)
        // Safe to use result.Doc concurrently
    }(i)
}
wg.Wait()
```

## 🛡️ Best Practices

1. **Use immutable mode by default** for thread safety
2. **Enable mutation only when performance is critical** and thread safety is guaranteed
3. **Prefer type-safe struct operations** when possible for compile-time safety
4. **Validate patches** in public APIs using the `Valid` function
5. **Handle errors appropriately** for production code
6. **Use `omitempty` tags** carefully - they affect merge behavior

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.