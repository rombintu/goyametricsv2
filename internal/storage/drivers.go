package storage

import (
	"errors"
	"strconv"
)

type CounterTable map[string]int64
type GaugeTable map[string]float64

type memDriver struct {
	data map[string]interface{}
}

func NewMemDriver() *memDriver {
	return &memDriver{
		data: make(
			map[string]interface{},
		),
	}
}

func (m *memDriver) Open() error {
	m.data = make(map[string]interface{})
	m.data["counter"] = make(CounterTable)
	m.data["gauge"] = make(GaugeTable)
	return nil
}

func (m *memDriver) Close() error {
	m.data = nil
	return nil
}

func (s *memDriver) Update(mtype, mname, mvalue string) (err error) {
	switch mtype {
	case gaugeType:
		var value float64
		if value, err = strconv.ParseFloat(mvalue, 64); err != nil {
			return err
		}
		s.UpdateGauge(mname, value)
	case counterType:
		var value int
		if value, err = strconv.Atoi(mvalue); err != nil {
			return err
		}
		s.UpdateCounter(mname, int64(value))
	default:
		return errors.New("invalid metric type")
	}
	return nil
}

func (s *memDriver) Get(mtype, mname string) (string, error) {
	switch mtype {
	case gaugeType:
		value, ok := s.GetGauge(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case counterType:
		value, ok := s.GetCounter(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatInt(value, 10), nil
	}
	return "", errors.New("invalid metric type")
}

func (m *memDriver) GetCounter(key string) (int64, bool) {
	data, ok := m.data["counter"].(CounterTable)
	if !ok {
		return 0, false
	}
	return data[key], true
}

func (m *memDriver) GetGauge(key string) (float64, bool) {
	data, ok := m.data["gauge"].(GaugeTable)
	if !ok {
		return 0, false
	}
	return data[key], true
}

func (m *memDriver) UpdateGauge(key string, value float64) {
	data, _ := m.data["gauge"].(GaugeTable)
	data[key] = value
	m.data["gauge"] = data
}

func (m *memDriver) UpdateCounter(key string, value int64) {
	data, _ := m.data["counter"].(CounterTable)
	oldValue := data[key]
	if oldValue == 0 {
		data[key] = value
	} else {
		value = oldValue + value
	}
	data[key] = value
	m.data["counter"] = data
}
