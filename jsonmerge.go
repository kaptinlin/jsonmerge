// Package jsonmerge provides RFC 7386 JSON Merge Patch implementation.
// It offers a simple, type-safe API for applying merge patches to JSON documents.
//
// Basic usage:
//
//	result, err := jsonmerge.Merge(target, patch)
//	if err != nil {
//		return err
//	}
//	// result.Doc contains the merged document
//
// The library supports multiple document types:
//   - Structs (with full type safety)
//   - map[string]any (dynamic documents)
//   - []byte (JSON bytes)
//   - string (JSON strings)
//
// All operations are immutable by default. Use WithMutate(true) for
// performance-critical scenarios where in-place modification is acceptable.
package jsonmerge

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/deepclone"
)

// Predefined errors - Define merge-specific errors clearly
var (
	ErrUnsupportedType       = errors.New("unsupported document type")
	ErrMergeFailed           = errors.New("merge operation failed")
	ErrInvalidJSON           = errors.New("invalid JSON document")
	ErrConversionFailed      = errors.New("type conversion failed")
	ErrMarshalFailed         = errors.New("JSON marshal operation failed")
	ErrUnmarshalFailed       = errors.New("JSON unmarshal operation failed")
	ErrPatchGenerationFailed = errors.New("patch generation failed")
)

// Merge applies a JSON Merge Patch (RFC 7386) to a target document.
// It returns a new Result containing the merged document and metadata.
// The operation is immutable by default unless WithMutate(true) is specified.
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error) {
	// Apply options
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	// Convert inputs to interface{} for processing
	targetInterface, err := convertToInterface(target)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert target document: %w", ErrUnsupportedType, err)
	}

	patchInterface, err := convertToInterface(patch)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert patch document: %w", ErrUnsupportedType, err)
	}

	// Clone target if not mutating
	if !options.Mutate {
		targetInterface = deepclone.Clone(targetInterface)
	}

	// Apply merge patch
	merged := mergePatch(targetInterface, patchInterface)

	// Convert back to original type
	result, err := convertFromInterface[T](merged)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert merged result: %w", ErrConversionFailed, err)
	}

	return &Result[T]{
		Doc: result,
	}, nil
}

// Generate creates a JSON Merge Patch between source and target documents.
// The generated patch can be applied to source to produce target.
func Generate[T Document](source, target T) (T, error) {
	// Convert inputs to interface{} for processing
	sourceInterface, err := convertToInterface(source)
	if err != nil {
		return *new(T), fmt.Errorf("%w: failed to convert source document: %w", ErrUnsupportedType, err)
	}

	targetInterface, err := convertToInterface(target)
	if err != nil {
		return *new(T), fmt.Errorf("%w: failed to convert target document: %w", ErrUnsupportedType, err)
	}

	// Generate patch
	patch := generatePatch(sourceInterface, targetInterface)

	// Convert back to original type
	result, err := convertFromInterface[T](patch)
	if err != nil {
		return *new(T), fmt.Errorf("%w: failed to convert generated patch: %w", ErrPatchGenerationFailed, err)
	}

	return result, nil
}

// Valid checks if a patch is a valid JSON Merge Patch.
// According to RFC 7386, any valid JSON value is a valid merge patch.
func Valid[T Document](patch T) bool {
	// Convert patch to interface{} for validation
	_, err := convertToInterface(patch)
	return err == nil
}

// mergePatch implements the core RFC 7386 algorithm.
// This function directly corresponds to the algorithm specified in RFC 7386 Section 2.
func mergePatch(target, patch interface{}) interface{} {
	// If patch is not an object, return patch (complete replacement)
	if !isObject(patch) {
		return patch
	}

	// If target is not an object, create empty object
	if !isObject(target) {
		target = make(map[string]interface{})
	}

	targetObj := target.(map[string]interface{})
	patchObj := patch.(map[string]interface{})

	// Apply patch operations
	for name, value := range patchObj {
		if value == nil {
			// null value means delete the field
			delete(targetObj, name)
		} else {
			// Recursive merge for nested objects
			targetObj[name] = mergePatch(targetObj[name], value)
		}
	}

	return targetObj
}

// generatePatch creates a patch that transforms source into target.
func generatePatch(source, target interface{}) interface{} {
	// If target is not an object, return target as complete replacement
	if !isObject(target) {
		return target
	}

	// If source is not an object, return target as complete replacement
	if !isObject(source) {
		return target
	}

	sourceObj := source.(map[string]interface{})
	targetObj := target.(map[string]interface{})
	patch := make(map[string]interface{})

	// Add fields that exist in target
	for key, targetValue := range targetObj {
		if sourceValue, exists := sourceObj[key]; exists {
			// Field exists in both - check if they need merging
			if isObject(sourceValue) && isObject(targetValue) {
				// Both are objects - recursive patch generation
				nestedPatch := generatePatch(sourceValue, targetValue)
				if nestedPatch != nil && len(nestedPatch.(map[string]interface{})) > 0 {
					patch[key] = nestedPatch
				}
			} else if !deepEqual(sourceValue, targetValue) {
				// Values are different - include in patch
				patch[key] = targetValue
			}
		} else {
			// Field only exists in target - add it
			patch[key] = targetValue
		}
	}

	// Add null values for fields that should be deleted (exist in source but not in target)
	for key := range sourceObj {
		if _, exists := targetObj[key]; !exists {
			patch[key] = nil
		}
	}

	return patch
}

// isObject checks if a value is a JSON object (map[string]interface{}).
func isObject(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

// deepEqual compares two values for deep equality.
func deepEqual(a, b interface{}) bool {
	// Use JSON marshaling for deep comparison
	aBytes, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aBytes) == string(bBytes)
}

// convertToInterface converts various document types to interface{} for processing.
func convertToInterface[T Document](doc T) (interface{}, error) {
	var v interface{} = doc

	switch typed := v.(type) {
	case []byte:
		var result interface{}
		if err := json.Unmarshal(typed, &result); err != nil {
			return nil, fmt.Errorf("%w: failed to unmarshal JSON bytes: %w", ErrInvalidJSON, err)
		}
		return result, nil

	case string:
		// First try to unmarshal as JSON
		var result interface{}
		if err := json.Unmarshal([]byte(typed), &result); err == nil {
			return result, nil
		}
		// If it's not valid JSON, treat it as a raw string value
		return typed, nil

	case map[string]any:
		return typed, nil

	case nil:
		return nil, nil

	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return typed, nil

	default:
		// For struct types, marshal then unmarshal to get map[string]interface{}
		data, err := json.Marshal(typed)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to marshal document to JSON: %w", ErrMarshalFailed, err)
		}

		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("%w: failed to unmarshal document from JSON: %w", ErrUnmarshalFailed, err)
		}
		return result, nil
	}
}

// convertFromInterface converts interface{} back to the original document type.
func convertFromInterface[T Document](val interface{}) (T, error) {
	var zero T

	switch any(zero).(type) {
	case []byte:
		// Convert to JSON bytes
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: failed to marshal value to bytes: %w", ErrMarshalFailed, err)
		}
		return any(data).(T), nil

	case string:
		// Convert to JSON string
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: failed to marshal value to string: %w", ErrMarshalFailed, err)
		}
		return any(string(data)).(T), nil

	case map[string]any:
		if m, ok := val.(map[string]interface{}); ok {
			return any(m).(T), nil
		}
		return zero, fmt.Errorf("%w: expected map[string]interface{}, got %T", ErrConversionFailed, val)

	default:
		// For struct types, marshal then unmarshal to target type
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: failed to marshal value for struct conversion: %w", ErrMarshalFailed, err)
		}

		var target T
		if err := json.Unmarshal(data, &target); err != nil {
			return zero, fmt.Errorf("%w: failed to unmarshal value to target struct type: %w", ErrUnmarshalFailed, err)
		}
		return target, nil
	}
}
