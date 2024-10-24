package internal

import (
	"fmt"

	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/myparser"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) SetValueOrDelta(s string) (err error) {
	// Optimizated
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

func (m *Metrics) setDelta(delta int64) {
	m.Delta = &delta
}

func (m *Metrics) setValue(value float64) {
	m.Value = &value
}
