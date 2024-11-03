// Package myparser provides utility functions for parsing strings into numeric types.
package myparser

import (
	"strconv"
)

// Str2Float64 parses a string into a float64 value.
//
// Parameters:
// - s: The string to be parsed.
//
// Returns:
// - The parsed float64 value and any error encountered during parsing.
func Str2Float64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Str2Int64 parses a string into an int64 value.
//
// Parameters:
// - s: The string to be parsed.
//
// Returns:
// - The parsed int64 value and any error encountered during parsing.
func Str2Int64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
