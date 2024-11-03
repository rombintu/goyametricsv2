package myparser

import "fmt"

// ExampleStr2Float64 demonstrates how to use the Str2Float64 function to parse a string into a float64 value.
func ExampleStr2Float64() {
	value, err := Str2Float64("3.14159")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Parsed value:", value)
	}
	// Output:
	// Parsed value: 3.14159
}

// ExampleStr2Int64 demonstrates how to use the Str2Int64 function to parse a string into an int64 value.
func ExampleStr2Int64() {
	value, err := Str2Int64("1234567890")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Parsed value:", value)
	}
	// Output:
	// Parsed value: 1234567890
}
