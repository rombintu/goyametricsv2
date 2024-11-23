// Package storage Storage
package storage

// Constants defining the types of metrics supported by the system.
const (
	GaugeType   = "gauge"   // Represents a gauge metric type.
	CounterType = "counter" // Represents a counter metric type.
)

// Constants defining the types of storage drivers supported by the system.
const (
	// MemDriver is a simple in-memory storage.
	MemDriver = "mem"

	// FileDriver is a simple file-based storage.
	FileDriver = "file"

	// PgxDriver is a PostgreSQL storage using the pgx library.
	PgxDriver = "pgx"
)

// memPath is the default path for in-memory storage.
const memPath string = ""

// Storage is an interface that defines the methods required for a storage implementation.
type Storage interface {
	// Update updates a metric of the specified type and name with the given value.
	Update(mtype, mname, mval string) error

	// Get retrieves the value of a metric of the specified type and name.
	Get(mtype, mname string) (string, error)

	// UpdateAll updates all metrics in the provided Data struct.
	UpdateAll(Data) error

	// GetAll retrieves all metrics stored in the storage.
	GetAll() Data

	// Save persists the current state of the storage to a persistent medium.
	Save() error

	// Restore loads the state of the storage from a persistent medium.
	Restore() error

	// Open initializes the storage, typically by opening connections or files.
	Open() error

	// Close gracefully shuts down the storage, typically by closing connections or files.
	Close() error

	// Ping checks the health of the storage, typically by testing connections.
	Ping() error
}

// NewStorage creates a new instance of the Storage interface based on the provided storage type and path.
// It initializes the appropriate storage driver based on the storageType parameter.
//
// Parameters:
// - storageType: The type of storage driver to use (e.g., "mem", "file", "pgx").
// - storepath: The path or connection string for the storage (e.g., file path, database URL).
//
// Returns:
// - An instance of the Storage interface.
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
