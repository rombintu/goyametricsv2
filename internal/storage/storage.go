package storage

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

type Storage interface {
	Update(mtype, mname, mval string) error
	Get(mtype, mname string) (string, error)
	GetAll() Data
	Save() error
	Restore() error
	Open() error
	Close() error
}

// type Storage struct {
// 	Driver Driver
// }

func NewStorage(storageType string, storepath string) Storage {
	var storage Storage
	switch storageType {
	default:
		storage = NewMemDriver(storepath)
	}
	return storage
}

// func (s *Storage) Open() {
// 	if err := s.Driver.Open(); err != nil {
// 		logger.Log.Error("cannot open storage", zap.Error(err))
// 	}
// }

// func (s *Storage) Close() {
// 	if err := s.Driver.Close(); err != nil {
// 		logger.Log.Error("cannot close storage", zap.Error(err))
// 	}
// }
