// Package ptrhelper provides utility functions for creating pointers to basic types.
// These functions are useful when you need to create pointers to int64 and float64 values.
package ptrhelper

// Int64Ptr creates a pointer to an int64 value.
// It is useful when you need to pass a pointer to an int64 value to a function or struct field.
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr creates a pointer to a float64 value.
// It is useful when you need to pass a pointer to a float64 value to a function or struct field.
func Float64Ptr(f float64) *float64 {
	return &f
}
