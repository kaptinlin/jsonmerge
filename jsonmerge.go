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
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/deepclone"
)

// Error represents a sentinel error type for the jsonmerge package.
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors for error checking with errors.Is.
const (
	// ErrMarshal indicates JSON marshaling failed.
	ErrMarshal Error = "marshal failed"

	// ErrUnmarshal indicates JSON unmarshaling failed.
	ErrUnmarshal Error = "unmarshal failed"

	// ErrConversion indicates type conversion between document types failed.
	ErrConversion Error = "type conversion failed"
)

// Merge applies a JSON Merge Patch (RFC 7386) to a target document.
// It returns a new Result containing the merged document.
// The operation is immutable by default unless WithMutate(true) is specified.
//
// Possible errors (checkable with errors.Is):
//   - ErrMarshal: JSON marshaling failed during type conversion
//   - ErrUnmarshal: JSON unmarshaling failed during type conversion
//   - ErrConversion: type conversion between document types failed
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	targetInterface, err := convertToInterface(target)
	if err != nil {
		return nil, fmt.Errorf("convert target: %w", err)
	}

	patchInterface, err := convertToInterface(patch)
	if err != nil {
		return nil, fmt.Errorf("convert patch: %w", err)
	}

	if !options.Mutate {
		targetInterface = deepclone.Clone(targetInterface)
	}

	merged := applyPatch(targetInterface, patchInterface)

	result, err := convertFromInterface[T](merged)
	if err != nil {
		return nil, fmt.Errorf("convert result: %w", err)
	}

	return &Result[T]{Doc: result}, nil
}

// Generate creates a JSON Merge Patch between source and target documents.
// The generated patch can be applied to source to produce target.
//
// Possible errors (checkable with errors.Is):
//   - ErrMarshal: JSON marshaling failed during type conversion
//   - ErrUnmarshal: JSON unmarshaling failed during type conversion
//   - ErrConversion: type conversion between document types failed
func Generate[T Document](source, target T) (T, error) {
	var zero T

	sourceInterface, err := convertToInterface(source)
	if err != nil {
		return zero, fmt.Errorf("convert source: %w", err)
	}

	targetInterface, err := convertToInterface(target)
	if err != nil {
		return zero, fmt.Errorf("convert target: %w", err)
	}

	patch := generatePatch(sourceInterface, targetInterface)

	result, err := convertFromInterface[T](patch)
	if err != nil {
		return zero, fmt.Errorf("convert patch: %w", err)
	}

	return result, nil
}

// Valid checks if a patch is a valid JSON Merge Patch.
// According to RFC 7386, any valid JSON value is a valid merge patch.
func Valid[T Document](patch T) bool {
	_, err := convertToInterface(patch)
	return err == nil
}

// applyPatch applies the RFC 7386 Section 2 merge patch algorithm.
//
// Implementation: if patch is not an object, return patch (complete replacement);
// if target is not an object, create empty object; then recursively merge fields,
// deleting on null.
func applyPatch(target, patch any) any {
	patchObj, isPatchObject := patch.(map[string]any)
	if !isPatchObject {
		return patch
	}

	targetObj, isTargetObject := target.(map[string]any)
	if !isTargetObject {
		targetObj = make(map[string]any)
	}

	for name, value := range patchObj {
		if value == nil {
			delete(targetObj, name)
		} else {
			targetObj[name] = applyPatch(targetObj[name], value)
		}
	}

	return targetObj
}

// generatePatch creates a patch that transforms source into target.
func generatePatch(source, target any) any {
	targetObj, isTargetObject := target.(map[string]any)
	if !isTargetObject {
		return target
	}

	sourceObj, isSourceObject := source.(map[string]any)
	if !isSourceObject {
		return target
	}

	patch := make(map[string]any)

	for key, targetValue := range targetObj {
		sourceValue, exists := sourceObj[key]
		if !exists {
			patch[key] = targetValue
			continue
		}

		sourceValueObj, isSourceObj := sourceValue.(map[string]any)
		targetValueObj, isTargetObj := targetValue.(map[string]any)
		if isSourceObj && isTargetObj {
			nestedPatch := generatePatch(sourceValueObj, targetValueObj)
			if m, ok := nestedPatch.(map[string]any); ok && len(m) > 0 {
				patch[key] = nestedPatch
			}
			continue
		}

		if !deepEqual(sourceValue, targetValue) {
			patch[key] = targetValue
		}
	}

	for key := range sourceObj {
		if _, exists := targetObj[key]; !exists {
			patch[key] = nil
		}
	}

	return patch
}

// deepEqual compares two values for deep equality.
// Uses JSON marshaling for comparison to avoid panics on uncomparable types.
func deepEqual(a, b any) bool {
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

// convertToInterface converts various document types to a common representation for processing.
func convertToInterface[T Document](doc T) (any, error) {
	var v any = doc

	switch typed := v.(type) {
	case []byte:
		var result any
		if err := json.Unmarshal(typed, &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
		}
		return result, nil

	case string:
		var result any
		if err := json.Unmarshal([]byte(typed), &result); err == nil {
			return result, nil
		}
		return typed, nil

	case map[string]any, nil,
		bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return typed, nil

	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMarshal, err)
		}

		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnmarshal, err)
		}
		return result, nil
	}
}

// convertFromInterface converts a common representation back to the original document type.
func convertFromInterface[T Document](val any) (T, error) {
	var zero T

	switch any(zero).(type) {
	case []byte:
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrMarshal, err)
		}
		return any(data).(T), nil

	case string:
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrMarshal, err)
		}
		return any(string(data)).(T), nil

	case map[string]any:
		if m, ok := val.(map[string]any); ok {
			return any(m).(T), nil
		}
		return zero, fmt.Errorf("%w: expected map[string]any, got %T", ErrConversion, val)

	default:
		data, err := json.Marshal(val)
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrMarshal, err)
		}

		var target T
		if err := json.Unmarshal(data, &target); err != nil {
			return zero, fmt.Errorf("%w: %w", ErrUnmarshal, err)
		}
		return target, nil
	}
}
