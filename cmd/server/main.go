package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/internal/server"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"go.uber.org/zap"
)

// main is the entry point of the application.
// It initializes the server, configures it, and starts the necessary workers.
// The application listens for termination signals to gracefully shut down.
func main() {
	// Load the server configuration
	conf := config.LoadServerConfig()

	// If the storage URL is provided and the storage driver is PgxDriver, set the storage path to the storage URL
	if conf.StorageURL != "" && conf.StorageDriver == storage.PgxDriver {
		conf.StoragePath = conf.StorageURL
	}

	// Create a new storage instance based on the configuration
	storage := storage.NewStorage(conf.StorageDriver, conf.StoragePath)

	// Create a new server instance with the storage and configuration
	server := server.NewServer(storage, conf)
	server.Configure()

	// Start the server in a separate goroutine
	go server.Run()

	// Create a channel to signal the completion of the application
	done := make(chan struct{})

	// If the sync mode is not enabled or the store interval is greater than 0, start a worker to synchronize the storage
	if !conf.SyncMode || conf.StoreInterval > 0 {
		go func() {
			// Create a ticker to trigger storage synchronization at the specified interval
			ticker := time.NewTicker(time.Duration(conf.StoreInterval) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					server.SyncStorageInterval()
				case <-done:
					logger.Log.Debug("worker is shutdown", zap.String("name", "sync_storage"))
					return
				}
			}
		}()
	}

	// Create a channel to capture termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for a termination signal
	<-sigChan

	// Signal the completion of the application
	close(done)

	// Gracefully shut down the server
	server.Shutdown()
	logger.Log.Info("All workers have shut down. Exiting program.")
}
