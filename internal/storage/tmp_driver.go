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

type tmpDriver struct {
	data      *Data
	storepath string
}

func NewTmpDriver(storepath string) *tmpDriver {
	return &tmpDriver{
		data:      &Data{},
		storepath: storepath,
	}
}

func (d *tmpDriver) Open() error {
	counters := make(Counters)
	gauges := make(Gauges)
	d.data = &Data{
		Counters: counters,
		Gauges:   gauges,
	}
	return nil
}

func (d *tmpDriver) Close() error {
	d.data = &Data{}
	return nil
}

func (d *tmpDriver) Ping() error {
	return nil
}

func (d *tmpDriver) Update(mtype, mname, mvalue string) (err error) {
	switch mtype {
	case GaugeType:
		var value float64
		if value, err = strconv.ParseFloat(mvalue, 64); err != nil {
			return err
		}
		d.updateGauge(mname, value)
	case CounterType:
		var value int64
		if value, err = strconv.ParseInt(mvalue, 10, 64); err != nil {
			return err
		}
		d.updateCounter(mname, value)
	default:
		return errors.New("invalid metric type")
	}
	return nil
}

func (d *tmpDriver) Get(mtype, mname string) (string, error) {
	switch mtype {
	case GaugeType:
		value, ok := d.getGauge(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case CounterType:
		value, ok := d.getCounter(mname)
		if !ok {
			return "", errors.New("not found")
		}
		return strconv.FormatInt(value, 10), nil
	}
	return "", errors.New("invalid metric type")
}

func (d *tmpDriver) getCounter(key string) (int64, bool) {
	value, ok := d.data.Counters[key]
	if !ok {
		return 0, false
	}
	return value, true
}

func (d *tmpDriver) getGauge(key string) (float64, bool) {
	value, ok := d.data.Gauges[key]
	if !ok {
		return 0, false
	}
	return value, true
}

func (d *tmpDriver) updateGauge(key string, value float64) {
	d.data.Gauges[key] = value
}

func (d *tmpDriver) updateCounter(key string, value int64) {
	oldValue, exist := d.getCounter(key)
	if !exist {
		d.data.Counters[key] = value
		return
	}
	d.data.Counters[key] = oldValue + value
}

func (d *tmpDriver) GetAll() Data {
	return *d.data
}

func (d *tmpDriver) UpdateAll(data Data) error {
	for k, v := range data.Counters {
		d.updateCounter(k, v)
	}
	for k, v := range data.Gauges {
		d.updateGauge(k, v)
	}
	return nil
}

func (d *tmpDriver) Save() error {

	if d.storepath == memPath {
		return nil
	}

	file, err := os.OpenFile(d.storepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(d.GetAll(), "", "\t")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *tmpDriver) Restore() error {

	if d.storepath == memPath {
		return nil
	}

	if _, err := os.Stat(d.storepath); errors.Is(err, os.ErrNotExist) {
		logger.Log.Info("no file found, skipping restore...")
		return nil
	}
	file, err := os.OpenFile(d.storepath, os.O_RDONLY|os.O_CREATE, 0660)
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

	err = json.Unmarshal(bytesData, &d.data)
	if err != nil {
		logger.Log.Error("error unmarshalling JSON data", zap.Error(err))
		return nil
	}
	return nil
}
