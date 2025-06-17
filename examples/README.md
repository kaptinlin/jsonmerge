# Examples

Examples for JSON Merge Patch (RFC 7386) implementation.

## Quick Start

Choose any example to run:
```bash
cd struct-merge && go run main.go
cd map-merge && go run main.go  
cd json-string-merge && go run main.go
cd json-bytes-merge && go run main.go
```

## Example Structure

Each directory focuses on one document type:

- **`struct-merge/`** - Type-safe struct operations
- **`map-merge/`** - Dynamic map data operations
- **`json-string-merge/`** - JSON string processing
- **`json-bytes-merge/`** - Byte array processing

## Core Concepts

### RFC 7386 Key Behaviors
1. **Recursive object merging** - Nested objects are merged recursively
2. **null deletes fields** - `null` values remove target fields
3. **Array replacement** - Arrays are replaced entirely, not element-wise
4. **New field addition** - New fields from patch are added to target

### Supported Document Types
- `struct` - Type-safe with compile-time checking
- `map[string]any` - Dynamic data with runtime flexibility
- `string` - JSON string format
- `[]byte` - JSON byte array format