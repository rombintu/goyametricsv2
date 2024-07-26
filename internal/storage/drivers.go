package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"strconv"

	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

type Counter struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}
type Gauge struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Data struct {
	Counters []Counter `json:"counter"`
	Gauges   []Gauge   `json:"gauge"`
}

type memDriver struct {
	data      Data
	storepath string
}

func NewMemDriver(storepath string) *memDriver {
	return &memDriver{
		data:      Data{},
		storepath: storepath,
	}
}

func (m *memDriver) Open() error {
	return nil
}

func (m *memDriver) Close() error {
	m.data = Data{}
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
	for _, c := range m.data.Counters {
		if c.Name == key {
			return c.Value, true
		}
	}
	return 0, false
}

func (m *memDriver) getGauge(key string) (float64, bool) {
	for _, g := range m.data.Gauges {
		if g.Name == key {
			return g.Value, true
		}
	}
	return 0, false
}

func (m *memDriver) updateGauge(key string, value float64) {
	// update gauge, if exist. Create if not exist
	for _, g := range m.data.Gauges {
		if g.Name == key {
			g.Value = value
			return
		}
	}
	m.data.Gauges = append(m.data.Gauges, Gauge{Name: key, Value: value})
}

func (m *memDriver) updateCounter(key string, value int64) {
	for _, c := range m.data.Counters {
		if c.Name == key {
			c.Value += value
			return
		}
	}
	m.data.Counters = append(m.data.Counters, Counter{Name: key, Value: value})
}

func (m *memDriver) GetAll() Data {
	return m.data
}

func (m *memDriver) Save() error {
	file, err := os.OpenFile(m.storepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(m.GetAll(), "", "\t")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (m *memDriver) Restore() error {
	if _, err := os.Stat(m.storepath); errors.Is(err, os.ErrNotExist) {
		logger.Log.Warn("No data found in store path, skipping restore...")
		return nil
	}
	file, err := os.OpenFile(m.storepath, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Log.Warn("Error getting file info, skipping...")
		return nil
	}

	if fileInfo.Size() == 0 {
		logger.Log.Warn("File is empty, skipping restore...")
		return nil
	}

	bytesData, err := io.ReadAll(file)
	if err != nil {
		logger.Log.Warn(err.Error())
		return nil
	}

	err = json.Unmarshal(bytesData, &m.data)
	if err != nil {
		logger.Log.Error("Error unmarshalling JSON data", zap.Error(err))
		return nil
	}
	return nil
}
