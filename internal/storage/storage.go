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

type Storage interface {
	Update(mtype, mname, mval string) error
	Get(mtype, mname string) (string, error)
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
	default:
		storage = NewtmpDriver(storepath)
	}
	return storage
}
