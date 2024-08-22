package internal

import (
	"strconv"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) SetValueOrDelta(s string) (err error) {
	switch m.MType {
	case storage.GaugeType:
		var vval float64
		if vval, err = strconv.ParseFloat(s, 64); err != nil {
			return err
		}
		m.setValue(vval)
	case storage.CounterType:
		var dval int64
		if dval, err = strconv.ParseInt(s, 10, 64); err != nil {
			return err
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
