package main

import (
	"fmt"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	fmt.Println("=== Map Merge Example ===")

	// Original document
	original := map[string]any{
		"name": "John Doe",
		"age":  30,
		"address": map[string]any{
			"city":    "New York",
			"country": "USA",
		},
		"hobbies": []string{"reading", "swimming"},
	}

	// RFC 7386 merge patch
	patch := map[string]any{
		"age": 31, // Update age
		"address": map[string]any{
			"city":    "Boston",      // Update city
			"street":  "123 Main St", // Add street
			"country": nil,           // Delete country (null value)
		},
		"hobbies": []string{"hiking", "coding"}, // Replace entire array
		"email":   "john@example.com",           // Add new field
	}

	// Apply merge
	result, err := jsonmerge.Merge(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:", prettyJSON(original))
	fmt.Println("Patch:   ", prettyJSON(patch))
	fmt.Println("Result:  ", prettyJSON(result.Doc))

	// Demonstrate key RFC 7386 behaviors
	fmt.Println("\n=== Key Behaviors ===")
	fmt.Println("• Objects are merged recursively")
	fmt.Println("• null values delete fields")
	fmt.Println("• Arrays are replaced entirely")
	fmt.Println("• New fields are added")
}

func prettyJSON(v any) string {
	data, _ := json.Marshal(v, jsontext.Multiline(true))
	return string(data)
}
