package jsonmerge

import (
	"testing"
)

// BenchmarkMerge benchmarks the merge operation
func BenchmarkMerge(b *testing.B) {
	target := map[string]interface{}{
		"name": "John",
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": 42,
				"array": []interface{}{1, 2, 3, 4, 5},
			},
		},
		"tags": []string{"developer", "golang", "json"},
	}
	patch := map[string]interface{}{
		"name": "Jane",
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": 100,
				"new":   "field",
			},
		},
		"email": "jane@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeWithMutate benchmarks the merge operation with mutation
func BenchmarkMergeWithMutate(b *testing.B) {
	originalTarget := map[string]interface{}{
		"name": "John",
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": 42,
				"array": []interface{}{1, 2, 3, 4, 5},
			},
		},
		"tags": []string{"developer", "golang", "json"},
	}
	patch := map[string]interface{}{
		"name": "Jane",
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": 100,
				"new":   "field",
			},
		},
		"email": "jane@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh copy for each iteration since we're mutating
		target := make(map[string]interface{})
		for k, v := range originalTarget {
			target[k] = v
		}

		_, err := Merge(target, patch, WithMutate(true))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerate benchmarks the Generate function
func BenchmarkGenerate(b *testing.B) {
	original := map[string]interface{}{
		"name": "John",
		"age":  30,
		"nested": map[string]interface{}{
			"value": 42,
		},
	}

	updated := map[string]interface{}{
		"name": "Jane",
		"age":  30,
		"nested": map[string]interface{}{
			"value": 100,
			"new":   "field",
		},
		"email": "jane@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Generate(original, updated)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeStructs benchmarks merge operations with struct types
func BenchmarkMergeStructs(b *testing.B) {
	type User struct {
		Name    string            `json:"name"`
		Email   string            `json:"email,omitempty"`
		Age     int               `json:"age"`
		Profile map[string]string `json:"profile,omitempty"`
		Tags    []string          `json:"tags,omitempty"`
	}

	target := User{
		Name:  "John",
		Email: "john@example.com",
		Age:   30,
		Profile: map[string]string{
			"city":    "New York",
			"country": "USA",
		},
		Tags: []string{"developer", "golang"},
	}

	patch := User{
		Name: "Jane",
		Age:  25,
		Profile: map[string]string{
			"city": "San Francisco",
		},
		Tags: []string{"engineer", "python"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeJSONStrings benchmarks merge operations with JSON strings
func BenchmarkMergeJSONStrings(b *testing.B) {
	target := `{"name":"John","age":30,"nested":{"value":42,"array":[1,2,3,4,5]},"tags":["developer","golang","json"]}`
	patch := `{"name":"Jane","nested":{"value":100,"new":"field"},"email":"jane@example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeJSONBytes benchmarks merge operations with JSON byte slices
func BenchmarkMergeJSONBytes(b *testing.B) {
	target := []byte(`{"name":"John","age":30,"nested":{"value":42,"array":[1,2,3,4,5]},"tags":["developer","golang","json"]}`)
	patch := []byte(`{"name":"Jane","nested":{"value":100,"new":"field"},"email":"jane@example.com"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeDeepNesting benchmarks merge operations with deeply nested structures
func BenchmarkMergeDeepNesting(b *testing.B) {
	target := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": map[string]interface{}{
							"value": "original",
						},
					},
				},
			},
		},
	}

	patch := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": map[string]interface{}{
							"value": "updated",
							"new":   "field",
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMergeLargeArrays benchmarks merge operations with large arrays
func BenchmarkMergeLargeArrays(b *testing.B) {
	// Create a large array
	largeArray := make([]interface{}, 1000)
	for i := range largeArray {
		largeArray[i] = map[string]interface{}{
			"id":    i,
			"value": i * 2,
		}
	}

	target := map[string]interface{}{
		"data": largeArray,
	}

	patch := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":    0,
				"value": 999,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Merge(target, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}
