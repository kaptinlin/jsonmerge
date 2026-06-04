// Package main demonstrates struct-based JSON Merge Patch usage.
package main

import (
	"fmt"
	"log"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/jsonmerge"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age"`
}

func main() {
	fmt.Println("=== Struct Merge Example ===")

	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	original := user

	patchValue := map[string]any{
		"name": "Jane Doe",
		"age":  25,
	}

	patch, err := jsonmerge.NewPatch(patchValue)
	if err != nil {
		log.Fatal(err)
	}

	user, err = jsonmerge.Apply(user, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:", prettyJSON(original))
	fmt.Println("Patch:   ", prettyJSON(patchValue))
	fmt.Println("Result:  ", prettyJSON(user))

	fmt.Printf("\nType-safe access: %s is %d years old\n",
		user.Name, user.Age)
}

func prettyJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
