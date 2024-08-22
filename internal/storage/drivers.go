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

type Counters map[string]int64
type Gauges map[string]float64

type Data struct {
	Counters Counters `json:"counters"`
	Gauges   Gauges   `json:"gauges"`
}

type memDriver struct {
	data      *Data
	storepath string
}

func NewMemDriver(storepath string) *memDriver {
	return &memDriver{
		data:      &Data{},
		storepath: storepath,
	}
}

func (m *memDriver) Open() error {
	counters := make(map[string]int64)
	gauges := make(map[string]float64)
	m.data = &Data{
		Counters: counters,
		Gauges:   gauges,
	}
	return nil
}

func (m *memDriver) Close() error {
	m.data = &Data{}
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
	value, ok := m.data.Counters[key]
	if !ok {
		return 0, false
	}
	return value, true
}

func (m *memDriver) getGauge(key string) (float64, bool) {
	value, ok := m.data.Gauges[key]
	if !ok {
		return 0, false
	}
	return value, true
}

func (m *memDriver) updateGauge(key string, value float64) {
	m.data.Gauges[key] = value
}

func (m *memDriver) updateCounter(key string, value int64) {
	oldValue, exist := m.getCounter(key)
	if !exist {
		m.data.Counters[key] = value
	} else {
		m.data.Counters[key] = oldValue + value
	}
}

func (m *memDriver) GetAll() Data {
	return *m.data
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
		logger.Log.Info("no file found, skipping restore...")
		return nil
	}
	file, err := os.OpenFile(m.storepath, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Log.Warn("error getting file info, skipping...")
		return nil
	}

	if fileInfo.Size() == 0 {
		logger.Log.Warn("file is empty, skipping restore...")
		return nil
	}

	bytesData, err := io.ReadAll(file)
	if err != nil {
		logger.Log.Warn("no data found in store path, skipping restore...")
		return nil
	}

	err = json.Unmarshal(bytesData, &m.data)
	if err != nil {
		logger.Log.Error("error unmarshalling JSON data", zap.Error(err))
		return nil
	}
	return nil
}
