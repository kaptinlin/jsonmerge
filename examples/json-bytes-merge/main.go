package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	fmt.Println("=== JSON Bytes Merge Example ===")

	// JSON data as byte arrays (common in HTTP requests, file I/O)
	original := []byte(`{
		"user": {
			"name": "John Doe",
			"email": "john@example.com"
		},
		"settings": {
			"theme": "light",
			"notifications": true
		},
		"version": "1.0"
	}`)

	patch := []byte(`{
		"user": {
			"email": null,
			"phone": "+1-555-0123"
		},
		"settings": {
			"theme": "dark",
			"language": "en"
		},
		"version": "1.1"
	}`)

	// Apply merge
	result, err := jsonmerge.Merge(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:")
	fmt.Println(string(original))
	fmt.Println("\nPatch:")
	fmt.Println(string(patch))
	fmt.Println("\nResult:")
	fmt.Println(string(result.Doc))

	// Generate patch example
	fmt.Println("\n=== Generate Patch ===")
	source := []byte(`{"name": "Alice", "age": 25}`)
	target := []byte(`{"name": "Alice", "age": 26, "city": "Boston"}`)

	generatedPatch, err := jsonmerge.Generate(source, target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Source: %s\n", source)
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Generated Patch: %s\n", generatedPatch)
}
