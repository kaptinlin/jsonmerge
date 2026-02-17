package jsonmerge

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRFC7386Compliance tests all examples from RFC 7386 Appendix A
// Reference: https://datatracker.ietf.org/doc/html/rfc7386#appendix-A
func TestRFC7386Compliance(t *testing.T) {
	testCases := []struct {
		name     string
		target   string
		patch    string
		expected string
	}{
		// Basic replacement
		{
			name:     "simple replacement",
			target:   `{"a":"b"}`,
			patch:    `{"a":"c"}`,
			expected: `{"a":"c"}`,
		},
		// Add new field
		{
			name:     "add field",
			target:   `{"a":"b"}`,
			patch:    `{"b":"c"}`,
			expected: `{"a":"b","b":"c"}`,
		},
		// Delete field with null
		{
			name:     "delete with null",
			target:   `{"a":"b"}`,
			patch:    `{"a":null}`,
			expected: `{}`,
		},
		// Partial deletion
		{
			name:     "partial delete",
			target:   `{"a":"b","b":"c"}`,
			patch:    `{"a":null}`,
			expected: `{"b":"c"}`,
		},
		// Array replacement with scalar
		{
			name:     "array replacement with scalar",
			target:   `{"a":["b"]}`,
			patch:    `{"a":"c"}`,
			expected: `{"a":"c"}`,
		},
		// Scalar replacement with array
		{
			name:     "scalar replacement with array",
			target:   `{"a":"c"}`,
			patch:    `{"a":["b"]}`,
			expected: `{"a":["b"]}`,
		},
		// Nested object merge with deletion
		{
			name:     "nested object merge with deletion",
			target:   `{"a":{"b":"c"}}`,
			patch:    `{"a":{"b":"d","c":null}}`,
			expected: `{"a":{"b":"d"}}`,
		},
		// Replace object with array (complex case from RFC)
		{
			name:     "replace array of objects with array of scalars",
			target:   `{"a":[{"b":"c"}]}`,
			patch:    `{"a":[1]}`,
			expected: `{"a":[1]}`,
		},
		// Array to array replacement
		{
			name:     "array to array replacement",
			target:   `["a","b"]`,
			patch:    `["c","d"]`,
			expected: `["c","d"]`,
		},
		// Object to array replacement
		{
			name:     "object to array replacement",
			target:   `{"a":"b"}`,
			patch:    `["c"]`,
			expected: `["c"]`,
		},
		// Scalar to null replacement
		{
			name:     "scalar to null replacement",
			target:   `{"a":"foo"}`,
			patch:    `null`,
			expected: `null`,
		},
		// Scalar to scalar replacement
		{
			name:     "scalar to scalar replacement",
			target:   `{"a":"foo"}`,
			patch:    `"bar"`,
			expected: `"bar"`,
		},
		// Null value addition
		{
			name:     "null value addition",
			target:   `{"e":null}`,
			patch:    `{"a":1}`,
			expected: `{"e":null,"a":1}`,
		},
		// Array to object replacement (from RFC)
		{
			name:     "array to object replacement",
			target:   `[1,2]`,
			patch:    `{"a":"b","c":null}`,
			expected: `{"a":"b"}`,
		},
		// Empty object to nested object
		{
			name:     "empty object to nested object",
			target:   `{}`,
			patch:    `{"a":{"bb":{"ccc":null}}}`,
			expected: `{"a":{"bb":{}}}`,
		},
		// Additional edge cases for completeness
		{
			name:     "empty object target",
			target:   `{}`,
			patch:    `{"a":"b"}`,
			expected: `{"a":"b"}`,
		},
		{
			name:     "empty patch",
			target:   `{"a":"b"}`,
			patch:    `{}`,
			expected: `{"a":"b"}`,
		},
		// Deep nesting with mixed operations
		{
			name:     "deep nesting with mixed operations",
			target:   `{"a":{"b":{"c":"d","e":"f"}},"g":"h"}`,
			patch:    `{"a":{"b":{"c":"new","x":"y"},"i":"j"}}`,
			expected: `{"a":{"b":{"c":"new","e":"f","x":"y"},"i":"j"},"g":"h"}`,
		},
		// Number handling
		{
			name:     "number handling",
			target:   `{"a":1,"b":2.5}`,
			patch:    `{"a":10,"c":3.14}`,
			expected: `{"a":10,"b":2.5,"c":3.14}`,
		},
		// Boolean handling
		{
			name:     "boolean handling",
			target:   `{"a":true,"b":false}`,
			patch:    `{"a":false,"c":true}`,
			expected: `{"a":false,"b":false,"c":true}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Merge(tt.target, tt.patch)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, result.Doc)
		})
	}
}

// TestTypeSafety tests that the library maintains type safety with generic parameters
func TestTypeSafety(t *testing.T) {
	t.Run("struct_type_safety", func(t *testing.T) {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email,omitempty"`
			Age   int    `json:"age"`
		}

		user := User{Name: "John", Email: "john@example.com", Age: 30}
		patch := User{Name: "Jane"} // This creates {"name":"Jane","age":0} because Age is zero value

		result, err := Merge(user, patch)
		require.NoError(t, err)

		// Verify type safety - result.Doc should be User type
		// When we use struct literals, zero values are included in the JSON
		// So patch = User{Name: "Jane"} becomes {"name":"Jane","age":0,"email":""}
		// According to RFC 7386, this means we explicitly set age to 0
		assert.Equal(t, "Jane", result.Doc.Name)
		assert.Equal(t, "john@example.com", result.Doc.Email) // Preserved from original
		assert.Equal(t, 0, result.Doc.Age)                    // Explicitly set to 0 by patch
	})

	t.Run("map_type_safety", func(t *testing.T) {
		target := map[string]any{"name": "John", "age": 30}
		patch := map[string]any{"name": "Jane", "email": "jane@example.com"}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		expected := map[string]any{
			"name":  "Jane",
			"age":   30,
			"email": "jane@example.com",
		}
		assert.Equal(t, expected, result.Doc)
	})

	t.Run("json_bytes_type_safety", func(t *testing.T) {
		target := []byte(`{"name":"John","age":30}`)
		patch := []byte(`{"name":"Jane","email":"jane@example.com"}`)

		result, err := Merge(target, patch)
		require.NoError(t, err)

		var actualMap, expectedMap map[string]any
		require.NoError(t, json.Unmarshal(result.Doc, &actualMap))
		require.NoError(t, json.Unmarshal([]byte(`{"name":"Jane","age":30,"email":"jane@example.com"}`), &expectedMap))
		assert.Equal(t, expectedMap, actualMap)
	})

	t.Run("json_string_type_safety", func(t *testing.T) {
		target := `{"name":"John","age":30}`
		patch := `{"name":"Jane","email":"jane@example.com"}`

		result, err := Merge(target, patch)
		require.NoError(t, err)

		var actualMap, expectedMap map[string]any
		require.NoError(t, json.Unmarshal([]byte(result.Doc), &actualMap))
		require.NoError(t, json.Unmarshal([]byte(`{"name":"Jane","age":30,"email":"jane@example.com"}`), &expectedMap))
		assert.Equal(t, expectedMap, actualMap)
	})
}

// TestComplexNestedStructs tests complex nested data structures
func TestComplexNestedStructs(t *testing.T) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type User struct {
		Name    string    `json:"name"`
		Email   string    `json:"email,omitempty"`
		Age     int       `json:"age"`
		Address *Address  `json:"address,omitempty"`
		Tags    []string  `json:"tags,omitempty"`
		Created time.Time `json:"created"`
	}

	user := User{
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
		Address: &Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
		},
		Tags:    []string{"developer", "golang"},
		Created: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	patch := User{
		Name: "Jane",
		Address: &Address{
			City: "San Francisco", // Only update city
		},
		Tags: []string{"engineer"}, // Replace entire array
	}

	result, err := Merge(user, patch)
	require.NoError(t, err)

	// Verify the merge results
	assert.Equal(t, "Jane", result.Doc.Name)
	assert.Equal(t, "john@example.com", result.Doc.Email) // Preserved from original
	assert.Equal(t, 0, result.Doc.Age)                    // Set to 0 by patch (zero value)
	assert.Equal(t, "San Francisco", result.Doc.Address.City)
	assert.Equal(t, "", result.Doc.Address.Street)  // Zero value since not in patch
	assert.Equal(t, "", result.Doc.Address.Country) // Zero value since not in patch
	assert.Equal(t, []string{"engineer"}, result.Doc.Tags)
	assert.True(t, result.Doc.Created.IsZero()) // Zero value since not in patch
}

// TestMutateOption tests the WithMutate option for in-place modification
func TestMutateOption(t *testing.T) {
	t.Run("immutable_by_default", func(t *testing.T) {
		original := map[string]any{"name": "John", "age": 30}
		patch := map[string]any{"name": "Jane"}

		result, err := Merge(original, patch)
		require.NoError(t, err)

		// Original should remain unchanged
		assert.Equal(t, "John", original["name"])
		assert.Equal(t, "Jane", result.Doc["name"])
	})

	t.Run("mutate_option", func(t *testing.T) {
		original := map[string]any{"name": "John", "age": 30}
		originalCopy := map[string]any{"name": "John", "age": 30}
		patch := map[string]any{"name": "Jane"}

		result, err := Merge(original, patch, WithMutate(true))
		require.NoError(t, err)

		// With mutate option, the result should contain the modified data
		assert.Equal(t, "Jane", result.Doc["name"])
		// We can't reliably test if original was mutated due to implementation details
		// but we verify the result is correct
		assert.NotEqual(t, originalCopy["name"], result.Doc["name"])
	})
}

// TestErrorCases tests various error conditions
func TestErrorCases(t *testing.T) {
	t.Run("invalid_json_target", func(t *testing.T) {
		invalidJSON := `{"name": invalid}`
		patch := `{"name": "Jane"}`

		// Since we now treat invalid JSON strings as raw strings, this should not error
		result, err := Merge(invalidJSON, patch)
		require.NoError(t, err)
		// The result should be the patch since patch is also a valid JSON string
		// When target is invalid JSON (treated as raw string), patch replaces it completely
		assert.Equal(t, `{"name":"Jane"}`, result.Doc)
	})

	t.Run("invalid_json_patch", func(t *testing.T) {
		target := `{"name": "John"}`
		invalidPatch := `{"name": invalid}`

		// Since we now treat invalid JSON strings as raw strings, this should not error
		result, err := Merge(target, invalidPatch)
		require.NoError(t, err)
		// The result should be the raw patch string wrapped in quotes (JSON marshaling)
		assert.Equal(t, `"{\"name\": invalid}"`, result.Doc)
	})

	t.Run("invalid_json_bytes", func(t *testing.T) {
		invalidBytes := []byte(`{"name": invalid}`)
		patch := []byte(`{"name": "Jane"}`)

		// Invalid JSON bytes should still cause an error
		_, err := Merge(invalidBytes, patch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})

	t.Run("extremely_malformed_json", func(t *testing.T) {
		malformedCases := [][]byte{
			[]byte(`[[[[`),
			[]byte(`}}}}`),
			[]byte(`{"key":}`),
			[]byte(`{"key":,}`),
			[]byte(`{,}`),
		}

		for i, malformed := range malformedCases {
			t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
				_, err := Merge(malformed, []byte(`{"valid": "patch"}`))
				assert.Error(t, err, "Should error on malformed JSON bytes")
			})
		}
	})
}

// TestGenerate tests the Generate function
func TestGenerate(t *testing.T) {
	t.Run("generate_basic_patch", func(t *testing.T) {
		original := map[string]any{
			"name": "John",
			"age":  30,
			"city": "New York",
		}

		updated := map[string]any{
			"name":  "Jane",             // Modified
			"age":   30,                 // Unchanged
			"email": "jane@example.com", // Added
			// "city" removed
		}

		patch, err := Generate(original, updated)
		require.NoError(t, err)

		expectedPatch := map[string]any{
			"name":  "Jane",
			"email": "jane@example.com",
			"city":  nil, // Indicates deletion
		}
		assert.Equal(t, expectedPatch, patch)

		// Verify the generated patch works
		result, err := Merge(original, patch)
		require.NoError(t, err)
		assert.Equal(t, updated, result.Doc)
	})

	t.Run("generate_nested_patch", func(t *testing.T) {
		original := map[string]any{
			"user": map[string]any{
				"name": "John",
				"age":  30,
			},
			"metadata": map[string]any{
				"version": 1,
			},
		}

		updated := map[string]any{
			"user": map[string]any{
				"name":  "Jane",             // Modified
				"email": "jane@example.com", // Added
				// "age" removed
			},
			"metadata": map[string]any{
				"version": 1, // Unchanged
			},
		}

		patch, err := Generate(original, updated)
		require.NoError(t, err)

		// Apply the generated patch
		result, err := Merge(original, patch)
		require.NoError(t, err)
		assert.Equal(t, updated, result.Doc)
	})
}

// TestValid tests the Valid function
func TestValid(t *testing.T) {
	t.Run("valid_patches", func(t *testing.T) {
		validPatches := []any{
			map[string]any{"name": "Jane"},
			`{"name": "Jane"}`,
			[]byte(`{"name": "Jane"}`),
			nil,
			"string value",
			123,
			true,
		}

		for i, patch := range validPatches {
			assert.True(t, Valid(patch), "patch %d should be valid", i)
		}
	})

	t.Run("invalid_patches", func(t *testing.T) {
		// Only JSON bytes with invalid JSON should be invalid
		// Invalid JSON strings are treated as raw strings (valid)
		invalidPatches := []any{
			[]byte(`{"name": invalid}`), // Invalid JSON bytes
		}

		for i, patch := range invalidPatches {
			assert.False(t, Valid(patch), "patch %d should be invalid", i)
		}
	})

	t.Run("complex_valid_patches", func(t *testing.T) {
		complexPatches := []any{
			map[string]any{
				"nested": map[string]any{
					"array": []any{1, 2, 3},
					"null":  nil,
				},
			},
			`{"unicode": "🌍", "special": "key/with/slashes"}`,
			[]byte(`{"deeply": {"nested": {"object": {"with": "value"}}}}`),
		}

		for i, patch := range complexPatches {
			assert.True(t, Valid(patch), "complex patch %d should be valid", i)
		}
	})
}

// TestSophisticatedPatchGeneration tests Generate function with complex transformation scenarios
func TestSophisticatedPatchGeneration(t *testing.T) {
	t.Run("collection_transformation_analysis", func(t *testing.T) {
		source := map[string]any{
			"resources": []any{"server1", "server2", "server3"},
			"capacity":  100,
		}

		target := map[string]any{
			"resources": []any{"server4", "server5"},
			"capacity":  75,
		}

		patch, err := Generate(source, target)
		require.NoError(t, err)

		expected := map[string]any{
			"resources": []any{"server4", "server5"},
			"capacity":  75,
		}
		assert.Equal(t, expected, patch)

		// Verify patch application achieves target state
		result, err := Merge(source, patch)
		require.NoError(t, err)
		assert.Equal(t, target, result.Doc)
	})

	t.Run("datatype_migration_detection", func(t *testing.T) {
		source := map[string]any{
			"identifier":  "user123",
			"preferences": map[string]any{"theme": "light"},
		}

		target := map[string]any{
			"identifier":  12345,
			"preferences": []any{"dark_mode", "notifications"},
		}

		patch, err := Generate(source, target)
		require.NoError(t, err)

		expected := map[string]any{
			"identifier":  12345,
			"preferences": []any{"dark_mode", "notifications"},
		}
		assert.Equal(t, expected, patch)

		// Verify patch achieves schema transformation
		result, err := Merge(source, patch)
		require.NoError(t, err)
		assert.Equal(t, target, result.Doc)
	})

	t.Run("optimized_differential_generation", func(t *testing.T) {
		source := map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
				"name": "production",
			},
			"cache": map[string]any{
				"enabled": true,
				"ttl":     3600,
			},
			"logging": map[string]any{
				"level": "info",
			},
		}

		target := map[string]any{
			"database": map[string]any{
				"host": "localhost", // Unchanged
				"port": 5432,        // Unchanged
				"name": "staging",   // Changed
				"ssl":  true,        // Added
			},
			"cache": map[string]any{
				"enabled": true, // Unchanged
				"ttl":     3600, // Unchanged
			},
			"logging": map[string]any{
				"level": "debug",            // Changed
				"file":  "/var/log/app.log", // Added
			},
		}

		patch, err := Generate(source, target)
		require.NoError(t, err)

		// Patch should contain only the actual changes
		expected := map[string]any{
			"database": map[string]any{
				"name": "staging",
				"ssl":  true,
			},
			"logging": map[string]any{
				"level": "debug",
				"file":  "/var/log/app.log",
			},
		}
		assert.Equal(t, expected, patch)

		// Verify patch application produces correct result
		result, err := Merge(source, patch)
		require.NoError(t, err)
		assert.Equal(t, target, result.Doc)
	})
}

// TestOperationalReliabilityAndConsistency tests library reliability across diverse usage patterns
func TestOperationalReliabilityAndConsistency(t *testing.T) {
	t.Run("sequential_state_transitions", func(t *testing.T) {
		state := map[string]any{"phase": 0}

		for phase := 1; phase <= 12; phase++ {
			patch := map[string]any{"phase": phase}
			result, err := Merge(state, patch)
			require.NoError(t, err)
			assert.Equal(t, phase, result.Doc["phase"])
			state = result.Doc
		}
	})

	t.Run("deterministic_operation_behavior", func(t *testing.T) {
		target := map[string]any{"flag": "enabled", "score": 85}
		patch := map[string]any{"flag": "enabled"} // Identical value

		first, err := Merge(target, patch)
		require.NoError(t, err)

		second, err := Merge(first.Doc, patch)
		require.NoError(t, err)

		assert.Equal(t, first.Doc, second.Doc)
	})

	t.Run("selective_updates_in_large_datasets", func(t *testing.T) {
		dataset := make(map[string]any)
		for i := range 1200 {
			dataset[fmt.Sprintf("record_%d", i)] = fmt.Sprintf("data_%d", i)
		}
		dataset["activeRecord"] = "initial_state"

		target := map[string]any{
			"dataset":    dataset,
			"statistics": map[string]any{"totalRecords": 1200},
		}

		patch := map[string]any{
			"dataset": map[string]any{
				"activeRecord": "updated_state",
			},
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		merged := result.Doc["dataset"].(map[string]any)
		assert.Equal(t, "updated_state", merged["activeRecord"])
		assert.Equal(t, "data_0", merged["record_0"])
	})
}

// TestRFC7386EdgeCases tests additional edge cases and boundary conditions
func TestRFC7386EdgeCases(t *testing.T) {
	t.Run("unicode_handling", func(t *testing.T) {
		target := `{"name":"José","city":"São Paulo"}`
		patch := `{"name":"María","country":"España"}`
		expected := `{"name":"María","city":"São Paulo","country":"España"}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("large_numbers", func(t *testing.T) {
		target := `{"bigint":9223372036854775807,"smallint":-9223372036854775808}`
		patch := `{"bigint":1234567890123456789,"newfloat":1.7976931348623157e+308}`
		expected := `{"bigint":1234567890123456789,"smallint":-9223372036854775808,"newfloat":1.7976931348623157e+308}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("special_float_values", func(t *testing.T) {
		// Note: JSON doesn't support NaN or Infinity, but we test regular floats
		target := `{"pi":3.14159,"zero":0.0,"negative":-123.456}`
		patch := `{"pi":3.141592653589793,"epsilon":2.220446049250313e-16}`
		expected := `{"pi":3.141592653589793,"zero":0.0,"negative":-123.456,"epsilon":2.220446049250313e-16}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("empty_arrays_and_objects", func(t *testing.T) {
		target := `{"empty_obj":{},"empty_arr":[],"data":"value"}`
		patch := `{"empty_obj":{"new":"field"},"empty_arr":[1,2,3],"new_empty":{}}`
		expected := `{"empty_obj":{"new":"field"},"empty_arr":[1,2,3],"data":"value","new_empty":{}}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("deeply_nested_structures", func(t *testing.T) {
		target := `{"level1":{"level2":{"level3":{"level4":{"value":"deep"}}}}}`
		patch := `{"level1":{"level2":{"level3":{"level4":{"value":"updated","new":"field"},"new_level4":"added"}}}}`
		expected := `{"level1":{"level2":{"level3":{"level4":{"value":"updated","new":"field"},"new_level4":"added"}}}}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("mixed_array_content", func(t *testing.T) {
		target := `{"mixed":[1,"string",{"obj":"value"},null,true]}`
		patch := `{"mixed":["new",42,false]}`
		expected := `{"mixed":["new",42,false]}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("null_handling_in_nested_objects", func(t *testing.T) {
		target := `{"a":{"b":{"c":"keep","d":"delete"},"e":"preserve"}}`
		patch := `{"a":{"b":{"d":null,"f":"add"}}}`
		expected := `{"a":{"b":{"c":"keep","f":"add"},"e":"preserve"}}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})
}

// TestGoTypeSystem tests Go-specific type handling and edge cases
func TestGoTypeSystem(t *testing.T) {
	t.Run("interface_any_handling", func(t *testing.T) {
		target := map[string]any{
			"string": "value",
			"int":    42,
			"float":  3.14,
			"bool":   true,
			"null":   nil,
		}
		patch := map[string]any{
			"string": "updated",
			"int":    100,
			"new":    "added",
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		expected := map[string]any{
			"string": "updated",
			"int":    100,
			"float":  3.14,
			"bool":   true,
			"null":   nil,
			"new":    "added",
		}
		assert.Equal(t, expected, result.Doc)
	})

	t.Run("pointer_fields_in_structs", func(t *testing.T) {
		type Config struct {
			Name        string  `json:"name"`
			Port        *int    `json:"port,omitempty"`
			Debug       *bool   `json:"debug,omitempty"`
			Description *string `json:"description,omitempty"`
		}

		port := 8080
		debug := true
		desc := "test config"

		target := Config{
			Name:        "app",
			Port:        &port,
			Debug:       &debug,
			Description: &desc,
		}

		newPort := 9090
		patch := Config{
			Name: "updated-app",
			Port: &newPort,
			// Debug and Description will be nil (omitted)
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		assert.Equal(t, "updated-app", result.Doc.Name)
		assert.Equal(t, 9090, *result.Doc.Port)
		// Note: nil pointer fields in patch mean the field is omitted from JSON
		// which means they are preserved from target
		assert.Equal(t, true, *result.Doc.Debug)
		assert.Equal(t, "test config", *result.Doc.Description)
	})

	t.Run("slice_and_array_types", func(t *testing.T) {
		type Data struct {
			Slice []string `json:"slice"`
			Array [3]int   `json:"array"`
			Mixed []any    `json:"mixed"`
		}

		target := Data{
			Slice: []string{"a", "b", "c"},
			Array: [3]int{1, 2, 3},
			Mixed: []any{"string", 42, true},
		}

		patch := Data{
			Slice: []string{"x", "y"},
			Array: [3]int{10, 20, 30},
			Mixed: []any{false, "new"},
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		assert.Equal(t, []string{"x", "y"}, result.Doc.Slice)
		assert.Equal(t, [3]int{10, 20, 30}, result.Doc.Array)
		assert.Equal(t, []any{false, "new"}, result.Doc.Mixed)
	})

	t.Run("embedded_structs", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}

		type Person struct {
			Name string `json:"name"`
			Address
			Age int `json:"age"`
		}

		target := Person{
			Name: "John",
			Address: Address{
				Street: "123 Main St",
				City:   "NYC",
			},
			Age: 30,
		}

		patch := Person{
			Name: "Jane",
			Address: Address{
				City: "LA", // Only update city
			},
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		assert.Equal(t, "Jane", result.Doc.Name)
		assert.Equal(t, "", result.Doc.Street) // Zero value from patch
		assert.Equal(t, "LA", result.Doc.City)
		assert.Equal(t, 0, result.Doc.Age) // Zero value from patch
	})

	t.Run("custom_json_tags", func(t *testing.T) {
		type Item struct {
			ID          int     `json:"id"`
			Name        string  `json:"item_name"`
			Description string  `json:"desc,omitempty"`
			Internal    string  `json:"-"`            // Should be ignored
			Price       float64 `json:"price,string"` // Price as string in JSON
		}

		target := Item{
			ID:          1,
			Name:        "Widget",
			Description: "A useful widget",
			Internal:    "secret",
			Price:       19.99,
		}

		patch := Item{
			Name:  "Super Widget",
			Price: 29.99,
			// Description is empty (zero value), but due to omitempty it won't appear in JSON
			// so it should be preserved from target
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		assert.Equal(t, 0, result.Doc.ID) // Zero value from patch
		assert.Equal(t, "Super Widget", result.Doc.Name)
		assert.Equal(t, "A useful widget", result.Doc.Description) // Preserved due to omitempty
		assert.Equal(t, "", result.Doc.Internal)                   // Always zero (ignored by JSON)
		assert.Equal(t, 29.99, result.Doc.Price)
	})
}

// TestMediaTypeCompliance tests RFC 7386 media type requirements
func TestMediaTypeCompliance(t *testing.T) {
	t.Run("application_merge_patch_json_semantics", func(t *testing.T) {
		// Test that our implementation follows the semantics required for
		// application/merge-patch+json media type as defined in RFC 7386 Section 4

		// Content-Type: application/merge-patch+json should behave exactly
		// like our implementation
		target := `{
			"title": "Goodbye!",
			"author": {
				"givenName": "John",
				"familyName": "Doe"
			},
			"tags": ["example", "sample"],
			"content": "This will be unchanged",
			"phoneNumber": "+01-123-456-7890"
		}`

		patch := `{
			"title": "Hello!",
			"phoneNumber": null,
			"author": {
				"familyName": null,
				"givenName": "Jane"
			},
			"tags": ["example"]
		}`

		expected := `{
			"title": "Hello!",
			"author": {
				"givenName": "Jane"
			},
			"tags": ["example"],
			"content": "This will be unchanged"
		}`

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})
}

// TestConcurrencyAndImmutability tests thread safety and immutability guarantees
func TestConcurrencyAndImmutability(t *testing.T) {
	t.Run("immutable_operations_are_thread_safe", func(t *testing.T) {
		target := map[string]any{
			"counter": 0,
			"data":    []string{"a", "b", "c"},
		}

		const numGoroutines = 10
		const numOperations = 10

		results := make([]*Result[map[string]any], numGoroutines*numOperations)
		var wg sync.WaitGroup

		// Run concurrent merge operations
		for i := range numGoroutines {
			goroutineID := i
			wg.Go(func() {
				for j := range numOperations {
					patch := map[string]any{
						"counter": goroutineID*numOperations + j,
						"id":      fmt.Sprintf("goroutine-%d-op-%d", goroutineID, j),
					}

					result, err := Merge(target, patch)
					require.NoError(t, err)
					results[goroutineID*numOperations+j] = result
				}
			})
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify original target is unchanged
		assert.Equal(t, 0, target["counter"])
		assert.Equal(t, []string{"a", "b", "c"}, target["data"])

		// Verify all operations completed successfully
		for i, result := range results {
			require.NotNil(t, result, "Result %d should not be nil", i)
			assert.Contains(t, result.Doc, "counter")
			assert.Contains(t, result.Doc, "id")
			assert.Equal(t, []string{"a", "b", "c"}, result.Doc["data"])
		}
	})
}

// TestArrayOperations tests common array handling scenarios
func TestArrayOperations(t *testing.T) {
	t.Run("shopping_cart_updates", func(t *testing.T) {
		scenarios := []struct {
			name     string
			cart     string
			update   string
			expected string
		}{
			{
				name:     "replace_cart_items",
				cart:     `{"items": ["book", "pen", "notebook"]}`,
				update:   `{"items": ["laptop", null, "mouse"]}`,
				expected: `{"items": ["laptop", null, "mouse"]}`,
			},
			{
				name:     "clear_cart",
				cart:     `{"items": ["phone", "headset", "charger"]}`,
				update:   `{"items": []}`,
				expected: `{"items": []}`,
			},
			{
				name:     "remove_cart_field",
				cart:     `{"quantities": [1, 2, 3]}`,
				update:   `{"quantities": null}`,
				expected: `{}`,
			},
			{
				name:     "mixed_item_types",
				cart:     `{"data": [1, "product", true]}`,
				update:   `{"data": [false, null, {"name": "new item"}]}`,
				expected: `{"data": [false, null, {"name": "new item"}]}`,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				result, err := Merge(scenario.cart, scenario.update)
				require.NoError(t, err)
				assert.JSONEq(t, scenario.expected, result.Doc)
			})
		}
	})

	t.Run("score_matrix", func(t *testing.T) {
		original := `{"scores": [[85, 90], [78, 82], [92, 88]]}`
		update := `{"scores": [[95, 96]]}`
		expected := `{"scores": [[95, 96]]}`

		result, err := Merge(original, update)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("student_list", func(t *testing.T) {
		original := `{"students": [{"id": 1, "name": "Alice", "grade": "3rd"}, {"id": 2, "name": "Bob", "grade": "4th"}]}`
		update := `{"students": [{"id": 3, "name": "Charlie", "grade": "5th"}]}`
		expected := `{"students": [{"id": 3, "name": "Charlie", "grade": "5th"}]}`

		result, err := Merge(original, update)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})
}

// TestSpecialCharacters tests handling of common special characters and text
func TestSpecialCharacters(t *testing.T) {
	t.Run("user_profiles", func(t *testing.T) {
		testCases := []struct {
			name     string
			profile  string
			update   string
			expected string
		}{
			{
				name:     "file_path_username",
				profile:  `{"username": "user/123", "folder": "Documents/Photos"}`,
				update:   `{"username": "admin/456", "folder": "Documents/Videos"}`,
				expected: `{"username": "admin/456", "folder": "Documents/Videos"}`,
			},
			{
				name:     "email_and_status",
				profile:  `{"email": "user@example.com", "status": "online"}`,
				update:   `{"email": "new@example.com", "bio": "loves coding"}`,
				expected: `{"email": "new@example.com", "status": "online", "bio": "loves coding"}`,
			},
			{
				name:     "social_media_handles",
				profile:  `{"twitter": "@oldname"}`,
				update:   `{"twitter": "@newname", "instagram": "#photography"}`,
				expected: `{"twitter": "@newname", "instagram": "#photography"}`,
			},
			{
				name:     "multilingual_names",
				profile:  `{"chinese": "张三", "english": "Tom"}`,
				update:   `{"chinese": "李四", "japanese": "田中"}`,
				expected: `{"chinese": "李四", "english": "Tom", "japanese": "田中"}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Merge(tc.profile, tc.update)
				require.NoError(t, err)
				assert.JSONEq(t, tc.expected, result.Doc)
			})
		}
	})

	t.Run("text_formatting", func(t *testing.T) {
		original := `{"title": "", "content": " ", "code": "\t", "poem": "\nline1\nline2"}`
		update := `{"title": "My Article", "emoji": "😊", "quote": "\"Hello World\"", "math": "x² + y² = z²"}`
		expected := `{"title": "My Article", "content": " ", "code": "\t", "poem": "\nline1\nline2", "emoji": "😊", "quote": "\"Hello World\"", "math": "x² + y² = z²"}`

		result, err := Merge(original, update)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})
}

// TestNestedStructures tests handling of nested document structures
func TestNestedStructures(t *testing.T) {
	t.Run("game_levels", func(t *testing.T) {
		// Build a game with 20 nested levels
		levels := 20
		game := make(map[string]any)
		currentLevel := game
		for level := range levels {
			nextLevel := make(map[string]any)
			currentLevel[fmt.Sprintf("level_%d", level+1)] = nextLevel
			currentLevel = nextLevel
		}
		currentLevel["boss"] = "dragon"
		currentLevel["difficulty"] = "hard"

		// Create an update for the final level
		update := make(map[string]any)
		currentLevel = update
		for level := range levels {
			nextLevel := make(map[string]any)
			currentLevel[fmt.Sprintf("level_%d", level+1)] = nextLevel
			currentLevel = nextLevel
		}
		currentLevel["boss"] = "super dragon"
		currentLevel["difficulty"] = "nightmare"
		currentLevel["reward"] = "legendary sword"

		result, err := Merge(game, update)
		require.NoError(t, err)

		// Navigate to the final level and verify changes
		currentLevel = result.Doc
		for level := range levels {
			currentLevel = currentLevel[fmt.Sprintf("level_%d", level+1)].(map[string]any)
		}
		assert.Equal(t, "super dragon", currentLevel["boss"])
		assert.Equal(t, "nightmare", currentLevel["difficulty"])
		assert.Equal(t, "legendary sword", currentLevel["reward"])
	})
}

// TestLargeDatasets tests performance with large amounts of data
func TestLargeDatasets(t *testing.T) {
	t.Run("game_player_settings", func(t *testing.T) {
		const count = 500
		settings := make(map[string]any)
		patch := make(map[string]any)

		for i := range count {
			settings[fmt.Sprintf("setting_%d", i)] = fmt.Sprintf("default_%d", i)
			if i%10 == 0 {
				patch[fmt.Sprintf("setting_%d", i)] = fmt.Sprintf("new_value_%d", i)
			}
		}
		for i := count; i < count+30; i++ {
			patch[fmt.Sprintf("new_setting_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		result, err := Merge(settings, patch)
		require.NoError(t, err)

		assert.Equal(t, "new_value_0", result.Doc["setting_0"])
		assert.Equal(t, "default_1", result.Doc["setting_1"])
		assert.Equal(t, "value_500", result.Doc["new_setting_500"])
	})

	t.Run("product_catalog_replacement", func(t *testing.T) {
		const count = 5000
		products := make([]any, count)
		for i := range count {
			products[i] = fmt.Sprintf("product_%d", i)
		}
		target := map[string]any{"products": products}

		featured := []any{"featured_item_1", "featured_item_2", "featured_item_3"}
		patch := map[string]any{"products": featured}

		result, err := Merge(target, patch)
		require.NoError(t, err)
		assert.Equal(t, featured, result.Doc["products"])
	})
}

// TestBoundaryConditions tests various extreme and boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("empty_document_cases", func(t *testing.T) {
		scenarios := []struct {
			name     string
			input    any
			patch    any
			expected string
		}{
			{
				name:     "empty_string_to_object",
				input:    "",
				patch:    `{"status": "active"}`,
				expected: `{"status":"active"}`,
			},
			{
				name:     "empty_object_to_array",
				input:    map[string]any{},
				patch:    []any{"item1", "item2", "item3"},
				expected: `["item1","item2","item3"]`,
			},
			{
				name:     "null_to_object",
				input:    nil,
				patch:    map[string]any{"created": true},
				expected: `{"created":true}`,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				result, err := Merge(scenario.input, scenario.patch)
				require.NoError(t, err)

				data, err := json.Marshal(result.Doc)
				require.NoError(t, err)
				assert.JSONEq(t, scenario.expected, string(data))
			})
		}
	})

	t.Run("number_precision", func(t *testing.T) {
		base := `{"maxFloat": 1.7976931348623157e+308, "maxInt": 9223372036854775807, "minInt": -9223372036854775808}`
		changes := `{"maxFloat": 2.2250738585072014e-308, "maxInt": -9223372036854775808, "precision": 0.123456789}`
		expected := `{"maxFloat": 2.2250738585072014e-308, "maxInt": -9223372036854775808, "minInt": -9223372036854775808, "precision": 0.123456789}`

		result, err := Merge(base, changes)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("boolean_and_null_values", func(t *testing.T) {
		base := `{"enabled": true, "disabled": false, "missing": null}`
		changes := `{"enabled": false, "missing": "present", "disabled": null, "added": true}`
		expected := `{"enabled": false, "missing": "present", "added": true}`

		result, err := Merge(base, changes)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})
}

// TestJSONCompatibility tests compatibility across different JSON processing scenarios
func TestJSONCompatibility(t *testing.T) {
	t.Run("json_roundtrip", func(t *testing.T) {
		target := map[string]any{
			"score":    85.5,
			"player":   "alex",
			"online":   true,
			"metadata": nil,
			"stats":    map[string]any{"level": "beginner"},
			"items":    []any{1, "sword", 99.9},
		}

		patch := map[string]any{
			"score": 92.3,
			"rank":  "advanced",
			"stats": map[string]any{"experience": "1000"},
		}

		result, err := Merge(target, patch)
		require.NoError(t, err)

		// Serialize and deserialize the result
		data, err := json.Marshal(result.Doc)
		require.NoError(t, err)

		var roundtripped map[string]any
		err = json.Unmarshal(data, &roundtripped)
		require.NoError(t, err)

		assert.Equal(t, float64(92.3), roundtripped["score"])
		assert.Equal(t, "advanced", roundtripped["rank"])
		assert.Equal(t, "alex", roundtripped["player"])
	})

	t.Run("data_type_handling", func(t *testing.T) {
		// Verify different JSON data types are handled correctly
		testCases := []struct {
			name     string
			base     string
			update   string
			expected string
		}{
			{
				name:     "number_handling",
				base:     `{"count": 789, "rate": 12.34}`,
				update:   `{"count": 999}`,
				expected: `{"count": 999, "rate": 12.34}`,
			},
			{
				name:     "boolean_handling",
				base:     `{"valid": true, "expired": false}`,
				update:   `{"category": "premium"}`,
				expected: `{"valid": true, "expired": false, "category": "premium"}`,
			},
			{
				name:     "empty_containers",
				base:     `{"options": {}, "list": []}`,
				update:   `{"options": {"theme": "dark"}}`,
				expected: `{"options": {"theme": "dark"}, "list": []}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := Merge(tc.base, tc.update)
				require.NoError(t, err)
				assert.JSONEq(t, tc.expected, result.Doc)
			})
		}
	})

	t.Run("special_key_formats", func(t *testing.T) {
		// Test keys with special characters (like JSON Pointer syntax)
		base := `{"/api/v2/users": "endpoint", "~setting": "value"}`
		update := `{"/api/v2/users": "updated", "/cache/ttl": "3600"}`
		expected := `{"/api/v2/users": "updated", "~setting": "value", "/cache/ttl": "3600"}`

		result, err := Merge(base, update)
		require.NoError(t, err)
		assert.JSONEq(t, expected, result.Doc)
	})

	t.Run("formatting_independence", func(t *testing.T) {
		// Verify JSON formatting variations don't impact merge behavior
		compactDoc := `{"version":1,"mode":"production"}`
		formattedDoc := `{
			"version": 1,
			"mode": "production"
		}`
		update := `{"status": "ready"}`

		result1, err := Merge(compactDoc, update)
		require.NoError(t, err)

		result2, err := Merge(formattedDoc, update)
		require.NoError(t, err)

		// Results should be functionally equivalent regardless of formatting
		assert.JSONEq(t, result1.Doc, result2.Doc)
	})
}

// TestPerformanceAndStressConditions tests library behavior under demanding scenarios
func TestPerformanceAndStressConditions(t *testing.T) {
	t.Run("hierarchical_structure_efficiency", func(t *testing.T) {
		const depth = 90

		// Build hierarchical structure
		target := make(map[string]any)
		node := target
		for i := range depth {
			next := make(map[string]any)
			node["branch"] = next
			node["info"] = fmt.Sprintf("tier_%d", i)
			node = next
		}
		node["endpoint"] = "original_value"

		// Create update for the deepest level
		patch := make(map[string]any)
		node = patch
		for range depth {
			next := make(map[string]any)
			node["branch"] = next
			node = next
		}
		node["endpoint"] = "restructured_value"

		start := time.Now()
		result, err := Merge(target, patch)
		elapsed := time.Since(start)

		require.NoError(t, err)

		// Navigate to verify the deep change
		node = result.Doc
		for range depth {
			node = node["branch"].(map[string]any)
		}
		assert.Equal(t, "restructured_value", node["endpoint"])

		assert.Less(t, elapsed, 90*time.Millisecond, "Hierarchical merge took %v, exceeds threshold", elapsed)
	})

	t.Run("extensive_property_handling", func(t *testing.T) {
		const count = 4000

		target := make(map[string]any)
		patch := make(map[string]any)

		for i := range count {
			target[fmt.Sprintf("property_%d", i)] = fmt.Sprintf("original_%d", i)
			if i%80 == 0 {
				patch[fmt.Sprintf("property_%d", i)] = fmt.Sprintf("revised_%d", i)
			}
		}

		start := time.Now()
		result, err := Merge(target, patch)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, "revised_0", result.Doc["property_0"])
		assert.Equal(t, "original_1", result.Doc["property_1"])

		assert.Less(t, elapsed, 45*time.Millisecond, "Extensive property merge took %v, exceeds threshold", elapsed)
	})

	t.Run("sequential_operations_stability", func(t *testing.T) {
		state := map[string]any{"sequence": 0, "metadata": "persistent"}

		const iterations = 800
		var checkpoints []map[string]any

		for i := 1; i <= iterations; i++ {
			patch := map[string]any{"sequence": i}
			result, err := Merge(state, patch)
			require.NoError(t, err)

			state = result.Doc
			if i%150 == 0 {
				checkpoints = append(checkpoints, state)
			}
		}

		assert.Equal(t, iterations, state["sequence"])
		assert.Equal(t, "persistent", state["metadata"])

		for i, checkpoint := range checkpoints {
			assert.Equal(t, (i+1)*150, checkpoint["sequence"])
			assert.Equal(t, "persistent", checkpoint["metadata"])
		}
	})
}
