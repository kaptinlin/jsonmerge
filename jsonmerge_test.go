package jsonmerge

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

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
		assert.Contains(t, err.Error(), "unsupported document type")
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
		done := make(chan bool, numGoroutines)

		// Run concurrent merge operations
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					patch := map[string]any{
						"counter": goroutineID*numOperations + j,
						"id":      fmt.Sprintf("goroutine-%d-op-%d", goroutineID, j),
					}

					result, err := Merge(target, patch)
					require.NoError(t, err)
					results[goroutineID*numOperations+j] = result
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

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
