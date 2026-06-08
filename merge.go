// Package jsonmerge implements RFC 7386 JSON Merge Patch for Go.
//
// Patches are explicit Patch values. Plain Go strings are JSON string values;
// use []byte or JSON when carrying encoded JSON text.
package jsonmerge

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/kaptinlin/deepclone"
)

var (
	// ErrInvalidJSON indicates encoded JSON text is malformed.
	ErrInvalidJSON = errors.New("invalid json")

	// ErrInvalidValue indicates a Go value cannot be represented as JSON.
	ErrInvalidValue = errors.New("invalid json value")

	// ErrCannotRepresent indicates a JSON result cannot be projected into the requested Go type.
	ErrCannotRepresent = errors.New("cannot represent json value as requested type")
)

// Parse parses encoded JSON text as a merge patch.
func Parse(data []byte) (Patch, error) {
	value, err := parseJSON(data)
	if err != nil {
		return Patch{}, fmt.Errorf("parse patch: %w", err)
	}
	return Patch{value: value}, nil
}

// NewPatch converts value into a canonical merge patch.
func NewPatch(value any) (Patch, error) {
	canonical, err := canonicalize(value)
	if err != nil {
		return Patch{}, fmt.Errorf("new patch: %w", err)
	}
	return Patch{value: canonical}, nil
}

// Apply applies patch to target according to RFC 7386.
func Apply[T any](target T, patch Patch) (T, error) {
	var zero T

	targetValue, err := canonicalize(target)
	if err != nil {
		return zero, fmt.Errorf("canonicalize target: %w", err)
	}

	merged, err := applyPatch(targetValue, patch.value)
	if err != nil {
		return zero, fmt.Errorf("apply patch: %w", err)
	}

	result, err := project[T](merged)
	if err != nil {
		return zero, fmt.Errorf("project result: %w", err)
	}

	return result, nil
}

// Diff returns a merge patch that transforms source into target.
func Diff(source, target any) (Patch, error) {
	sourceValue, err := canonicalize(source)
	if err != nil {
		return Patch{}, fmt.Errorf("canonicalize source: %w", err)
	}

	targetValue, err := canonicalize(target)
	if err != nil {
		return Patch{}, fmt.Errorf("canonicalize target: %w", err)
	}

	patchValue, err := generatePatch(sourceValue, targetValue, true)
	if err != nil {
		return Patch{}, fmt.Errorf("generate patch: %w", err)
	}

	return Patch{value: patchValue}, nil
}

// MarshalJSON returns the canonical JSON encoding of p.
func (p Patch) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(p.value)
	if err != nil {
		return nil, wrapError("marshal patch", ErrInvalidValue, err)
	}
	return data, nil
}

func canonicalize(value any) (any, error) {
	switch typed := value.(type) {
	case Patch:
		return cloneJSONValue(typed.value)
	case JSON:
		value, err := parseJSON([]byte(typed))
		if err != nil {
			return nil, fmt.Errorf("parse json text: %w", err)
		}
		return value, nil
	case []byte:
		value, err := parseJSON(typed)
		if err != nil {
			return nil, fmt.Errorf("parse json bytes: %w", err)
		}
		return value, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, wrapError("marshal value", ErrInvalidValue, err)
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, wrapError("unmarshal value", ErrInvalidValue, err)
	}
	return result, nil
}

func parseJSON(data []byte) (any, error) {
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, wrapError("unmarshal json text", ErrInvalidJSON, err)
	}
	return result, nil
}

func project[T any](value any) (T, error) {
	var zero T

	switch any(zero).(type) {
	case JSON:
		data, err := json.Marshal(value)
		if err != nil {
			return zero, wrapError("marshal json text result", ErrInvalidValue, err)
		}
		return any(JSON(data)).(T), nil
	case []byte:
		data, err := json.Marshal(value)
		if err != nil {
			return zero, wrapError("marshal bytes result", ErrInvalidValue, err)
		}
		return any(data).(T), nil
	case map[string]any:
		if value == nil {
			return zero, nil
		}
		object, ok := value.(map[string]any)
		if !ok {
			return zero, cannotRepresent("expected object result for map target, got %T", value)
		}
		cloned, err := cloneJSONValue(object)
		if err != nil {
			return zero, err
		}
		return any(cloned).(T), nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return zero, wrapError("marshal result", ErrInvalidValue, err)
	}

	var result T
	if err := json.Unmarshal(data, &result, json.RejectUnknownMembers(true)); err != nil {
		return zero, wrapError("unmarshal result", ErrCannotRepresent, err)
	}

	roundtrip, err := canonicalize(result)
	if err != nil {
		return zero, wrapError("verify result", ErrCannotRepresent, err)
	}
	if !equalJSON(value, roundtrip) {
		return zero, cannotRepresent("projecting into %T changes json value", result)
	}

	return result, nil
}

func applyPatch(target, patch any) (any, error) {
	patchObj, isPatchObject := patch.(map[string]any)
	if !isPatchObject {
		return cloneJSONValue(patch)
	}

	targetObj, targetIsObject := target.(map[string]any)
	if !targetIsObject || targetObj == nil {
		targetObj = make(map[string]any, len(patchObj))
	}

	for name, value := range patchObj {
		if value == nil {
			delete(targetObj, name)
			continue
		}
		merged, err := applyPatch(targetObj[name], value)
		if err != nil {
			return nil, err
		}
		targetObj[name] = merged
	}

	return targetObj, nil
}

func generatePatch(source, target any, preserveEmptyObject bool) (any, error) {
	targetObj, targetIsObject := target.(map[string]any)
	if !targetIsObject {
		return cloneJSONValue(target)
	}

	sourceObj, sourceIsObject := source.(map[string]any)
	if !sourceIsObject {
		return cloneJSONValue(target)
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
			cloned, err := cloneJSONValue(targetValue)
			if err != nil {
				return nil, err
			}
			setPatch(key, cloned)
			continue
		}

		sourceObject, sourceIsObject := sourceValue.(map[string]any)
		targetObject, targetIsObject := targetValue.(map[string]any)
		if sourceIsObject && targetIsObject {
			nestedPatch, err := generatePatch(sourceObject, targetObject, false)
			if err != nil {
				return nil, err
			}
			if nestedPatch != nil {
				setPatch(key, nestedPatch)
			}
			continue
		}

		if !equalJSON(sourceValue, targetValue) {
			cloned, err := cloneJSONValue(targetValue)
			if err != nil {
				return nil, err
			}
			setPatch(key, cloned)
		}
	}

	for key := range sourceObj {
		if _, exists := targetObj[key]; !exists {
			setPatch(key, nil)
		}
	}

	if patch != nil {
		return patch, nil
	}
	if preserveEmptyObject {
		return map[string]any{}, nil
	}
	return nil, nil
}

func equalJSON(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	switch av := a.(type) {
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !equalJSON(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for key, avalue := range av {
			bvalue, ok := bv[key]
			if !ok || !equalJSON(avalue, bvalue) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func cloneJSONValue[T any](value T) (T, error) {
	cloned, err := deepclone.Clone(value)
	if err != nil {
		return value, wrapError("clone json value", ErrInvalidValue, err)
	}
	return cloned, nil
}

func wrapError(stage string, sentinel, err error) error {
	return fmt.Errorf("%s: %w", stage, errors.Join(sentinel, err))
}

func cannotRepresent(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrCannotRepresent)
}
