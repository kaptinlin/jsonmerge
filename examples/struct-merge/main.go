package main

import (
	"fmt"
	"log"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/jsonmerge"
)

// User represents a simple user profile
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

	patch := User{
		Name: "Jane Doe",
		Age:  25,
	}

	result, err := jsonmerge.Merge(user, patch)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original:", prettyJSON(user))
	fmt.Println("Patch:   ", prettyJSON(patch))
	fmt.Println("Result:  ", prettyJSON(result.Doc))

	fmt.Printf("\nType-safe access: %s is %d years old\n",
		result.Doc.Name, result.Doc.Age)
}

func prettyJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
