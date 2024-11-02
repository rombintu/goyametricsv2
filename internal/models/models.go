// Package internal models
package internal

import (
	"fmt"

	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/myparser"
)

// Metrics represents a struct that holds the details of a metric.
// It includes the metric's ID, type, and value (either Delta for counter or Value for gauge).
type Metrics struct {
	ID    string   `json:"id"`              // The name of the metric
	MType string   `json:"type"`            // The type of the metric, which can be "gauge" or "counter"
	Delta *int64   `json:"delta,omitempty"` // The value of the metric if it is a counter
	Value *float64 `json:"value,omitempty"` // The value of the metric if it is a gauge
}

// SetValueOrDelta sets the appropriate field (Delta or Value) of the Metrics struct based on its type.
// It parses the provided string value into the appropriate type and sets the corresponding field.
//
// Parameters:
// - s: The string representation of the metric value to be parsed and set.
//
// Returns:
// - An error if the parsing fails, otherwise nil.
func (m *Metrics) SetValueOrDelta(s string) (err error) {
	// Optimized switch case to set the appropriate field based on the metric type
	switch m.MType {
	case storage.GaugeType:
		vval, err := myparser.Str2Float64(s)
		if err != nil {
			return fmt.Errorf("failed to parse gauge value: %w", err)
		}
		m.setValue(vval)
	case storage.CounterType:
		dval, err := myparser.Str2Int64(s)
		if err != nil {
			return fmt.Errorf("failed to parse counter value: %w", err)
		}
		m.setDelta(dval)
	}
	return nil
}

// setDelta sets the Delta field of the Metrics struct to the provided int64 value.
//
// Parameters:
// - delta: The int64 value to set as the Delta.
func (m *Metrics) setDelta(delta int64) {
	m.Delta = &delta
}

// setValue sets the Value field of the Metrics struct to the provided float64 value.
//
// Parameters:
// - value: The float64 value to set as the Value.
func (m *Metrics) setValue(value float64) {
	m.Value = &value
}
