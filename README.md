# JSON Merge Patch for Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.26-blue.svg)](https://golang.org/)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/jsonmerge.svg)](https://pkg.go.dev/github.com/kaptinlin/jsonmerge)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A type-safe RFC 7386 JSON Merge Patch library for Go that preserves the caller's document type

## Features

- **RFC 7386 semantics**: Object patches merge recursively, `null` deletes fields, non-object patches replace the target, and arrays replace as a whole
- **Type preservation**: `Merge` and `Generate` return the same document type they receive
- **Flexible document forms**: Work with `map[string]any`, JSON `[]byte`, JSON `string`, structs, and scalar values
- **Safe-by-default map merges**: Preserve `map[string]any` inputs unless you opt into `WithMutate(true)`
- **Small API**: Learn `Merge`, `Generate`, `Valid`, and `WithMutate`
- **Benchmarked and tested**: Includes RFC Appendix A coverage, benchmarks, and runnable examples

## Installation

```bash
go get github.com/kaptinlin/jsonmerge
```

Requires **Go 1.26+**.

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	target := map[string]any{"name": "John", "age": 30}
	patch := map[string]any{"age": 31, "email": "john@example.com"}

	result, err := jsonmerge.Merge(target, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Doc["name"])
	fmt.Println(result.Doc["age"])
	fmt.Println(result.Doc["email"])
}
```

## Document Forms

| Form | Behavior |
| --- | --- |
| `map[string]any` | Canonical in-memory object form |
| `[]byte` | Must contain valid JSON |
| `string` | Parsed as JSON when valid; otherwise treated as a raw string scalar |
| Structs and typed Go values | Converted through JSON before merge or generation |
| Scalars and `nil` | Accepted when they fit the call's static type `T` |

Invalid JSON bytes fail. Invalid JSON strings remain valid raw string values.

## Core Operations

| API | Description |
| --- | --- |
| `Merge[T Document](target, patch T, opts ...Option)` | Apply a merge patch and return `*Result[T]` |
| `Generate[T Document](source, target T)` | Build a merge patch that transforms `source` into `target` |
| `Valid[T Document](patch T)` | Report whether a value is accepted as a patch input |
| `WithMutate(true)` | Allow in-place updates for `map[string]any` targets during object merges |

## Struct Patch Semantics

Struct patches follow JSON marshaling rules. Zero values overwrite fields once they are present in the marshaled patch.

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age"`
}

user := User{Name: "John", Email: "john@example.com", Age: 30}
patch := User{Name: "Jane"}

result, _ := jsonmerge.Merge(user, patch)
fmt.Println(result.Doc.Name)  // Jane
fmt.Println(result.Doc.Email) // john@example.com
fmt.Println(result.Doc.Age)   // 0
```

Use pointer fields or `map[string]any` when you need omission-versus-zero-value distinction.

## Error Handling

The package wraps failures with sentinel errors so callers can use `errors.Is`:

- `ErrMarshal`
- `ErrUnmarshal`
- `ErrConversion`

Invalid JSON bytes return an unmarshal error. Invalid JSON strings are accepted as raw string scalar values.

## Examples

Run the example programs from the repository root:

```bash
go run ./examples/map-merge
go run ./examples/struct-merge
go run ./examples/json-string-merge
go run ./examples/json-bytes-merge
```

See [`examples/`](examples/) for complete programs.

## Performance

`WithMutate(true)` reduces allocations for `map[string]any` object merges by allowing in-place updates.
Run benchmarks on your hardware with:

```bash
task bench
```

## Development

```bash
task test          # Run all tests with race detection
task lint          # Run golangci-lint and tidy checks
task markdownlint  # Lint Markdown files
task verify        # Run deps, fmt, vet, lint, test, and vuln
```

For development guidelines and package contracts, see [AGENTS.md](AGENTS.md).

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
