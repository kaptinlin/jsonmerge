# JSON Merge Patch for Go

[![Go Module](https://img.shields.io/badge/go-module-blue.svg)](https://golang.org/)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/jsonmerge.svg)](https://pkg.go.dev/github.com/kaptinlin/jsonmerge)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A small RFC 7386 JSON Merge Patch library for Go with explicit patch values and lossless typed results

## Features

- **RFC 7386 semantics**: Objects merge recursively, `null` deletes object members, non-object patches replace the target, and arrays replace as a whole
- **Explicit patches**: Build a `Patch` with `Parse` for JSON text or `NewPatch` for Go values
- **Clear string semantics**: Plain `string` is a JSON string scalar; use `[]byte` or `jsonmerge.JSON` for encoded JSON text
- **Lossless results**: `Apply` returns the requested Go type when the merged JSON value can be represented without loss
- **Shared JSON semantics**: `Diff` and `Apply` share one normalized JSON value model across maps, structs, bytes, and JSON text
- **Stable patch encoding**: `Patch.MarshalJSON` emits deterministic compact JSON and preserves encoded JSON number literals
- **Small API**: Learn `Patch`, `Parse`, `NewPatch`, `Apply`, `Diff`, and `MarshalJSON`

## Installation

```bash
go get github.com/kaptinlin/jsonmerge
```

Requires Go 1.26.4 or newer.

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

	patch, err := jsonmerge.NewPatch(map[string]any{
		"age":   31,
		"email": "john@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := jsonmerge.Apply(target, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result["name"])
	fmt.Println(result["age"])
	fmt.Println(result["email"])
}
```

## Document Forms

| Form | Behavior |
| --- | --- |
| `map[string]any` | Object document form |
| `[]byte` | Encoded JSON text; malformed text returns `ErrInvalidJSON` |
| `jsonmerge.JSON` | Encoded JSON text carried as a string |
| `string` | JSON string scalar, never parsed as JSON text |
| Structs and typed Go values | Converted through JSON, then projected back without loss |
| Scalars and `nil` | Accepted when the result can be represented as the requested type |

If the merged JSON value cannot be represented by the requested Go type, `Apply` returns `ErrCannotRepresent`.

## Core Operations

| API | Description |
| --- | --- |
| `Parse(data []byte) (Patch, error)` | Parse encoded JSON text as a merge patch |
| `NewPatch(value any) (Patch, error)` | Convert a Go value into a merge patch |
| `Apply[T any](target T, patch Patch) (T, error)` | Apply a patch and return `T` when projection is lossless |
| `Diff(source, target any) (Patch, error)` | Build a patch that transforms source into target in the normalized JSON model |
| `(Patch).MarshalJSON() ([]byte, error)` | Encode a patch for storage or transport |

## Struct Targets

Use sparse object patches for typed targets. Struct zero values are normal JSON values when you put the struct itself into `NewPatch`.

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age"`
}

user := User{Name: "John", Email: "john@example.com", Age: 30}
patch, _ := jsonmerge.NewPatch(map[string]any{"name": "Jane"})

user, _ = jsonmerge.Apply(user, patch)
fmt.Println(user.Name)  // Jane
fmt.Println(user.Email) // john@example.com
fmt.Println(user.Age)   // 30
```

Projection is strict. If a patch adds a member that the struct cannot represent, `Apply` fails with `ErrCannotRepresent` instead of silently dropping data.

## JSON Text

Use `Parse` for patch text and `jsonmerge.JSON` or `[]byte` for document text.

```go
doc := jsonmerge.JSON(`{"name":"John","age":30}`)
patch, _ := jsonmerge.Parse([]byte(`{"name":"Jane"}`))

doc, _ = jsonmerge.Apply(doc, patch)
fmt.Println(doc)
```

Malformed text fails when parsed as text. The same bytes inside a Go `string` are just a string scalar when passed to `NewPatch`.
Encoded JSON number literals are preserved through `Parse`, `Apply`, and `Patch.MarshalJSON`; large integers are not rounded while carried as JSON text.

## Error Handling

The package wraps failures with sentinel errors so callers can use `errors.Is`:

- `ErrInvalidJSON`
- `ErrInvalidValue`
- `ErrCannotRepresent`

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

The public API is pure by default and does not mutate caller-owned maps.
Run benchmarks on your hardware with:

```bash
task bench
```

## Development

```bash
task test          # Run all tests with race detection
task lint          # Run golangci-lint and tidy checks
task verify        # Run deps, fmt, vet, lint, test, and vuln
```

For development guidelines, see [AGENTS.md](AGENTS.md). For detailed package contracts, see [`SPECS/`](SPECS/).

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
