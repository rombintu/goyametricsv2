package storage

import (
	"errors"
	"strconv"
)

type MemStorage struct {
	data Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: Metrics{
			GaugeMetrics:   make([]GaugeMetric, 0),
			CounterMetrics: make([]CounterMetric, 0),
		},
	}
}

func (s *MemStorage) Fetch(mtype, mname string) (Metric, error) {
	switch mtype {
	case "gauge":
		for _, m := range s.data.GaugeMetrics {
			if m.Name == mname {
				return Metric{
					Type:     m.Type,
					Name:     m.Name,
					ValueStr: m.ValueStr,
				}, nil
			}
		}
	case "counter":
		for _, m := range s.data.CounterMetrics {
			if m.Name == mname {
				return Metric{
					Type:     m.Type,
					Name:     m.Name,
					ValueStr: m.ValueStr,
				}, nil
			}
		}
	}
	return Metric{}, errors.New("not found")
}

func (s *MemStorage) Update(mtype, mname, mvalue string) (err error) {
	switch mtype {
	case "gauge":
		var value float64
		if value, err = strconv.ParseFloat(mvalue, 64); err != nil {
			return err
		}
		s.data.GaugeMetrics = append(s.data.GaugeMetrics, GaugeMetric{
			Metric: Metric{
				Type:     mtype,
				Name:     mname,
				ValueStr: mvalue,
			},
			Value: value,
		})
	case "counter":
		var value int
		if value, err = strconv.Atoi(mvalue); err != nil {
			return err
		}
		s.data.CounterMetrics = append(s.data.CounterMetrics, CounterMetric{
			Metric: Metric{
				Type:     mtype,
				Name:     mname,
				ValueStr: mvalue,
			},
			Value: int64(value),
		})
	default:
		return errors.New("invalid metric type")
	}
	return nil
}
