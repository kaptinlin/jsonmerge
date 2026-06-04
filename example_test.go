package jsonmerge_test

import (
	"fmt"
	"log"

	"github.com/kaptinlin/jsonmerge"
)

func ExampleApply() {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{Name: "John", Age: 30}
	patch, err := jsonmerge.NewPatch(map[string]any{"name": "Jane"})
	if err != nil {
		log.Fatal(err)
	}

	user, err = jsonmerge.Apply(user, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(user.Name)
	fmt.Println(user.Age)
	// Output:
	// Jane
	// 30
}

func ExampleApply_jsonText() {
	doc := jsonmerge.JSON(`{"name":"John"}`)
	patch, err := jsonmerge.Parse([]byte(`{"name":"Jane"}`))
	if err != nil {
		log.Fatal(err)
	}

	doc, err = jsonmerge.Apply(doc, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc)
	// Output:
	// {"name":"Jane"}
}

func ExampleNewPatch() {
	patch, err := jsonmerge.NewPatch("raw string")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := jsonmerge.Apply("draft", patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc)
	// Output:
	// raw string
}

func ExampleDiff() {
	before := map[string]any{"name": "John", "age": 30}
	after := map[string]any{"name": "Jane", "age": 30}

	patch, err := jsonmerge.Diff(before, after)
	if err != nil {
		log.Fatal(err)
	}

	data, err := patch.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Output:
	// {"name":"Jane"}
}
