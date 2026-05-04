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

	patch := map[string]any{
		"age": 31,
		"address": map[string]any{
			"city":    "Boston",
			"street":  "123 Main St",
			"country": nil,
		},
		"hobbies": []string{"hiking", "coding"},
		"email":   "john@example.com",
	}

	result, err := jsonmerge.Merge(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:", prettyJSON(original))
	fmt.Println("Patch:   ", prettyJSON(patch))
	fmt.Println("Result:  ", prettyJSON(result.Doc))

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
