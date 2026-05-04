// Package jsonmerge implements RFC 7386 JSON Merge Patch for Go.
//
// It supports structs, map[string]any, []byte, and string documents.
// Use WithMutate(true) to allow in-place updates when the caller prefers
// speed over preserving map inputs.
//
// See https://datatracker.ietf.org/doc/html/rfc7386.
package jsonmerge

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/deepclone"
)

// Error is a sentinel error string.
type Error string

// Error returns e as a string.
func (e Error) Error() string { return string(e) }

const (
	// ErrMarshal indicates JSON marshaling failed.
	ErrMarshal Error = "marshal failed"

	// ErrUnmarshal indicates JSON unmarshaling failed.
	ErrUnmarshal Error = "unmarshal failed"

	// ErrConversion indicates type conversion between document types failed.
	ErrConversion Error = "type conversion failed"
)

// Merge applies patch to target according to RFC 7386.
// By default it preserves map targets; use WithMutate(true) to update them in place.
// If it fails, the error matches ErrMarshal, ErrUnmarshal, or ErrConversion.
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

	_, patchIsObject := patchInterface.(map[string]any)
	targetMap, targetIsMap := any(target).(map[string]any)
	if !options.Mutate && patchIsObject && targetIsMap && targetMap != nil {
		targetInterface = deepclone.Clone(targetInterface)
	}

	merged := applyPatch(targetInterface, patchInterface)

	result, err := convertFromInterface[T](merged)
	if err != nil {
		return nil, fmt.Errorf("convert result: %w", err)
	}

	return &Result[T]{Doc: result}, nil
}

// Generate returns a merge patch that transforms source into target.
// If it fails, the error matches ErrMarshal, ErrUnmarshal, or ErrConversion.
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

	patch := generatePatch(sourceInterface, targetInterface, true)

	result, err := convertFromInterface[T](patch)
	if err != nil {
		return zero, fmt.Errorf("convert patch: %w", err)
	}

	return result, nil
}

// Valid reports whether patch is accepted as a JSON Merge Patch value.
func Valid[T Document](patch T) bool {
	_, err := convertToInterface(patch)
	return err == nil
}

func applyPatch(target, patch any) any {
	patchObj, isPatchObject := patch.(map[string]any)
	if !isPatchObject {
		return patch
	}

	targetObj, isTargetObject := target.(map[string]any)
	if !isTargetObject || targetObj == nil {
		targetObj = make(map[string]any)
	}

	for name, value := range patchObj {
		if value == nil {
			delete(targetObj, name)
			continue
		}

		targetObj[name] = applyPatch(targetObj[name], value)
	}

	return targetObj
}

func generatePatch(source, target any, preserveEmptyObject bool) any {
	targetObj, isTargetObject := target.(map[string]any)
	if !isTargetObject {
		return target
	}

	sourceObj, isSourceObject := source.(map[string]any)
	if !isSourceObject {
		return target
	}

	var patch map[string]any
	setPatch := func(key string, value any) {
		if patch == nil {
			patch = make(map[string]any)
		}
		patch[key] = value
	}

	for key, targetValue := range targetObj {
		sourceValue, exists := sourceObj[key]
		if !exists {
			setPatch(key, targetValue)
			continue
		}

		sourceValueObj, isSourceObj := sourceValue.(map[string]any)
		targetValueObj, isTargetObj := targetValue.(map[string]any)
		if isSourceObj && isTargetObj {
			nestedPatch := generatePatch(sourceValueObj, targetValueObj, false)
			if nestedPatch != nil {
				setPatch(key, nestedPatch)
			}
			continue
		}

		if !deepEqual(sourceValue, targetValue) {
			setPatch(key, targetValue)
		}
	}

	for key := range sourceObj {
		if _, exists := targetObj[key]; !exists {
			setPatch(key, nil)
		}
	}

	if patch != nil {
		return patch
	}
	if preserveEmptyObject {
		return map[string]any{}
	}
	return nil
}

func deepEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	switch va := a.(type) {
	case bool:
		vb, ok := b.(bool)
		return ok && va == vb
	case float64:
		vb, ok := b.(float64)
		return ok && va == vb
	case int:
		vb, ok := b.(int)
		return ok && va == vb
	case int64:
		vb, ok := b.(int64)
		return ok && va == vb
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case []any:
		vb, ok := b.([]any)
		return ok && slices.EqualFunc(va, vb, deepEqual)
	case map[string]any:
		vb, ok := b.(map[string]any)
		return ok && maps.EqualFunc(va, vb, deepEqual)
	default:
		return reflect.DeepEqual(a, b)
	}
}

func convertToInterface[T Document](doc T) (any, error) {
	switch typed := any(doc).(type) {
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
