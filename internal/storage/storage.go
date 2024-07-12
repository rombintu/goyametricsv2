package storage

// type metricType string
// type metricName string
// type metricValue string

type Metric struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	ValueStr string `json:"value"`
}

type CounterMetric struct {
	Metric
	Value int64 `json:"value"`
}

type GaugeMetric struct {
	Metric
	Value float64 `json:"value"`
}

type Metrics struct {
	GaugeMetrics   []GaugeMetric
	CounterMetrics []CounterMetric
}

type Driver interface {
	Update(mtype, mname, mval string) error
	Fetch(mtype, mname string) (Metric, error)
}

type Storage struct {
	Driver Driver
}

func NewStorage(storageType string) *Storage {
	var driver Driver
	switch storageType {
	default:
		driver = NewMemStorage()
	}
	return &Storage{
		Driver: driver,
	}
}
