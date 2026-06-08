// Package main demonstrates map-based JSON Merge Patch usage.
package main

import (
	"fmt"
	"log"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	fmt.Println("=== Map Merge Example ===")

	original := map[string]any{
		"name": "John Doe",
		"age":  30,
		"address": map[string]any{
			"city":    "New York",
			"country": "USA",
		},
		"hobbies": []string{"reading", "swimming"},
	}

	patchValue := map[string]any{
		"age": 31,
		"address": map[string]any{
			"city":    "Boston",
			"street":  "123 Main St",
			"country": nil,
		},
		"hobbies": []string{"hiking", "coding"},
		"email":   "john@example.com",
	}

	patch, err := jsonmerge.NewPatch(patchValue)
	if err != nil {
		log.Fatal(err)
	}

	result, err := jsonmerge.Apply(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	originalJSON, err := prettyJSON(original)
	if err != nil {
		log.Fatal(err)
	}
	patchJSON, err := prettyJSON(patchValue)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, err := prettyJSON(result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:", originalJSON)
	fmt.Println("Patch:   ", patchJSON)
	fmt.Println("Result:  ", resultJSON)

	fmt.Println("\n=== Key Behaviors ===")
	fmt.Println("• Objects are merged recursively")
	fmt.Println("• null values delete fields")
	fmt.Println("• Arrays are replaced entirely")
	fmt.Println("• New fields are added")
}

func prettyJSON(v any) (string, error) {
	data, err := json.Marshal(v, jsontext.Multiline(true))
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}
	return string(data), nil
}
