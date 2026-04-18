# Examples

Examples for the JSON Merge Patch (RFC 7386) implementation.

## Quick Start

Choose an example to run:

```bash
go run ./examples/struct-merge
go run ./examples/map-merge
go run ./examples/json-string-merge
go run ./examples/json-bytes-merge
```

## Example Structure

Each directory focuses on one document form:

- `struct-merge` — type-safe struct operations
- `map-merge` — dynamic map data operations
- `json-string-merge` — JSON string processing
- `json-bytes-merge` — JSON byte processing

## Core Concepts

### RFC 7386 Key Behaviors

1. **Recursive object merging** — nested objects are merged recursively.
2. **`null` deletes fields** — `null` values remove target fields.
3. **Array replacement** — arrays are replaced entirely, not element by element.
4. **New field addition** — new fields from the patch are added to the target.

### Supported Document Types

- `struct` — type-safe documents converted through JSON
- `map[string]any` — dynamic object form
- `string` — valid JSON strings or raw string scalar values
- `[]byte` — valid JSON byte slices
