package storage

// type metricType string
// type metricName string
// type metricValue string

const (
	gaugeType   = "gauge"
	counterType = "counter"
)

type Driver interface {
	Update(mtype, mname, mval string) error
	Get(mtype, mname string) (string, error)
	Open() error
	Close() error
}

type Storage struct {
	Driver Driver
}

func NewStorage(storageType string) *Storage {
	var driver Driver
	switch storageType {
	default:
		driver = NewMemDriver()
	}
	return &Storage{
		Driver: driver,
	}
}

func (s *Storage) Open() {
	if err := s.Driver.Open(); err != nil {
		panic(err)
	}
}
