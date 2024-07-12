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
	m.data[CounterType] = make(CounterTable)
	m.data[GaugeType] = make(GaugeTable)
	return nil
}

func (m *memDriver) Close() error {
	m.data = nil
	return nil
}

func (m *memDriver) Update(mtype, mname, mvalue string) (err error) {
	switch mtype {
	case GaugeType:
		var value float64
		if value, err = strconv.ParseFloat(mvalue, 64); err != nil {
			return err
		}
		m.updateGauge(mname, value)
	case CounterType:
		var value int
		if value, err = strconv.Atoi(mvalue); err != nil {
			return err
		}
		m.updateCounter(mname, int64(value))
	default:
		return errors.New("invalid metric type")
	}
	return nil
}

func (m *memDriver) Get(mtype, mname string) (string, error) {
	switch mtype {
	case GaugeType:
		value, ok := m.getGauge(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case CounterType:
		value, ok := m.getCounter(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatInt(value, 10), nil
	}
	return "", errors.New("invalid metric type")
}

func (m *memDriver) getCounter(key string) (int64, bool) {
	data, ok := m.data[CounterType].(CounterTable)
	if !ok {
		return 0, false
	}
	return data[key], true
}

func (m *memDriver) getGauge(key string) (float64, bool) {
	data, ok := m.data[GaugeType].(GaugeTable)
	if !ok {
		return 0, false
	}
	return data[key], true
}

func (m *memDriver) updateGauge(key string, value float64) {
	data, _ := m.data[GaugeType].(GaugeTable)
	data[key] = value
	m.data[GaugeType] = data
}

func (m *memDriver) updateCounter(key string, value int64) {
	data, _ := m.data[CounterType].(CounterTable)
	oldValue := data[key]
	if oldValue == 0 {
		data[key] = value
	} else {
		value = oldValue + value
	}
	data[key] = value
	m.data[CounterType] = data
}
