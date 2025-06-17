package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	fmt.Println("=== JSON String Merge Example ===")

	// JSON strings
	original := `{
		"name": "John Doe",
		"age": 30,
		"skills": ["Go", "Python"],
		"address": {
			"city": "New York",
			"zip": "10001"
		}
	}`

	patch := `{
		"age": 31,
		"skills": ["Go", "JavaScript", "Rust"],
		"address": {
			"zip": "10002",
			"country": "USA"
		},
		"email": "john@example.com"
	}`

	// Apply merge
	result, err := jsonmerge.Merge(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:")
	fmt.Println(original)
	fmt.Println("\nPatch:")
	fmt.Println(patch)
	fmt.Println("\nResult:")
	fmt.Println(result.Doc)

	// Demonstrate validation
	fmt.Println("\n=== Validation ===")
	validPatch := `{"name": "Updated"}`
	invalidPatch := `{"name": invalid}`

	fmt.Printf("Valid patch: %t\n", jsonmerge.Valid(validPatch))
	fmt.Printf("Invalid patch: %t\n", jsonmerge.Valid(invalidPatch))
}
