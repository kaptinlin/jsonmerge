---
description: Go High-Performance Development Rules for Cursor Code Generation
globs:
alwaysApply: true
---
# Go Development Rules for JSON Merge Patch (RFC 7386)

## Core Principles

### 1. Code Quality Standards
- **All comments and documentation must be in English**
- **Follow Go conventions and idioms strictly**
- **Prioritize RFC 7386 compliance over performance optimizations**
- **Simple, elegant implementation over complex optimizations**
- **Type safety first - leverage Go 1.18+ generics comprehensively**
- **Immutable by default with optional performance mode**

### 2. RFC 7386 Implementation Philosophy
- **Strict RFC compliance** - Implement the algorithm exactly as specified in RFC 7386 Section 2
- **Minimal API surface** - Only 3 core functions: Merge, Generate, Valid
- **Zero tolerance for RFC deviations** - No custom behaviors that break standard compliance
- **Comprehensive type support** - Handle all common Go types safely and efficiently

### 3. Performance and Memory Management
- **Immutable operations by default** - Use deep cloning to prevent side effects
- **Optional in-place mutation** - Provide `WithMutate(true)` for performance-critical scenarios
- **Efficient type conversions** - Minimize JSON marshal/unmarshal cycles
- **Memory-conscious design** - Avoid unnecessary allocations in hot paths

## Dependencies and Import Strategy

### Required Dependencies
```go
// For deep cloning when immutable operations are needed
import "github.com/kaptinlin/deepclone"

// For JSON operations - use experimental json for better performance and features
import "github.com/go-json-experiment/json"
```

### Dependency Constraints
- **Minimal external dependencies** - Only the two specified packages beyond standard library
- **No utility libraries** - Implement required utilities internally
- **Standard library preference** - Use Go standard library for all other functionality
- **Lightweight design** - Keep the implementation minimal and focused

## Error Handling Standards

### Error Definition Requirements
- **Predefined static errors** - Define all common error types as package-level variables
- **Descriptive error names** - Use clear, specific error names (e.g., `ErrMarshalFailed`, `ErrInvalidJSON`)
- **Consistent error messages** - Use standard terminology across all errors
- **Hierarchical error types** - Organize errors by operation type (marshal, unmarshal, conversion, merge)

### Error Wrapping Standards
- **Static errors** - Return predefined errors directly for known failure modes
- **Dynamic errors** - Use `fmt.Errorf("%w: context: %w", StaticError, originalError)` format
- **Contextual information** - Always provide specific context about what operation failed
- **Error chain preservation** - Maintain full error chain for debugging and introspection

### Required Error Types
- `ErrUnsupportedType` - Document type not supported for merge operations
- `ErrMergeFailed` - General merge operation failure
- `ErrInvalidJSON` - Invalid JSON document structure or format
- `ErrConversionFailed` - Type conversion between document formats failed
- `ErrMarshalFailed` - JSON marshaling operation failed
- `ErrUnmarshalFailed` - JSON unmarshaling operation failed
- `ErrPatchGenerationFailed` - Patch generation process failed

### Error Context Patterns
- **Target/Source identification** - Clearly identify which document caused the error
- **Operation specificity** - Specify the exact operation that failed
- **Type information** - Include type details when type conversion fails
- **JSON path context** - When possible, include JSON path information for nested errors

## ✅ RFC 7386 Core Algorithm Implementation

### MergePatch Function Requirements
- **Algorithm fidelity** - Implement RFC 7386 Section 2 algorithm exactly
- **Object handling** - Recursive merging for JSON objects only
- **Null semantics** - `null` values must delete fields from target
- **Non-object replacement** - Complete replacement for all non-object values
- **Array handling** - Arrays must be replaced entirely, never element-wise merged

### Critical RFC 7386 Rules
1. **Object merging**: Objects are merged recursively following the algorithm
2. **Null deletion**: `null` values delete corresponding fields from target
3. **Complete replacement**: Non-objects completely replace target values  
4. **Array replacement**: Arrays are replaced entirely, not element-wise merged
5. **Primitive handling**: All primitive values replace corresponding target values

### Type System Integration
- **Generic constraints** - Use type constraints to limit supported document types
- **Type conversion consistency** - Maintain type fidelity throughout merge operations
- **Interface standardization** - Convert all types to `interface{}` for algorithm processing
- **Result type preservation** - Convert results back to original generic type

## Type Conversion Architecture

### Document Type Support Matrix
- **Structs** - Full support with JSON tags, maintaining type safety
- **map[string]any** - Native support as primary working format
- **[]byte** - JSON bytes with automatic parsing and generation
- **string** - JSON strings with validation and conversion
- **Primitives** - Direct support for bool, int variants, float variants
- **nil handling** - Proper null value handling across all types

### Conversion Function Requirements
- **Bidirectional conversion** - Convert to and from interface{} format
- **Error propagation** - Proper error handling for invalid conversions
- **Type preservation** - Maintain original type information for result conversion
- **JSON compliance** - Ensure all conversions produce valid JSON structures

### Performance Optimization Guidelines
- **Minimize marshaling** - Avoid unnecessary JSON marshal/unmarshal cycles
- **Direct type handling** - Handle map[string]any directly without conversion
- **Efficient cloning** - Use deepclone only when immutable operations are required
- **String optimization** - Handle JSON strings efficiently with proper validation

## API Design Standards

### Function Signature Requirements
- **Generic type parameters** - Use generics for compile-time type safety
- **Document constraint** - Apply Document interface constraint to all generic parameters
- **Options pattern** - Use variadic options for configuration
- **Result wrapping** - Wrap results in typed Result struct for additional metadata

### Core API Functions
1. **Merge[T Document]** - Apply merge patch to target document
2. **Generate[T Document]** - Create merge patch between two documents  
3. **Valid[T Document]** - Validate if document is valid merge patch

### Options Design
- **Mutate option** - Single boolean option for in-place modification
- **Extensible pattern** - Use function options pattern for future extensibility
- **Minimal configuration** - Avoid complex configuration structures
- **Clear documentation** - Document performance implications of each option

## Testing Requirements

### RFC 7386 Compliance Testing
- **All RFC examples** - Implement every test case from RFC 7386 Appendix A
- **Edge case coverage** - Test nested objects, arrays, nulls, and primitives
- **Type safety verification** - Ensure generic types work correctly across all scenarios
- **Error condition testing** - Verify proper error handling for invalid inputs

### Performance Testing Standards
- **Benchmark coverage** - Benchmark all core operations with realistic data
- **Memory profiling** - Verify memory usage patterns and identify leaks
- **Mutation vs immutable** - Compare performance characteristics of both modes
- **Large document testing** - Test performance with complex, nested JSON documents

### Type System Testing
- **All supported types** - Test every document type in the support matrix
- **Conversion roundtrips** - Verify type conversion fidelity
- **Generic type inference** - Test that Go's type inference works correctly
- **Error type verification** - Ensure errors are properly typed and wrapped

## Documentation Standards

### Package Documentation Requirements
- **RFC 7386 reference** - Clear reference to the implemented standard
- **Usage examples** - Comprehensive examples for all supported document types
- **Type safety explanation** - Explain generic type system benefits
- **Performance guidance** - Document when to use mutation options

### Function Documentation Standards
- **Algorithm reference** - Reference specific RFC sections where applicable
- **Type constraint explanation** - Explain generic type constraints and their purpose
- **Error documentation** - Document all possible error conditions and types
- **Example code** - Provide working examples in documentation

### Code Comment Requirements
- **Algorithm steps** - Comment each step of the RFC 7386 algorithm implementation
- **Type conversion rationale** - Explain why specific type conversions are needed
- **Performance considerations** - Comment on performance implications of design choices
- **Edge case handling** - Document how edge cases are handled

## Code Organization Principles

### File Structure Standards
- **Single main file** - Keep core implementation in one cohesive file
- **Logical grouping** - Group related functions together (core algorithm, type conversion, etc.)
- **Clear separation** - Separate public API from internal implementation details
- **Consistent ordering** - Follow consistent function ordering pattern

### Naming Conventions
- **Exported functions** - Use clear, action-oriented names (Merge, Generate, Valid)
- **Internal functions** - Use descriptive names explaining purpose (convertToInterface, isObject)
- **Error variables** - Use Err prefix with descriptive names (ErrInvalidJSON)
- **Type names** - Use clear, Go-idiomatic names (Document, Result, Options)

### Code Style Requirements
- **gofmt compliance** - All code must pass gofmt without changes
- **golint compliance** - Follow all golint recommendations
- **Consistent indentation** - Use tabs for indentation, spaces for alignment
- **Line length limits** - Keep lines under 100 characters when practical

## Security and Robustness

### Input Validation
- **JSON structure validation** - Validate JSON structure before processing
- **Type assertion safety** - Use safe type assertions with ok checks where needed
- **Nil pointer protection** - Handle nil values gracefully throughout the codebase
- **Panic prevention** - Avoid panics in all public API functions

### Memory Safety
- **Deep clone verification** - Ensure deep cloning prevents unintended mutations
- **Resource cleanup** - Properly handle resources and avoid memory leaks
- **Large document handling** - Handle large documents without excessive memory usage
- **Circular reference protection** - Handle potential circular references in nested structures

## Development Workflow Standards

### Code Generation Guidelines for Cursor
- **Follow RFC strictly** - Never deviate from RFC 7386 algorithm specification
- **Prioritize type safety** - Use generics and type constraints comprehensively
- **Maintain error handling** - Implement consistent error handling patterns
- **Preserve immutability** - Default to immutable operations with optional mutation

### Quality Assurance Requirements
- **100% test coverage** - Achieve complete test coverage for all public APIs
- **Benchmark verification** - Verify performance characteristics meet requirements
- **Documentation completeness** - Ensure all public APIs are fully documented
- **RFC compliance verification** - Regular verification against RFC 7386 test cases

## Success Criteria

### Must Achieve (Non-negotiable)
- ✅ **100% RFC 7386 compliance** - Pass all RFC test cases without exception
- ✅ **Type safety guarantee** - Compile-time type safety for all operations
- ✅ **Zero regression tolerance** - No breaking changes to RFC compliance
- ✅ **Complete error handling** - Comprehensive error coverage with proper types

### Should Achieve (High Priority)
- ✅ **High performance** - Efficient operation for typical merge scenarios
- ✅ **Memory efficiency** - Minimal memory overhead in normal operations
- ✅ **Production readiness** - Robust handling of edge cases and error conditions
- ✅ **Developer ergonomics** - Clear, intuitive API that follows Go conventions

### Could Achieve (Nice to Have)
- ✅ **Advanced optimizations** - Performance optimizations for specific use cases
- ✅ **Extended validation** - Additional validation utilities beyond basic Valid function
- ✅ **Detailed metrics** - Performance and usage metrics for monitoring

## Key Development Principle

**RFC 7386 compliance is paramount. Every implementation decision must prioritize standard compliance over performance, convenience, or feature additions. The library exists to provide a correct, type-safe implementation of JSON Merge Patch with the ergonomics and safety guarantees expected in modern Go development.**
