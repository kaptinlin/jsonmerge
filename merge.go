// Package jsonmerge implements RFC 7386 JSON Merge Patch for Go.
//
// It supports structs, map[string]any, []byte, and string documents.
// Use WithMutate(true) to allow in-place updates when the caller prefers
// speed over preserving map inputs.
//
// See https://datatracker.ietf.org/doc/html/rfc7386.
package jsonmerge

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/deepclone"
)

var (
	// ErrMarshal indicates JSON marshaling failed.
	ErrMarshal = errors.New("marshal failed")

	// ErrUnmarshal indicates JSON unmarshaling failed.
	ErrUnmarshal = errors.New("unmarshal failed")

	// ErrConversion indicates type conversion between document types failed.
	ErrConversion = errors.New("type conversion failed")
)

// Merge applies patch to target according to RFC 7386.
// By default it preserves map targets; use WithMutate(true) to update them in place.
// If it fails, the error matches ErrMarshal, ErrUnmarshal, or ErrConversion.
func Merge[T Document](target, patch T, opts ...Option) (*Result[T], error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	targetValue, err := toJSONValue(target)
	if err != nil {
		return nil, fmt.Errorf("convert target: %w", err)
	}

	patchValue, err := toJSONValue(patch)
	if err != nil {
		return nil, fmt.Errorf("convert patch: %w", err)
	}

	_, patchIsObject := patchValue.(map[string]any)
	if patchMap, patchIsMap := any(patch).(map[string]any); patchIsMap && patchMap != nil {
		patchValue = deepclone.Clone(patchValue)
	}

	targetMap, targetIsMap := any(target).(map[string]any)
	if !options.Mutate && patchIsObject && targetIsMap && targetMap != nil {
		targetValue = deepclone.Clone(targetValue)
	}

	merged := applyPatch(targetValue, patchValue)

	result, err := fromJSONValue[T](merged)
	if err != nil {
		return nil, fmt.Errorf("convert result: %w", err)
	}

	return &Result[T]{Doc: result}, nil
}

// Generate returns a merge patch that transforms source into target.
// If it fails, the error matches ErrMarshal, ErrUnmarshal, or ErrConversion.
func Generate[T Document](source, target T) (T, error) {
	var zero T

	sourceValue, err := toJSONValue(source)
	if err != nil {
		return zero, fmt.Errorf("convert source: %w", err)
	}

	targetValue, err := toJSONValue(target)
	if err != nil {
		return zero, fmt.Errorf("convert target: %w", err)
	}

	patch := generatePatch(sourceValue, targetValue, true)

	result, err := fromJSONValue[T](patch)
	if err != nil {
		return zero, fmt.Errorf("convert patch: %w", err)
	}

	return result, nil
}

// Valid reports whether patch is accepted as a JSON Merge Patch value.
func Valid[T Document](patch T) bool {
	_, err := toJSONValue(patch)
	return err == nil
}

func wrapError(stage string, sentinel, err error) error {
	return fmt.Errorf("%s: %w", stage, errors.Join(sentinel, err))
}

func isJSONNull(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func applyPatch(target, patch any) any {
	patchObj, isPatchObject := patch.(map[string]any)
	if !isPatchObject {
		return patch
	}

	targetObj, targetIsObject := target.(map[string]any)
	if !targetIsObject || targetObj == nil {
		targetObj = make(map[string]any, len(patchObj))
	}

	for name, value := range patchObj {
		if isJSONNull(value) {
			delete(targetObj, name)
			continue
		}

		targetObj[name] = applyPatch(targetObj[name], value)
	}

	return targetObj
}

func generatePatch(source, target any, preserveEmptyObject bool) any {
	targetObj, targetIsObject := target.(map[string]any)
	if !targetIsObject {
		return target
	}

	sourceObj, sourceIsObject := source.(map[string]any)
	if !sourceIsObject {
		return target
	}

	var patch map[string]any
	setPatch := func(key string, value any) {
		if patch == nil {
			patch = make(map[string]any, len(targetObj))
		}
		patch[key] = value
	}

	for key, targetValue := range targetObj {
		sourceValue, exists := sourceObj[key]
		if !exists {
			setPatch(key, targetValue)
			continue
		}

		sourceObject, sourceIsObject := sourceValue.(map[string]any)
		targetObject, targetIsObject := targetValue.(map[string]any)
		if sourceIsObject && targetIsObject {
			nestedPatch := generatePatch(sourceObject, targetObject, false)
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

func toJSONValue[T Document](doc T) (any, error) {
	switch typed := any(doc).(type) {
	case []byte:
		var result any
		if err := json.Unmarshal(typed, &result); err != nil {
			return nil, wrapError("unmarshal bytes", ErrUnmarshal, err)
		}
		return result, nil

	case string:
		var result any
		if err := json.Unmarshal([]byte(typed), &result); err == nil {
			return result, nil
		}
		return typed, nil

	case map[string]any:
		if typed != nil {
			if _, err := json.Marshal(typed); err != nil {
				return nil, wrapError("marshal map", ErrMarshal, err)
			}
		}
		return typed, nil

	case nil,
		bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return typed, nil

	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return nil, wrapError("marshal document", ErrMarshal, err)
		}

		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, wrapError("unmarshal document", ErrUnmarshal, err)
		}
		return result, nil
	}
}

func fromJSONValue[T Document](val any) (T, error) {
	var zero T

	if _, ok := any(zero).(map[string]any); ok {
		m, ok := val.(map[string]any)
		if !ok {
			return zero, fmt.Errorf("expected map[string]any, got %T: %w", val, ErrConversion)
		}
		return any(m).(T), nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return zero, wrapError("marshal result", ErrMarshal, err)
	}

	switch any(zero).(type) {
	case []byte:
		return any(data).(T), nil
	case string:
		return any(string(data)).(T), nil
	default:
		var target T
		if err := json.Unmarshal(data, &target); err != nil {
			return zero, wrapError("unmarshal result", ErrUnmarshal, err)
		}
		return target, nil
	}
}
