// Package main demonstrates explicit JSON text and string scalar usage.
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func main() {
	fmt.Println("=== JSON Text Merge Example ===")

	original := jsonmerge.JSON(`{
		"name": "John Doe",
		"age": 30,
		"skills": ["Go", "Python"],
		"address": {
			"city": "New York",
			"zip": "10001"
		}
	}`)

	patchData := []byte(`{
		"age": 31,
		"skills": ["Go", "JavaScript", "Rust"],
		"address": {
			"zip": "10002",
			"country": "USA"
		},
		"email": "john@example.com"
	}`)

	patch, err := jsonmerge.Parse(patchData)
	if err != nil {
		log.Fatal(err)
	}

	result, err := jsonmerge.Apply(original, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:")
	fmt.Println(original)
	fmt.Println("\nPatch:")
	fmt.Println(string(patchData))
	fmt.Println("\nResult:")
	fmt.Println(result)

	fmt.Println("\n=== String Scalar ===")
	scalarPatch, err := jsonmerge.NewPatch(`{"name": invalid}`)
	if err != nil {
		log.Fatal(err)
	}

	scalar, err := jsonmerge.Apply("draft", scalarPatch)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Scalar result: %s\n", scalar)
}
