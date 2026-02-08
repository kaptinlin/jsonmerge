package jsonmerge_test

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func ExampleMerge() {
	target := map[string]any{"name": "John", "age": 30}
	patch := map[string]any{"age": 31}

	result, err := jsonmerge.Merge(target, patch)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Doc["name"])
	fmt.Println(result.Doc["age"])
	// Output:
	// John
	// 31
}

func ExampleMerge_withMutate() {
	target := map[string]any{"name": "John", "age": 30}
	patch := map[string]any{"age": 31}

	result, err := jsonmerge.Merge(target, patch, jsonmerge.WithMutate(true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Doc["age"])
	// Output: 31
}

func ExampleGenerate() {
	source := map[string]any{"name": "John", "age": 30, "city": "NYC"}
	target := map[string]any{"name": "Jane", "age": 30}

	patch, err := jsonmerge.Generate(source, target)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(patch["name"])
	fmt.Println(patch["city"])
	// Output:
	// Jane
	// <nil>
}

func ExampleValid() {
	fmt.Println(jsonmerge.Valid(map[string]any{"name": "Jane"}))
	fmt.Println(jsonmerge.Valid([]byte(`{"name": "Jane"}`)))
	fmt.Println(jsonmerge.Valid([]byte(`{invalid}`)))
	// Output:
	// true
	// true
	// false
}
