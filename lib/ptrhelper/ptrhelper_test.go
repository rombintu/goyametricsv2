package ptrhelper

import (
	"testing"
)

func TestInt64Ptr(t *testing.T) {
	// Тестируем функцию Int64Ptr
	tests := []struct {
		name string
		val  int64
	}{
		{"positive_value", 42},
		{"negative_value", -42},
		{"zero_value", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := Int64Ptr(tt.val)
			if ptr == nil {
				t.Error("Expected a non-nil pointer, got nil")
			}
		})
	}
}

func TestFloat64Ptr(t *testing.T) {
	// Тестируем функцию Float64Ptr
	tests := []struct {
		name string
		val  float64
	}{
		{"positive_value", 42.42},
		{"negative_value", -42.42},
		{"zero_value", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := Float64Ptr(tt.val)
			if ptr == nil {
				t.Error("Expected a non-nil pointer, got nil")
			}
		})
	}
}
