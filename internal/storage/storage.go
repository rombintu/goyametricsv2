package storage

const (
	GaugeType   = "gauge"
	CounterType = "counter"
)

const (
	// MemDriver is a simple in-memory storage.
	MemDriver = "mem"

	// FileDriver is a simple file-based storage.
	FileDriver = "file"

	// PgxDriver is a PostgreSQL storage using pgx library.
	PgxDriver = "pgx"
	// InfluxDriver is a InfluxDB storage using influxdb-client-go library.
)

const memPath string = ""

type Storage interface {
	Update(mtype, mname, mval string) error
	Get(mtype, mname string) (string, error)

	// inc 12
	UpdateAll(Data) error

	GetAll() Data
	Save() error
	Restore() error
	Open() error
	Close() error
	Ping() error
}

func NewStorage(storageType string, storepath string) Storage {
	var storage Storage
	switch storageType {
	case PgxDriver:
		storage = NewPgxDriver(storepath)
	case FileDriver:
		storage = NewTmpDriver(storepath)
	default:
		storage = NewTmpDriver(memPath)
	}
	return storage
}
