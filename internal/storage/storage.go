package storage

import "strconv"

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
	InfluxDriver = "influxdb"
	// ElasticsearchDriver is a Elasticsearch storage using elasticsearch-go library.
	ElasticsearchDriver = "elasticsearch"
	// ConsulDriver is a Consul storage using consul-go-api library.
	ConsulDriver = "consul"
	// RedisDriver is a Redis storage using redis-go library.
	RedisDriver = "redis"
	// S3Driver is a S3 storage using minio library.
	S3Driver = "s3"
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

// Lib tools
func counters2Any(source Counters) AnyMetrics {
	newMap := make(AnyMetrics)
	for k, v := range source {
		newMap[k] = strconv.FormatInt(v, 10)
	}
	return newMap
}

func gauges2Any(source Gauges) AnyMetrics {
	newMap := make(AnyMetrics)
	for k, v := range source {
		newMap[k] = strconv.FormatFloat(v, 'g', -1, 64)
	}
	return newMap
}

// func checkUniqueCounters(counters Counters) bool {

// }
